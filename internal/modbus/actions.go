package modbus

import (
	"fmt"

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
