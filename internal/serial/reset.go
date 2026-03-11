package serial

import (
	"bms/v2/internal"
	"fmt"
	"time"
)

// ResetBMS sends "Reset BMS V0106" over serial to factory reset the BMS.
// This removes the Serial Number, calibration and Charge Cycles.
// The BMS needs 1 second to process the reset.
func ResetBMS(serialPort string) {
	if internal.Debug {
		fmt.Printf("[DEBUG] ResetBMS: sending \"Reset BMS V0106\" to port=%s\n", serialPort)
	}
	sendGPIOCommand(serialPort, "Reset BMS V0106")
	if internal.Debug {
		fmt.Println("[DEBUG] ResetBMS: waiting 1s for BMS to process reset")
	}
	time.Sleep(1 * time.Second)
}

// ResetESN sends "Reset ESN" over serial to clear the Electronic Serial Number.
// The BMS responds with "Done" on success or "Reset ESN fail" on failure.
func ResetESN(serialPort string) {
	if internal.Debug {
		fmt.Printf("[DEBUG] ResetESN: sending \"Reset ESN\" to port=%s\n", serialPort)
	}
	sendGPIOCommand(serialPort, "Reset ESN", "Done")
}
