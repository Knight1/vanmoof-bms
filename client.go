package main

import (
	"time"

	"github.com/simonvetter/modbus"
)

func createModbusClient() (*modbus.ModbusClient, error) {
	// for an RTU (serial) device/bus
	return modbus.NewClient(&modbus.ClientConfiguration{
		URL:      "rtu:///dev/serial0",
		Speed:    9600,
		DataBits: 8,
		Parity:   modbus.PARITY_NONE,
		StopBits: 2,
		Timeout:  3000 * time.Millisecond,
	})
}
