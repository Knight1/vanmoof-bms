package serial

import (
	"bms/v2/internal"
	"fmt"
)

// ClearLog sends "Log Clear" over serial to clear the BMS log.
// The BMS responds with "OK" on success.
func ClearLog(serialPort string) {
	if internal.Debug {
		fmt.Printf("[DEBUG] ClearLog: sending \"Log Clear\" to port=%s\n", serialPort)
	}
	sendGPIOCommand(serialPort, "Log Clear", "OK")
}
