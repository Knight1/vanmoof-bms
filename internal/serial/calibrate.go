package serial

import (
	"fmt"
	"log"
)

// CalibrateDischargeCurrent sends "DSG CAL=<mA>" over serial to calibrate the discharge current.
// The value is the current in mA.
func CalibrateDischargeCurrent(serialPort string, mA int) {
	if mA <= 0 {
		log.Fatal("Calibration current must be a positive value in mA")
	}
	sendGPIOCommand(serialPort, fmt.Sprintf("DSG CAL=%d", mA))
}

// CalibrateChargeCurrent sends "CHG CAL=<mA>" over serial to calibrate the charge current.
// The value is the current in mA.
func CalibrateChargeCurrent(serialPort string, mA int) {
	if mA <= 0 {
		log.Fatal("Calibration current must be a positive value in mA")
	}
	sendGPIOCommand(serialPort, fmt.Sprintf("CHG CAL=%d", mA))
}
