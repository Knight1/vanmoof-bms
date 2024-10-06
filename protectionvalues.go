package main

import (
	"fmt"
)

func getAndShowProtectionBMSValues() {

	fmt.Println("-- BEGIN TRIGGER AND PROTECTION VALUES --")

	// Checking Proteection Statusses
	for register, value := range regs {
		switch register {
		case 2:
			if value == 0 {
				fmt.Println("BMS STATUS OK!")
			} else {
				fmt.Println("BMS SHUTDOWN!")
			}
		case 03:
			// TODO:
		case 45:
			// TODO: check values 45 to 48 about plausibility
			bmsUndervoltageCellProtection1 = value
			fmt.Println("Undervoltage Cell Protection 1 Trigger Value:", bmsUndervoltageCellProtection1, "mV")
		case 46:
			bmsUndervoltageCellProtection2 = value
			fmt.Println("Undervoltage Cell Protection 2 Trigger Value:", bmsUndervoltageCellProtection2, "mV")
		case 47:
			bmsUndervoltageCellShutdown = value
			fmt.Println("Undervoltage Cell Shutdown Trigger Value:", bmsUndervoltageCellShutdown, "mV")
		case 48:
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
			fmt.Println("(PDSCT) Peak Discharge Source/safety? Current Protection:", value)
		case 85:
			fmt.Println("(MOTP) MOSFET (Metal Oxide Semiconductor Field-Effect Transistors) Output Temperature Protection:", value)
		case 86:
			fmt.Println("(SCP) Source/safety? Current Protection:", value)
		default:
			continue
		}
	}

	fmt.Println("-- END TRIGGER AND PROTECTION VALUES --")
}
