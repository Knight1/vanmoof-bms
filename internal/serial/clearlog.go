package serial

// ClearLog sends "Log Clear" over serial to clear the BMS log.
// The BMS responds with "OK" on success.
func ClearLog(serialPort string) {
	sendGPIOCommand(serialPort, "Log Clear", "OK")
}
