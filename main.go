package main

import (
	"bms/v2/internal"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/simonvetter/modbus"
)

func main() {
	var client *modbus.ModbusClient

	flag.BoolVar(&internal.Debug, "debug", false, "Enable Debug Output")
	serialPort := flag.String("serial-port", "/dev/serial0", "Serial device URL (e.g., /dev/serial0)")
	action := flag.String("action", "show", "Action to perform (clearPF, live, show or showPorts)")
	//firmwareFile := flag.String("firmwareFile", "", "Firmware File to flash to BMS Chip.")
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

	if *action == "clearPF" {
		internal.ClearPF(*serialPort)
		os.Exit(0)
	} else if *action == "showPorts" {
		internal.ShowSerialPorts()
	}

	var err error

	// Creates the Modbus connection with all relevant parameters and the port to use
	client, err = internal.CreateModbusClient(*serialPort)
	if err != nil {
		log.Fatalf("Failed to create Modbus client. Maybe the Probe is disconnected? Check the Address of the Device! Error: %v", err)
	}
	defer client.Close()

	// DEBUG
	if internal.Debug {
		fmt.Println("Modbus client created")
	}

	// Loop for connecting to the bms. Loops until it reaches the end of connectionRetries
	if _, err := internal.ConnectToBMS(client, internal.Debug); err != nil {
		log.Fatalf("Failed to connect to BMS: %v", err)
	}

	// Actions that need ModBus to be initialized
	if *action == "debug" {
		internal.TurnDebugOn(client)
		os.Exit(0)
	} else if *action == "debugoff" {
		internal.TurnDebugOff(client)
		os.Exit(0)
	} else if *action == "discharge" {
		internal.TurnDischargingOn(client)
		os.Exit(0)
	} else if *action == "dischargeoff" {
		internal.TurnDischargingOff(client)
		os.Exit(0)
	}

	if internal.Registers, err = internal.ReadRegisters(client, 0, 95); err != nil {
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
		internal.LiveData(client, internal.Debug)
	}

	if *overview {
		internal.ShowOverview()
		os.Exit(0)
	}

	internal.GetAndShowPassiveBMSData()

	internal.GetAndShowFlashBMSData()

	internal.GetAndShowProtectionBMSValues()

	internal.GetAndShowPassiveVoltages()
}
