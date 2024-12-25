package main

import (
	"fmt"
)

func showOverview() {

	fmt.Println("-- BEGIN BMS OVERVIEW --")

	// Hardware
	// Warnings
	// ESN

	for register, value := range regs {
		switch register {
		case RegisterFault:
			checkFaults(value)
		case 3:
			fmt.Println("Battery Temperature:", calculateCelsius(value), "°C")
		case 10:
			fmt.Printf("Hardware Version: %04X\n", value)
		case 11:
			fmt.Printf("Software Version: %04X\n", value)
		case 19:
			fmt.Println("Cycle Count:", value)
		}
	}

	fmt.Println("-- END BMS OVERVIEW --")

	getAndShowPassiveVoltages()
}
