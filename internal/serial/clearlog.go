package serial

import (
	"bms/v2/internal"
	"fmt"
	"time"
)

// ClearLog sends "Log Clear" over serial to clear the BMS log.
// The BMS responds with "OK" on success.
// The BMS needs 3 seconds to process the log clear.
func ClearLog(serialPort string) {
	if internal.Debug {
		fmt.Printf("[DEBUG] ClearLog: sending \"Log Clear\" to port=%s\n", serialPort)
	}
	sendGPIOCommand(serialPort, "Log Clear", "OK")
	if internal.Debug {
		fmt.Println("[DEBUG] ClearLog: waiting 3s for BMS to process")
	}
	time.Sleep(3 * time.Second)
}
