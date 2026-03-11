package serial

// ResetBMS sends "Reset BMS V0106" over serial to factory reset the BMS.
// This removes the Serial Number, calibration and Charge Cycles.
func ResetBMS(serialPort string) {
	sendGPIOCommand(serialPort, "Reset BMS V0106")
}

// ResetESN sends "Reset ESN" over serial to clear the Electronic Serial Number.
// The BMS responds with "Done" on success or "Reset ESN fail" on failure.
func ResetESN(serialPort string) {
	sendGPIOCommand(serialPort, "Reset ESN")
}
