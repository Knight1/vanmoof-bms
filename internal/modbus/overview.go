package modbus

import (
	"bms/v2/internal"
	"fmt"
)

func ShowOverview() {

	fmt.Println("-- BEGIN BMS OVERVIEW --")

	// Hardware
	// Warnings
	// ESN

	for register, value := range internal.Registers {
		switch register {
		case internal.RegisterFault:
			CheckFaults(value)
		case 3:
			fmt.Println("Battery Temperature:", internal.CalculateCelsius(value), "°C")
		case 10: // 0x0A
			fmt.Printf("Hardware Version: %04X\n", value)
		case 11: // 0x0B
			fmt.Printf("Software Version: %04X\n", value)
		case 25: // 0x19
			fmt.Println("Cycle Count:", value)
		}
	}

	fmt.Println("-- END BMS OVERVIEW --")

	GetAndShowPassiveVoltages()
}
