package modbus

import (
	"bms/v2/internal"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/simonvetter/modbus"
)

// logRegister defines a DataFlash register used in log entries.
type logRegister struct {
	Address uint16
	Name    string
	Unit    string
	// Formula: "temp" for (x-2731)/10, "current" for x*10, "" for raw value
	Formula string
}

// DataFlash registers 0x30-0x44 that the BMS stores in each log entry.
var logRegisters = []logRegister{
	{0x30, "Fault Status-Record", "hex", ""},
	{0x31, "BAT Temp1-Record", "C", "temp"},
	{0x32, "BAT Temp2-Record", "C", "temp"},
	{0x33, "MOS Temp-Record", "C", "temp"},
	{0x34, "Battery Voltage-Record", "mV", ""},
	{0x35, "Current", "mA", "current"},
	{0x36, "Full Charge Capacity", "mAh", ""},
	{0x37, "Remaining Capacity", "mAh", ""},
	{0x38, "RSOC", "%", ""},
	{0x39, "Absolute SOC", "%", ""},
	{0x3A, "Cycle Count", "count", ""},
	{0x3B, "Cell 1", "mV", ""},
	{0x3C, "Cell 2", "mV", ""},
	{0x3D, "Cell 3", "mV", ""},
	{0x3E, "Cell 4", "mV", ""},
	{0x3F, "Cell 5", "mV", ""},
	{0x40, "Cell 6", "mV", ""},
	{0x41, "Cell 7", "mV", ""},
	{0x42, "Cell 8", "mV", ""},
	{0x43, "Cell 9", "mV", ""},
	{0x44, "Cell 10", "mV", ""},
}

// ExportReadLog reads 100 log entries from the BMS and exports them to a CSV file.
// The BMS expects register 0x0F45 to be written with the log ID (0-99) to select the log entry.
// Then each DataFlash register (0x30-0x44) is read from the 0x0F00 offset address space.
func ExportReadLog(client *modbus.ModbusClient, filename string) {
	if internal.Debug {
		fmt.Println("[DEBUG] ExportReadLog: starting log export")
		fmt.Printf("[DEBUG] ExportReadLog: reading %d registers per log entry, 100 log entries\n", len(logRegisters))
	}

	if filename == "" {
		filename = fmt.Sprintf("bms_log_%s.csv", time.Now().Format("20060102_150405"))
	}

	if internal.Debug {
		fmt.Printf("[DEBUG] ExportReadLog: output file=%s\n", filename)
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create CSV file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Failed to close CSV file: %v", err)
		}
	}()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	header := []string{"DateTime", "LogID"}
	for _, reg := range logRegisters {
		header = append(header, reg.Name)
	}
	if err := writer.Write(header); err != nil {
		log.Fatalf("Failed to write CSV header: %v", err)
	}

	for logID := 0; logID < 100; logID++ {
		fmt.Printf("Reading Log ID %d\n", logID)

		if internal.Debug {
			fmt.Printf("[DEBUG] ExportReadLog: writing register 0x0F45=%d to select log entry\n", logID)
		}

		// Write register 0x0F45 with the log ID to select which log entry to read
		if err := client.WriteRegister(0x0F45, uint16(logID)); err != nil {
			fmt.Printf("Failed to set log ID %d: %v\n", logID, err)
			continue
		}

		row := []string{
			time.Now().Format("2006-01-02 15:04:05.000"),
			strconv.Itoa(logID),
		}

		for _, reg := range logRegisters {
			// Read from the log address space: 0x0F00 + register address
			logAddr := 0x0F00 + reg.Address
			if internal.Debug {
				fmt.Printf("[DEBUG] ExportReadLog: reading register 0x%04X (%s)\n", logAddr, reg.Name)
			}
			regs, err := client.ReadRegisters(logAddr, 1, modbus.HOLDING_REGISTER)
			if err != nil {
				if internal.Debug {
					fmt.Printf("[DEBUG] ExportReadLog: read error at 0x%04X: %v\n", logAddr, err)
				}
				row = append(row, "Read Err")
				continue
			}

			value := regs[0]
			if internal.Debug {
				fmt.Printf("[DEBUG] ExportReadLog: register 0x%04X raw=0x%04X (%d)\n", logAddr, value, value)
			}
			row = append(row, formatLogValue(value, reg))
		}

		if err := writer.Write(row); err != nil {
			fmt.Printf("Failed to write log row %d: %v\n", logID, err)
		}

		// Print to console as well
		fmt.Println(row)
	}

	fmt.Printf("Log exported to %s\n", filename)
}

func formatLogValue(value uint16, reg logRegister) string {
	switch reg.Formula {
	case "temp":
		return fmt.Sprintf("%.1f", internal.CalculateCelsius(value))
	case "current":
		return fmt.Sprintf("%.0f", internal.CalculateAmperes(value))
	default:
		if reg.Unit == "hex" {
			return fmt.Sprintf("0x%04X", value)
		}
		return strconv.FormatUint(uint64(value), 10)
	}
}
