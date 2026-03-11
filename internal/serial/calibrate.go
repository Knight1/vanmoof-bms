package serial

import (
	"bms/v2/internal"
	"fmt"
	"log"
	"time"
)

// CalibrateDischargeCurrent sends "DSG CAL=<mA>" over serial to calibrate the discharge current.
// The value is the current in mA.
// The BMS needs 7 seconds to process the calibration.
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
	if internal.Debug {
		fmt.Println("[DEBUG] CalibrateDischargeCurrent: waiting 7s for BMS to process calibration")
	}
	time.Sleep(7 * time.Second)
}

// CalibrateChargeCurrent sends "CHG CAL=<mA>" over serial to calibrate the charge current.
// The value is the current in mA.
// The BMS needs 7 seconds to process the calibration.
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
	if internal.Debug {
		fmt.Println("[DEBUG] CalibrateChargeCurrent: waiting 7s for BMS to process calibration")
	}
	time.Sleep(7 * time.Second)
}
