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

// TurnChargeMOSOn enables the charge MOSFET by writing register 0x1A=1.
func TurnChargeMOSOn(client *modbus.ModbusClient) {
	if err := client.WriteRegister(0x1A, 1); err != nil {
		fmt.Println("Error setting Charge MOS to On. Error:", err)
	} else {
		fmt.Println("Charge MOS set to On!")
	}
}

// TurnChargeMOSOff disables the charge MOSFET by writing register 0x1A=0.
func TurnChargeMOSOff(client *modbus.ModbusClient) {
	if err := client.WriteRegister(0x1A, 0); err != nil {
		fmt.Println("Error setting Charge MOS to Off. Error:", err)
	} else {
		fmt.Println("Charge MOS set to Off!")
	}
}

// ResetESNModbus clears the Electronic Serial Number via Modbus by writing register 0x0A=0.
func ResetESNModbus(client *modbus.ModbusClient) {
	if err := client.WriteRegister(0x0A, 0); err != nil {
		fmt.Println("Error resetting ESN via Modbus. Error:", err)
	} else {
		fmt.Println("ESN reset via Modbus!")
	}
}

// ShipMode puts the BMS into ship mode by writing register 0x01=0.
func ShipMode(client *modbus.ModbusClient) {
	if err := client.WriteRegister(0x01, 0); err != nil {
		fmt.Println("Error setting Ship mode. Error:", err)
	} else {
		fmt.Println("Ship mode set!")
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
