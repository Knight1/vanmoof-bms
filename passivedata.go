package main

import (
	"fmt"
	"math"
)

func getAndShowPassiveBMSData() {
	fmt.Println("-- BEGIN BMS PASSIVE STATUS --")

	for register, value := range regs {
		switch register {
		case 2:
			if value == 0 {
				fmt.Println("BMS STATUS OK!")
			} else {
				fmt.Println("BMS SHUTDOWN!")
				fmt.Printf("Register 0x%X ('Fault Status'): 0x%04X\n", 0x2, value)

				// Decode flags (bitwise operations)
				flags := []string{"DOTP", "DUTP", "COTP", "CUTP", "DOCP1", "DOCP2", "COCP1", "COCP2", "OVP1", "OVP2", "UVP1", "UVP2", "PDOCP", "PDSCP", "MOTP", "SCP"}
				for i, flag := range flags {
					if value&(1<<i) != 0 {
						fmt.Printf(" - %s is set\n", flag)
					}
				}
			}
		case 3:
			batteryTemperature := calculateCelsius(value)
			fmt.Println("Battery Temperature:", batteryTemperature, "°C")
		case 4:
			batteryVoltage := value
			fmt.Println("Battery Voltage:", batteryVoltage, "mV")
		case 5:
			fmt.Println("Real State of Charge:", value, "%")
		case 6:
			fmt.Println("Current:", calculateAmperes(value), "mA")
		case 7:
			// TODO: should be a hex output
			fmt.Println("Charging Status:", value)
			fmt.Printf("Register 0x%X ('Charging Status'): 0x%04X\n", 0x7, value)

			// Decode flags (bitwise operations)
			flags := []string{"RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "CHG_IN", "Fault", "CHG"}
			for i, flag := range flags {
				if value&(1<<i) != 0 {
					fmt.Printf(" - %s is set\n", flag)
				}
			}
		case 8:
			fmt.Println("Discharging on/off:", value)
			fmt.Printf("Register 0x%X ('Discharging Status'): 0x%04X\n", 0x8, value)

			// Decode flags (bitwise operations)
			flags := []string{"RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "DSG"}
			for i, flag := range flags {
				if value&(1<<i) != 0 {
					fmt.Printf(" - %s is set\n", flag)
				}
			}
		case 9:
			fmt.Println("Test Mode:", value)
			fmt.Printf("Register 0x%X ('Test Mode'): 0x%04X\n", 0x9, value)
		case 10:
			// TODO: Compare this. Hardware and Software Version are broken
			fmt.Println("Hardware Version:", value)
			fmt.Printf("Register 0x%X ('Hardware Version'): 0x%04X\n", 0x10, value)
		case 11:
			// TODO: Compare this. Hardware and Software Version are broken
			fmt.Println("Software Version:", value)
			fmt.Printf("Register 0x%X ('Software Version'): 0x%04X\n", 0x11, value)
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
			fmt.Printf("Register 0x%X ('CHG MOS Control'): 0x%04X\n", 0x26, value)

			// Decode flags (bitwise operations)
			flags := []string{"RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "CHG"}
			for i, flag := range flags {
				if value&(1<<i) != 0 {
					fmt.Printf(" - %s is set\n", flag)
				}
			}
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
			fmt.Println("Warning Status:", value)
			fmt.Printf("Register 0x%X ('Warning Status'): 0x%04X\n", 0x40, value)

			// Decode flags (bitwise operations)
			flags := []string{"DOTPW", "DUTPW", "COTPW", "CUTPW", "DOCPW", "RSVD", "COCPW", "RSVD", "OVP1W", "RSVD", "UVP1W", "SOC", "PDOCPW", "RSVD", "MOTPW", "RSVD"}
			for i, flag := range flags {
				if value&(1<<i) != 0 {
					fmt.Printf(" - %s is set\n", flag)
				}
			}
		case 41:
			cellVoltageHighest = value
			fmt.Println("Maximum Battery Voltage:", cellVoltageHighest, "mV")
		case 42:
			cellVoltageLowest = value
			fmt.Println("Minimum Battery Voltage:", cellVoltageLowest, "mV")
			if math.Abs(float64(cellVoltageHighest-cellVoltageLowest)) > 20 {
				fmt.Println("WARNING: Voltage Imbalance in Cells!")
			}
		case 43:
			fmt.Println("Cell Balance:", value)
		case 44:
			//TODO: check if this is correct
			fmt.Println("Bootloader Version:", value)
			fmt.Printf("Register 0x%X ('Bootloader Version'): 0x%04X\n", 0x44, value)
		default:
			continue
		}
	}

	fmt.Println("-- END BMS PASSIVE STATUS --")
}
