package main

import "time"

// build flags
var (
	BuildTime  string
	CommitHash string
	GitTag     string
	GOOS       string
	GOARCH     string
	GoVersion  string
)

const (
	connectionRetries      uint8 = 5
	connectionRetryDelay         = 500 * time.Millisecond
	DynaPackVanMoofSlaveID uint8 = 170

	// Define Thresholds (own!)
	cellVoltageLow  uint16 = 2500
	cellVoltageHigh uint16 = 4300
	packVoltageLow  uint16 = 25000
	packVoltageHigh uint16 = 43000
)

var (
	regs       []uint16
	err        error
	milliVolts float64

	// Define holding registers
	cellVoltageHighest = uint16(0)
	cellVoltageLowest  = uint16(0)

	// Initialize Global BMS Error Statusses
	bmsUndervoltageCellProtection1 uint16 = 0
	bmsUndervoltageCellProtection2 uint16 = 0
	bmsUndervoltageCellShutdown    uint16 = 0
	bmsOvervoltageCellProtection1  uint16 = 0
	bmsStatus                      bool   = false

	// Define Thresholds (own!)
	cellVoltageImbalance uint16 = 5
	cellVoltagePrevious  uint16 = 0
)
