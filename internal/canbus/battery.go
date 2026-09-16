package canbus

import (
	"bms/v2/internal"
	"encoding/binary"
	"fmt"
	"sync"
	"time"
)

// Battery pack Object-Dictionary (OD) telemetry CAN IDs (S5 / S6).
//
// On the S5 / S6 the battery pack is OD node 0xA4 and broadcasts a run of
// telemetry signals on the main CAN bus. Each signal travels on its own 29-bit
// extended CAN ID built from the OD address fields:
//
//	CAN_ID = node<<21 | signal<<13 | port<<5 | class
//
// with node = 0xA4, port = 0x82 and class = 0, i.e.
//
//	CAN_ID = 0xA4<<21 | signal<<13 | 0x1040
//
// The signal index (0x03..0x0A) selects the measurement group. All multi-byte
// fields are big-endian on the wire (the pack byte-swaps to host order on the
// receiving side).
const (
	BatteryChargingID    uint32 = 0x14807040 // signal 0x03: charge current / charge voltage
	BatteryCellID        uint32 = 0x14809040 // signal 0x04: per-cell voltages (multi-frame)
	BatteryCapacityID    uint32 = 0x1480B040 // signal 0x05: RSOC / remaining capacity / average current
	BatteryWarningID     uint32 = 0x1480D040 // signal 0x06: warning flags + cell imbalance
	BatteryStatusID      uint32 = 0x1480F040 // signal 0x07: mode + status flags
	BatteryVoltageID     uint32 = 0x14811040 // signal 0x08: total / min / max cell voltage
	BatteryTemperatureID uint32 = 0x14813040 // signal 0x09: temperatures + discharge current
	BatteryHealthID      uint32 = 0x14815040 // signal 0x0A: nominal / full capacity, SOH, cycles
)

// BatteryTelemetryTimeout is how long to wait for each broadcast signal frame.
const BatteryTelemetryTimeout = 2000 * time.Millisecond

// BatteryIdentifyTimeout is how long to listen for an A5/S5/S6 battery signature
// before giving up. It must comfortably exceed the pack's heartbeat/telemetry
// broadcast period.
const BatteryIdentifyTimeout = 3000 * time.Millisecond

// NumBatteryCells is the number of series cell voltages the pack reports.
const NumBatteryCells = 12

// BatteryHeartbeatID is the A5/S5/S6 battery pack's heartbeat CAN ID.
//
// The bus heartbeat format is 0x01111000 | ((PFSA & 0x7F) << 5); the battery is
// node 0xA4, giving 0x01111000 | (0x24 << 5) = 0x01111480. (Cross-checked against
// the elock heartbeat 0x01111820 for node 0xC1.)
const BatteryHeartbeatID uint32 = 0x01111480

// batterySignatureIDs are the CAN IDs that positively identify an A5/S5/S6
// battery on the bus. The S3/S4 battery talks Modbus over UART and never appears
// on CAN, so observing any of these confirms an A5/S5/S6 pack. Keyed by CAN ID
// with a human-readable description of the signal.
func batterySignatureIDs() map[uint32]string {
	return map[uint32]string{
		BatteryHeartbeatID:   "battery heartbeat (node 0xA4)",
		IDHeartbeat:          "BMS command/heartbeat", // 0x10801000
		BatteryChargingID:    "charging telemetry",
		BatteryCellID:        "cell-voltage telemetry",
		BatteryCapacityID:    "capacity telemetry",
		BatteryWarningID:     "warning telemetry",
		BatteryStatusID:      "status telemetry",
		BatteryVoltageID:     "voltage telemetry",
		BatteryTemperatureID: "temperature telemetry",
		BatteryHealthID:      "health telemetry",
	}
}

