package internal

import (
	"log"
	"time"

	"github.com/simonvetter/modbus"
)

// very fragile!
// TODO: GRAB Screen and update only that part
func LiveData(client *modbus.ModbusClient, debug bool) {

	for {
		// 0x0000 is start otherwise the case statement will not work because 2 gets 0 there.
		if Regs, err = ReadRegisters(client, 0, 44); err != nil {
			log.Printf("Failed to read registers: %v", err)
			if _, err := ConnectToBMS(client, debug); err != nil {
				log.Printf("Failed to connect to BMS: %v", err)
			}
		}
		GetAndShowPassiveBMSData()
		time.Sleep(1 * time.Second)
	}

}
