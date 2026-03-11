package modbus

import (
	"bms/v2/internal"
	"fmt"
	"os"
	"time"

	"github.com/simonvetter/modbus"
)

func ConnectToBMS(client *modbus.ModbusClient, debug bool) (fault []uint16, err error) {
	var connectErr error

	// Read all BMS ModBus Addresses
	for attempt := 0; attempt < int(internal.ConnectionRetries); attempt++ {
		if debug {
			fmt.Println("Trying to connect to BMS via ModBus. Attempt:", attempt+1)
		}

		// Try to establish a connection to the BMS. If it fails, retry until we reach the connectionRetries limit.
		connectErr = client.Open()
		if connectErr != nil {
			if debug {
				fmt.Println("Failure opening client. Waiting and retrying in 500ms.")
			}
			time.Sleep(internal.ConnectionRetryDelay)
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
		if err := client.SetUnitId(internal.DynaPackVanMoofSlaveID); err != nil {
			if debug {
				fmt.Println("Failed to set unit ID. Error:", err)
			}
			continue
		}

		// Getting Fault Status to check if BMS is answering
		fault, connectErr = client.ReadRegisters(0x0002, 1, modbus.HOLDING_REGISTER)
		if connectErr != nil {
			if debug {
				fmt.Println("Failed to read registers. Error:", connectErr)
			}
			continue
		} else {
			break
		}
	}

	if connectErr != nil || client == nil {
		fmt.Println("Retry Counter exceeded. Giving Up. Retry counter:", internal.ConnectionRetries)
		fmt.Println("Failed to connect to BMS. Check if VCC on SWD Interface has 2.5Volts!")
		fmt.Println("Verify that RX/TX is connected correctly via JTAG BMS Version Output!")
		fmt.Println("Also make sure TEST is connected to GND. Otherwise the BMS will sleep and not respond!")
		fmt.Println("Thanks for keeping the World a better place!")
		os.Exit(1)
	}

	return fault, nil
}

func ReadRegisters(client *modbus.ModbusClient, startAddress, quantity uint16) ([]uint16, error) {
	regs, readErr := client.ReadRegisters(startAddress, quantity, modbus.HOLDING_REGISTER)
	if readErr != nil {
		fmt.Println("Failed to read registers. Error:", readErr)
		return regs, readErr
	}

	internal.Registers = regs
	return internal.Registers, nil
}
