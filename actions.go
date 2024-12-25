package main

import (
	"bufio"
	"fmt"
	"log"
	"strings"

	"go.bug.st/serial"
)

// actions

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
	//defer port.Close()

	fmt.Println("Serial port opened")

	// Write the string "PF=0" to the serial port
	_, err = port.Write([]byte("PF=0"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Sent PF=0 to serial port")

	// Buffer to store the response
	buf := make([]byte, 1000) // Adjust buffer size as needed

	scanner := bufio.NewScanner(port)
	for scanner.Scan() {
		fmt.Println("Got response")
		fmt.Println(scanner.Text()) // Println will add back the final '\n'
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// Read the response from the serial port
	n, err := port.Read(buf)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Read response from serial port")

	// Convert the response to a string
	response := string(buf[:n])
	fmt.Println("Reading response from serial port. buf:", buf)

	fmt.Println("Response converted to string")
	fmt.Print("Response: ", response)

	// Check if the response contains "OK"
	if strings.Contains(response, "OK") {
		fmt.Println("Response contains OK")
	} else {
		fmt.Println("Response does not contain OK")
	}

	fmt.Printf("Received response: %s\n", response)

	// TODO: readRegisters(client, 2, 1) and check if status is 0. If not, retry
}
