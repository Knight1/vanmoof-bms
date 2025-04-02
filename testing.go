package main

import (
	"fmt"
	"github.com/simonvetter/modbus"
	"os"
)

func testingFunc(client *modbus.ModbusClient) {
	registers, _ := client.ReadRegisters(0x2, 1, modbus.HOLDING_REGISTER)
	// Process and display the register value
	if len(registers) > 0 {
		value := registers[0]
		fmt.Printf("Register 0x%X ('Fault Status'): 0x%04X\n", 0x2, value)

		// Decode flags (bitwise operations)
		flags := []string{"DOTP", "DUTP", "COTP", "CUTP", "DOCP1", "DOCP2", "COCP1", "COCP2", "OVP1", "OVP2", "UVP1", "UVP2", "PDOCP", "PDSCP", "MOTP", "SCP"}
		for i, flag := range flags {
			if value&(1<<i) != 0 {
				fmt.Printf(" - %s is set\n", flag)
			}
		}
	} else {
		fmt.Println("No data returned")
	}
	os.Exit(0)
}
