package serial

import (
	"bms/v2/internal"
	"bms/v2/internal/modbus"
	"fmt"
	"log"

	"go.bug.st/serial"
)

// ClearPF sends PF=0 over serial to clear all Power Failures.
// This might need some tries also we might need to clear the Log first.
func ClearPF(serialPort string) {
	if internal.Debug {
		fmt.Printf("[DEBUG] ClearPF: port=%s loop=%v retries=%d\n", serialPort, internal.Loop, internal.ConnectionRetries)
	}

	mode := &serial.Mode{
		BaudRate: 9600,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	fmt.Println("Opening serial port", serialPort)

	port, err := serial.Open(serialPort, mode)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if port != nil {
			_ = port.Close()
		}
	}()

	fmt.Println("Serial port opened")

	for attempt := 0; internal.Loop || attempt < internal.ConnectionRetries; attempt++ {
		if internal.Debug {
			fmt.Printf("[DEBUG] ClearPF: attempt %d - sending PF=0\n", attempt+1)
		}

		_, err = port.Write([]byte("PF=0"))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Sent PF=0 to serial port")

		// Close the serial port before switching to Modbus
		if internal.Debug {
			fmt.Println("[DEBUG] ClearPF: closing serial port for Modbus connection")
		}
		if err = port.Close(); err != nil {
			log.Fatalf("Failed to close serial port: %v", err)
		}
		port = nil

		client, err := modbus.CreateModbusClient(serialPort)
		if err != nil {
			log.Fatalf("Failed to create Modbus client. Maybe the Probe is disconnected? Check the Address of the Device! Error: %v", err)
		}

		if internal.Debug {
			fmt.Println("[DEBUG] ClearPF: Modbus client created, connecting to BMS")
		}

		if _, err := modbus.ConnectToBMS(client, internal.Debug); err != nil {
			_ = client.Close()
			log.Fatalf("Failed to connect to BMS: %v", err)
		}

		if internal.Debug {
			fmt.Println("[DEBUG] ClearPF: BMS connected, closing Modbus client")
		}

		if err := client.Close(); err != nil {
			log.Fatalf("Failed to close Modbus client: %v", err)
		}

		// Re-open serial port for next attempt
		if internal.Loop || attempt < internal.ConnectionRetries-1 {
			if internal.Debug {
				fmt.Println("[DEBUG] ClearPF: re-opening serial port for next attempt")
			}
			port, err = serial.Open(serialPort, mode)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
