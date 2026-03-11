package serial

import (
	"bms/v2/internal"
	"fmt"
	"log"
	"strings"
	"time"

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

// sendGPIOCommand sends a command over serial and reads the response.
// If expectedResponse is provided, the response is checked against it.
func sendGPIOCommand(serialPort string, command string, expectedResponse ...string) {
	if internal.Debug {
		fmt.Printf("[DEBUG] sendGPIOCommand: port=%s command=%q expectedResponse=%v\n", serialPort, command, expectedResponse)
	}

	mode := &serial.Mode{
		BaudRate: 9600,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	if internal.Debug {
		fmt.Printf("[DEBUG] sendGPIOCommand: opening serial port with baud=9600 data=8 parity=none stop=1\n")
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

	if internal.Debug {
		fmt.Printf("[DEBUG] sendGPIOCommand: writing %d bytes: %q\n", len(command), command)
	}

	_, err = port.Write([]byte(command))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Sent %s to serial port\n", command)

	// Read response from BMS
	if err := port.SetReadTimeout(2 * time.Second); err != nil {
		log.Fatal(err)
	}

	if internal.Debug {
		fmt.Println("[DEBUG] sendGPIOCommand: reading response with 2s timeout")
	}

	buf := make([]byte, 256)
	var response []byte
	for {
		n, err := port.Read(buf)
		if err != nil || n == 0 {
			break
		}
		response = append(response, buf[:n]...)
		if internal.Debug {
			fmt.Printf("[DEBUG] sendGPIOCommand: read %d bytes (total: %d)\n", n, len(response))
		}
	}

	if len(response) > 0 {
		respStr := strings.TrimSpace(string(response))
		fmt.Printf("Response: %s\n", respStr)

		if internal.Debug {
			fmt.Printf("[DEBUG] sendGPIOCommand: raw response bytes: %v\n", response)
		}

		if len(expectedResponse) > 0 {
			found := false
			for _, expected := range expectedResponse {
				if strings.Contains(respStr, expected) {
					found = true
					if internal.Debug {
						fmt.Printf("[DEBUG] sendGPIOCommand: matched expected response %q\n", expected)
					}
					break
				}
			}
			if !found {
				fmt.Printf("WARNING: Expected response containing %q, got %q\n", expectedResponse, respStr)
			}
		}
	} else {
		fmt.Println("No response from BMS")
	}
}
