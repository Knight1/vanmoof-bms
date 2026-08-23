package modbus

import (
	"bms/v2/internal"
	"fmt"
)

// Firmware-side protection context (batteryware docs/protection-config.md,
// thresholds live in cfg_blk @ 0x200028D0, byte-identical in 1.14.1 / 1.17.1):
//   - Temp shutdown: fg_threshold_check trips when the MOSFET/board thermistor
//     byte (0x20002588) >= 135 (default), recovers at 125. An active
//     over-temperature fault surfaces on reg 2 as 0x0002 (FAULT_TS).
//   - Cell voltage window is effectively [40,110] in the firmware's byte scale
//     with an inner warn at 85 and a critical guard at 135.
//   - Over-current (discharge/charge) trips at |I| >= 500 internal units; bits
//     6/7 are the only faults that fire the pyro fuse (PB7). Over/under-voltage
//     is recoverable (charge FET opened).
//   - Autonomous hardware protections (ML5236 AFE, docs sec 5): short-circuit at
//     ~150 mV across the 1 mOhm shunt (~150 A) with a 2 s comm watchdog; cell
//     over-voltage is caught by the S-8215AAD stage (4.35 V/cell, 2 s).
// Scaling of the reg 71-86 values below (byte->mV, unit->A) is NOT decoded in
// the decomp, so they are printed raw; treat them as best-guess trip points.
func GetAndShowProtectionBMSValues() {

	fmt.Println("-- BEGIN TRIGGER AND PROTECTION VALUES --")
	fmt.Println("Trigger Values are best guess. DynaPack does not specify them.")
	fmt.Println("Reg 71-86 raw (scaling undecoded). Temp shutdown ~135 byte-scale; AFE short-circuit ~150 mV/2 s (hardware).")

	// Checking Proteection Statusses
	for register, value := range internal.Registers {
		switch register {
		case 71: // 0x47
			fmt.Println("(DOTP) Discharge Over Temperature Protection:", value)
		case 72: // 0x48
			fmt.Println("(DUTP) Discharge Under Temperature Protection:", value)
		case 73: // 0x49
			fmt.Println("(COTP) Charging Over Temperature Protection:", value)
		case 74: // 0x4A
			fmt.Println("(CUTP) Current Under Temperature Protection:", value)
		case 75: // 0x4B
			fmt.Println("(DOCP1) Discharge Over Current Protection 1:", value)
		case 76: // 0x4C
			fmt.Println("(DOCP2) Discharge Over Current Protection 2:", value)
		case 77: // 0x4D
			fmt.Println("(COCP1) Charging Over Current Protection 1:", value)
		case 78: // 0x4E
			fmt.Println("(COCP2) Charging Over Current Protection 2:", value)
		case 79: // 0x4F
			fmt.Println("(OVP1) Over Voltage Protection 1:", value)
		case 80: // 0x50
			fmt.Println("(OVP2) Over Voltage Protection 2:", value)
		case 81: // 0x51
			fmt.Println("(UVP1) Under Voltage Protection 1:", value)
		case 82: // 0x52
			fmt.Println("(UVP2) Under Voltage Protection 2:", value)
		case 83: // 0x53
			fmt.Println("(PDOCP) Peak Discharge Over Current Protection:", value)
		case 84: // 0x54
			fmt.Println("(PDSCP) Peak Discharge Short Circuit Protection:", value)
		case 85: // 0x55
			fmt.Println("(MOTP) MOSFET (Metal Oxide Semiconductor Field-Effect Transistors) Over Temperature Protection:", value)
		case 86: // 0x56
			fmt.Println("(SCP) Short Circuit Protection:", value)
		default:
			continue
		}
	}

	fmt.Println("-- END TRIGGER AND PROTECTION VALUES --")
}
