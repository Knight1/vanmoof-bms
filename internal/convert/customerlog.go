package convert

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// CustomerLog converts a BMS customer log text file to CSV.
// The input file contains lines with "#" followed by semicolon-separated values.
// Record types: 2=fault/voltage/current, 22=capacity, 37=temperatures/warnings.
// Output CSV columns: ts(ms), fault(register), voltage(mV), rsoc(%), current(mA),
// current_abs(mA), full_charge(mAh), remaining_charge(mAh), temp1(C), temp2(C),
// dsg_mos_temp(C), warn(register), min_bat_v(mV), max_bat_v(mV)
func CustomerLog(inputFile string) {
	if inputFile == "" {
		log.Fatal("Input file is required for convertLog action")
	}

	data, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("Failed to read input file: %v", err)
	}

	dir := filepath.Dir(inputFile)
	base := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
	outputFile := filepath.Join(dir, base+".csv")

	header := []string{
		"ts(ms)", "fault(register)", "voltage(mV)", "rsoc(%)",
		"current(mA)", "current_abs(mA)", "full_charge(mAh)",
		"remaining_charge(mAh)", "temp1(C)", "temp2(C)",
		"dsg_mos_temp(C)", "warn(register)", "min_bat_v(mV)", "max_bat_v(mV)",
	}
	colCount := len(header)

	// Group records by timestamp (rounded to seconds)
	records := make(map[int][]string)
	var keys []int

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "#", 2)
		if len(parts) < 2 {
			continue
		}

		fields := strings.Split(parts[1], ";")
		if len(fields) < 2 {
			continue
		}

		ts, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		recordType, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}

		// Group by second (round down to nearest 1000ms)
		key := (ts / 1000) * 1000

		if _, exists := records[key]; !exists {
			records[key] = make([]string, colCount)
			keys = append(keys, key)
		}

		row := records[key]

		switch {
		case len(fields) > 5 && recordType == 2:
			row[1] = fields[2] // fault
			voltage, _ := strconv.Atoi(fields[4])
			if voltage >= 0 {
				row[2] = fields[4] // voltage
			} else {
				row[2] = strconv.Itoa(1 + voltage + int(math.MaxUint16))
			}
			row[3] = fields[5] // rsoc
			current, _ := strconv.Atoi(fields[6])
			row[4] = strconv.Itoa(current * 10)                         // current (mA)
			row[5] = strconv.Itoa(int(math.Abs(float64(current))) * 10) // current_abs (mA)

		case len(fields) > 3 && recordType == 22:
			row[6] = fields[2] // full_charge
			row[7] = fields[3] // remaining_charge

		case len(fields) > 5 && recordType == 37:
			temp1, _ := strconv.Atoi(fields[2])
			row[8] = strconv.Itoa((temp1 - 2731) / 10) // temp1
			temp2, _ := strconv.Atoi(fields[3])
			row[9] = strconv.Itoa((temp2 - 2731) / 10) // temp2
			mosTemp, _ := strconv.Atoi(fields[4])
			row[10] = strconv.Itoa((mosTemp - 2731) / 10) // dsg_mos_temp
			row[11] = fields[5]                           // warn
			row[12] = fields[6]                           // min_bat_v
			row[13] = fields[7]                           // max_bat_v
		}
	}

	// Write CSV
	file, err := os.Create(outputFile)
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

	if err := writer.Write(header); err != nil {
		log.Fatalf("Failed to write CSV header: %v", err)
	}

	for _, key := range keys {
		row := records[key]
		row[0] = strconv.Itoa(key) // ts
		if err := writer.Write(row); err != nil {
			fmt.Printf("Failed to write row: %v\n", err)
		}
	}

	fmt.Printf("Customer log converted to %s\n", outputFile)
}
