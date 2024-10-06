package main

import (
	"fmt"
	"time"

	"github.com/simonvetter/modbus"
)

// very fragile!
func liveData(client *modbus.ModbusClient, debug bool) {

	for {
		// 0x0000 is start otherwise the case statement will not work because 2 gets 0 there.
		if regs, err = readRegisters(client, 0, 44); err != nil {
			fmt.Println("Failed to read registers: %v", err)
			if err = connectToBMS(client, debug); err != nil {
				fmt.Println("Failed to connect to BMS: %v", err)
			}
		}
		getAndShowPassiveBMSData()
		time.Sleep(1 * time.Second)
	}

}
