package main

import (
	"fmt"
	"github.com/simonvetter/modbus"
	"log"

	"go.bug.st/serial"
)

// actions

// We send PF=0 over serial to clear all Power Failures.
// This might need some tries also we might need to clear the Log first.
func clearPF(serialPort string) {
	// Open the serial port
	mode := &serial.Mode{
		BaudRate: 9600, // Set baud rate (adjust as needed)
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	fmt.Println("Opening serial port", serialPort)

	port, err := serial.Open(serialPort, mode) // Use the appropriate serial port
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Serial port opened")

	for attempt := 0; attempt < int(connectionRetries); attempt++ {

		// Write the string "PF=0" to the serial port
		_, err = port.Write([]byte("PF=0"))
		if err != nil {
			log.Fatal(err)
		}

		port.Close()

		fmt.Println("Sent PF=0 to serial port")

		// Creates the Modbus connection with all relevant parameters and the port to use
		client, err := createModbusClient(serialPort)
		if err != nil {
			log.Fatalf("Failed to create Modbus client. Maybe the Probe is disconnected? Check the Address of the Device! Error: %v", err)
		}
		defer client.Close()

		if debug {
			fmt.Println("Modbus client created")
		}

		// Loop for connecting to the bms. Loops until it reaches the end of connectionRetries
		if err, _ := connectToBMS(client, debug); err != nil {
			log.Fatalf("Failed to connect to BMS: %v", err)
		}

		//if fault == "0x0000" {

		//}
	}
}

func turnDebugOn(client *modbus.ModbusClient) {
	if err = client.WriteRegister(0x9, 1); err != nil {
		fmt.Println("Error setting Debug to On. Error:", err)
	} else {
		fmt.Println("Debug set to On!")
	}
}

func turnDebugOff(client *modbus.ModbusClient) {
	if err = client.WriteRegister(0x9, 0); err != nil {
		fmt.Println("Error setting Debug to Off. Error:", err)
	} else {
		fmt.Println("Debug set to Off!")
	}
}

func turnDischargingOn(client *modbus.ModbusClient) {
	if err = client.WriteRegister(0x8, 0); err != nil {
		fmt.Println("Error setting Discharging to Off. Error:", err)
	} else {
		fmt.Println("Discharging set to Off!")
	}
}

func turnDischargingOff(client *modbus.ModbusClient) {
	if err = client.WriteRegister(0x8, 0); err != nil {
		fmt.Println("Error setting Discharging to Off. Error:", err)
	} else {
		fmt.Println("Discharging set to Off!")
	}
}
