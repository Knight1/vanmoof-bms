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
	const DynaPackVanMoofSlaveID uint8 = 170
	const connectionRetries uint8 = 5
	const connectionRetryDelay = 500 * time.Millisecond

	// Define Threasholds
	const cellVoltageLow uint16 = 2500
	const cellVoltageHigh uint16 = 4300
	var cellVoltageImbalance uint16 = 5
	var cellVoltagePrevious uint16 = 0
	const packVoltageLow uint16 = 25000
	const packVoltageHigh uint16 = 43000

	// Initialize Global BMS Error Statusses
	var bmsUndervoltageCellProtection1 uint16 = 0
	var bmsUndervoltageCellProtection2 uint16 = 0
	var bmsUndervoltageCellShutdown uint16 = 0
	var bmsOvervoltageCellProtection1 uint16 = 0

	// Define holding registers
	var cellVoltageHighest = uint16(0)
	var cellVoltageLowest = uint16(0)

	fmt.Println("Starting VanMoof / DynaPack BMS Toolkit")
	fmt.Println("Go version:", runtime.Version(), "Version:", GoVersion, "BuildTime:", BuildTime, "CommitHash:", CommitHash, "Git:", GitTag, "GOOS:", GOOS, "GOARCH:", GOARCH)

	client, err = createModbusClient()
	if err != nil {
		log.Fatalf("Failed to create Modbus client.Maybe the Probe is disconnected? Check the Address of the Device! Error: %v", err)
	}
	defer client.Close()

	// DEBUG
	fmt.Println("Modbus client created")

	// Read all BMS ModBus Addresses
	for attempt := 0; attempt < int(connectionRetries); attempt++ {
		fmt.Println("Trying to connect to BMS via ModBus. Attempt:", attempt+1)
		// Try to establish a connection to the BMS. If it fails, retry.
		err = client.Open()
		if err != nil {
			fmt.Println("Failure opening client. Waiting and retrying in 500ms.")
			time.Sleep(connectionRetryDelay)
			continue
		}

		defer client.Close()

		fmt.Println("Modbus client opened")
		//DEBUG
		fmt.Println("Client:", client)
		fmt.Println("Reading Registers... Please wait!")

		// VanMoof / DynaPack BMS uses slave-id 170
		client.SetUnitId(DynaPackVanMoofSlaveID)

		regs, err = client.ReadRegisters(0x0, 95, modbus.HOLDING_REGISTER)
		if err != nil {
			fmt.Println("Failed to read registers. Error:", err)
			continue
		} else {
			break
		}
	}

	if err != nil || client == nil {
		fmt.Println("Failed to connect to BMS. Check if VCC on SWD Interface has 2.5Volts!")
		fmt.Println("Verify that RX/TX is connected correctly via JTAG BMS Version Output!")
		fmt.Println("Also make sure TEST is connected to GND. Otherwise the BMS will sleep and not respond!")
		fmt.Println("Thanks for keeping the World a better place!")
		os.Exit(1)
	}

	// Debug Output
	fmt.Println("-- BEGIN DEBUG --")
	fmt.Println("BMS ModBus Addresses 0 to 95")
	for register, reg := range regs {
		fmt.Println("Register:", register, "Value:", reg)
	}

	fmt.Println("-- END DEBUG --")

	fmt.Println("-- BEGIN BMS PASSIVE STATUS --")

	for register, reg := range regs {
		switch register {
		case 2:
			// TODO: Check what this is, should be a hex output
			fmt.Println("BMS Fault Status:", reg)
		case 3:
			batteryTemperature := (reg - 2731) / 10
			fmt.Println("Battery Temperature:", batteryTemperature, "°C")
		case 4:
			batteryVoltage := reg
			fmt.Println("Battery Voltage:", batteryVoltage, "mV")
		case 5:
			fmt.Println("Real State of Charge:", reg, "%")
		case 6:
			fmt.Println("Current:", (reg * 10), "mA")
		case 7:
			// TODO: should be a hex output
			fmt.Println("Charging Status:", reg)
		case 8:
			fmt.Println("Charging on/off:", reg)
		case 9:
			fmt.Println("Test Mode:", reg)
		case 10:
			// TODO: Compare this. Hardware and Software Version are broken
			fmt.Println("Hardware Version:", reg)
		case 11:
			// TODO: Compare this. Hardware and Software Version are broken
			fmt.Println("Software Version:", reg)
		case 15:
			fmt.Println("Normal Capacity:", reg, "mAh")
		case 19:
			fmt.Println("Cycle Count:", reg)
		case 27:
			fmt.Println("Cell 1 Voltage:", reg, "mV")
		case 28:
			fmt.Println("Cell 2 Voltage:", reg, "mV")
		case 29:
			fmt.Println("Cell 3 Voltage:", reg, "mV")
		case 30:
			fmt.Println("Cell 4 Voltage:", reg, "mV")
		case 31:
			fmt.Println("Cell 5 Voltage:", reg, "mV")
		case 32:
			fmt.Println("Cell 6 Voltage:", reg, "mV")
		case 33:
			fmt.Println("Cell 7 Voltage:", reg, "mV")
		case 34:
			fmt.Println("Cell 8 Voltage:", reg, "mV")
		case 35:
			fmt.Println("Cell 9 Voltage:", reg, "mV")
		case 36:
			fmt.Println("Cell 10 Voltage:", reg, "mV")
		case 37:
			fmt.Println("Temperature Sensor 1:", (reg-2731)/10, "°C")
		case 38:
			fmt.Println("Temperature Sensor 2:", (reg-2731)/10, "°C")
		case 39:
			fmt.Println("Discharge MOSFET Temperature:", (reg-2731)/10, "°C")
		case 40:
			fmt.Println("Warning Status:", reg)
		case 41:
			cellVoltageHighest = reg
			fmt.Println("Maximum Battery Voltage:", cellVoltageHighest, "mV")
		case 42:
			cellVoltageLowest = reg
			fmt.Println("Minimum Battery Voltage:", cellVoltageLowest, "mV")
			if math.Abs(float64(cellVoltageHighest-cellVoltageLowest)) > 20 {
				fmt.Println("WARNING: Voltage Imbalance in Cells!")
			}
		case 43:
			fmt.Println("Cell Balance:", reg)
		case 44:
			//TODO: check if this is correct
			fmt.Println("Bootloader Version:", reg)
		default:
			continue
		}
	}

	fmt.Println("-- END BMS PASSIVE STATUS --")

	fmt.Println("-- BEGIN BMS FLASH STATUS --")

	//
	for register, reg := range regs {
		switch register {
		case 48:
			// TODO: Check what this is, should be a hex output
			fmt.Println("Fault Status:", reg)
		case 49:
			// TODO: Reports records that are too high
			fmt.Println("Battery Temperature Sensor 1 Record:", (reg-2731)/10, "°C")
		case 50:
			// TODO: Reports records that are too high
			fmt.Println("Battery Temperature Sensor 2 Record:", (reg-2731)/10, "°C")
		case 51:
			// TODO: Reports records that are too high
			fmt.Println("MOSFET Temperature Record:", (reg-2731)/10, "°C")
		case 52:
			fmt.Println("Battery Voltage Record:", reg, "mV")
		case 53:
			fmt.Println("Current: ", (reg * 10), "mA")
		case 54:
			fmt.Println("Full Charge Capacity:", reg, "%")
		case 55:
			fmt.Println("Remaining Capacity:", reg, "%")
		case 56:
			// TODO: shows 0 insted of Real State of Charge
		case 57:
			// TODO: shows 0 insted of Absolute State of Charge
		case 58:
			// TODO: shows 0 instead of Cycle Count
		case 59:
			// register 59 to 68 are Cell Voltages from flash
			// Values are either 0 or something odd
		case 87:
			fmt.Println("Maximal recorded Charging Current:", (reg * 10), "mA")
		case 88:
			fmt.Println("Maximal recorded Discharging Current:", (reg * 10), "mA")
		case 89:
			fmt.Println("Maximal recorded Cell Temperature:", (reg-2731)/10, "°C")
		case 90:
			fmt.Println("Minimal recorded Cell Temperature:", (reg-2731)/10, "°C")
		case 91:
			fmt.Println("Maximal recorded MOSFET Temperature:", (reg-2731)/10, "°C")
		default:
			continue
		}
	}

	fmt.Println("-- END BMS FLASH STATUS --")

	fmt.Println("-- BEGIN TRIGGER VALUES --")

	// Checking Proteection Statusses
	for register, reg := range regs {
		switch register {
		case 03:

		case 45:
			bmsUndervoltageCellProtection1 = reg
			fmt.Println("Undervoltage Cell Protection 1 Trigger Value:", bmsUndervoltageCellProtection1, "mV")
		case 46:
			bmsUndervoltageCellProtection2 = reg
			fmt.Println("Undervoltage Cell Protection 2 Trigger Value:", bmsUndervoltageCellProtection2, "mV")
		case 47:
			bmsUndervoltageCellShutdown = reg
			fmt.Println("Undervoltage Cell Shutdown Trigger Value:", bmsUndervoltageCellShutdown, "mV")
		case 48:
			bmsOvervoltageCellProtection1 = reg
			fmt.Println("Overvoltage Cell Protection 1 Trigger Value:", bmsOvervoltageCellProtection1, "mV")
		case 71:
			fmt.Println("Discharge Over Temperature Protection:", reg)
		case 72:
			fmt.Println("Discharge Under Temperature Protection:", reg)
		case 73:
			fmt.Println("Charging Over Temperature Protection:", reg)
		case 74:
			fmt.Println("Current Under Temperature Protection:", reg)
		case 75:
			fmt.Println("Discharge Over Current Protection 1:", reg)
		case 76:
			fmt.Println("Discharge Over Current Protection 2:", reg)
		case 77:
			fmt.Println("Charging Over Current Protection 1:", reg)
		case 78:
			fmt.Println("Charging Over Current Protection 2:", reg)
		case 79:
			fmt.Println("Over Voltage Cell Protection 1:", reg)
		case 80:
			fmt.Println("Over Voltage Cell Protection 2:", reg)
		case 81:
			fmt.Println("Under Voltage Cell Protection 1:", reg)
		case 82:
			fmt.Println("Under Voltage Cell Protection 2:", reg)
		case 83:
			fmt.Println("Peak Discharge Over Current Protection:", reg)
		case 84:
			fmt.Println("Peak Discharge Source/safety? Current Protection:", reg)
		case 85:
			fmt.Println("MOSFET (Metal Oxide Semiconductor Field-Effect Transistors) Output Temperature Protection:", reg)
		case 86:
			fmt.Println("Source/safety? Current Protection:", reg)
		default:
			continue
		}
	}

	fmt.Println("-- END TRIGGER VALUES --")
	// TODO: Sort Output
	time.Sleep(time.Millisecond * 50)

	fmt.Println("-- BEGIN LIVE VOLTAGES --")

	// Voltage Cell Pack Monitoring
	for register, reg := range regs {
		milliVolts = float64(reg) / 1000
		switch register {
		case 4:
			fmt.Println("Pack:", milliVolts, "Volts")
		case 27:
			fmt.Println("Cell 1:", milliVolts, "Volts")
		case 28:
			fmt.Println("Cell 2:", milliVolts, "Volts")
		case 29:
			fmt.Println("Cell 3:", milliVolts, "Volts")
		case 30:
			fmt.Println("Cell 4:", milliVolts, "Volts")
		case 31:
			fmt.Println("Cell 5:", milliVolts, "Volts")
		case 32:
			fmt.Println("Cell 6:", milliVolts, "Volts")
		case 33:
			fmt.Println("Cell 7:", milliVolts, "Volts")
		case 34:
			fmt.Println("Cell 8:", milliVolts, "Volts")
		case 35:
			fmt.Println("Cell 9:", milliVolts, "Volts")
		case 36:
			fmt.Println("Cell 10:", milliVolts, "Volts")

			// TODO: Check if one cell is below these values

		default:
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

	fmt.Println("-- END LIVE VOLTAGES --")

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
