package internal

import (
	"fmt"
)

func GetAndShowFlashBMSData() {
	fmt.Println("-- BEGIN BMS FLASH STATUS --")

	//
	for register, value := range Registers {
		switch register {
		case 48: // 0x30
			// TODO: Check what this is, should be a hex output

			if value != 0 {
				fmt.Printf("Register 0x%X ('Flash Fault Status'): 0x%04X\n", 0x30, value)

				// Decode flags (bitwise operations)
				flags := []string{"DOTP", "DUTP", "COTP", "CUTP", "DOCP1", "DOCP2", "COCP1", "COCP2", "OVP1", "OVP2", "UVP1", "UVP2", "PDOCP", "PDSCP", "MOTP", "SCP"}
				for i, flag := range flags {
					if value&(1<<i) != 0 {
						fmt.Printf(" - %s is set\n", flag)
					}
				}
			}
		case 49: // 0x31
			// TODO: Reports records that are too high (65535)
			if value < 3731 && value > 2431 {
				fmt.Println("Battery Temperature Sensor 1 Record:", calculateCelsius(value), "°C")
			} else {
				fmt.Println("Battery Temperature Sensor 1 Record to high or to low!")
			}

		case 50: // 0x32
			if value < 3731 && value > 2431 {
				fmt.Println("Battery Temperature Sensor 2 Record:", calculateCelsius(value), "°C")
			} else {
				fmt.Println("Battery Temperature Sensor 2 Record to high or to low!")
			}
		case 51: // 0x33
			if value > 0 {
				fmt.Println("MOSFET Temperature Record:", calculateCelsius(value), "°C")
			} else {
				fmt.Println("MOSFET Temperature Record is set to 0!")
			}
		case 52: // 0x34
			if value < 4000 {
				fmt.Println("Battery Voltage Record is below 4 Volts!")
			} else {
				fmt.Println("Battery Voltage Record:", value, "mV")
			}

		case 53: // 0x35
			fmt.Println("Current: ", calculateAmperes(value), "mA")
		case 54: // 0x36
			if value < 10000 {
				fmt.Println("Full Charge Capacity below 10000 mAh!")
			}
			// Does not make sense
			fmt.Println("Full Charge Capacity:", value, "mAh")
		case 55: // 0x37
			// Does not make sense
			fmt.Println("Remaining Capacity:", value, "mAh")
		case 56: // 0x38
			fmt.Println("RSOC: ", value, "%")
			// TODO: shows 0 or something to high instead of Real State of Charge
		case 57: // 0x39
			fmt.Println("Absolute SOC", value, "%")
			// TODO: shows 0 or something to high instead of Absolute State of Charge
		case 58: // 0x3A
			if value != 0 {
				fmt.Println("Cycle Count:", value)
			} else {
				fmt.Println("No Cycle Count set in Flash. Check Passive Data Values!")
			}

			if len(Registers) > 25 && value != Registers[25] {
				fmt.Println("Flash Data and Passive Cycle Count mismatch! Real Cycle Count:", Registers[25])
			}

		case 59: // 0x3B
			fmt.Println("Flash Cell 1 Voltage:", value, "mV")
		case 60: // 0x3C
			fmt.Println("Flash Cell 2 Voltage:", value, "mV")
		case 61: // 0x3D
			fmt.Println("Flash Cell 3 Voltage:", value, "mV")
		case 62: // 0x3E
			fmt.Println("Flash Cell 4 Voltage:", value, "mV")
		case 63: // 0x3F
			fmt.Println("Flash Cell 5 Voltage:", value, "mV")
		case 64: // 0x40
			fmt.Println("Flash Cell 6 Voltage:", value, "mV")
		case 65: // 0x41
			fmt.Println("Flash Cell 7 Voltage:", value, "mV")
		case 66: // 0x42
			fmt.Println("Flash Cell 8 Voltage:", value, "mV")
		case 67: // 0x43
			fmt.Println("Flash Cell 9 Voltage:", value, "mV")
		case 68: // 0x44
			fmt.Println("Flash Cell 10 Voltage:", value, "mV")
		case 69: // 0x45
			fmt.Println("Flash Max Vbatt Voltage:", value, "mV")
		case 70: // 0x46
			fmt.Println("Flash Min Vbatt Voltage:", value, "mV")
		case 87: // 0x57
			fmt.Println("Maximum recorded Charging Current:", calculateAmperes(value), "mA")
		case 88: // 0x58
			fmt.Println("Maximum recorded Discharging Current:", calculateAmperes(value), "mA")
		case 89: // 0x59
			if value < 3731 && value > 2431 {
				fmt.Println("Maximum recorded Cell Temperature:", calculateCelsius(value), "°C")
			} else {
				fmt.Println("Maximum recorded Cell Temperature is set to high!")
			}
		case 90: // 0x5A
			if value < 3731 && value > 2431 {
				fmt.Println("Minimal recorded Cell Temperature:", calculateCelsius(value), "°C")
			} else {
				fmt.Println("Minimal recorded Cell Temperature is set to high!")
			}
		case 91: // 0x5B
			if value < 3731 && value > 2431 {
				fmt.Println("Maximum recorded MOSFET Temperature:", calculateCelsius(value), "°C")
			} else {
				fmt.Println("Maximum recorded MOSFET Temperature is set to high!")
			}
		default:
			continue
		}
	}

	fmt.Println("-- END BMS FLASH STATUS --")
}
