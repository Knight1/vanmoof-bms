package modbus

import (
	"bms/v2/internal"
	"fmt"
	"math"
)

func ShowOverview() {
	regs := internal.Registers

	fmt.Println("-- BEGIN BMS OVERVIEW --")

	// Fault Status
	if len(regs) > int(internal.RegisterFault) {
		CheckFaults(regs[internal.RegisterFault])
	}

	// Warnings
	if len(regs) > 40 {
		checkWarnings(regs[40])
	}

	// Hardware / Software / Bootloader
	if len(regs) > 11 {
		fmt.Printf("Hardware Version: %04X\n", regs[10])
		fmt.Printf("Software Version: %04X\n", regs[11])
	}
	if len(regs) > 44 {
		fmt.Printf("Bootloader Version: %04X\n", regs[44])
	}

	// ESN
	if len(regs) >= 19 {
		bytes := make([]byte, 0, 14)
		for _, reg := range regs[12:19] {
			bytes = append(bytes, byte(reg>>8), byte(reg&0xFF))
		}
		fmt.Printf("ESN: %s\n", string(bytes))
	}

	// Manufacture Date
	if len(regs) >= 21 {
		dateBytes := make([]byte, 0, 4)
		for _, reg := range regs[19:21] {
			dateBytes = append(dateBytes, byte(reg>>8), byte(reg&0xFF))
		}
		fmt.Printf("Manufacture Date: %d/%d/%d\n", dateBytes[1], dateBytes[2], dateBytes[3])
	}

	// Temperatures
	if len(regs) > 39 {
		fmt.Println("Battery Temperature:", internal.CalculateCelsius(regs[3]), "°C")
		fmt.Println("Temperature Sensor 1:", internal.CalculateCelsius(regs[37]), "°C")
		fmt.Println("Temperature Sensor 2:", internal.CalculateCelsius(regs[38]), "°C")
		fmt.Println("Discharge MOSFET Temperature:", internal.CalculateCelsius(regs[39]), "°C")
	}

	// Voltage / Current / SOC
	if len(regs) > 24 {
		fmt.Println("Battery Voltage:", regs[4], "mV")
		fmt.Println("Current:", internal.CalculateAmperes(regs[6]), "mA")
		fmt.Println("Real State of Charge:", regs[5], "%")
		fmt.Println("Absolute SOC:", regs[24], "%")
	}

	// Capacity
	if len(regs) > 23 {
		fmt.Println("Full Charge Capacity:", regs[22], "mAh")
		fmt.Println("Remaining Capacity:", regs[23], "mAh")
	}

	// Cycle Count
	if len(regs) > 25 {
		fmt.Println("Cycle Count:", regs[25])
	}

	// Charging / Discharging / MOS Control
	if len(regs) > 26 {
		checkChargingStatus(regs[7])
		checkDischargingStatus(regs[8])
		checkMOSControl(regs[26])
	}

	// Cell Voltage Min/Max
	if len(regs) > 42 {
		maxV := regs[41]
		minV := regs[42]
		fmt.Println("Maximum Cell Voltage:", maxV, "mV")
		fmt.Println("Minimum Cell Voltage:", minV, "mV")
		imbalance := math.Abs(float64(int(maxV) - int(minV)))
		if imbalance > 20 {
			fmt.Printf("WARNING: Cell Voltage Imbalance: %.0f mV!\n", imbalance)
		}
	}

	// Cell Balance
	if len(regs) > 43 {
		fmt.Println("Cell Balance:", regs[43])
	}

	fmt.Println("-- END BMS OVERVIEW --")

	GetAndShowPassiveVoltages()
}
