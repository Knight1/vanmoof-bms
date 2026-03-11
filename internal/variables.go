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
	ConnectionRetryDelay         = 500 * time.Millisecond
	DynaPackVanMoofSlaveID uint8 = 170

	// Define Thresholds (own!)
	CellVoltageLow  uint16 = 2500
	CellVoltageHigh uint16 = 4300
	PackVoltageLow  uint16 = 25000
	PackVoltageHigh uint16 = 43000
)

var (
	ConnectionRetries int  = 5
	Loop              bool = false
	Registers         []uint16
	Debug             bool = false

	// Define holding registers
	CellVoltageHighest = uint16(0)
	CellVoltageLowest  = uint16(0)

	// Define Thresholds (own!)
	CellVoltageImbalance uint16 = 5
	CellVoltagePrevious  uint16 = 0
)

// ModBus Registers
var (
	RegisterFault = 2
)

func CalculateCelsius(value uint16) float32 {
	return float32(int16(value)-2731) / 10
}

func CalculateAmperes(value uint16) float64 {
	return float64(int16(value)) * 10
}
