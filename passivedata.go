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
			// TODO: Check what this is, should be a hex output
			fmt.Println("BMS Fault Status:", value)
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
		case 8:
			fmt.Println("Charging on/off:", value)
		case 9:
			fmt.Println("Test Mode:", value)
		case 10:
			// TODO: Compare this. Hardware and Software Version are broken
			fmt.Println("Hardware Version:", value)
		case 11:
			// TODO: Compare this. Hardware and Software Version are broken
			fmt.Println("Software Version:", value)
		case 15:
			fmt.Println("Normal Capacity:", value, "mAh")
		case 19:
			fmt.Println("Cycle Count:", value)
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
		default:
			continue
		}
	}

	fmt.Println("-- END BMS PASSIVE STATUS --")
}
