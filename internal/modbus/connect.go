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

	if debug {
		fmt.Printf("[DEBUG] ConnectToBMS: loop=%v retries=%d retryDelay=%v slaveID=0x%02X\n",
			internal.Loop, internal.ConnectionRetries, internal.ConnectionRetryDelay, internal.DynaPackVanMoofSlaveID)
	}

	// Read all BMS ModBus Addresses
	for attempt := 0; internal.Loop || attempt < internal.ConnectionRetries; attempt++ {
		if debug {
			fmt.Printf("[DEBUG] ConnectToBMS: attempt %d - opening Modbus client\n", attempt+1)
		}

		// Try to establish a connection to the BMS. If it fails, retry until we reach the connectionRetries limit.
		connectErr = client.Open()
		if connectErr != nil {
			if debug {
				fmt.Printf("[DEBUG] ConnectToBMS: open failed: %v, retrying in %v\n", connectErr, internal.ConnectionRetryDelay)
			}
			time.Sleep(internal.ConnectionRetryDelay)
			continue
		}

		if debug {
			fmt.Println("[DEBUG] ConnectToBMS: Modbus client opened")
			fmt.Println("[DEBUG] ConnectToBMS: client:", client)
		}

		// VanMoof / DynaPack BMS uses slave-id 170
		if debug {
			fmt.Printf("[DEBUG] ConnectToBMS: setting unit ID to 0x%02X (%d)\n", internal.DynaPackVanMoofSlaveID, internal.DynaPackVanMoofSlaveID)
		}
		if err := client.SetUnitId(internal.DynaPackVanMoofSlaveID); err != nil {
			if debug {
				fmt.Printf("[DEBUG] ConnectToBMS: failed to set unit ID: %v\n", err)
			}
			continue
		}

		// Getting Fault Status to check if BMS is answering
		if debug {
			fmt.Println("[DEBUG] ConnectToBMS: reading fault status register 0x0002 to verify BMS is responding")
		}
		fault, connectErr = client.ReadRegisters(0x0002, 1, modbus.HOLDING_REGISTER)
		if connectErr != nil {
			if debug {
				fmt.Printf("[DEBUG] ConnectToBMS: fault status read failed: %v\n", connectErr)
			}
			continue
		} else {
			if debug {
				fmt.Printf("[DEBUG] ConnectToBMS: BMS responding, fault status=0x%04X\n", fault[0])
			}
			break
		}
	}

	if connectErr != nil || client == nil {
		fmt.Println("Retry Counter exceeded. Giving Up. Retry counter:", internal.ConnectionRetries)
		fmt.Println("Failed to connect to BMS. Check if")
		fmt.Println("-> VCC on SWD Interface has 2.5Volts!")
		fmt.Println("-> RX/TX is connected correctly via JTAG BMS Version Output in minicom/putty!")
		fmt.Println("-> TEST is connected to GND. Otherwise the BMS will sleep and not respond!")
		fmt.Println("-> IF the Battery has no errors, check DSG Voltage.")
		fmt.Println("Thanks for keeping the World a better place! ❤️")
		os.Exit(1)
	}

	return fault, nil
}

func ReadRegisters(client *modbus.ModbusClient, startAddress, quantity uint16) ([]uint16, error) {
	if internal.Debug {
		fmt.Printf("[DEBUG] ReadRegisters: startAddress=0x%04X quantity=%d\n", startAddress, quantity)
	}

	regs, readErr := client.ReadRegisters(startAddress, quantity, modbus.HOLDING_REGISTER)
	if readErr != nil {
		fmt.Println("Failed to read registers. Error:", readErr)
		return regs, readErr
	}

	if internal.Debug {
		fmt.Printf("[DEBUG] ReadRegisters: read %d registers successfully\n", len(regs))
	}

	internal.Registers = regs
	return internal.Registers, nil
}
