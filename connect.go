package main

import (
	"fmt"
	"os"
	"time"

	"github.com/simonvetter/modbus"
)

func connectToBMS(client *modbus.ModbusClient) (error, []uint16) {
	// Read all BMS ModBus Addresses
	for attempt := 0; attempt < int(connectionRetries); attempt++ {
		fmt.Println("Trying to connect to BMS via ModBus. Attempt:", attempt+1)
		// Try to establish a connection to the BMS. If it fails, retry.
		err = client.Open()
		if err != nil {
			fmt.Println("Failure opening client. Waiting and retrying in 500ms.")
			time.Sleep(connectionRetryDelay)
			continue
		}

		defer client.Close()

		fmt.Println("Modbus client opened")
		//DEBUG
		fmt.Println("Client:", client)
		fmt.Println("Reading Registers... Please wait!")

		// VanMoof / DynaPack BMS uses slave-id 170
		client.SetUnitId(DynaPackVanMoofSlaveID)

		regs, err = client.ReadRegisters(0x0, 95, modbus.HOLDING_REGISTER)
		if err != nil {
			fmt.Println("Failed to read registers. Error:", err)
			continue
		} else {
			break
		}
	}

	if err != nil || client == nil {
		fmt.Println("Failed to connect to BMS. Check if VCC on SWD Interface has 2.5Volts!")
		fmt.Println("Verify that RX/TX is connected correctly via JTAG BMS Version Output!")
		fmt.Println("Also make sure TEST is connected to GND. Otherwise the BMS will sleep and not respond!")
		fmt.Println("Thanks for keeping the World a better place!")
		os.Exit(1)
	}

	return nil, regs
}
