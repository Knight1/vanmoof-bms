package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/simonvetter/modbus"
)

func main() {
	var client *modbus.ModbusClient
	var err error
	var regs []uint16
	var milliVolts float64

	// Define Threasholds
	var cellVoltageLow uint16 = 2500
	var cellVoltageHigh uint16 = 4300
	var cellVoltageImbalance uint16 = 1
	var cellVoltagePrevious uint16 = 0
	var packVoltageLow uint16 = 25000
	var packVoltageHigh uint16 = 43000

	// for an RTU (serial) device/bus
	client, err = modbus.NewClient(&modbus.ClientConfiguration{
		URL:      "rtu:///dev/serial0",
		Speed:    9600,
		DataBits: 8,
		Parity:   modbus.PARITY_NONE,
		StopBits: 2,
		Timeout:  300 * time.Millisecond,
	})

	if err != nil {
		// error out if client creation failed
	}

	// now that the client is created and configured, attempt to connect
	err = client.Open()
	if err != nil {
		// error out if we failed to connect/open the device
		// note: multiple Open() attempts can be made on the same client until
		// the connection succeeds (i.e. err == nil), calling the constructor again
		// is unnecessary.
		// likewise, a client can be opened and closed as many times as needed.
		fmt.Printf("failed to connect: %v\n", err)
		os.Exit(2)
	}

	client.SetUnitId(170)

	// Read all BMS ModBus Addresses
	regs, err = client.ReadRegisters(0x0, 95, modbus.HOLDING_REGISTER)
	if err != nil {
		fmt.Printf("failed to read registers: %v\n", err)
	}

	fmt.Println("BMS ModBus Addresses 0 to 95")
	for register, reg := range regs {
		fmt.Println("Register:", register, "Value:", reg)
	}

	fmt.Println("END DEBUG")

	// Voltage Cell Pack Monitoring
	for register, reg := range regs {
		milliVolts = float64(reg) / 1000
		if register == 4 {
			fmt.Println("Pack:", milliVolts, "Volts")
		} else if register == 27 {
			fmt.Println("Cell 1:", milliVolts, "Volts")
		} else if register == 28 {
			fmt.Println("Cell 2:", milliVolts, "Volts")
		} else if register == 29 {
			fmt.Println("Cell 3:", milliVolts, "Volts")
		} else if register == 30 {
			fmt.Println("Cell 4:", milliVolts, "Volts")
		} else if register == 31 {
			fmt.Println("Cell 5:", milliVolts, "Volts")
		} else if register == 32 {
			fmt.Println("Cell 6:", milliVolts, "Volts")
		} else if register == 33 {
			fmt.Println("Cell 7:", milliVolts, "Volts")
		} else if register == 34 {
			fmt.Println("Cell 8:", milliVolts, "Volts")
		} else if register == 35 {
			fmt.Println("Cell 9:", milliVolts, "Volts")
		} else if register == 36 {
			fmt.Println("Cell 10:", milliVolts, "Volts")
		} else if register <= 3 || register <= 26 || register >= 37 {
			// Skip everything else
			continue
		}

		// Verify if Voltage is to low, to high
		if register != 4 && reg < cellVoltageLow {
			fmt.Println("Voltage in Cell to LOW!", register)
		} else if register != 4 && reg > cellVoltageHigh {
			fmt.Println("Voltage in Cell to HIGH!", register)
		}

		// TODO: Define Pack Voltage Limits from BMS!
		if register == 4 && reg < packVoltageLow {
			fmt.Println("Voltage in Pack to LOW!")
		} else if register == 4 && reg > packVoltageHigh {
			fmt.Println("Voltage in Pack to HIGH!")
		}

		// TODO: Check Voltage imbalance trigger
		// Check for Voltage Imbalances in the Cells
		diff := math.Abs(float64(cellVoltagePrevious - reg))
		if diff > float64(cellVoltageImbalance) {
			//fmt.Println("Voltage Imbalance in Cells!")
			fmt.Println("DEBUG: Value ", reg, "differs from previous value by more than", diff)
		}
		cellVoltagePrevious = reg
		//fmt.Println("DEBUG: Value", reg, "differs from previous value by", diff)

	}

	for register, reg := range regs {
		if register == 2 {
			if reg != 0 {
				fmt.Println("ALARM: ", reg)
				fmt.Println("BMS SHUTDOWN!")
			} else {
				fmt.Println("BMS OK!")
			}
		}

	}

	// close the TCP connection/serial port
	client.Close()
}
