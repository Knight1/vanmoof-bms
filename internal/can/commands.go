package can

import (
	"encoding/binary"
	"fmt"
	"time"

	"golang.org/x/sys/unix"
)

// CAN IDs for commands sent to the BMS.
const (
	idDynaPack    = uint32(0x1082FF00) // DynaPack proprietary commands (needs prefix)
	idFeatureCtrl = uint32(0x000002FF) // Feature control (no prefix needed)
)

// runCAN asserts GPIO17, brings up the CAN interface, sends the BMS wake
// sequence, then calls fn with the open socket fd. Tears everything down on return.
func runCAN(iface string, fn func(fd int)) {
	releaseGPIO := HoldWakeLine()
	defer releaseGPIO()

	time.Sleep(100 * time.Millisecond)
	trySlcand()

	fd, err := openCAN(iface)
	if err != nil {
		fmt.Printf("[CAN] Error: %v\n", err)
		return
	}
	defer unix.Close(fd)

	// Wake sequence: heartbeat + AP version request + RequestBMSInfo
	hb := buildFrame(idPCHeartbeat, []byte{0x01, 0x00, 0x00, 0x00})
	writeFrame(fd, hb)
	time.Sleep(5 * time.Millisecond)
	writeFrame(fd, buildFrame(idAPVersionReq, nil))
	time.Sleep(5 * time.Millisecond)
	writeFrame(fd, buildFrame(idSetStatus, []byte{0x40, 0x00}))
	time.Sleep(300 * time.Millisecond) // let BMS wake

	fn(fd)
}

// sendDynaPack sends the "DynaPack" magic prefix then the command frame on
// 0x1082FF00. data is zero-padded to 8 bytes automatically.
func sendDynaPack(fd int, data []byte) {
	writeFrame(fd, buildFrame(idDynaPack, []byte("DynaPack")))
	time.Sleep(5 * time.Millisecond)
	cmd := make([]byte, 8)
	copy(cmd, data)
	writeFrame(fd, buildFrame(idDynaPack, cmd))
}

// sendDynaPackRead sends a DynaPack command and waits up to 2 s for a reply
// frame on 0x1082FF00. Returns the 8 reply bytes, or nil on timeout.
func sendDynaPackRead(fd int, data []byte) []byte {
	writeFrame(fd, buildFrame(idDynaPack, []byte("DynaPack")))
	time.Sleep(5 * time.Millisecond)
	cmd := make([]byte, 8)
	copy(cmd, data)
	writeFrame(fd, buildFrame(idDynaPack, cmd))

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if f, ok := readFrame(fd); ok {
			if f.ID&^unix.CAN_EFF_FLAG == idDynaPack {
				reply := make([]byte, 8)
				copy(reply, f.Data[:])
				return reply
			}
		}
	}
	return nil
}

// sendFeature sends a single-byte feature control command on 0x000002FF.
func sendFeature(fd int, cmd byte) {
	writeFrame(fd, buildFrame(idFeatureCtrl, []byte{cmd}))
}

// sendSetStatus sends a two-byte SetStatus frame on idSetStatus (0x14803470).
func sendSetStatus(fd int, cmd byte) {
	writeFrame(fd, buildFrame(idSetStatus, []byte{cmd, 0x00}))
	time.Sleep(50 * time.Millisecond)
}

// ---------------------------------------------------------------------------
// DynaPack proprietary commands (0x1082FF00)
// ---------------------------------------------------------------------------

// TurnDischargeOn enables the discharge MOSFET (DynaPack cmd 0x14, arg 0x01).
func TurnDischargeOn(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Turning discharge ON ...")
		sendDynaPack(fd, []byte{0x14, 0x01})
		fmt.Println("[CAN] Sent.")
	})
}

// TurnDischargeOff disables the discharge MOSFET (DynaPack cmd 0x14, arg 0x00).
func TurnDischargeOff(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Turning discharge OFF ...")
		sendDynaPack(fd, []byte{0x14, 0x00})
		fmt.Println("[CAN] Sent.")
	})
}

// TurnChargeOn enables the charge MOSFET (DynaPack cmd 0x15, arg 0x01).
func TurnChargeOn(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Turning charge ON ...")
		sendDynaPack(fd, []byte{0x15, 0x01})
		fmt.Println("[CAN] Sent.")
	})
}

// TurnChargeOff disables the charge MOSFET (DynaPack cmd 0x15, arg 0x00).
func TurnChargeOff(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Turning charge OFF ...")
		sendDynaPack(fd, []byte{0x15, 0x00})
		fmt.Println("[CAN] Sent.")
	})
}

