package canbus

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/brutella/can"
)

// frame builds an 8-byte CAN frame with the given ID and payload.
func frame(id uint32, b ...byte) can.Frame {
	var f can.Frame
	f.ID = id
	f.Length = uint8(len(b))
	copy(f.Data[:], b)
	return f
}

// TestShowBatteryState drives ShowBatteryState against a simulated battery pack
// that periodically broadcasts its OD telemetry signal run, and checks that the
// decoded output matches the crafted values (including the multi-frame per-cell
// voltage record).
func TestShowBatteryState(t *testing.T) {
	c := &Client{waiters: make(map[uint32]chan can.Frame)}

	// The pack broadcasts every signal each cycle; the three cell frames are
	// sent back-to-back so CollectBurst assembles them into one 12-cell record.
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
			}
			c.handleFrame(frame(BatteryStatusID, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x13, 0x00))
			c.handleFrame(frame(BatteryWarningID, 0x00, 0x00, 0x00, 0x00, 0x40, 0x00, 0x00, 0x00))
			c.handleFrame(frame(BatteryCapacityID, 0x00, 0x37, 0x13, 0x88, 0xFB, 0x50, 0x00, 0x00))
			c.handleFrame(frame(BatteryVoltageID, 0x8C, 0xA0, 0x0D, 0xFC, 0x0E, 0x24, 0x00, 0x00))
			c.handleFrame(frame(BatteryTemperatureID, 0x19, 0x1A, 0x1E, 0x1C, 0x03, 0x20, 0x3A, 0x98))
			c.handleFrame(frame(BatteryChargingID, 0x05, 0xDC, 0xA4, 0x10, 0x00, 0x00, 0x00, 0x00))
			c.handleFrame(frame(BatteryHealthID, 0x17, 0x70, 0x17, 0x0C, 0x00, 0x62, 0x00, 0x2A))
			// Per-cell voltages: 12 cells of 3601..3612 mV across three frames.
			c.handleFrame(frame(BatteryCellID, 0x0E, 0x11, 0x0E, 0x12, 0x0E, 0x13, 0x0E, 0x14))
			c.handleFrame(frame(BatteryCellID, 0x0E, 0x15, 0x0E, 0x16, 0x0E, 0x17, 0x0E, 0x18))
			c.handleFrame(frame(BatteryCellID, 0x0E, 0x19, 0x0E, 0x1A, 0x0E, 0x1B, 0x0E, 0x1C))
			time.Sleep(15 * time.Millisecond)
		}
	}()

	out := captureStdout(t, func() { ShowBatteryState(c) })
	close(stop)

	want := []string{
		"Mode index : 1",
		"Status flags : 0x10",
		"Status bytes : 01 02 03 04 05 06",
		"Warning flags : 00 00 00 00",
		"Cell imbalance : 64 mV",
		"RSOC : 55 %",
		"Remaining capacity : 5000 mAh",
		"Average current : -1200 mA",
		"Total voltage : 36000 mV",
		"Min. cell voltage : 3580 mV",
		"Max. cell voltage : 3620 mV",
		"Cell 01 : 3601 mV",
		"Cell 12 : 3612 mV",
		"Temperature cell 1 : 25 °C",
		"Temperature discharge MOS : 28 °C",
		"Discharge current : 800 mA",
		"Discharge current limit : 15000 mA",
		"Charge current : 1500 mA",
		"Charge voltage : 42000 mV",
		"Nominal capacity : 6000 mAh",
		"Full charge capacity : 5900 mAh",
		"SOH : 98 %",
		"Cycle count : 42",
	}
	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("output missing %q\n--- full output ---\n%s", w, out)
		}
	}

	// The record must decode exactly 12 cells (not a repeated single frame).
	if n := strings.Count(out, " mV\n"); n < 12 {
		t.Errorf("expected at least 12 cell lines, got %d\n%s", n, out)
	}
	if strings.Contains(out, "Cell 13") {
		t.Errorf("decoded more than 12 cells:\n%s", out)
	}
}

