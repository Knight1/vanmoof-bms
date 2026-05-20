package can

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

// BMS broadcast CAN IDs (29-bit extended frames, BMS enc=0x480, dest enc=0x460)
const (
	idStatus   = 0x14807460 // byte0=status flags (0x80=PF active), byte1=mode
	idStandby  = 0x14809460 // byte4=sub-state counter
	idSoCCurr  = 0x1480B460 // byte0=SoC%, bytes4-5=current signed int16 LE ×10mA
	idVoltages = 0x1480D460 // bytes0-1=pack mV LE, bytes2-3=cell_min mV (lowest of 13), bytes4-5=cell_max mV (highest of 13)
	idTemps    = 0x14811460 // bytes0-3=temp raw[4], byte4=rolling ctr, byte7=SoC cross-check
	idStartup  = 0x14815460 // bytes0-1=design voltage Q8.8 ÷256 V, byte4=rated cap Ah
	idFWReply  = 0x14901460 // byte0=hw_rev, bytes1-3=fw_ver BCD
	idModelA   = 0x1490F460 // ASCII model string bytes 0-7
	idModelB   = 0x1490F461 // ASCII model string bytes 8-15

	// Frames we send to keep BMS alive / trigger broadcasts
	idPCHeartbeat  = 0x01111460 // power-control heartbeat
	idAPVersionReq = 0x14901470 // triggers BMS broadcast burst
	idSetStatus    = 0x14803470 // SetStatus RequestBMSInfo payload [0x40,0x00]

	// Imbalance threshold: BMS enters PF mode above this
	cellImbalanceShutdownMV = 250
)

// canFrame mirrors struct can_frame from the kernel (always 16 bytes).
type canFrame struct {
	ID   uint32
	DLC  uint8
	pad  [3]uint8
	Data [8]uint8
}

const canFrameSize = int(unsafe.Sizeof(canFrame{}))

// State holds the latest decoded values from all BMS broadcast frames.
// Updated under mu by the reader goroutine; read under mu by the display loop.
type State struct {
	mu sync.Mutex

	// 0x1480D460
	PackmV    uint16 // total pack voltage mV
	CellMinMV uint16 // lowest cell voltage of 13 cells, mV
	CellMaxMV uint16 // highest cell voltage of 13 cells, mV

	// 0x1480B460
	SoC       uint8 // state of charge %
	CurrentmA int32 // signed, in mA (raw int16 ×10mA)

	// 0x14807460
	StatusByte uint8
	ModeByte   uint8
	PFActive   bool

	// 0x14809460
	SubState uint8

	// 0x14811460
	TempRaw    [4]uint8 // raw bytes; °C = raw - 40 (Kelvin offset from -40°C baseline)
	TempC      [4]int16 // decoded temperatures in °C (same indexing as TempRaw)
	RollingCtr uint8

	// 0x14815460 (one-shot at startup)
	DesignV float32
	RatedAh uint8

	// 0x14901460
	HWRev uint8
	FWVer [3]uint8

	// 0x1490F460/461
	ModelStr string

	// Diagnostics
	Frames    int
	StartTime time.Time
	LastFrame time.Time

	modelBuf [16]byte
}

func (s *State) update(f canFrame) {
	rawID := f.ID &^ unix.CAN_EFF_FLAG
	d := f.Data[:]
	dlc := int(f.DLC)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.Frames++
	s.LastFrame = time.Now()

	switch rawID {
	case idStatus:
		if dlc >= 1 {
			s.StatusByte = d[0]
			s.PFActive = (d[0] & 0x80) != 0
		}
		if dlc >= 2 {
			s.ModeByte = d[1]
		}

	case idStandby:
		if dlc >= 5 {
			s.SubState = d[4]
		}

	case idSoCCurr:
		if dlc >= 1 {
			s.SoC = d[0]
		}
		if dlc >= 6 {
			raw := binary.LittleEndian.Uint16(d[4:6])
			s.CurrentmA = int32(int16(raw)) * 10
		}

	case idVoltages:
		if dlc >= 2 {
			s.PackmV = binary.LittleEndian.Uint16(d[0:2])
		}
		if dlc >= 4 {
			s.CellMinMV = binary.LittleEndian.Uint16(d[2:4])
		}
		if dlc >= 6 {
			s.CellMaxMV = binary.LittleEndian.Uint16(d[4:6])
		}

	case idTemps:
		if dlc >= 4 {
			copy(s.TempRaw[:], d[0:4])
			// °C = raw_byte - 40  (Kelvin offset from −40°C baseline; empirically
			// confirmed: raw=58–59 at ~18°C room temperature)
			for i := 0; i < 4; i++ {
				s.TempC[i] = int16(d[i]) - 40
			}
		}
		if dlc >= 5 {
			s.RollingCtr = d[4]
		}
		// byte7 = SoC cross-check; prefer idSoCCurr byte0 for SoC
		// but use it if SoC hasn't been set yet
		if dlc >= 8 && s.SoC == 0 {
			s.SoC = d[7]
		}

	case idStartup:
		if dlc >= 2 {
			raw := binary.LittleEndian.Uint16(d[0:2])
			s.DesignV = float32(raw) / 256.0
		}
		if dlc >= 5 {
			s.RatedAh = d[4]
		}

	case idFWReply:
		if dlc >= 1 {
			s.HWRev = d[0]
		}
		if dlc >= 4 {
			s.FWVer[0] = d[1]
			s.FWVer[1] = d[2]
			s.FWVer[2] = d[3]
		}

	case idModelA:
		if dlc > 0 {
			copy(s.modelBuf[0:], d[:dlc])
		}
	case idModelB:
		if dlc > 0 {
			copy(s.modelBuf[8:], d[:dlc])
		}
		s.ModelStr = nullTermStr(s.modelBuf[:])
	}
}