// UnlockPF temporarily unlocks Protection Failure mode to allow commands
// (DynaPack cmd 0x0E, arg 0x01).
func UnlockPF(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Unlocking PF mode ...")
		sendDynaPack(fd, []byte{0x0E, 0x01})
		fmt.Println("[CAN] Sent.")
	})
}

// ClearLog erases the BMS error log (DynaPack cmd 0x22).
func ClearLog(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Clearing error log ...")
		sendDynaPack(fd, []byte{0x22})
		fmt.Println("[CAN] Sent.")
	})
}

// ResetBMS resets the BMS controller (DynaPack cmd 0xFF).
// This removes calibration, ESN and cycle data.
func ResetBMS(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Resetting BMS ...")
		sendDynaPack(fd, []byte{0xFF})
		fmt.Println("[CAN] Sent.")
	})
}

// CoulombCounterCheck triggers a coulomb counter verification (DynaPack cmd 0x23).
func CoulombCounterCheck(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Triggering coulomb counter check ...")
		sendDynaPack(fd, []byte{0x23})
		fmt.Println("[CAN] Sent.")
	})
}

// SetRTC sets the BMS real-time clock to the current system time
// (DynaPack cmd 0x00, bytes 1-6: ss/mm/hh/dd/MM/yy).
func SetRTC(iface string) {
	runCAN(iface, func(fd int) {
		now := time.Now()
		ss := byte(now.Second())
		mm := byte(now.Minute())
		hh := byte(now.Hour())
		dd := byte(now.Day())
		mo := byte(now.Month())
		yy := byte(now.Year() % 100)
		fmt.Printf("[CAN] Setting RTC to 20%02d/%02d/%02d %02d:%02d:%02d ...\n",
			yy, mo, dd, hh, mm, ss)
		sendDynaPack(fd, []byte{0x00, ss, mm, hh, dd, mo, yy, 0x00})
		fmt.Println("[CAN] Sent.")
	})
}

// ReadRTC reads the BMS real-time clock (DynaPack cmd 0x13).
// Reply bytes: [1]=ss [2]=mm [3]=HH [4]=dd [5]=MM [6]=yy
func ReadRTC(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Reading RTC ...")
		reply := sendDynaPackRead(fd, []byte{0x13})
		if reply == nil {
			fmt.Println("[CAN] No response (timeout).")
			return
		}
		fmt.Printf("[CAN] RTC: 20%02d/%02d/%02d  %02d:%02d:%02d\n",
			reply[6], reply[5], reply[4], reply[3], reply[2], reply[1])
	})
}

// SetSN writes a 13-character serial number to the BMS.
// Sent as two DynaPack frames: cmd 0x1B (chars 0-6) and cmd 0x1C (chars 7-12),
// each preceded by their own "DynaPack" prefix.
func SetSN(iface string, sn string) {
	if len(sn) != 13 {
		fmt.Printf("[CAN] Error: serial number must be exactly 13 characters (got %d)\n", len(sn))
		return
	}
	runCAN(iface, func(fd int) {
		fmt.Printf("[CAN] Setting serial number to %q ...\n", sn)

		p1 := make([]byte, 8)
		p1[0] = 0x1B
		copy(p1[1:], sn[:7])
		writeFrame(fd, buildFrame(idDynaPack, []byte("DynaPack")))
		time.Sleep(10 * time.Millisecond)
		writeFrame(fd, buildFrame(idDynaPack, p1))
		time.Sleep(10 * time.Millisecond)

		p2 := make([]byte, 8)
		p2[0] = 0x1C
		copy(p2[1:], sn[7:])
		writeFrame(fd, buildFrame(idDynaPack, []byte("DynaPack")))
		time.Sleep(10 * time.Millisecond)
		writeFrame(fd, buildFrame(idDynaPack, p2))

		fmt.Println("[CAN] Sent.")
	})
}

// ReadChargeCurrentOffset reads the charge current calibration offset
// (DynaPack cmd 0x0F). Reply bytes [0-1] = int16 LE.
func ReadChargeCurrentOffset(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Reading charge current offset ...")
		reply := sendDynaPackRead(fd, []byte{0x0F})
		if reply == nil {
			fmt.Println("[CAN] No response (timeout).")
			return
		}
		val := int16(binary.LittleEndian.Uint16(reply[0:2]))
		fmt.Printf("[CAN] Charge current offset: %d\n", val)
	})
}

// ReadDischargeCurrentOffset reads the discharge current calibration offset
// (DynaPack cmd 0x10). Reply bytes [0-1] = int16 LE.
func ReadDischargeCurrentOffset(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Reading discharge current offset ...")
		reply := sendDynaPackRead(fd, []byte{0x10})
		if reply == nil {
			fmt.Println("[CAN] No response (timeout).")
			return
		}
		val := int16(binary.LittleEndian.Uint16(reply[0:2]))
		fmt.Printf("[CAN] Discharge current offset: %d\n", val)
	})
}

