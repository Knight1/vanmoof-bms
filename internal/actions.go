package internal

import (
	"fmt"
	"log"

	"github.com/simonvetter/modbus"

	"go.bug.st/serial"
)

// actions

// We send PF=0 over serial to clear all Power Failures.
// This might need some tries also we might need to clear the Log first.
func ClearPF(serialPort string) {
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
	defer func() {
		if port != nil {
			_ = port.Close()
		}
	}()

	fmt.Println("Serial port opened")

	for attempt := 0; attempt < int(ConnectionRetries); attempt++ {

		// Write the string "PF=0" to the serial port
		_, err = port.Write([]byte("PF=0"))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Sent PF=0 to serial port")

		// Close the serial port before switching to Modbus
		if err = port.Close(); err != nil {
			log.Fatalf("Failed to close serial port: %v", err)
		}
		port = nil

		// Creates the Modbus connection with all relevant parameters and the port to use
		client, err := CreateModbusClient(serialPort)
		if err != nil {
			log.Fatalf("Failed to create Modbus client. Maybe the Probe is disconnected? Check the Address of the Device! Error: %v", err)
		}

		if Debug {
			fmt.Println("Modbus client created")
		}

		// Loop for connecting to the bms. Loops until it reaches the end of connectionRetries
		if _, err := ConnectToBMS(client, Debug); err != nil {
			_ = client.Close()
			log.Fatalf("Failed to connect to BMS: %v", err)
		}

		if err := client.Close(); err != nil {
			log.Fatalf("Failed to close Modbus client: %v", err)
		}

		// Re-open serial port for next attempt
		if attempt < int(ConnectionRetries)-1 {
			port, err = serial.Open(serialPort, mode)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func TurnDebugOn(client *modbus.ModbusClient) {
	if err := client.WriteRegister(0x9, 1); err != nil {
		fmt.Println("Error setting Debug to On. Error:", err)
	} else {
		fmt.Println("Debug set to On!")
	}
}

func TurnDebugOff(client *modbus.ModbusClient) {
	if err := client.WriteRegister(0x9, 0); err != nil {
		fmt.Println("Error setting Debug to Off. Error:", err)
	} else {
		fmt.Println("Debug set to Off!")
	}
}

func TurnDischargingOn(client *modbus.ModbusClient) {
	if err := client.WriteRegister(0x8, 1); err != nil {
		fmt.Println("Error setting Discharging to On. Error:", err)
	} else {
		fmt.Println("Discharging set to On!")
	}
}

func TurnDischargingOff(client *modbus.ModbusClient) {
	if err := client.WriteRegister(0x8, 0); err != nil {
		fmt.Println("Error setting Discharging to Off. Error:", err)
	} else {
		fmt.Println("Discharging set to Off!")
	}
}