// IdentifyBattery listens on the CAN bus for any A5/S5/S6 battery signature and
// reports whether one is present, returning a description of the first signature
// seen. A false result means no A5/S5/S6 battery was found (e.g. the bus is
// silent, or an S3/S4 battery which is Modbus/UART only and not addressable over
// CAN).
func IdentifyBattery(client *Client, timeout time.Duration) (string, bool) {
	sigs := batterySignatureIDs()
	ids := make([]uint32, 0, len(sigs))
	for id := range sigs {
		ids = append(ids, id)
	}

	if internal.Debug {
		fmt.Printf("[DEBUG] IdentifyBattery: listening %v for A5/S5/S6 battery CAN signatures\n", timeout)
	}

	frame, err := client.WaitForAnyID(ids, timeout)
	if err != nil {
		if internal.Debug {
			fmt.Printf("[DEBUG] IdentifyBattery: %v\n", err)
		}
		return "", false
	}

	id := frame.ID & canEFFMask
	desc := sigs[id]
	if desc == "" {
		desc = fmt.Sprintf("CAN ID 0x%08X", id)
	}
	if internal.Debug {
		fmt.Printf("[DEBUG] IdentifyBattery: detected 0x%08X (%s)\n", id, desc)
	}
	return desc, true
}

// Battery power command channel (A5/S5/S6 main bus).
//
// The battery pack is powered on / off through the power_control node's command
// channel: CAN object 0x1a4 -> CAN ID 0x14603040 (node 0xA3 -> battery node
// 0xA4, signal a1=0x01), carrying a 1-byte opcode in the first data byte. This
// is the wire form of the bike controller's `battery on` / `battery off`
// whole-bike power request.
const BatteryCommandID uint32 = 0x14603040

// Battery command opcodes (first data byte of a BatteryCommandID frame).
const (
	BatteryCmdOff             byte = 0x00 // battery power OFF
	BatteryCmdOn              byte = 0x01 // battery power ON
	BatteryCmdIdentifyCharger byte = 0x05 // identify charger
	BatteryCmdReset           byte = 0x06 // battery reset
	BatteryCmdShipping        byte = 0x08 // shipping mode
	BatteryCmdClearFaults     byte = 0x09 // clear fault flags
)

// SendBatteryCommand transmits a single battery power-command frame {opcode, 0}
// on the battery command channel, matching the controller's mode-command frame.
func (c *Client) SendBatteryCommand(opcode byte) error {
	if internal.Debug {
		fmt.Printf("[DEBUG] SendBatteryCommand: ID=0x%08X opcode=0x%02X\n", BatteryCommandID, opcode)
	}

	data := padTo8([]byte{opcode, 0x00})
	if err := c.SendFrame(BatteryCommandID, data); err != nil {
		return fmt.Errorf("send battery command 0x%02X: %w", opcode, err)
	}

	time.Sleep(DelayAfterCommand)
	return nil
}

// PowerBatteryOn requests the battery to power on — the whole-bike power-on
// request the controller issues for `battery on`.
func PowerBatteryOn(client *Client) error {
	fmt.Println("Requesting battery power ON...")
	if err := client.SendBatteryCommand(BatteryCmdOn); err != nil {
		fmt.Println("Error requesting battery power on:", err)
		return err
	}
	fmt.Println("Battery power ON command sent.")
	fmt.Println("(Run --action batteryState to confirm the battery is up.)")
	return nil
}

// PowerBatteryOff requests the battery to power off — the `battery off` request.
func PowerBatteryOff(client *Client) error {
	fmt.Println("Requesting battery power OFF...")
	fmt.Println("WARNING: this cuts the battery's output power — only do this with the bike at rest.")
	if err := client.SendBatteryCommand(BatteryCmdOff); err != nil {
		fmt.Println("Error requesting battery power off:", err)
		return err
	}
	fmt.Println("Battery power OFF command sent.")
	return nil
}

// batterySignal names a single telemetry signal for the concurrent reader.
type batterySignal struct {
	id   uint32
	name string
}

