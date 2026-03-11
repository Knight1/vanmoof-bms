package modbus

import (
	"bms/v2/internal"
	"fmt"
	"log"
	"time"

	"github.com/simonvetter/modbus"
)

// very fragile!
// TODO: GRAB Screen and update only that part
func LiveData(client *modbus.ModbusClient, debug bool) {
	if internal.Debug {
		fmt.Println("[DEBUG] LiveData: starting continuous read loop")
	}

	for {
		if internal.Debug {
			fmt.Println("[DEBUG] LiveData: reading registers 0-44")
		}
		// 0x0000 is start otherwise the case statement will not work because 2 gets 0 there.
		if _, readErr := ReadRegisters(client, 0, 45); readErr != nil {
			log.Printf("Failed to read registers: %v", readErr)
			if internal.Debug {
				fmt.Println("[DEBUG] LiveData: reconnecting to BMS")
			}
			if _, err := ConnectToBMS(client, debug); err != nil {
				log.Printf("Failed to connect to BMS: %v", err)
			}
			time.Sleep(1 * time.Second)
			continue
		}
		GetAndShowPassiveBMSData()
		time.Sleep(1 * time.Second)
	}

}
