package main

import (
	"fmt"
	"go.bug.st/serial/enumerator"
	"log"
	"os"
)

func showSerialPorts() {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Fatal(err)
	}

	if len(ports) == 0 {
		fmt.Println("No serial ports found!")
		os.Exit(1)
	}

	for _, port := range ports {
		fmt.Printf("Found port: %s\n", port.Name)
		if port.IsUSB {
			fmt.Printf("   USB ID     %s:%s\n", port.VID, port.PID)
			fmt.Printf("   USB serial %s\n", port.SerialNumber)
		}
	}
	os.Exit(0)
}
