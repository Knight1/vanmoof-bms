package main

import (
	"fmt"
	"math"
)

func getAndShowPassiveVoltages() {
	fmt.Println("-- BEGIN LIVE VOLTAGES --")

	// Voltage Cell Pack Monitoring
	for register, value := range regs {
		milliVolts = float64(value) / 1000
		switch register {
		case 4:
			fmt.Println("Pack:", milliVolts, "Volts")
		case 27:
			fmt.Println("Cell 1:", milliVolts, "Volts")
		case 28:
			fmt.Println("Cell 2:", milliVolts, "Volts")
		case 29:
			fmt.Println("Cell 3:", milliVolts, "Volts")
		case 30:
			fmt.Println("Cell 4:", milliVolts, "Volts")
		case 31:
			fmt.Println("Cell 5:", milliVolts, "Volts")
		case 32:
			fmt.Println("Cell 6:", milliVolts, "Volts")
		case 33:
			fmt.Println("Cell 7:", milliVolts, "Volts")
		case 34:
			fmt.Println("Cell 8:", milliVolts, "Volts")
		case 35:
			fmt.Println("Cell 9:", milliVolts, "Volts")
		case 36:
			fmt.Println("Cell 10:", milliVolts, "Volts")

			// TODO: Check if one cell is below these values

		default:
			// Skip everything else
			continue
		}

		// Verify if Voltage is to low, to high
		if register != 4 && value < cellVoltageLow {
			fmt.Println("Voltage in Cell to LOW!", register)
		} else if register != 4 && value > cellVoltageHigh {
			fmt.Println("Voltage in Cell to HIGH!", register)
		}

		// TODO: Define Pack Voltage Limits from BMS!
		if register == 4 && value < packVoltageLow {
			fmt.Println("Voltage in Pack to LOW!")
		} else if register == 4 && value > packVoltageHigh {
			fmt.Println("Voltage in Pack to HIGH!")
		}

		// TODO: Check Voltage imbalance trigger
		// Check for Voltage Imbalances in the Cells
		diff := math.Abs(float64(cellVoltagePrevious - value))
		if diff > float64(cellVoltageImbalance) {
			//fmt.Println("Voltage Imbalance in Cells!")
			fmt.Println("DEBUG: Value ", value, "differs from previous value by more than", diff)
		}
		cellVoltagePrevious = value
		//fmt.Println("DEBUG: Value", value, "differs from previous value by", diff)

	}

	fmt.Println("-- END LIVE VOLTAGES --")
}