func nullTermStr(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}

// openCAN opens a raw SocketCAN socket bound to iface.
func openCAN(iface string) (int, error) {
	fd, err := unix.Socket(unix.AF_CAN, unix.SOCK_RAW, unix.CAN_RAW)
	if err != nil {
		return 0, fmt.Errorf("socket: %w", err)
	}
	ni, err := net.InterfaceByName(iface)
	if err != nil {
		unix.Close(fd)
		return 0, fmt.Errorf("interface %q: %w", iface, err)
	}
	if err := unix.Bind(fd, &unix.SockaddrCAN{Ifindex: ni.Index}); err != nil {
		unix.Close(fd)
		return 0, fmt.Errorf("bind: %w", err)
	}
	// 100 ms read timeout so the goroutine can check ctx.Done() regularly
	tv := unix.Timeval{Sec: 0, Usec: 100_000}
	unix.SetsockoptTimeval(fd, unix.SOL_SOCKET, unix.SO_RCVTIMEO, &tv)
	return fd, nil
}

func buildFrame(id uint32, data []byte) canFrame {
	f := canFrame{ID: id | unix.CAN_EFF_FLAG, DLC: uint8(len(data))}
	copy(f.Data[:], data)
	return f
}

func writeFrame(fd int, f canFrame) {
	buf := (*[canFrameSize]byte)(unsafe.Pointer(&f))[:]
	unix.Write(fd, buf)
}

func readFrame(fd int) (canFrame, bool) {
	buf := make([]byte, canFrameSize)
	n, err := unix.Read(fd, buf)
	if err != nil || n != canFrameSize {
		return canFrame{}, false
	}
	var f canFrame
	copy((*[canFrameSize]byte)(unsafe.Pointer(&f))[:], buf)
	return f, true
}

// trySlcand brings up can0 via slcand on /dev/ttyACM0 if can0 is absent.
func trySlcand() {
	if _, err := net.InterfaceByName("can0"); err == nil {
		return
	}
	if _, err := os.Stat("/dev/ttyACM0"); err != nil {
		return
	}
	fmt.Println("[CAN] can0 absent — trying slcand on /dev/ttyACM0 ...")
	exec.Command("slcand", "-o", "-s8", "-t", "hw", "/dev/ttyACM0", "can0").Run()
	time.Sleep(500 * time.Millisecond)
	exec.Command("ip", "link", "set", "can0", "up").Run()
	time.Sleep(200 * time.Millisecond)
}

// ReadBMS is the main entry point for CAN monitoring.
// It holds GPIO17 HIGH, opens the CAN socket, runs a reader goroutine,
// and refreshes a live display until SIGINT/SIGTERM.
func ReadBMS(iface string) {
	// 1. Assert wake line
	releaseGPIO := HoldWakeLine()
	defer releaseGPIO()

	// Brief settle time for BMS to notice wake line
	time.Sleep(100 * time.Millisecond)

	// 2. Bring up CAN interface if needed
	trySlcand()

	// 3. Open socket
	fd, err := openCAN(iface)
	if err != nil {
		fmt.Printf("[CAN] Error: %v\n", err)
		return
	}
	defer unix.Close(fd)

	fmt.Printf("[CAN] Socket open on %s — sending wake sequence ...\n", iface)

	// 4. Wake sequence
	hb := buildFrame(idPCHeartbeat, []byte{0x01, 0x00, 0x00, 0x00})
	apReq := buildFrame(idAPVersionReq, nil)
	ssInfo := buildFrame(idSetStatus, []byte{0x40, 0x00})

	writeFrame(fd, hb)
	time.Sleep(5 * time.Millisecond)
	writeFrame(fd, apReq)
	time.Sleep(5 * time.Millisecond)
	writeFrame(fd, ssInfo)

	// 5. Signal handling
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 6. State + reader goroutine
	s := &State{StartTime: time.Now()}

	go func() {
		hbTick := time.NewTicker(250 * time.Millisecond)
		defer hbTick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-hbTick.C:
				writeFrame(fd, hb)
			default:
				if f, ok := readFrame(fd); ok {
					s.update(f)
				}
			}
		}
	}()

	// 7. Display loop: refresh every 500 ms
	displayTick := time.NewTicker(500 * time.Millisecond)
	defer displayTick.Stop()

	fmt.Println("[CAN] Monitoring BMS — press Ctrl+C to stop")
	time.Sleep(300 * time.Millisecond) // let first frames arrive

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n[CAN] Stopped.")
			printState(s)
			return
		case <-displayTick.C:
			printState(s)
		}
	}
}

