package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/simonvetter/modbus"
)

func main() {
	var client *modbus.ModbusClient

	flag.BoolVar(&debug, "debug", false, "Enable Debug Output")
	serialPort := flag.String("serial-port", "/dev/serial0", "Serial device URL (e.g., /dev/serial0)")
	action := flag.String("action", "show", "Action to perform (clearPF, live, show or showPorts)")
	//firmwareFile := flag.String("firmwareFile", "", "Firmware File to flash to BMS Chip.")
	loop := flag.Bool("loop", false, "Enable loop for connecting to bms.")
	overview := flag.Bool("overview", false, "Only show an overview of the essentials and exit.")
	flag.Parse()

	fmt.Println("Starting VanMoof / DynaPack BMS Toolkit")
	fmt.Println("Go version:", runtime.Version(), "Version:", GoVersion, "BuildTime:", BuildTime, "CommitHash:", CommitHash, "GOOS:", GOOS, "GOARCH:", GOARCH)
	fmt.Println("debug Mode:", debug, "serial Port:", *serialPort, "action", *action, "loop:", *loop)

	if *loop {
		// Should be enough?
		connectionRetries = 999999999
	}

	if *action == "clearPF" {
		clearPF(*serialPort)
	} else if *action == "showPorts" {
		showSerialPorts()
	}

	// Creates the Modbus connection with all relevant parameters and the port to use
	client, err = createModbusClient(*serialPort)
	if err != nil {
		log.Fatalf("Failed to create Modbus client. Maybe the Probe is disconnected? Check the Address of the Device! Error: %v", err)
	}
	defer client.Close()

	// DEBUG
	if debug {
		fmt.Println("Modbus client created")
	}

	// Loop for connecting to the bms. Loops until it reaches the end of connectionRetries
	if err = connectToBMS(client, debug); err != nil {
		log.Fatalf("Failed to connect to BMS: %v", err)
	}

	if regs, err = readRegisters(client, 0, 95); err != nil {
		log.Fatalf("Failed to read registers: %v", err)
	}

	// Debug Output
	if debug {
		fmt.Println("-- BEGIN DEBUG --")
		fmt.Println("BMS ModBus Addresses 0 to 94")
		for register, reg := range regs {
			fmt.Println("Register:", register, "Value:", reg)
		}

		fmt.Println("-- END DEBUG --")
	}

	if *action == "live" {
		liveData(client, debug)
	}

	if *overview {
		showOverview()
		os.Exit(0)
	}

	// defere here to make sure the client is closed after the live output is not used
	defer client.Close()

	getAndShowPassiveBMSData()

	getAndShowFlashBMSData()

	getAndShowProtectionBMSValues()

	getAndShowPassiveVoltages()
}

func calculateCelsius(value uint16) float32 {
	return float32(value-2731) / 10
}

func calculateAmperes(value uint16) float64 {
	return float64(value) / 10
}
