package modbus

import (
	"fmt"
	"strings"
)

// runStateWord maps register 2 (the run-state word emitted by report_arm_value)
// to its meaning. Register 2 is NOT a raw protection bitfield: the firmware
// passes the live state byte g_bms_state (0..0x19) through
// report_arm_value(g_bms_state-3), yielding a one-hot state word. Decode it by
// exact value, not bit-by-bit - several codes (0x00C0, 0x0030, 0xFFFF) are
// deliberate multi-bit values whose meaning is NOT the OR of their bits.
// Because states 0/1/2/4/5/6 all map to 0x0000, a 0x0000 reg-2 only means
// "running, no active protection" (normal / charge / discharge); the exact
// state index cannot be recovered from reg 2 alone.
// Source: batteryware docs/protocol.md, "Register 2 - run-state word".
var runStateWord = map[uint16]string{
	0x0000: "running - no active protection (normal / charge / discharge)",
	0x0080: "OVP1 - cell over-voltage protection 1",
	0x0040: "OVP2 - cell over-voltage protection 2",
	0x0020: "UVP1 - cell under-voltage protection 1",
	0x0010: "UVP2 - cell under-voltage protection 2",
	0x0200: "OVP1 charge-disable recovery latch",
	0x0100: "OVP2 charge-disable recovery latch",
	0x0800: "protection (prot-status bit 6, threshold-gated)",
	0x0400: "protection (prot-status bit 7, threshold-gated)",
	0x0008: "protection (prot-status bit 8, OVP-recovery class)",
	0x0004: "protection (prot-status bit 9, OVP-recovery class)",
	0x0001: "forced / pending protection (catch-all)",
	0x2000: "pack-1 cell over-voltage charge cutoff",
	0x1000: "pack-2 cell over-voltage charge cutoff",
	0x8000: "runtime OVP1 fault (FAULT_OVP1)",
	0x4000: "runtime OVP2 fault (FAULT_OVP2)",
	0x0002: "runtime over-temperature fault (FAULT_TS)",
	0xFFFF: "MOS FAILURE - over-current latched, secondary pyro fuse FIRED",
	0x00C0: "MOS FAILURE - hard protection, fuse NOT fired",
	0x0030: "recoverable protection (fuse never fires)",
}

// state3Composite decodes the four bits report_arm_value composes onto the wire
// when g_bms_state == 3: g_fault_flags 0x0001->0x2000 (DOTP), 0x0002->0x1000
// (DUTP), 0x0100->0x0200 (OVP1), 0x0200->0x0100 (OVP2). Used only as a fallback
// for reg-2 values not present in runStateWord, which can occur when more than
// one of these composes at once (protocol.md, "Register 2 in state 3").
var state3Composite = []struct {
	bit  uint16
	name string
}{
	{0x2000, "DOTP (discharge over-temperature)"},
	{0x1000, "DUTP (discharge under-temperature)"},
	{0x0200, "OVP1 (cell over-voltage 1)"},
	{0x0100, "OVP2 (cell over-voltage 2)"},
}

// runStateLabel returns a human-readable description of register-2's run-state
// word, or a best-effort state-3 composite/unknown description.
func runStateLabel(value uint16) string {
	if meaning, ok := runStateWord[value]; ok {
		return meaning
	}
	var parts []string
	for _, c := range state3Composite {
		if value&c.bit != 0 {
			parts = append(parts, c.name)
		}
	}
	if len(parts) > 0 {
		return "state-3 composite: " + strings.Join(parts, " + ")
	}
	return "unrecognized run-state word (not in protocol.md table)"
}

func CheckFaults(value uint16) {
	if value == 0 {
		fmt.Println("BMS STATUS OK! (reg 2 = 0x0000, running - no active protection)")
		return
	}
	fmt.Println("BMS SHUTDOWN / PROTECTION ACTIVE!")
	fmt.Printf("Register 0x%X ('Run State / Protection Word'): 0x%04X\n", 0x2, value)
	fmt.Printf(" - %s\n", runStateLabel(value))
}

func checkWarnings(value uint16) {
	fmt.Printf("Warning Status: %04X\n", value)

	// Decode flags (bitwise operations)
	flags := []string{"DOTPW", "DUTPW", "COTPW", "CUTPW", "DOCPW", "RSVD", "COCPW", "RSVD", "OVP1W", "RSVD", "UVP1W", "SOC", "PDOCPW", "RSVD", "MOTPW", "RSVD"}
	for i, flag := range flags {
		if value&(1<<i) != 0 {
			fmt.Printf(" - %s is set\n", flag)
		}
	}
}

func checkMOSControl(value uint16) {
	fmt.Printf("CHG MOS Control: %04X\n", value)

	// Decode flags (bitwise operations)
	flags := []string{"RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "CHG"}
	for i, flag := range flags {
		if value&(1<<i) != 0 {
			fmt.Printf(" - %s is set\n", flag)
		}
	}

}

func checkChargingStatus(value uint16) {
	fmt.Printf("Charging Status: %04X\n", value)

	// Decode flags (bitwise operations)
	flags := []string{"RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "CHG_IN", "Fault", "CHG"}
	for i, flag := range flags {
		if value&(1<<i) != 0 {
			fmt.Printf(" - %s is set\n", flag)
		}
	}
}

func checkDischargingStatus(value uint16) {
	fmt.Println("Discharging on/off:", value)
	fmt.Printf("Discharging Status: %04X\n", value)

	// Decode flags (bitwise operations)
	flags := []string{"RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "DSG"}
	for i, flag := range flags {
		if value&(1<<i) != 0 {
			fmt.Printf(" - %s is set\n", flag)
		}
	}
}