func printState(s *State) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear screen + move cursor home
	fmt.Print("\033[2J\033[H")

	uptime := time.Since(s.StartTime).Round(time.Second)
	age := "—"
	if !s.LastFrame.IsZero() {
		age = fmt.Sprintf("%.0fs ago", time.Since(s.LastFrame).Seconds())
		if time.Since(s.LastFrame) < time.Second {
			age = "live"
		}
	}

	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║      VanMoof / DynaPack BMS — Live CAN Monitor           ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Printf("  Uptime: %-10v  Frames: %-8d  Last frame: %s\n",
		uptime, s.Frames, age)
	fmt.Println()

	// Protection status
	pfStr := "NO"
	if s.PFActive {
		pfStr = "YES ← Protection Failure active"
	}
	fmt.Printf("  PF mode      : %s\n", pfStr)
	fmt.Printf("  Status       : 0x%02X   Mode: 0x%02X   Sub-state: 0x%02X\n",
		s.StatusByte, s.ModeByte, s.SubState)
	fmt.Println()

	// Voltages
	packV := float64(s.PackmV) / 1000.0
	cellMinV := float64(s.CellMinMV) / 1000.0
	cellMaxV := float64(s.CellMaxMV) / 1000.0

	fmt.Printf("  Pack voltage : %7.3f V   (%d mV)   nominal 47.97 V  (13S)\n", packV, s.PackmV)

	// CAN broadcasts the lowest and highest cell voltage across all 13 cells.
	if s.CellMinMV == 0 {
		fmt.Printf("  Cell min     : suppressed in PF mode (0 mV)\n")
	} else {
		fmt.Printf("  Cell min     : %7.3f V   (%d mV)   weakest cell\n", cellMinV, s.CellMinMV)
	}
	fmt.Printf("  Cell max     : %7.3f V   (%d mV)   strongest cell\n", cellMaxV, s.CellMaxMV)

	if s.CellMinMV > 0 && s.CellMaxMV > 0 {
		spread := int32(s.CellMaxMV) - int32(s.CellMinMV)
		if spread < 0 {
			spread = -spread
		}
		marker := ""
		if spread > cellImbalanceShutdownMV {
			marker = "  ← exceeds 250 mV PF threshold (expected)"
		}
		fmt.Printf("  Cell spread  : %7d mV%s\n", spread, marker)
	}

	fmt.Println()
	fmt.Printf("  State of charge: %d %%\n", s.SoC)
	currentA := float64(s.CurrentmA) / 1000.0
	dir := "discharging"
	if s.CurrentmA > 0 {
		dir = "charging"
	} else if s.CurrentmA == 0 {
		dir = "idle"
	}
	fmt.Printf("  Current        : %+.3f A  (%+d mA)  [%s]\n", currentA, s.CurrentmA, dir)

	fmt.Println()
	fmt.Printf("  Temperature    : %d °C  %d °C  %d °C  %d °C   rolling ctr: %d\n",
		s.TempC[0], s.TempC[1], s.TempC[2], s.TempC[3], s.RollingCtr)
	fmt.Printf("  Temp (raw)     : %d  %d  %d  %d\n",
		s.TempRaw[0], s.TempRaw[1], s.TempRaw[2], s.TempRaw[3])

	fmt.Println()
	if s.DesignV > 0 {
		fmt.Printf("  Design voltage : %.4f V\n", s.DesignV)
	}
	if s.RatedAh > 0 {
		fmt.Printf("  Rated capacity : %d Ah\n", s.RatedAh)
	}
	if s.ModelStr != "" {
		fmt.Printf("  Serial number  : %s\n", s.ModelStr)
	}
	if s.HWRev > 0 {
		fmt.Printf("  HW rev         : 0x%02X\n", s.HWRev)
		fmt.Printf("  FW version     : %d.%d.%d\n", s.FWVer[0], s.FWVer[1], s.FWVer[2])
	}

	fmt.Println()
	fmt.Println("  NOTE: CAN reports min + max cell voltage across all 13 cells.")
	fmt.Println()
	fmt.Println("  Press Ctrl+C to stop.")
}
