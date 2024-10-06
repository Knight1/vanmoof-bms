package main

import (
	"fmt"
	"time"

	"github.com/simonvetter/modbus"
)

func createModbusClient(device string) (*modbus.ModbusClient, error) {
	// for an RTU (serial) device/bus
	return modbus.NewClient(&modbus.ClientConfiguration{
		URL:      fmt.Sprintf("rtu://%s", device),
		Speed:    9600,
		DataBits: 8,
		Parity:   modbus.PARITY_NONE,
		StopBits: 2,
		Timeout:  3000 * time.Millisecond,
	})
}
