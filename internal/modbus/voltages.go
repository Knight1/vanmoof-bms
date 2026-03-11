package modbus

import (
	"bms/v2/internal"
	"fmt"
	"math"
)

// passive means live
func GetAndShowPassiveVoltages() {
	fmt.Println("-- BEGIN LIVE VOLTAGES --")

	cellImbalanced := false

	// Voltage Cell Pack Monitoring
	for register, value := range internal.Registers {
		internal.MilliVolts = float64(value) / 1000
		switch register {
		case 4:
			fmt.Println("Pack: ", internal.MilliVolts, "Volts")
		case 27: // 0x1B
			fmt.Println("Cell 1:", internal.MilliVolts, "Volts")
		case 28: // 0x1C
			fmt.Println("Cell 2:", internal.MilliVolts, "Volts")
		case 29: // 0x1D
			fmt.Println("Cell 3:", internal.MilliVolts, "Volts")
		case 30: // 0x1E
			fmt.Println("Cell 4:", internal.MilliVolts, "Volts")
		case 31: // 0x1F
			fmt.Println("Cell 5:", internal.MilliVolts, "Volts")
		case 32: // 0x20
			fmt.Println("Cell 6:", internal.MilliVolts, "Volts")
		case 33: // 0x21
			fmt.Println("Cell 7:", internal.MilliVolts, "Volts")
		case 34: // 0x22
			fmt.Println("Cell 8:", internal.MilliVolts, "Volts")
		case 35: // 0x23
			fmt.Println("Cell 9:", internal.MilliVolts, "Volts")
		case 36: // 0x24
			fmt.Println("Cell 10:", internal.MilliVolts, "Volts")

		default:
			// Skip everything else
			continue
		}

		// Verify if Voltage is too low, too high
		if register != 4 && value < internal.CellVoltageLow {
			fmt.Println("Voltage in Cell to LOW!", register)
		} else if register != 4 && value > internal.CellVoltageHigh {
			fmt.Println("Voltage in Cell to HIGH!", register)
		}

		// Check if Pack Voltage is too low or too high
		if register == 4 && value < internal.PackVoltageLow {
			fmt.Println("Voltage in Pack to LOW!")
		} else if register == 4 && value > internal.PackVoltageHigh {
			fmt.Println("Voltage in Pack to HIGH!")
		}

		// Check for Voltage Imbalances in the Cells from live Values by own means.
		diff := math.Abs(float64(int(internal.CellVoltagePrevious) - int(value)))
		if diff > float64(internal.CellVoltageImbalance) {
			// Ignore Pack Voltage and the first Cell. Because there is nothing to compare it to.
			if register != 4 && register != 27 {
				cellImbalanced = true
				if internal.Debug {
					fmt.Println("DEBUG: Value ", value, "differs from previous value", internal.CellVoltagePrevious, "by more than", diff)
				}

			}
		}

		// Ignore Pack Voltage and 0 Values
		if register != 4 && value != 0 {
			internal.CellVoltagePrevious = value
		}
	}

	if cellImbalanced == true {
		fmt.Println("WARNING: Voltage Imbalance in Cells!")
	}

	fmt.Println("-- END LIVE VOLTAGES --")
}
