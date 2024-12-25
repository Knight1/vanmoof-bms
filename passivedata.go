package main

import (
	"fmt"
	"math"
)

func getAndShowPassiveBMSData() {
	fmt.Println("-- BEGIN BMS PASSIVE STATUS --")

	for register, value := range regs {
		switch register {
		case RegisterFault:
			checkFaults(value)
		case 3:
			fmt.Println("Battery Temperature:", calculateCelsius(value), "°C")
		case 4:
			fmt.Println("Battery Voltage:", value, "mV")
		case 5:
			fmt.Println("Real State of Charge:", value, "%")
		case 6:
			fmt.Println("Current:", calculateAmperes(value), "mA")
		case 7:
			checkChargingStatus(value)
		case 8:
			checkDischargingStatus(value)
		case 9:
			fmt.Printf("Test Mode: %04X\n", value)
		case 10:
			fmt.Printf("Hardware Version:%04X\n", value)
		case 11:
			fmt.Printf("Software Version: %04X\n", value)
		case 12:
			// Convert register data to ASCII string

			// Read ESN (14 characters -> 7 registers)
			esnRegisters := uint16(7) // 7 registers for 14 characters. Automatically includes Capacity and manufacturing Date

			// Convert register data to ASCII string
			bytes := make([]byte, 0, esnRegisters*2)
			// Maybe 12 to 17. 18 seems blank.
			for _, reg := range regs[12:19] {
				bytes = append(bytes, byte(reg>>8), byte(reg&0xFF)) // High and low bytes
			}

			esn := string(bytes)
			fmt.Printf("ESN: %s\n", esn)

		case 13:
			// Slice the range for manufacture date (registers 18 and 19)
			dateRegs := regs[14:16]

			// Allocate space for bytes (4 characters = 2 registers * 2 bytes per register)
			dateBytes := make([]byte, 0, len(dateRegs)*2)

			// Convert registers to bytes
			for _, reg := range dateRegs {
				dateBytes = append(dateBytes, byte(reg>>8), byte(reg&0xFF)) // High byte, Low byte
			}

			// Convert byte slice to ASCII string
			manufactureDate := string(dateBytes)

			// Print the manufacture date in YYWW format
			fmt.Printf("Manufacture Date: %s\n", manufactureDate)
		case 14:
			// Manufacturer Date uses 2 Bytes so this is the second Part of the Manufacturing Date
		case 15:
			fmt.Println("Normal Capacity:", value, "mAh")
		case 19:
			fmt.Println("Cycle Count:", value)
		case 26:
			checkMOSControl(value)
		case 27:
			fmt.Println("Cell 1 Voltage:", value, "mV")
		case 28:
			fmt.Println("Cell 2 Voltage:", value, "mV")
		case 29:
			fmt.Println("Cell 3 Voltage:", value, "mV")
		case 30:
			fmt.Println("Cell 4 Voltage:", value, "mV")
		case 31:
			fmt.Println("Cell 5 Voltage:", value, "mV")
		case 32:
			fmt.Println("Cell 6 Voltage:", value, "mV")
		case 33:
			fmt.Println("Cell 7 Voltage:", value, "mV")
		case 34:
			fmt.Println("Cell 8 Voltage:", value, "mV")
		case 35:
			fmt.Println("Cell 9 Voltage:", value, "mV")
		case 36:
			fmt.Println("Cell 10 Voltage:", value, "mV")
		case 37:
			fmt.Println("Temperature Sensor 1:", calculateCelsius(value), "°C")
		case 38:
			fmt.Println("Temperature Sensor 2:", calculateCelsius(value), "°C")
		case 39:
			fmt.Println("Discharge MOSFET Temperature:", calculateCelsius(value), "°C")
		case 40:
			checkWarnings(value)
		case 41:
			cellVoltageHighest = value
			fmt.Println("Maximum Battery Voltage:", value, "mV")
		case 42:
			cellVoltageLowest = value
			fmt.Println("Minimum Battery Voltage:", cellVoltageLowest, "mV")
			if math.Abs(float64(cellVoltageHighest-cellVoltageLowest)) > 20 {
				fmt.Println("WARNING: Voltage Imbalance in Cells!")
			}
		case 43:
			fmt.Println("Cell Balance:", value)
		case 44:
			fmt.Printf("Bootloader Version: %04X\n", value)
		default:
			continue
		}
	}

	fmt.Println("-- END BMS PASSIVE STATUS --")
}
