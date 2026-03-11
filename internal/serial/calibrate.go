package serial

import (
	"bms/v2/internal"
	"fmt"
	"log"
)

// CalibrateDischargeCurrent sends "DSG CAL=<mA>" over serial to calibrate the discharge current.
// The value is the current in mA.
func CalibrateDischargeCurrent(serialPort string, mA int) {
	if internal.Debug {
		fmt.Printf("[DEBUG] CalibrateDischargeCurrent: port=%s mA=%d\n", serialPort, mA)
	}
	if mA <= 0 {
		log.Fatal("Calibration current must be a positive value in mA")
	}
	cmd := fmt.Sprintf("DSG CAL=%d", mA)
	if internal.Debug {
		fmt.Printf("[DEBUG] CalibrateDischargeCurrent: sending %q\n", cmd)
	}
	sendGPIOCommand(serialPort, cmd)
}

// CalibrateChargeCurrent sends "CHG CAL=<mA>" over serial to calibrate the charge current.
// The value is the current in mA.
func CalibrateChargeCurrent(serialPort string, mA int) {
	if internal.Debug {
		fmt.Printf("[DEBUG] CalibrateChargeCurrent: port=%s mA=%d\n", serialPort, mA)
	}
	if mA <= 0 {
		log.Fatal("Calibration current must be a positive value in mA")
	}
	cmd := fmt.Sprintf("CHG CAL=%d", mA)
	if internal.Debug {
		fmt.Printf("[DEBUG] CalibrateChargeCurrent: sending %q\n", cmd)
	}
	sendGPIOCommand(serialPort, cmd)
}
