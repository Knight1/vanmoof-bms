package main

import (
	"fmt"
)

func getAndShowProtectionBMSValues() {

	fmt.Println("-- BEGIN TRIGGER AND PROTECTION VALUES --")
	fmt.Println("Trigger Values are best guess. DynaPack does not specify them.")

	// Checking Proteection Statusses
	for register, value := range regs {
		switch register {
		case 45:
			// GUESS
			bmsUndervoltageCellProtection1 = value
			fmt.Println("Undervoltage Cell Protection 1 Trigger Value:", bmsUndervoltageCellProtection1, "mV")
		case 46:
			// GUESS
			bmsUndervoltageCellProtection2 = value
			fmt.Println("Undervoltage Cell Protection 2 Trigger Value:", bmsUndervoltageCellProtection2, "mV")
		case 47:
			// GUESS
			bmsUndervoltageCellShutdown = value
			fmt.Println("Undervoltage Cell Shutdown Trigger Value:", bmsUndervoltageCellShutdown, "mV")
		case 48:
			// GUESS
			bmsOvervoltageCellProtection1 = value
			fmt.Println("Overvoltage Cell Protection 1 Trigger Value:", bmsOvervoltageCellProtection1, "mV")
		case 71:
			fmt.Println("(DOTP) Discharge Over Temperature Protection:", value)
		case 72:
			fmt.Println("(DUTP) Discharge Under Temperature Protection:", value)
		case 73:
			fmt.Println("(COTP) Charging Over Temperature Protection:", value)
		case 74:
			fmt.Println("(CUTP) Current Under Temperature Protection:", value)
		case 75:
			fmt.Println("(DOCP1) Discharge Over Current Protection 1:", value)
		case 76:
			fmt.Println("(DOCP2) Discharge Over Current Protection 2:", value)
		case 77:
			fmt.Println("(COCP1) Charging Over Current Protection 1:", value)
		case 78:
			fmt.Println("(COCP2) Charging Over Current Protection 2:", value)
		case 79:
			fmt.Println("(OVP1) Over Voltage Protection 1:", value)
		case 80:
			fmt.Println("(OVP2) Over Voltage Protection 2:", value)
		case 81:
			fmt.Println("(UVP1) Under Voltage Protection 1:", value)
		case 82:
			fmt.Println("(UVP2) Under Voltage Protection 2:", value)
		case 83:
			fmt.Println("(PDOCP) Peak Discharge Over Current Protection:", value)
		case 84:
			fmt.Println("(PDSCT) Peak Discharge Short Circuit Protection:", value)
		case 85:
			fmt.Println("(MOTP) MOSFET (Metal Oxide Semiconductor Field-Effect Transistors) Over Temperature Protection:", value)
		case 86:
			fmt.Println("(SCP) Short Circuit Protection:", value)
		default:
			continue
		}
	}

	fmt.Println("-- END TRIGGER AND PROTECTION VALUES --")
}
