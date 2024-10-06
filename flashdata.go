package main

import "fmt"

func getAndShowFlashBMSData() {
	fmt.Println("-- BEGIN BMS FLASH STATUS --")

	//
	for register, value := range regs {
		switch register {
		case 48:
			// TODO: Check what this is, should be a hex output
			fmt.Println("Fault Status:", value)
		case 49:
			// TODO: Reports records that are too high
			fmt.Println("Battery Temperature Sensor 1 Record:", calculateCelsius(value), "°C")
		case 50:
			// TODO: Reports records that are too high
			fmt.Println("Battery Temperature Sensor 2 Record:", calculateCelsius(value), "°C")
		case 51:
			// TODO: Reports records that are too high
			fmt.Println("MOSFET Temperature Record:", calculateCelsius(value), "°C")
		case 52:
			fmt.Println("Battery Voltage Record:", value, "mV")
		case 53:
			fmt.Println("Current: ", calculateAmperes(value), "mA")
		case 54:
			fmt.Println("Full Charge Capacity:", value, "%")
		case 55:
			fmt.Println("Remaining Capacity:", value, "%")
		case 56:
			// TODO: shows 0 insted of Real State of Charge
		case 57:
			// TODO: shows 0 insted of Absolute State of Charge
		case 58:
			// TODO: shows 0 instead of Cycle Count
		case 59:
			// register 59 to 68 are Cell Voltages from flash
			// Values are either 0 or something odd
		case 87:
			fmt.Println("Maximal recorded Charging Current:", calculateAmperes(value), "mA")
		case 88:
			fmt.Println("Maximal recorded Discharging Current:", calculateAmperes(value), "mA")
		case 89:
			fmt.Println("Maximal recorded Cell Temperature:", calculateCelsius(value), "°C")
		case 90:
			fmt.Println("Minimal recorded Cell Temperature:", calculateCelsius(value), "°C")
		case 91:
			fmt.Println("Maximal recorded MOSFET Temperature:", calculateCelsius(value), "°C")
		default:
			continue
		}
	}

	fmt.Println("-- END BMS FLASH STATUS --")
}
