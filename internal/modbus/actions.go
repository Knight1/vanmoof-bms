package modbus

import (
	"fmt"
	"time"

	"github.com/simonvetter/modbus"
)

func TurnDebugOn(client *modbus.ModbusClient) {
	if err := client.WriteRegister(0x9, 1); err != nil {
		fmt.Println("Error setting Debug to On. Error:", err)
	} else {
		fmt.Println("Debug set to On!")
	}
}

func TurnDebugOff(client *modbus.ModbusClient) {
	if err := client.WriteRegister(0x9, 0); err != nil {
		fmt.Println("Error setting Debug to Off. Error:", err)
	} else {
		fmt.Println("Debug set to Off!")
	}
}

func TurnDischargingOn(client *modbus.ModbusClient) {
	if err := client.WriteRegister(0x8, 1); err != nil {
		fmt.Println("Error setting Discharging to On. Error:", err)
	} else {
		fmt.Println("Discharging set to On!")
	}
}

func TurnDischargingOff(client *modbus.ModbusClient) {
	if err := client.WriteRegister(0x8, 0); err != nil {
		fmt.Println("Error setting Discharging to Off. Error:", err)
	} else {
		fmt.Println("Discharging set to Off!")
	}
}

// ShipAndDischargeTurnOff puts the BMS into ship mode by writing register 0x01=0
// and then disabling discharge by writing register 0x08=0.
func ShipAndDischargeTurnOff(client *modbus.ModbusClient) {
	if err := client.WriteRegister(0x01, 0); err != nil {
		fmt.Println("Error setting Ship mode. Error:", err)
		return
	}
	fmt.Println("Ship mode set.")

	time.Sleep(20 * time.Millisecond)

	if err := client.WriteRegister(0x08, 0); err != nil {
		fmt.Println("Error disabling discharge. Error:", err)
	} else {
		fmt.Println("Discharge disabled. Ship and Discharge turned off!")
	}
}
