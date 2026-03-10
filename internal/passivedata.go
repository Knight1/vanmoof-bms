package internal

import (
	"fmt"
	"math"
)

func GetAndShowPassiveBMSData() {
	fmt.Println("-- BEGIN BMS PASSIVE STATUS --")

	for register, value := range Registers {
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
		case 10: // 0x0A
			fmt.Printf("Hardware Version:%04X\n", value)
		case 11: // 0x0B
			fmt.Printf("Software Version: %04X\n", value)
		case 12: // 0x0C
			// ESN: 14 bytes (7 registers 0x0C-0x12) as ASCII
			if len(Registers) < 19 {
				fmt.Println("ESN: insufficient register data")
				continue
			}

			bytes := make([]byte, 0, 14)
			for _, reg := range Registers[12:19] {
				bytes = append(bytes, byte(reg>>8), byte(reg&0xFF))
			}
			fmt.Printf("ESN: %s\n", string(bytes))

		case 19: // 0x13
			// Manufacture Date: 4 bytes (2 registers 0x13-0x14) as DATE
			if len(Registers) < 21 {
				fmt.Println("Manufacture Date: insufficient register data")
				continue
			}

			dateBytes := make([]byte, 0, 4)
			for _, reg := range Registers[19:21] {
				dateBytes = append(dateBytes, byte(reg>>8), byte(reg&0xFF))
			}
			fmt.Printf("Manufacture Date: %s\n", string(dateBytes))

		case 21: // 0x15
			fmt.Println("Normal Capacity:", value, "mAh")
		case 25: // 0x19
			fmt.Println("Cycle Count:", value)
		case 26: // 0x1A
			checkMOSControl(value)
		case 27: // 0x1B
			fmt.Println("Cell 1 Voltage:", value, "mV")
		case 28: // 0x1C
			fmt.Println("Cell 2 Voltage:", value, "mV")
		case 29: // 0x1D
			fmt.Println("Cell 3 Voltage:", value, "mV")
		case 30: // 0x1E
			fmt.Println("Cell 4 Voltage:", value, "mV")
		case 31: // 0x1F
			fmt.Println("Cell 5 Voltage:", value, "mV")
		case 32: // 0x20
			fmt.Println("Cell 6 Voltage:", value, "mV")
		case 33: // 0x21
			fmt.Println("Cell 7 Voltage:", value, "mV")
		case 34: // 0x22
			fmt.Println("Cell 8 Voltage:", value, "mV")
		case 35: // 0x23
			fmt.Println("Cell 9 Voltage:", value, "mV")
		case 36: // 0x24
			fmt.Println("Cell 10 Voltage:", value, "mV")
		case 37: // 0x25
			fmt.Println("Temperature Sensor 1:", calculateCelsius(value), "°C")
		case 38: // 0x26
			fmt.Println("Temperature Sensor 2:", calculateCelsius(value), "°C")
		case 39: // 0x27
			fmt.Println("Discharge MOSFET Temperature:", calculateCelsius(value), "°C")
		case 40: // 0x28
			checkWarnings(value)
		case 41: // 0x29
			cellVoltageHighest = value
			fmt.Println("Maximum Battery Voltage:", value, "mV")
		case 42: // 0x2A
			cellVoltageLowest = value
			fmt.Println("Minimum Battery Voltage:", cellVoltageLowest, "mV")
			if math.Abs(float64(int(cellVoltageHighest)-int(cellVoltageLowest))) > 20 {
				fmt.Println("WARNING: Voltage Imbalance in Cells!")
			}
		case 43: // 0x2B
			fmt.Println("Cell Balance:", value)
		case 44: // 0x2C
			fmt.Printf("Bootloader Version: %04X\n", value)
		default:
			continue
		}
	}

	fmt.Println("-- END BMS PASSIVE STATUS --")
}
