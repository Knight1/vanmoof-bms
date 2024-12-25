package main

import (
	"fmt"
)

func checkFaults(value uint16) {
	if value == 0 {
		fmt.Println("BMS STATUS OK!")
	} else {
		fmt.Println("BMS SHUTDOWN!")
		fmt.Printf("Register 0x%X ('Fault Status'): 0x%04X\n", 0x2, value)

		// Decode flags (bitwise operations)
		flags := []string{"DOTP", "DUTP", "COTP", "CUTP", "DOCP1", "DOCP2", "COCP1", "COCP2", "OVP1", "OVP2", "UVP1", "UVP2", "PDOCP", "PDSCP", "MOTP", "SCP"}
		for i, flag := range flags {
			if value&(1<<i) != 0 {
				fmt.Printf(" - %s is set\n", flag)
			}
		}
	}
}

func checkWarnings(value uint16) {
	fmt.Printf("Warning Status: %04X\n", value)

	// Decode flags (bitwise operations)
	flags := []string{"DOTPW", "DUTPW", "COTPW", "CUTPW", "DOCPW", "RSVD", "COCPW", "RSVD", "OVP1W", "RSVD", "UVP1W", "SOC", "PDOCPW", "RSVD", "MOTPW", "RSVD"}
	for i, flag := range flags {
		if value&(1<<i) != 0 {
			fmt.Printf(" - %s is set\n", flag)
		}
	}
}

func checkMOSControl(value uint16) {
	fmt.Printf("CHG MOS Control: %04X\n", value)

	// Decode flags (bitwise operations)
	flags := []string{"RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "CHG"}
	for i, flag := range flags {
		if value&(1<<i) != 0 {
			fmt.Printf(" - %s is set\n", flag)
		}
	}

}

func checkChargingStatus(value uint16) {
	fmt.Printf("Charging Status: %04X\n", value)

	// Decode flags (bitwise operations)
	flags := []string{"RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "CHG_IN", "Fault", "CHG"}
	for i, flag := range flags {
		if value&(1<<i) != 0 {
			fmt.Printf(" - %s is set\n", flag)
		}
	}
}

func checkDischargingStatus(value uint16) {
	fmt.Println("Discharging on/off:", value)
	fmt.Printf("Discharging Status: %04X\n", value)

	// Decode flags (bitwise operations)
	flags := []string{"RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "RSVD", "DSG"}
	for i, flag := range flags {
		if value&(1<<i) != 0 {
			fmt.Printf(" - %s is set\n", flag)
		}
	}
}