// readBatterySignal waits for a single broadcast frame carrying the given
// battery telemetry signal and returns its data bytes.
func readBatterySignal(client *Client, id uint32, name string) ([]byte, bool) {
	if internal.Debug {
		fmt.Printf("[DEBUG] readBatterySignal: waiting for %s signal (CAN ID 0x%08X)\n", name, id)
	}

	frame, err := client.WaitForID(id, BatteryTelemetryTimeout)
	if err != nil {
		if internal.Debug {
			fmt.Printf("[DEBUG] readBatterySignal: %s timed out: %v\n", name, err)
		}
		return nil, false
	}

	data := frame.Data[:frame.Length]
	if internal.Debug {
		fmt.Printf("[DEBUG] readBatterySignal: %s data=%X\n", name, data)
	}
	return data, true
}

// readBatteryCells collects the per-cell voltage signal (0x04), which the pack
// sends as a multi-frame record, and decodes up to NumBatteryCells big-endian
// cell voltages in millivolts. Frame payloads seen on the cell CAN ID within
// one broadcast burst are concatenated in arrival order; assembly stops on the
// first idle gap or when the record loops, so a repeated single frame is not
// mistaken for extra cells.
func readBatteryCells(client *Client) []uint16 {
	if internal.Debug {
		fmt.Printf("[DEBUG] readBatteryCells: collecting cell frames (CAN ID 0x%08X)\n", BatteryCellID)
	}

	// 12 cells * 2 bytes = 24 bytes -> up to 3 eight-byte frames per record.
	frames := client.CollectBurst(BatteryCellID, 4, BatteryTelemetryTimeout, 200*time.Millisecond)
	if len(frames) == 0 {
		if internal.Debug {
			fmt.Println("[DEBUG] readBatteryCells: no cell frames received")
		}
		return nil
	}

	var buf []byte
	for _, f := range frames {
		buf = append(buf, f.Data[:f.Length]...)
	}
	if internal.Debug {
		fmt.Printf("[DEBUG] readBatteryCells: assembled %d bytes from %d frame(s): %X\n", len(buf), len(frames), buf)
	}

	var cells []uint16
	for i := 0; i+2 <= len(buf) && len(cells) < NumBatteryCells; i += 2 {
		cells = append(cells, binary.BigEndian.Uint16(buf[i:i+2]))
	}
	return cells
}

