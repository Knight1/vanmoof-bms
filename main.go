package main

import (
	"bms/v2/internal"
	"bms/v2/internal/modbus"
	"bms/v2/internal/serial"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	mbclient "github.com/simonvetter/modbus"
)

func main() {
	var client *mbclient.ModbusClient

	flag.BoolVar(&internal.Debug, "debug", false, "Enable Debug Output")
	serialPort := flag.String("serial-port", "/dev/serial0", "Serial device URL (e.g., /dev/serial0)")
	action := flag.String("action", "show", "Action to perform (clearPF, detectOn, detectOff, gpioOn, gpioOff, keyInOn, keyInOff, live, resetBMS, resetESN, show, showPorts or writeESN)")
	//firmwareFile := flag.String("firmwareFile", "", "Firmware File to flash to BMS Chip.")
	esn := flag.String("esn", "", "Electronic Serial Number (14 characters)")
	esnDate := flag.String("esn-date", "", "Manufacture date as YYYYMMDD")
	loop := flag.Bool("loop", false, "Enable loop for connecting to bms.")
	overview := flag.Bool("overview", false, "Only show an overview of the essentials and exit.")
	flag.Parse()

	fmt.Println("Starting VanMoof / DynaPack BMS Toolkit")
	fmt.Println("Go version:", runtime.Version(), "Version:", internal.GoVersion, "BuildTime:", internal.BuildTime, "CommitHash:", internal.CommitHash, "GOOS:", internal.GOOS, "GOARCH:", internal.GOARCH)
	fmt.Println("debug Mode:", internal.Debug, "serial Port:", *serialPort, "action", *action, "loop:", *loop)

	if *loop {
		// Should be enough?
		internal.ConnectionRetries = 999999999
	}

	// Serial string commands (no Modbus needed)
	if *action == "clearPF" {
		serial.ClearPF(*serialPort)
		os.Exit(0)
	} else if *action == "gpioOn" {
		serial.SetGPIOOn(*serialPort)
		os.Exit(0)
	} else if *action == "gpioOff" {
		serial.SetGPIOOff(*serialPort)
		os.Exit(0)
	} else if *action == "detectOn" {
		serial.SetDetectPinOn(*serialPort)
		os.Exit(0)
	} else if *action == "detectOff" {
		serial.SetDetectPinOff(*serialPort)
		os.Exit(0)
	} else if *action == "keyInOn" {
		serial.SetKeyInOn(*serialPort)
		os.Exit(0)
	} else if *action == "keyInOff" {
		serial.SetKeyInOff(*serialPort)
		os.Exit(0)
	} else if *action == "resetBMS" {
		serial.ResetBMS(*serialPort)
		os.Exit(0)
	} else if *action == "resetESN" {
		serial.ResetESN(*serialPort)
		os.Exit(0)
	} else if *action == "showPorts" {
		serial.ShowSerialPorts()
	}

	var err error

	// Creates the Modbus connection with all relevant parameters and the port to use
	client, err = modbus.CreateModbusClient(*serialPort)
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
	if *action == "debug" {
		modbus.TurnDebugOn(client)
		os.Exit(0)
	} else if *action == "debugoff" {
		modbus.TurnDebugOff(client)
		os.Exit(0)
	} else if *action == "discharge" {
		modbus.TurnDischargingOn(client)
		os.Exit(0)
	} else if *action == "dischargeoff" {
		modbus.TurnDischargingOff(client)
		os.Exit(0)
	} else if *action == "writeESN" {
		modbus.WriteESNAndDate(client, *esn, *esnDate)
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

	if *action == "live" {
		modbus.LiveData(client, internal.Debug)
	}

	if *overview {
		modbus.ShowOverview()
		os.Exit(0)
	}

	modbus.GetAndShowPassiveBMSData()

	modbus.GetAndShowFlashBMSData()

	modbus.GetAndShowProtectionBMSValues()

	modbus.GetAndShowPassiveVoltages()
}