// TestIdentifyBattery confirms an A5/S5/S6 battery is recognised from any of its
// CAN signatures, and that a silent (or non-A5/S5/S6) bus is rejected.
func TestIdentifyBattery(t *testing.T) {
	// Positive: the pack broadcasts a telemetry signal -> identified.
	t.Run("detected", func(t *testing.T) {
		c := &Client{waiters: make(map[uint32]chan can.Frame)}
		stop := make(chan struct{})
		go func() {
			for {
				select {
				case <-stop:
					return
				default:
				}
				c.handleFrame(frame(BatteryVoltageID, 0x8C, 0xA0, 0x0D, 0xFC, 0x0E, 0x24, 0x00, 0x00))
				time.Sleep(10 * time.Millisecond)
			}
		}()
		desc, ok := IdentifyBattery(c, BatteryIdentifyTimeout)
		close(stop)
		if !ok {
			t.Fatal("expected an A5/S5/S6 battery to be identified")
		}
		if !strings.Contains(desc, "telemetry") {
			t.Errorf("unexpected signature description: %q", desc)
		}
	})

	// Positive: only the BMS command heartbeat is present -> identified.
	t.Run("heartbeat", func(t *testing.T) {
		c := &Client{waiters: make(map[uint32]chan can.Frame)}
		stop := make(chan struct{})
		go func() {
			for {
				select {
				case <-stop:
					return
				default:
				}
				c.handleFrame(frame(BatteryHeartbeatID, 0x00))
				time.Sleep(10 * time.Millisecond)
			}
		}()
		_, ok := IdentifyBattery(c, BatteryIdentifyTimeout)
		close(stop)
		if !ok {
			t.Fatal("expected the battery heartbeat to identify an A5/S5/S6 battery")
		}
	})

	// Negative: a silent bus (no A5/S5/S6 signatures) is rejected.
	t.Run("silent", func(t *testing.T) {
		c := &Client{waiters: make(map[uint32]chan can.Frame)}
		if _, ok := IdentifyBattery(c, 150*time.Millisecond); ok {
			t.Fatal("expected a silent bus to be rejected")
		}
	})

	// Negative: unrelated CAN traffic (not a battery signature) is rejected.
	t.Run("foreign-traffic", func(t *testing.T) {
		c := &Client{waiters: make(map[uint32]chan can.Frame)}
		stop := make(chan struct{})
		go func() {
			for {
				select {
				case <-stop:
					return
				default:
				}
				c.handleFrame(frame(0x18209820, 0x01)) // elock status, not a battery
				time.Sleep(10 * time.Millisecond)
			}
		}()
		_, ok := IdentifyBattery(c, 200*time.Millisecond)
		close(stop)
		if ok {
			t.Fatal("expected non-battery traffic to be rejected")
		}
	})
}

// TestBatteryPowerCommands verifies the battery on/off requests emit the correct
// command frame: CAN ID 0x14603040 with the opcode in the first data byte.
func TestBatteryPowerCommands(t *testing.T) {
	cases := []struct {
		name   string
		send   func(*Client) error
		opcode byte
	}{
		{"on", PowerBatteryOn, BatteryCmdOn},
		{"off", PowerBatteryOff, BatteryCmdOff},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var sent []can.Frame
			c := &Client{
				waiters: make(map[uint32]chan can.Frame),
				publish: func(f can.Frame) error { sent = append(sent, f); return nil },
			}

			_ = captureStdout(t, func() {
				if err := tc.send(c); err != nil {
					t.Errorf("send: %v", err)
				}
			})

			if len(sent) != 1 {
				t.Fatalf("expected 1 frame sent, got %d", len(sent))
			}
			f := sent[0]
			if got := f.ID &^ 0x80000000; got != BatteryCommandID { // strip the EFF flag
				t.Errorf("CAN ID = 0x%08X, want 0x%08X", got, BatteryCommandID)
			}
			if f.Length != 8 {
				t.Errorf("frame length = %d, want 8", f.Length)
			}
			if f.Data[0] != tc.opcode {
				t.Errorf("opcode = 0x%02X, want 0x%02X", f.Data[0], tc.opcode)
			}
			if f.Data[1] != 0x00 {
				t.Errorf("data[1] = 0x%02X, want 0x00", f.Data[1])
			}
		})
	}

	// Guard the wire constants against accidental change.
	if BatteryCommandID != 0x14603040 || BatteryCmdOn != 0x01 || BatteryCmdOff != 0x00 {
		t.Errorf("battery command wire constants changed: ID=0x%08X on=0x%02X off=0x%02X",
			BatteryCommandID, BatteryCmdOn, BatteryCmdOff)
	}
}

// captureStdout runs fn with os.Stdout redirected to a pipe and returns what it
// wrote.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()

	fn()

	w.Close()
	os.Stdout = orig
	return <-done
}