// ShowBatteryState reads the S5 / S6 battery pack's OD telemetry from the CAN
// bus and prints the decoded state (mode, state of charge, capacity, pack and
// per-cell voltages, temperatures, currents and health), mirroring the battery
// status dump the bike's controller produces.
//
// The pack broadcasts each signal autonomously, so the signals are read
// concurrently (every CAN ID has its own receive slot) to capture them within a
// single broadcast cycle. No commands are sent to the pack.
func ShowBatteryState(client *Client) {
	fmt.Println("-- BEGIN BATTERY STATE --")
	if internal.Debug {
		fmt.Println("[DEBUG] ShowBatteryState: reading battery OD telemetry from node 0xA4")
	}

	signals := []batterySignal{
		{BatteryStatusID, "status"},
		{BatteryWarningID, "warning"},
		{BatteryCapacityID, "capacity"},
		{BatteryVoltageID, "voltage"},
		{BatteryTemperatureID, "temperature"},
		{BatteryChargingID, "charging"},
		{BatteryHealthID, "health"},
	}

	var mu sync.Mutex
	results := make(map[uint32][]byte)
	var cells []uint16

	var wg sync.WaitGroup
	for _, sig := range signals {
		wg.Add(1)
		go func(s batterySignal) {
			defer wg.Done()
			if data, ok := readBatterySignal(client, s.id, s.name); ok {
				mu.Lock()
				results[s.id] = data
				mu.Unlock()
			}
		}(sig)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		cells = readBatteryCells(client)
	}()
	wg.Wait()

	got := false

	// Mode + status flags (signal 0x07).
	if data := results[BatteryStatusID]; len(data) >= 7 {
		got = true
		fmt.Printf("Mode index : %d\n", int(data[6]&0x0f)-2)
		fmt.Printf("Status flags : 0x%02X\n", data[6]&0xf0)
		fmt.Printf("Status bytes : % X\n", data[0:6])
	}

	// Warning flags + cell imbalance (signal 0x06).
	if data := results[BatteryWarningID]; len(data) >= 5 {
		got = true
		fmt.Printf("Warning flags : % X\n", data[0:4])
		fmt.Printf("Cell imbalance : %d mV\n", data[4])
	}

	// State of charge + capacity (signal 0x05).
	if data := results[BatteryCapacityID]; len(data) >= 6 {
		got = true
		fmt.Printf("RSOC : %d %%\n", binary.BigEndian.Uint16(data[0:2]))
		fmt.Printf("Remaining capacity : %d mAh\n", binary.BigEndian.Uint16(data[2:4]))
		fmt.Printf("Average current : %d mA\n", int16(binary.BigEndian.Uint16(data[4:6])))
	}

	// Pack + cell-voltage envelope (signal 0x08).
	if data := results[BatteryVoltageID]; len(data) >= 6 {
		got = true
		fmt.Printf("Total voltage : %d mV\n", binary.BigEndian.Uint16(data[0:2]))
		fmt.Printf("Min. cell voltage : %d mV\n", binary.BigEndian.Uint16(data[2:4]))
		fmt.Printf("Max. cell voltage : %d mV\n", binary.BigEndian.Uint16(data[4:6]))
	}

	// Per-cell voltages (signal 0x04, multi-frame).
	if len(cells) > 0 {
		got = true
		for i, mv := range cells {
			fmt.Printf("Cell %02d : %d mV\n", i+1, mv)
		}
	}

	// Temperatures + discharge current (signal 0x09).
	if data := results[BatteryTemperatureID]; len(data) >= 8 {
		got = true
		fmt.Printf("Temperature cell 1 : %d °C\n", int8(data[0]))
		fmt.Printf("Temperature cell 2 : %d °C\n", int8(data[1]))
		fmt.Printf("Temperature charge MOS : %d °C\n", int8(data[2]))
		fmt.Printf("Temperature discharge MOS : %d °C\n", int8(data[3]))
		fmt.Printf("Discharge current : %d mA\n", binary.BigEndian.Uint16(data[4:6]))
		fmt.Printf("Discharge current limit : %d mA\n", binary.BigEndian.Uint16(data[6:8]))
	}

	// Charge current + voltage (signal 0x03).
	if data := results[BatteryChargingID]; len(data) >= 4 {
		got = true
		fmt.Printf("Charge current : %d mA\n", binary.BigEndian.Uint16(data[0:2]))
		fmt.Printf("Charge voltage : %d mV\n", binary.BigEndian.Uint16(data[2:4]))
	}

	// Health (signal 0x0A).
	if data := results[BatteryHealthID]; len(data) >= 8 {
		got = true
		fmt.Printf("Nominal capacity : %d mAh\n", int16(binary.BigEndian.Uint16(data[0:2])))
		fmt.Printf("Full charge capacity : %d mAh\n", int16(binary.BigEndian.Uint16(data[2:4])))
		fmt.Printf("SOH : %d %%\n", int16(binary.BigEndian.Uint16(data[4:6])))
		fmt.Printf("Cycle count : %d\n", int16(binary.BigEndian.Uint16(data[6:8])))
	}

	if !got {
		fmt.Println("No battery telemetry received. Check that:")
		fmt.Println("-> The CAN probe is on the bike's main bus (battery is node 0xA4)")
		fmt.Println("-> The bike / battery is powered on and broadcasting telemetry")
		fmt.Println("-> The CAN interface bitrate matches the bus (1 Mbps on S5 / S6)")
	}

	fmt.Println("-- END BATTERY STATE --")
}
