package canbus

import (
	"bms/v2/internal"
	"encoding/binary"
	"fmt"
)

// ParseRegisterValue extracts and converts a register value from CAN frame data.
func ParseRegisterValue(data []byte, reg Register) float64 {
	if reg.Offset+reg.Length > len(data) {
		return 0
	}

	var raw int64
	switch reg.Length {
	case 1:
		raw = int64(data[reg.Offset])
	case 2:
		raw = int64(binary.LittleEndian.Uint16(data[reg.Offset : reg.Offset+2]))
	case 4:
		raw = int64(binary.LittleEndian.Uint32(data[reg.Offset : reg.Offset+4]))
	default:
		raw = int64(data[reg.Offset])
	}

	switch reg.Formula {
	case "temp":
		return float64(raw-2731) / 10.0
	case "current":
		return float64(raw) * 10.0
	default:
		return float64(raw)
	}
}

// FormatRegisterValue formats a parsed register value for display.
func FormatRegisterValue(value float64, reg Register) string {
	switch reg.Unit {
	case "hex":
		return fmt.Sprintf("0x%04X", int(value))
	case "°C":
		return fmt.Sprintf("%.1f °C", value)
	case "mA":
		return fmt.Sprintf("%.0f mA", value)
	default:
		if value == float64(int(value)) {
			return fmt.Sprintf("%d %s", int(value), reg.Unit)
		}
		return fmt.Sprintf("%.2f %s", value, reg.Unit)
	}
}

// ShowOverview displays a summary of BMS data from CAN bus registers.
// This is a placeholder that will work once registers are populated.
func ShowOverview(client *Client) {
	fmt.Println("-- BEGIN CAN BUS BMS OVERVIEW --")

	passiveRegs := GetPassiveRegisters()
	if len(passiveRegs) == 0 {
		fmt.Println("No passive registers defined yet.")
		fmt.Println("Register definitions must be added from the VMS5Manager.xml configuration.")
		fmt.Println("-- END CAN BUS BMS OVERVIEW --")
		return
	}

	// Request BMS info to trigger a broadcast
	if err := client.RequestBMSInfo(); err != nil {
		fmt.Printf("Failed to request BMS info: %v\n", err)
	}

	// Read responses for each known CAN ID
	for canID, subRegs := range passiveRegs {
		frame, err := client.WaitForID(canID, TimeoutCommand)
		if err != nil {
			if internal.Debug {
				fmt.Printf("[DEBUG] ShowOverview: no response for CAN ID 0x%08X: %v\n", canID, err)
			}
			continue
		}

		for _, reg := range subRegs {
			value := ParseRegisterValue(frame.Data[:frame.Length], reg)
			formatted := FormatRegisterValue(value, reg)
			fmt.Printf("%s: %s\n", reg.Name, formatted)
		}
	}

	fmt.Println("-- END CAN BUS BMS OVERVIEW --")
}
