package main

// Modbus / serial action dispatch for the S3/S4 DynaPack BMS (packages
// internal/modbus and internal/serial), selected with `--protocol modbus`
// (the default).

import (
	"bms/v2/internal"
	"bms/v2/internal/modbus"
	"bms/v2/internal/serial"
	"fmt"
	"log"
	"os"

	mbclient "github.com/simonvetter/modbus"
)

func runModbus(serialPort, action, firmwareFile, logFile, esn, esnDate string, calibrateCurrent int, overview bool) {
	// Serial UART string commands (S3/S4 only, no Modbus connection needed)
	if action == "clearPF" {
		serial.ClearPF(serialPort)
		os.Exit(0)
	} else if action == "gpioOn" {
		serial.SetGPIOOn(serialPort)
		os.Exit(0)
	} else if action == "gpioOff" {
		serial.SetGPIOOff(serialPort)
		os.Exit(0)
	} else if action == "detectOn" {
		serial.SetDetectPinOn(serialPort)
		os.Exit(0)
	} else if action == "detectOff" {
		serial.SetDetectPinOff(serialPort)
		os.Exit(0)
	} else if action == "keyInOn" {
		serial.SetKeyInOn(serialPort)
		os.Exit(0)
	} else if action == "keyInOff" {
		serial.SetKeyInOff(serialPort)
		os.Exit(0)
	} else if action == "resetBMS" {
		serial.ResetBMS(serialPort)
		os.Exit(0)
	} else if action == "resetESN" {
		serial.ResetESN(serialPort)
		os.Exit(0)
	} else if action == "clearLog" {
		serial.ClearLog(serialPort)
		os.Exit(0)
	} else if action == "calibrateDSG" {
		serial.CalibrateDischargeCurrent(serialPort, calibrateCurrent)
		os.Exit(0)
	} else if action == "calibrateCHG" {
		serial.CalibrateChargeCurrent(serialPort, calibrateCurrent)
		os.Exit(0)
	}

	var client *mbclient.ModbusClient

	var err error

	// Creates the Modbus connection with all relevant parameters and the port to use
	client, err = modbus.CreateModbusClient(serialPort)
	if err != nil {
		log.Fatalf("Failed to create Modbus client. Maybe the Probe is disconnected? Check the Address of the Device! Error: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Failed to close Modbus client: %v", err)
		}
	}()

	// DEBUG
	if internal.Debug {
		fmt.Println("Modbus client created")
	}

	// Loop for connecting to the bms. Loops until it reaches the end of connectionRetries
	if _, err := modbus.ConnectToBMS(client, internal.Debug); err != nil {
		log.Fatalf("Failed to connect to BMS: %v", err)
	}

	// Modbus register write commands
	if action == "debug" {
		modbus.TurnDebugOn(client)
		os.Exit(0)
	} else if action == "debugoff" {
		modbus.TurnDebugOff(client)
		os.Exit(0)
	} else if action == "discharge" {
		modbus.TurnDischargingOn(client)
		os.Exit(0)
	} else if action == "dischargeoff" {
		modbus.TurnDischargingOff(client)
		os.Exit(0)
	} else if action == "chargeOn" {
		modbus.TurnChargeMOSOn(client)
		os.Exit(0)
	} else if action == "chargeOff" {
		modbus.TurnChargeMOSOff(client)
		os.Exit(0)
	} else if action == "writeESN" {
		modbus.WriteESNAndDate(client, esn, esnDate)
		os.Exit(0)
	} else if action == "resetESNModbus" {
		modbus.ResetESNModbus(client)
		os.Exit(0)
	} else if action == "resetMCU" {
		modbus.ResetMCU(client)
		os.Exit(0)
	} else if action == "shipMode" {
		modbus.ShipMode(client)
		os.Exit(0)
	} else if action == "ship" {
		modbus.ShipAndDischargeTurnOff(client)
		os.Exit(0)
	} else if action == "exportLog" {
		modbus.ExportReadLog(client, logFile)
		os.Exit(0)
	} else if action == "updateFirmware" {
		modbus.UpdateFirmware(client, firmwareFile)
		os.Exit(0)
	}

	if internal.Registers, err = modbus.ReadRegisters(client, 0, 95); err != nil {
		log.Fatalf("Failed to read registers: %v", err)
	}

	// Debug Output
	if internal.Debug {
		fmt.Println("-- BEGIN DEBUG --")
		fmt.Println("BMS ModBus Addresses 0 to 94")
		for register, reg := range internal.Registers {
			fmt.Println("Register:", register, "Value:", reg)
		}

		fmt.Println("-- END DEBUG --")
	}

	if action == "live" {
		modbus.LiveData(client, internal.Debug)
	}

	if overview {
		modbus.ShowOverview()
		os.Exit(0)
	}

	modbus.GetAndShowPassiveBMSData()

	modbus.GetAndShowFlashBMSData()

	modbus.GetAndShowProtectionBMSValues()

	modbus.GetAndShowPassiveVoltages()
}
