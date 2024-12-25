package main

import (
	"fmt"
	"os"
	"time"

	"github.com/simonvetter/modbus"
)

func connectToBMS(client *modbus.ModbusClient, debug bool) error {
	// Read all BMS ModBus Addresses
	for attempt := 0; attempt < int(connectionRetries); attempt++ {
		if debug {
			fmt.Println("Trying to connect to BMS via ModBus. Attempt:", attempt+1)
		}

		// Try to establish a connection to the BMS. If it fails, retry until we reach the connectionRetries limit.
		err = client.Open()
		if err != nil {
			if debug {
				fmt.Println("Failure opening client. Waiting and retrying in 500ms.")
			}
			time.Sleep(connectionRetryDelay)
			continue
		}

		if debug {
			fmt.Println("Modbus client opened")
		}

		//DEBUG
		if debug {
			fmt.Println("Client:", client)
			fmt.Println("Reading Registers... Please wait!")
		}
		// VanMoof / DynaPack BMS uses slave-id 170
		client.SetUnitId(DynaPackVanMoofSlaveID)

		// Getting Fault Status to check if BMS is answering
		_, err = client.ReadRegisters(0x0002, 1, modbus.HOLDING_REGISTER)
		if err != nil {
			if debug {
				fmt.Println("Failed to read registers. Error:", err)
			}
			continue
		} else {
			break
		}
	}

	if err != nil || client == nil {
		fmt.Println("Retry Counter exceeded. Giving Up. Retry counter:", connectionRetries)
		fmt.Println("Failed to connect to BMS. Check if VCC on SWD Interface has 2.5Volts!")
		fmt.Println("Verify that RX/TX is connected correctly via JTAG BMS Version Output!")
		fmt.Println("Also make sure TEST is connected to GND. Otherwise the BMS will sleep and not respond!")
		fmt.Println("Thanks for keeping the World a better place!")
		os.Exit(1)
	}

	return nil
}

func readRegisters(client *modbus.ModbusClient, startAddress, quantity uint16) ([]uint16, error) {
	regs, err = client.ReadRegisters(startAddress, quantity, modbus.HOLDING_REGISTER)
	if err != nil {
		fmt.Println("Failed to read registers. Error:", err)
	}

	return regs, err
}
