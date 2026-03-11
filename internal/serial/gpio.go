package serial

import (
	"fmt"
	"log"

	"go.bug.st/serial"
)

// GPIO controls the charge port.
// Send "GPIO.PF2=1." for ON and "GPIO.PF2=0." for OFF over serial.

func SetGPIOOn(serialPort string) {
	sendGPIOCommand(serialPort, "GPIO.PF2=1.")
}

func SetGPIOOff(serialPort string) {
	sendGPIOCommand(serialPort, "GPIO.PF2=0.")
}

// DetectPin controls the detect pin (IO2)
// Send "GPIO.IO2=1." for ON and "GPIO.IO2=0." for OFF over serial.

func SetDetectPinOn(serialPort string) {
	sendGPIOCommand(serialPort, "GPIO.IO2=1.")
}

func SetDetectPinOff(serialPort string) {
	sendGPIOCommand(serialPort, "GPIO.IO2=0.")
}

// KeyIn controls the key input pin (IO1).
// Send "GPIO.IO1=1." for ON and "GPIO.IO1=0." for OFF over serial.

func SetKeyInOn(serialPort string) {
	sendGPIOCommand(serialPort, "GPIO.IO1=1.")
}

func SetKeyInOff(serialPort string) {
	sendGPIOCommand(serialPort, "GPIO.IO1=0.")
}

func sendGPIOCommand(serialPort string, command string) {
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

	_, err = port.Write([]byte(command))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Sent %s to serial port\n", command)
}
