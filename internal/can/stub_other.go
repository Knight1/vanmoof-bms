//go:build !linux

// Package can talks to the DynaPack BMS over SocketCAN and the Linux GPIO
// character device, both of which are Linux-kernel-only. On every other
// platform (e.g. macOS during development) the real implementation cannot
// compile, so this file provides no-op stubs with the same exported API.
// Each stub prints a notice and returns, allowing the Modbus/serial paths of
// the toolkit to be built and run locally while CAN actions are unavailable.
package can

import "fmt"

// unsupported prints a consistent notice for a CAN action invoked on a
// non-Linux build.
func unsupported(action string) {
	fmt.Printf("[CAN] %q is only supported on Linux (SocketCAN/GPIO); ignoring on this platform.\n", action)
}

// HoldWakeLine is a no-op on non-Linux platforms. It returns a cleanup func so
// callers can use it identically to the Linux implementation.
func HoldWakeLine() func() {
	unsupported("HoldWakeLine")
	return func() {}
}

func ReadBMS(iface string, snapshot bool)     { unsupported("ReadBMS") }
func TurnDischargeOn(iface string)            { unsupported("TurnDischargeOn") }
func TurnDischargeOff(iface string)           { unsupported("TurnDischargeOff") }
func TurnChargeOn(iface string)               { unsupported("TurnChargeOn") }
func TurnChargeOff(iface string)              { unsupported("TurnChargeOff") }
func UnlockPF(iface string)                   { unsupported("UnlockPF") }
func ClearLog(iface string)                   { unsupported("ClearLog") }
func ResetBMS(iface string)                   { unsupported("ResetBMS") }
func CoulombCounterCheck(iface string)        { unsupported("CoulombCounterCheck") }
func SetRTC(iface string)                     { unsupported("SetRTC") }
func ReadRTC(iface string)                    { unsupported("ReadRTC") }
func SetSN(iface string, sn string)           { unsupported("SetSN") }
func ReadChargeCurrentOffset(iface string)    { unsupported("ReadChargeCurrentOffset") }
func ReadDischargeCurrentOffset(iface string) { unsupported("ReadDischargeCurrentOffset") }
func ReadChargeADCOffset(iface string)        { unsupported("ReadChargeADCOffset") }
func ReadChargeVoltageOffset(iface string)    { unsupported("ReadChargeVoltageOffset") }
func SetNormalMode(iface string)              { unsupported("SetNormalMode") }
func SetChargeMode(iface string)              { unsupported("SetChargeMode") }
func SetSleepMode(iface string)               { unsupported("SetSleepMode") }
func SetStandbyMode(iface string)             { unsupported("SetStandbyMode") }
func SetStatusChargeOn(iface string)          { unsupported("SetStatusChargeOn") }
func SetStatusChargeOff(iface string)         { unsupported("SetStatusChargeOff") }
func EnableCellBalance(iface string)          { unsupported("EnableCellBalance") }
func DisableCellBalance(iface string)         { unsupported("DisableCellBalance") }
func EnablePDSCP(iface string)                { unsupported("EnablePDSCP") }
func DisablePDSCP(iface string)               { unsupported("DisablePDSCP") }
func EnableCellOffline(iface string)          { unsupported("EnableCellOffline") }
func DisableCellOffline(iface string)         { unsupported("DisableCellOffline") }
func ResetSystemFunction(iface string)        { unsupported("ResetSystemFunction") }
