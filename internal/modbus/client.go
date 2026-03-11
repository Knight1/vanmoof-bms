package modbus

import (
	"bms/v2/internal"
	"fmt"
	"time"

	"github.com/simonvetter/modbus"
)

func CreateModbusClient(device string) (*modbus.ModbusClient, error) {
	url := fmt.Sprintf("rtu://%s", device)

	if internal.Debug {
		fmt.Printf("[DEBUG] CreateModbusClient: url=%s speed=9600 dataBits=8 parity=none stopBits=1 timeout=3000ms\n", url)
	}

	// for an RTU (serial) device/bus
	return modbus.NewClient(&modbus.ClientConfiguration{
		URL:      url,
		Speed:    9600,
		DataBits: 8,
		Parity:   modbus.PARITY_NONE,
		StopBits: 1,
		Timeout:  3000 * time.Millisecond,
	})
}
