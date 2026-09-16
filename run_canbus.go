package main

// CAN action dispatch. Two independent CAN paths live here:
//
//   - runRawCAN: A5/S5 DynaPack BMS over raw SocketCAN (package internal/can),
//     selected with `-action can*`.
//   - runCANBus: A5/S5/S6 over the brutella/can library (package internal/canbus),
//     selected with `--protocol canbus`.

import (
	"bms/v2/internal/can"
	"bms/v2/internal/canbus"
	"fmt"
	"log"
	"os"
)

// runRawCAN handles the A5/S5 DynaPack BMS commands over raw SocketCAN
// (internal/can). It exits the process on a matching action and returns
// otherwise, so the caller can continue dispatching.
func runRawCAN(action, iface string, snapshot bool, esn string) {
	switch action {
	case "can":
		can.ReadBMS(iface, snapshot)
		os.Exit(0)
	// DynaPack proprietary commands
	case "canDischarge":
		can.TurnDischargeOn(iface)
		os.Exit(0)
	case "canDischargeOff":
		can.TurnDischargeOff(iface)
		os.Exit(0)
	case "canChargeOn":
		can.TurnChargeOn(iface)
		os.Exit(0)
	case "canChargeOff":
		can.TurnChargeOff(iface)
		os.Exit(0)
	case "canUnlockPF":
		can.UnlockPF(iface)
		os.Exit(0)
	case "canClearLog":
		can.ClearLog(iface)
		os.Exit(0)
	case "canResetBMS":
		can.ResetBMS(iface)
		os.Exit(0)
	case "canCoulombCheck":
		can.CoulombCounterCheck(iface)
		os.Exit(0)
	case "canSetRTC":
		can.SetRTC(iface)
		os.Exit(0)
	case "canReadRTC":
		can.ReadRTC(iface)
		os.Exit(0)
	case "canSetSN":
		can.SetSN(iface, esn)
		os.Exit(0)
	case "canReadChargeOffset":
		can.ReadChargeCurrentOffset(iface)
		os.Exit(0)
	case "canReadDischargeOffset":
		can.ReadDischargeCurrentOffset(iface)
		os.Exit(0)
	case "canReadADCOffset":
		can.ReadChargeADCOffset(iface)
		os.Exit(0)
	case "canReadVoltageOffset":
		can.ReadChargeVoltageOffset(iface)
		os.Exit(0)
	// SetStatus commands
	case "canNormalMode":
		can.SetNormalMode(iface)
		os.Exit(0)
	case "canChargeMode":
		can.SetChargeMode(iface)
		os.Exit(0)
	case "canSleep":
		can.SetSleepMode(iface)
		os.Exit(0)
	case "canStandby":
		can.SetStandbyMode(iface)
		os.Exit(0)
	case "canStatusChargeOn":
		can.SetStatusChargeOn(iface)
		os.Exit(0)
	case "canStatusChargeOff":
		can.SetStatusChargeOff(iface)
		os.Exit(0)
	// Feature control commands
	case "canEnableBalance":
		can.EnableCellBalance(iface)
		os.Exit(0)
	case "canDisableBalance":
		can.DisableCellBalance(iface)
		os.Exit(0)
	case "canEnablePDSCP":
		can.EnablePDSCP(iface)
		os.Exit(0)
	case "canDisablePDSCP":
		can.DisablePDSCP(iface)
		os.Exit(0)
	case "canEnableCellOffline":
		can.EnableCellOffline(iface)
		os.Exit(0)
	case "canDisableCellOffline":
		can.DisableCellOffline(iface)
		os.Exit(0)
	case "canResetSystem":
		can.ResetSystemFunction(iface)
		os.Exit(0)
	}
}

// runCANBus handles the A5/S5/S6 actions over the brutella/can library
// (internal/canbus), selected with `--protocol canbus`.
func runCANBus(canInterface, action, firmwareFile string, overview bool) {
	client, err := canbus.NewClient(canInterface)
	if err != nil {
		log.Fatalf("Failed to create CAN bus client on %s: %v", canInterface, err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Failed to close CAN bus client: %v", err)
		}
	}()

	// Main-bus battery actions (telemetry read + power on/off request). These
	// live on the battery pack's own CAN node and do not depend on the BMS
	// command heartbeat, so handle them directly without that connection gate.
	if action == "batteryState" || action == "batteryOn" || action == "batteryOff" {
		if err := client.Connect(); err != nil {
			log.Fatalf("Failed to start CAN bus: %v", err)
		}
		// The CAN protocol only applies to A5/S5/S6 batteries. Verify one is
		// present before acting (the S3/S4 battery is Modbus/UART only).
		if desc, ok := canbus.IdentifyBattery(client, canbus.BatteryIdentifyTimeout); ok {
			fmt.Printf("A5/S5/S6 battery detected (%s).\n", desc)
		} else {
			fmt.Println("No A5/S5/S6 battery detected on the CAN bus — refusing to continue.")
			fmt.Println("The S3/S4 battery is Modbus/UART only; use --protocol modbus for it.")
			os.Exit(1)
		}

		switch action {
		case "batteryState":
			canbus.ShowBatteryState(client)
		case "batteryOn":
			canbus.PowerBatteryOn(client)
		case "batteryOff":
			canbus.PowerBatteryOff(client)
		}
		return
	}

	if err := canbus.ConnectToBMS(client); err != nil {
		log.Fatalf("Failed to connect to BMS via CAN bus: %v", err)
	}

	// Start heartbeat monitoring
	go canbus.MonitorHeartbeat(client)

	switch action {
	case "chargeOn":
		client.SetChargeOn()
	case "chargeOff":
		client.SetChargeOff()
	case "shipMode":
		client.SetShipMode()
	case "standby":
		client.SetStandby()
	case "wakeup":
		client.Wakeup()
	case "updateFirmware":
		client.UpdateFirmware(firmwareFile)
	case "updateFirmwareAP":
		client.UpdateFirmwareAP(firmwareFile)
	case "enterBootloader":
		client.EnterBootloader()
	case "exitBootloader":
		client.ExitBootloader()
	case "resetMCU":
		if err := client.SendFrame(canbus.IDAPReboot, make([]byte, 8)); err != nil {
			fmt.Println("Error sending reboot:", err)
		} else {
			fmt.Println("MCU reboot sent!")
		}
	case "show":
		canbus.ShowOverview(client)
	default:
		if overview {
			canbus.ShowOverview(client)
		} else {
			fmt.Printf("Action '%s' is not supported over CAN bus.\n", action)
			fmt.Println("Supported CAN bus actions: batteryState, batteryOn, batteryOff, chargeOn, chargeOff, shipMode, standby, wakeup, updateFirmware, updateFirmwareAP, enterBootloader, exitBootloader, resetMCU, show")
		}
	}
}
