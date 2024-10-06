package main

import (
	"fmt"
	"log"
	"math"
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

const (
	connectionRetries      uint8 = 5
	connectionRetryDelay         = 500 * time.Millisecond
	DynaPackVanMoofSlaveID uint8 = 170
)

var (
	regs []uint16
	err  error
)

func main() {
	var client *modbus.ModbusClient
	var milliVolts float64

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

	// TODO: make it possible to pass the device as a parameter
	client, err = createModbusClient("/dev/serial0")
	if err != nil {
		log.Fatalf("Failed to create Modbus client.Maybe the Probe is disconnected? Check the Address of the Device! Error: %v", err)
	}
	defer client.Close()

	// DEBUG
	fmt.Println("Modbus client created")

	if err, regs = connectToBMS(client); err != nil {
		log.Fatalf("Failed to connect to BMS: %v", err)
	}

	// Debug Output
	fmt.Println("-- BEGIN DEBUG --")
	fmt.Println("BMS ModBus Addresses 0 to 95")
	for register, reg := range regs {
		fmt.Println("Register:", register, "Value:", reg)
	}

	fmt.Println("-- END DEBUG --")

	fmt.Println("-- BEGIN BMS PASSIVE STATUS --")

	for register, value := range regs {
		switch register {
		case 2:
			// TODO: Check what this is, should be a hex output
			fmt.Println("BMS Fault Status:", value)
		case 3:
			batteryTemperature := calculateCelsius(value)
			fmt.Println("Battery Temperature:", batteryTemperature, "°C")
		case 4:
			batteryVoltage := value
			fmt.Println("Battery Voltage:", batteryVoltage, "mV")
		case 5:
			fmt.Println("Real State of Charge:", value, "%")
		case 6:
			fmt.Println("Current:", calculateAmperes(value), "mA")
		case 7:
			// TODO: should be a hex output
			fmt.Println("Charging Status:", value)
		case 8:
			fmt.Println("Charging on/off:", value)
		case 9:
			fmt.Println("Test Mode:", value)
		case 10:
			// TODO: Compare this. Hardware and Software Version are broken
			fmt.Println("Hardware Version:", value)
		case 11:
			// TODO: Compare this. Hardware and Software Version are broken
			fmt.Println("Software Version:", value)
		case 15:
			fmt.Println("Normal Capacity:", value, "mAh")
		case 19:
			fmt.Println("Cycle Count:", value)
		case 27:
			fmt.Println("Cell 1 Voltage:", value, "mV")
		case 28:
			fmt.Println("Cell 2 Voltage:", value, "mV")
		case 29:
			fmt.Println("Cell 3 Voltage:", value, "mV")
		case 30:
			fmt.Println("Cell 4 Voltage:", value, "mV")
		case 31:
			fmt.Println("Cell 5 Voltage:", value, "mV")
		case 32:
			fmt.Println("Cell 6 Voltage:", value, "mV")
		case 33:
			fmt.Println("Cell 7 Voltage:", value, "mV")
		case 34:
			fmt.Println("Cell 8 Voltage:", value, "mV")
		case 35:
			fmt.Println("Cell 9 Voltage:", value, "mV")
		case 36:
			fmt.Println("Cell 10 Voltage:", value, "mV")
		case 37:
			fmt.Println("Temperature Sensor 1:", calculateCelsius(value), "°C")
		case 38:
			fmt.Println("Temperature Sensor 2:", calculateCelsius(value), "°C")
		case 39:
			fmt.Println("Discharge MOSFET Temperature:", calculateCelsius(value), "°C")
		case 40:
			fmt.Println("Warning Status:", value)
		case 41:
			cellVoltageHighest = value
			fmt.Println("Maximum Battery Voltage:", cellVoltageHighest, "mV")
		case 42:
			cellVoltageLowest = value
			fmt.Println("Minimum Battery Voltage:", cellVoltageLowest, "mV")
			if math.Abs(float64(cellVoltageHighest-cellVoltageLowest)) > 20 {
				fmt.Println("WARNING: Voltage Imbalance in Cells!")
			}
		case 43:
			fmt.Println("Cell Balance:", value)
		case 44:
			//TODO: check if this is correct
			fmt.Println("Bootloader Version:", value)
		default:
			continue
		}
	}

	fmt.Println("-- END BMS PASSIVE STATUS --")

	fmt.Println("-- BEGIN BMS FLASH STATUS --")

	//
	for register, value := range regs {
		switch register {
		case 48:
			// TODO: Check what this is, should be a hex output
			fmt.Println("Fault Status:", value)
		case 49:
			// TODO: Reports records that are too high
			fmt.Println("Battery Temperature Sensor 1 Record:", calculateCelsius(value), "°C")
		case 50:
			// TODO: Reports records that are too high
			fmt.Println("Battery Temperature Sensor 2 Record:", calculateCelsius(value), "°C")
		case 51:
			// TODO: Reports records that are too high
			fmt.Println("MOSFET Temperature Record:", calculateCelsius(value), "°C")
		case 52:
			fmt.Println("Battery Voltage Record:", value, "mV")
		case 53:
			fmt.Println("Current: ", calculateAmperes(value), "mA")
		case 54:
			fmt.Println("Full Charge Capacity:", value, "%")
		case 55:
			fmt.Println("Remaining Capacity:", value, "%")
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
			fmt.Println("Maximal recorded Charging Current:", calculateAmperes(value), "mA")
		case 88:
			fmt.Println("Maximal recorded Discharging Current:", calculateAmperes(value), "mA")
		case 89:
			fmt.Println("Maximal recorded Cell Temperature:", calculateCelsius(value), "°C")
		case 90:
			fmt.Println("Minimal recorded Cell Temperature:", calculateCelsius(value), "°C")
		case 91:
			fmt.Println("Maximal recorded MOSFET Temperature:", calculateCelsius(value), "°C")
		default:
			continue
		}
	}

	fmt.Println("-- END BMS FLASH STATUS --")

	fmt.Println("-- BEGIN TRIGGER VALUES --")

	// Checking Proteection Statusses
	for register, value := range regs {
		switch register {
		case 03:

		case 45:
			bmsUndervoltageCellProtection1 = value
			fmt.Println("Undervoltage Cell Protection 1 Trigger Value:", bmsUndervoltageCellProtection1, "mV")
		case 46:
			bmsUndervoltageCellProtection2 = value
			fmt.Println("Undervoltage Cell Protection 2 Trigger Value:", bmsUndervoltageCellProtection2, "mV")
		case 47:
			bmsUndervoltageCellShutdown = value
			fmt.Println("Undervoltage Cell Shutdown Trigger Value:", bmsUndervoltageCellShutdown, "mV")
		case 48:
			bmsOvervoltageCellProtection1 = value
			fmt.Println("Overvoltage Cell Protection 1 Trigger Value:", bmsOvervoltageCellProtection1, "mV")
		case 71:
			fmt.Println("Discharge Over Temperature Protection:", value)
		case 72:
			fmt.Println("Discharge Under Temperature Protection:", value)
		case 73:
			fmt.Println("Charging Over Temperature Protection:", value)
		case 74:
			fmt.Println("Current Under Temperature Protection:", value)
		case 75:
			fmt.Println("Discharge Over Current Protection 1:", value)
		case 76:
			fmt.Println("Discharge Over Current Protection 2:", value)
		case 77:
			fmt.Println("Charging Over Current Protection 1:", value)
		case 78:
			fmt.Println("Charging Over Current Protection 2:", value)
		case 79:
			fmt.Println("Over Voltage Cell Protection 1:", value)
		case 80:
			fmt.Println("Over Voltage Cell Protection 2:", value)
		case 81:
			fmt.Println("Under Voltage Cell Protection 1:", value)
		case 82:
			fmt.Println("Under Voltage Cell Protection 2:", value)
		case 83:
			fmt.Println("Peak Discharge Over Current Protection:", value)
		case 84:
			fmt.Println("Peak Discharge Source/safety? Current Protection:", value)
		case 85:
			fmt.Println("MOSFET (Metal Oxide Semiconductor Field-Effect Transistors) Output Temperature Protection:", value)
		case 86:
			fmt.Println("Source/safety? Current Protection:", value)
		default:
			continue
		}
	}

	fmt.Println("-- END TRIGGER VALUES --")
	// TODO: Sort Output
	time.Sleep(time.Millisecond * 50)

	fmt.Println("-- BEGIN LIVE VOLTAGES --")

	// Voltage Cell Pack Monitoring
	for register, value := range regs {
		milliVolts = float64(value) / 1000
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
		if register != 4 && value < cellVoltageLow {
			fmt.Println("Voltage in Cell to LOW!", register)
		} else if register != 4 && value > cellVoltageHigh {
			fmt.Println("Voltage in Cell to HIGH!", register)
		}

		// TODO: Define Pack Voltage Limits from BMS!
		if register == 4 && value < packVoltageLow {
			fmt.Println("Voltage in Pack to LOW!")
		} else if register == 4 && value > packVoltageHigh {
			fmt.Println("Voltage in Pack to HIGH!")
		}

		// TODO: Check Voltage imbalance trigger
		// Check for Voltage Imbalances in the Cells
		diff := math.Abs(float64(cellVoltagePrevious - value))
		if diff > float64(cellVoltageImbalance) {
			//fmt.Println("Voltage Imbalance in Cells!")
			fmt.Println("DEBUG: Value ", value, "differs from previous value by more than", diff)
		}
		cellVoltagePrevious = value
		//fmt.Println("DEBUG: Value", value, "differs from previous value by", diff)

	}

	fmt.Println("-- END LIVE VOLTAGES --")

	for register, value := range regs {
		if register == 2 {
			if value != 0 {
				fmt.Println("ALARM: ", value)
				fmt.Println("BMS SHUTDOWN!")
			} else {
				fmt.Println("BMS STATUS OK!")
			}
		}

	}

	// close the TCP connection/serial port
	client.Close()
}

func calculateCelsius(value uint16) float64 {
	return float64(value-2731) / 10
}

func calculateAmperes(value uint16) float64 {
	return float64(value) / 10
}
