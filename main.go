package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/simonvetter/modbus"
)

// build flags
var (
	BuildTime  string
	CommitHash string
	GitTag     string
	GOOS       string
	GOARCH     string
	GoVersion  string
)

func main() {
	var client *modbus.ModbusClient
	var err error
	var regs []uint16
	var milliVolts float64

	// Define Threasholds
	var cellVoltageLow uint16 = 2500
	var cellVoltageHigh uint16 = 4300
	var cellVoltageImbalance uint16 = 5
	var cellVoltagePrevious uint16 = 0
	var packVoltageLow uint16 = 25000
	var packVoltageHigh uint16 = 43000

	// Initialize Global BMS Error Statusses
	var bmsUndervoltageCellProtection1 uint16 = 0
	var bmsUndervoltageCellProtection2 uint16 = 0
	var bmsUndervoltageCellShutdown uint16 = 0
	var bmsOvervoltageCellProtection1 uint16 = 0

	fmt.Println("Starting VanMoof / DynaPack BMS Toolkit")
	fmt.Println("Go version:", runtime.Version(), "Version:", GoVersion, "BuildTime:", BuildTime, "CommitHash:", CommitHash, "Git:", GitTag, "GOOS:", GOOS, "GOARCH:", GOARCH)

	// for an RTU (serial) device/bus
	client, err = modbus.NewClient(&modbus.ClientConfiguration{
		URL:      "rtu:///dev/serial0",
		Speed:    9600,
		DataBits: 8,
		Parity:   modbus.PARITY_NONE,
		StopBits: 2,
		Timeout:  3000 * time.Millisecond,
	})

	if err != nil {
		// error out if client creation failed
		log.Fatal("failed to create Modbus client. Maybe the Probe is disconnected? Check the Address of the Device! Error:", err)
	}

	// DEBUG
	fmt.Println("Modbus client created")

	// Read all BMS ModBus Addresses
	for i := 0; i < 5; i++ {
		// Try to establish a connection to the BMS. If it fails, retry.
		for i := 0; i < 50; i++ {
			err = client.Open()
			if err != nil {
				time.Sleep(time.Millisecond * 500)
				continue
			} else {
				break
			}
		}

		// if the client is nil, error out
		if client == nil {
			log.Fatal("Failure opening client")
		}

		fmt.Println("Modbus client opened")
		//DEBUG
		fmt.Println("Client:", client)
		fmt.Println("Reading Registers... Attempts:")

		// VanMoof / DynaPack BMS uses slave-id 170
		client.SetUnitId(170)

		regs, err = client.ReadRegisters(0x0, 95, modbus.HOLDING_REGISTER)
		if err != nil {
			continue
		} else {
			break
		}
	}

	if err != nil {
		fmt.Println("Failed to connect to BMS. Check if VCC on SWD Interface has 2.5Volts!")
		fmt.Println("Verify that RX/TX is connected correctly via JTAG BMS Version Output!")
		fmt.Println("Also make sure TEST is connected to GND. Otherwise the BMS will sleep and not respond!")
		fmt.Println("Thanks for keeping the World a better place!")
		client.Close()
		os.Exit(1)
	}

	// Debug Output

	fmt.Println("-- BEGIN DEBUG --")
	fmt.Println("BMS ModBus Addresses 0 to 95")
	for register, reg := range regs {
		fmt.Println("Register:", register, "Value:", reg)
	}

	fmt.Println("-- END DEBUG --")
	fmt.Println("-- BEGIN TRIGGER VALUES --")

	// Checking Cell Voltages

	// Checking Proteection Statusses
	for register, reg := range regs {
		if register == 45 {
			bmsUndervoltageCellProtection1 = reg
			fmt.Println("Undervoltage Cell Protection 1 Trigger Value:", bmsUndervoltageCellProtection1, "mV")
		} else if register == 46 {
			bmsUndervoltageCellProtection2 = reg
			fmt.Println("Undervoltage Cell Protection 2 Trigger Value:", bmsUndervoltageCellProtection2, "mV")
		} else if register == 47 {
			bmsUndervoltageCellShutdown = reg
			fmt.Println("Undervoltage Cell Shutdown Trigger Value:", bmsUndervoltageCellShutdown, "mV")
		} else if register == 48 {
			bmsOvervoltageCellProtection1 = reg
			fmt.Println("Overvoltage Cell Protection 1 Trigger Value:", bmsOvervoltageCellProtection1, "mV")
		} else if register == 67 {
			fmt.Println("Discharge Over Temperature Protection:", reg)
		} else if register == 68 {
			fmt.Println("Discharge Under Temperature Protection:", reg)
		} else if register == 69 {
			fmt.Println("Charging Over Temperature Protection:", reg)
		} else if register == 70 {
			fmt.Println("Current Under Temperature Protection:", reg)
		} else if register == 71 {
			fmt.Println("Discharge Over Current Protection 1:", reg)
		} else if register == 72 {
			fmt.Println("Discharge Over Current Protection 2:", reg)
		} else if register == 73 {
			fmt.Println("Charging Over Current Protection 1:", reg)
		} else if register == 74 {
			fmt.Println("Charging Over Current Protection 2:", reg)
		} else if register == 75 {
			fmt.Println("Over Voltage Protection 1:", reg)
		} else if register == 76 {
			fmt.Println("Over Voltage Protection 2:", reg)
		} else if register == 77 {
			fmt.Println("Undervoltage Cell Protection 1:", reg)
		} else if register == 78 {
			fmt.Println("Undervoltage Cell Protection 2:", reg)
		} else if register == 79 {
			fmt.Println("Peak Discharge Over Current Protection:", reg)
		} else if register == 80 {
			fmt.Println("Peak Discharge Source/safety? Current Protection:", reg)
		} else if register == 81 {
			fmt.Println("MOSFET (Metal Oxide Semiconductor Field-Effect Transistors) Output Temperature Protection:", reg)
		} else if register == 82 {
			fmt.Println("Source/safety? Current Protection:", reg)
		} else {
			continue
		}
	}

	fmt.Println("-- END TRIGGER VALUES --")
	// TODO: Sort Output
	time.Sleep(time.Millisecond * 50)

	fmt.Println("-- BEGIN VOLTAGES --")

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

			// TODO: Check if one cell is below these values

		} else {
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

	fmt.Println("-- END VOLTAGES --")

	for register, reg := range regs {
		if register == 2 {
			if reg != 0 {
				fmt.Println("ALARM: ", reg)
				fmt.Println("BMS SHUTDOWN!")
			} else {
				fmt.Println("BMS STATUS OK!")
			}
		}

	}

	// close the TCP connection/serial port
	client.Close()
}