// ReadChargeADCOffset reads the charge ADC current calibration offset
// (DynaPack cmd 0x12). Reply bytes [0-1] = uint16 LE.
func ReadChargeADCOffset(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Reading charge ADC current offset ...")
		reply := sendDynaPackRead(fd, []byte{0x12})
		if reply == nil {
			fmt.Println("[CAN] No response (timeout).")
			return
		}
		val := binary.LittleEndian.Uint16(reply[0:2])
		fmt.Printf("[CAN] Charge ADC current offset: %d\n", val)
	})
}

// ReadChargeVoltageOffset reads the charge voltage calibration offset
// (DynaPack cmd 0x21). Reply bytes [0-1] = int16 LE.
func ReadChargeVoltageOffset(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Reading charge voltage offset ...")
		reply := sendDynaPackRead(fd, []byte{0x21})
		if reply == nil {
			fmt.Println("[CAN] No response (timeout).")
			return
		}
		val := int16(binary.LittleEndian.Uint16(reply[0:2]))
		fmt.Printf("[CAN] Charge voltage offset: %d\n", val)
	})
}

// ---------------------------------------------------------------------------
// SetStatus commands (0x14803470)
// ---------------------------------------------------------------------------

// SetNormalMode puts the BMS into normal operating mode (SetStatus 0x01).
func SetNormalMode(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Setting normal mode ...")
		sendSetStatus(fd, 0x01)
		fmt.Println("[CAN] Sent.")
	})
}

// SetChargeMode puts the BMS into charge mode (SetStatus 0x02).
func SetChargeMode(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Setting charge mode ...")
		sendSetStatus(fd, 0x02)
		fmt.Println("[CAN] Sent.")
	})
}

// SetSleepMode puts the BMS into sleep / shipping mode (SetStatus 0x03).
func SetSleepMode(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Setting sleep/shipping mode ...")
		sendSetStatus(fd, 0x03)
		fmt.Println("[CAN] Sent.")
	})
}

// SetStandbyMode puts the BMS into standby mode (SetStatus 0x08).
func SetStandbyMode(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Setting standby mode ...")
		sendSetStatus(fd, 0x08)
		fmt.Println("[CAN] Sent.")
	})
}

// SetStatusChargeOn sets normal mode with charge MOSFET open (SetStatus 0x05).
func SetStatusChargeOn(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Normal mode + charge MOSFET open ...")
		sendSetStatus(fd, 0x05)
		fmt.Println("[CAN] Sent.")
	})
}

// SetStatusChargeOff sets normal mode with charge MOSFET closed (SetStatus 0x0D).
func SetStatusChargeOff(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Normal mode + charge MOSFET closed ...")
		sendSetStatus(fd, 0x0D)
		fmt.Println("[CAN] Sent.")
	})
}

// ---------------------------------------------------------------------------
// Feature control commands (0x000002FF)
// ---------------------------------------------------------------------------

// EnableCellBalance enables cell balancing (feature cmd 0x07).
func EnableCellBalance(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Enabling cell balancing ...")
		sendFeature(fd, 0x07)
		fmt.Println("[CAN] Sent.")
	})
}

// DisableCellBalance disables cell balancing (feature cmd 0x06).
func DisableCellBalance(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Disabling cell balancing ...")
		sendFeature(fd, 0x06)
		fmt.Println("[CAN] Sent.")
	})
}

// EnablePDSCP enables power-down SCP protection (feature cmd 0x03).
func EnablePDSCP(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Enabling PDSCP ...")
		sendFeature(fd, 0x03)
		fmt.Println("[CAN] Sent.")
	})
}

// DisablePDSCP disables power-down SCP protection (feature cmd 0x02).
func DisablePDSCP(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Disabling PDSCP ...")
		sendFeature(fd, 0x02)
		fmt.Println("[CAN] Sent.")
	})
}

// EnableCellOffline enables cell offline detection (feature cmd 0x05).
func EnableCellOffline(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Enabling cell offline detection ...")
		sendFeature(fd, 0x05)
		fmt.Println("[CAN] Sent.")
	})
}

// DisableCellOffline disables cell offline detection (feature cmd 0x04).
func DisableCellOffline(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Disabling cell offline detection ...")
		sendFeature(fd, 0x04)
		fmt.Println("[CAN] Sent.")
	})
}

// ResetSystemFunction resets BMS system functions (feature cmd 0xFE).
func ResetSystemFunction(iface string) {
	runCAN(iface, func(fd int) {
		fmt.Println("[CAN] Resetting system functions ...")
		sendFeature(fd, 0xFE)
		fmt.Println("[CAN] Sent.")
	})
}
