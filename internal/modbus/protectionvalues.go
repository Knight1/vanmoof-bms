package modbus

import (
	"bms/v2/internal"
	"fmt"
)

func GetAndShowProtectionBMSValues() {

	fmt.Println("-- BEGIN TRIGGER AND PROTECTION VALUES --")
	fmt.Println("Trigger Values are best guess. DynaPack does not specify them.")

	// Checking Proteection Statusses
	for register, value := range internal.Registers {
		switch register {
		case 71: // 0x47
			fmt.Println("(DOTP) Discharge Over Temperature Protection:", value)
		case 72: // 0x48
			fmt.Println("(DUTP) Discharge Under Temperature Protection:", value)
		case 73: // 0x49
			fmt.Println("(COTP) Charging Over Temperature Protection:", value)
		case 74: // 0x4A
			fmt.Println("(CUTP) Current Under Temperature Protection:", value)
		case 75: // 0x4B
			fmt.Println("(DOCP1) Discharge Over Current Protection 1:", value)
		case 76: // 0x4C
			fmt.Println("(DOCP2) Discharge Over Current Protection 2:", value)
		case 77: // 0x4D
			fmt.Println("(COCP1) Charging Over Current Protection 1:", value)
		case 78: // 0x4E
			fmt.Println("(COCP2) Charging Over Current Protection 2:", value)
		case 79: // 0x4F
			fmt.Println("(OVP1) Over Voltage Protection 1:", value)
		case 80: // 0x50
			fmt.Println("(OVP2) Over Voltage Protection 2:", value)
		case 81: // 0x51
			fmt.Println("(UVP1) Under Voltage Protection 1:", value)
		case 82: // 0x52
			fmt.Println("(UVP2) Under Voltage Protection 2:", value)
		case 83: // 0x53
			fmt.Println("(PDOCP) Peak Discharge Over Current Protection:", value)
		case 84: // 0x54
			fmt.Println("(PDSCP) Peak Discharge Short Circuit Protection:", value)
		case 85: // 0x55
			fmt.Println("(MOTP) MOSFET (Metal Oxide Semiconductor Field-Effect Transistors) Over Temperature Protection:", value)
		case 86: // 0x56
			fmt.Println("(SCP) Short Circuit Protection:", value)
		default:
			continue
		}
	}

	fmt.Println("-- END TRIGGER AND PROTECTION VALUES --")
}
