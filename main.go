package main

import (
	"bms/v2/internal"
	"bms/v2/internal/convert"
	"bms/v2/internal/serial"
	"flag"
	"fmt"
	"os"
	"runtime"
)

func main() {
	flag.BoolVar(&internal.Debug, "debug", false, "Enable Debug Output")
	serialPort := flag.String("serial-port", "/dev/serial0", "Serial device URL (e.g., /dev/serial0)")
	canInterface := flag.String("can-interface", "can0", "CAN bus interface for --protocol canbus (e.g., can0, vcan0)")
	protocol := flag.String("protocol", "modbus", "Communication protocol: modbus or canbus")
	canIface := flag.String("can-iface", "can0", "SocketCAN interface for -action can (e.g., can0)")
	canSnapshot := flag.Bool("can-snapshot", false, "Read BMS values once and exit (use with -action can)")
	action := flag.String("action", "show", "Action to perform.\n"+
		"  Modbus/serial (S3/S4): calibrateCHG, calibrateDSG, chargeOn, chargeOff, clearLog, clearPF,\n"+
		"    convertLog, detectOn, detectOff, discharge, dischargeoff, exportLog, gpioOn, gpioOff,\n"+
		"    keyInOn, keyInOff, live, resetBMS, resetESN, resetESNModbus, resetMCU, ship, shipMode,\n"+
		"    show, showPorts, updateFirmware, writeESN\n"+
		"  CAN raw SocketCAN (A5/S5, -action can*): can, canChargeMode, canChargeOn, canChargeOff, canClearLog,\n"+
		"    canCoulombCheck, canDisableCellOffline, canDisableBalance, canDisablePDSCP,\n"+
		"    canDischarge, canDischargeOff, canEnableCellOffline, canEnableBalance, canEnablePDSCP,\n"+
		"    canNormalMode, canReadADCOffset, canReadChargeOffset, canReadDischargeOffset,\n"+
		"    canReadRTC, canReadVoltageOffset, canResetBMS, canResetSystem, canSetRTC,\n"+
		"    canSetSN, canSleep, canStandby, canStatusChargeOff, canStatusChargeOn, canUnlockPF\n"+
		"  CAN brutella (A5/S5/S6, --protocol canbus): batteryState, batteryOn, batteryOff, chargeOn,\n"+
		"    chargeOff, shipMode, standby, wakeup, enterBootloader, exitBootloader, updateFirmware,\n"+
		"    updateFirmwareAP, resetMCU, show")
	firmwareFile := flag.String("firmware-file", "", "Firmware .bin file to flash to BMS")
	logFile := flag.String("log-file", "", "Output CSV file for exportLog (default: bms_log_<timestamp>.csv)")
	logInput := flag.String("log-input", "", "Input text file for convertLog action")
	calibrateCurrent := flag.Int("calibrate-current", 0, "Current in mA for calibrateDSG / calibrateCHG actions")
	esn := flag.String("esn", "", "Electronic Serial Number (14 characters)")
	esnDate := flag.String("esn-date", "", "Manufacture date as YYYYMMDD")
	loop := flag.Bool("loop", false, "Enable loop for connecting to bms.")
	overview := flag.Bool("overview", false, "Only show an overview of the essentials and exit.")
	showVersion := flag.Bool("version", false, "Print version information and exit")
	flag.BoolVar(showVersion, "v", false, "Print version information and exit (shorthand)")
	flag.Parse()

	fmt.Println("Starting VanMoof / DynaPack BMS Toolkit")
	fmt.Println("Go version:", runtime.Version(), "Version:", internal.GoVersion, "BuildTime:", internal.BuildTime, "CommitHash:", internal.CommitHash, "GOOS:", internal.GOOS, "GOARCH:", internal.GOARCH)

	// Print version/build info and exit before touching any hardware.
	if *showVersion {
		os.Exit(0)
	}

	if internal.Debug {
		fmt.Println("-- CLI Arguments --")
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Printf("  --%s=%s (default: %s)\n", f.Name, f.Value.String(), f.DefValue)
		})
		fmt.Println("-------------------")
	}

	if *loop {
		internal.Loop = true
	}

	// Raw SocketCAN commands (A5/S5 DynaPack BMS, internal/can). Exits the process
	// on a matching action; returns so dispatch can continue otherwise.
	runRawCAN(*action, *canIface, *canSnapshot, *esn)

	// File utility commands (no serial, Modbus, or CAN needed)
	if *action == "convertLog" {
		convert.CustomerLog(*logInput)
		os.Exit(0)
	} else if *action == "showPorts" {
		serial.ShowSerialPorts()
		os.Exit(0)
	}

	// Route to CAN bus or Modbus based on --protocol flag
	if *protocol == "canbus" {
		runCANBus(*canInterface, *action, *firmwareFile, *overview)
	} else {
		runModbus(*serialPort, *action, *firmwareFile, *logFile, *esn, *esnDate, *calibrateCurrent, *overview)
	}
}
