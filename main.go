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

	fmt.Println("Starting VanMoof / DynaPack BMS Toolkit")
	fmt.Println("Go version:", runtime.Version(), "Version:", GoVersion, "BuildTime:", BuildTime, "CommitHash:", CommitHash, "Git:", GitTag, "GOOS:", GOOS, "GOARCH:", GOARCH)

	// TODO: make it possible to pass the device as a parameter
	client, err = createModbusClient("/dev/serial0")
	if err != nil {
		log.Fatalf("Failed to create Modbus client. Maybe the Probe is disconnected? Check the Address of the Device! Error: %v", err)
	}
	defer client.Close()

	// DEBUG
	fmt.Println("Modbus client created")

	if err, regs = connectToBMS(client); err != nil {
		log.Fatalf("Failed to connect to BMS: %v", err)
	}

	// Debug Output
	if *debug {
		fmt.Println("-- BEGIN DEBUG --")
		fmt.Println("BMS ModBus Addresses 0 to 95")
		for register, reg := range regs {
			fmt.Println("Register:", register, "Value:", reg)
		}

		fmt.Println("-- END DEBUG --")
	}

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
