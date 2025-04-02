package main

import (
	"log"
	"time"

	"github.com/simonvetter/modbus"
)

// very fragile!
// TODO: GRAB Screen and update only that part
func liveData(client *modbus.ModbusClient, debug bool) {

	for {
		// 0x0000 is start otherwise the case statement will not work because 2 gets 0 there.
		if regs, err = readRegisters(client, 0, 44); err != nil {
			log.Printf("Failed to read registers: %v", err)
			if _, err := connectToBMS(client, debug); err != nil {
				log.Printf("Failed to connect to BMS: %v", err)
			}
		}
		getAndShowPassiveBMSData()
		time.Sleep(1 * time.Second)
	}

}
