package main

import "fmt"

func getAndShowFlashBMSData() {
	fmt.Println("-- BEGIN BMS FLASH STATUS --")

	//
	for register, value := range regs {
		switch register {
		case 48:
			// TODO: Check what this is, should be a hex output

			if value != 0 {
				fmt.Printf("Register 0x%X ('Fault Status'): 0x%04X\n", 0x2, value)

				// Decode flags (bitwise operations)
				flags := []string{"DOTP", "DUTP", "COTP", "CUTP", "DOCP1", "DOCP2", "COCP1", "COCP2", "OVP1", "OVP2", "UVP1", "UVP2", "PDOCP", "PDSCP", "MOTP", "SCP"}
				for i, flag := range flags {
					if value&(1<<i) != 0 {
						fmt.Printf(" - %s is set\n", flag)
					}
				}
			}
		case 49:
			// TODO: Reports records that are too high (65535)
			if value < 3731 && value > 2431 {
				fmt.Println("Battery Temperature Sensor 1 Record:", calculateCelsius(value), "°C")
			} else {
				fmt.Println("Battery Temperature Sensor 1 Record to high or to low!")
			}

		case 50:
			if value < 3731 && value > 2431 {
				fmt.Println("Battery Temperature Sensor 2 Record:", calculateCelsius(value), "°C")
			} else {
				fmt.Println("Battery Temperature Sensor 2 Record to high or to low!")
			}
		case 51:
			if value > 0 {
				fmt.Println("MOSFET Temperature Record:", calculateCelsius(value), "°C")
			} else {
				fmt.Println("MOSFET Temperature Record is set to 0!")
			}
		case 52:
			if value < 4000 {
				fmt.Println("Battery Voltage Record is below 4 Volts!")
			} else {
				fmt.Println("Battery Voltage Record:", value, "mV")
			}

		case 53:
			fmt.Println("Current: ", calculateAmperes(value), "mA")
		case 54:
			if value < 10000 {
				fmt.Println("Full Charge Capacity below 10000 mAh!")
			}
			// Does not make sense
			fmt.Println("Full Charge Capacity:", value, "mAh")
		case 55:
			// Does not make sense
			fmt.Println("Remaining Capacity:", value, "mAh")
		case 56:
			fmt.Println("RSOC: ", value, "%")
			// TODO: shows 0 or something to high instead of Real State of Charge
		case 57:
			fmt.Println("Absolute SOC", value, "%")
			// TODO: shows 0 or something to high instead of Absolute State of Charge
		case 58:
			if value != 0 {
				fmt.Println("Cycle Count:", value)
			} else {
				fmt.Println("No Cycle Count set in Flash. Check Passive Data Values!")
			}

			if value != regs[19] {
				fmt.Println("Flash Data and Passive Cycle Count mismatch! Real Cycle Count:", regs[19])
			}

		case 59:
			// register 59 to 68 are Cell Voltages from flash
			// Values are either 0 or something odd
		case 87:
			fmt.Println("Maximum recorded Charging Current:", calculateAmperes(value), "mA")
		case 88:
			fmt.Println("Maximum recorded Discharging Current:", calculateAmperes(value), "mA")
		case 89:
			if value < 3731 && value > 2431 {
				fmt.Println("Maximum recorded Cell Temperature:", calculateCelsius(value), "°C")
			} else {
				fmt.Println("Maximum recorded Cell Temperature is set to high!")
			}
		case 90:
			if value < 3731 && value > 2431 {
				fmt.Println("Minimal recorded Cell Temperature:", calculateCelsius(value), "°C")
			} else {
				fmt.Println("Minimal recorded Cell Temperature is set to high!")
			}
		case 91:
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
