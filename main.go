package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"

	"github.com/simonvetter/modbus"
)

func main() {
	var client *modbus.ModbusClient

	debug := flag.Bool("debug", false, "Enable Debug Output")
	serialPort := flag.String("serial-port", "/dev/serial0", "Serial device URL (e.g., /dev/serial0)")
	live := flag.Bool("live", false, "Enable Live Output")
	flag.Parse()

	fmt.Println("Starting VanMoof / DynaPack BMS Toolkit")
	fmt.Println("Go version:", runtime.Version(), "Version:", GoVersion, "BuildTime:", BuildTime, "CommitHash:", CommitHash, "Git:", GitTag, "GOOS:", GOOS, "GOARCH:", GOARCH)

	// TODO: make it possible to pass the device as a parameter
	client, err = createModbusClient(*serialPort)
	if err != nil {
		log.Fatalf("Failed to create Modbus client. Maybe the Probe is disconnected? Check the Address of the Device! Error: %v", err)
	}
	defer client.Close()

	// DEBUG
	if *debug {
		fmt.Println("Modbus client created")
	}

	if err = connectToBMS(client, *debug); err != nil {
		log.Fatalf("Failed to connect to BMS: %v", err)
	}

	if regs, err = readRegisters(client, 0, 95); err != nil {
		log.Fatalf("Failed to read registers: %v", err)
	}

	// Debug Output
	if *debug {
		fmt.Println("-- BEGIN DEBUG --")
		fmt.Println("BMS ModBus Addresses 0 to 94")
		for register, reg := range regs {
			fmt.Println("Register:", register, "Value:", reg)
		}

		fmt.Println("-- END DEBUG --")
	}

	if *live {
		liveData(client, *debug)
	}

	// defere here to make sure the client is closed after the live output is not used
	defer client.Close()

	getAndShowPassiveBMSData()

	getAndShowFlashBMSData()

	getAndShowProtectionBMSValues()

	getAndShowPassiveVoltages()
}

func calculateCelsius(value uint16) float64 {
	return float64(value-2731) / 10
}

func calculateAmperes(value uint16) float64 {
	return float64(value) / 10
}
