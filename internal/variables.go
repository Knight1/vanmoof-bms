package internal

import "time"

// build flags
var (
	BuildTime  string
	CommitHash string
	GOOS       string
	GOARCH     string
	GoVersion  string
)

const (
	connectionRetryDelay         = 500 * time.Millisecond
	DynaPackVanMoofSlaveID uint8 = 170

	// Define Thresholds (own!)
	cellVoltageLow  uint16 = 2500
	cellVoltageHigh uint16 = 4300
	packVoltageLow  uint16 = 25000
	packVoltageHigh uint16 = 43000
)

var (
	ConnectionRetries uint64 = 5
	Registers         []uint16
	milliVolts        float64
	Debug             bool = false

	// Define holding registers
	cellVoltageHighest = uint16(0)
	cellVoltageLowest  = uint16(0)

	// Initialize Global BMS Error Statusses
	bmsUndervoltageCellProtection1 uint16 = 0
	bmsUndervoltageCellProtection2 uint16 = 0
	bmsUndervoltageCellShutdown    uint16 = 0
	bmsOvervoltageCellProtection1  uint16 = 0

	// Define Thresholds (own!)
	cellVoltageImbalance uint16 = 5
	cellVoltagePrevious  uint16 = 0
)

// ModBus Registers
var (
	RegisterFault = 2
)

func calculateCelsius(value uint16) float32 {
	return float32(value-2731) / 10
}

func calculateAmperes(value uint16) float64 {
	return float64(value) / 10
}
