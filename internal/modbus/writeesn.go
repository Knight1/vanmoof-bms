package modbus

import (
	"fmt"
	"log"
	"strconv"

	"github.com/simonvetter/modbus"
)

// WriteESNAndDate writes a 14-character ESN and a manufacture date to registers 0x0C-0x14.
// ESN: 14 bytes (7 registers 0x0C-0x12) as ASCII.
// Date: 4 bytes (2 registers 0x13-0x14) as [0x00, year, month, day].
// date must be in YYYYMMDD format.
func WriteESNAndDate(client *modbus.ModbusClient, esn string, date string) {
	if esn == "" || date == "" {
		log.Fatal("writeESN requires --esn (14 chars) and --esn-date (YYYYMMDD)")
	}

	if len(esn) != 14 {
		log.Fatalf("ESN must be exactly 14 characters, got %d", len(esn))
	}

	for i, c := range esn {
		if c > 127 {
			log.Fatalf("ESN contains non-ASCII character at position %d", i)
		}
	}

	if len(date) != 8 {
		log.Fatal("Date must be in YYYYMMDD format (e.g., 20220315)")
	}

	year, err1 := strconv.Atoi(date[0:4])
	month, err2 := strconv.Atoi(date[4:6])
	day, err3 := strconv.Atoi(date[6:8])
	if err1 != nil || err2 != nil || err3 != nil {
		log.Fatal("Date must be in YYYYMMDD format (e.g., 20220315)")
	}

	// Build 18 bytes: 14 ESN bytes + 4 date bytes [0x00, year, month, day]
	data := make([]byte, 18)
	copy(data[0:14], []byte(esn))
	data[14] = 0x00
	data[15] = byte(year)
	data[16] = byte(month)
	data[17] = byte(day)

	// Pack bytes into 9 uint16 registers (big-endian)
	regs := make([]uint16, 9)
	for i := 0; i < 9; i++ {
		regs[i] = uint16(data[i*2])<<8 | uint16(data[i*2+1])
	}

	if err := client.WriteRegisters(0x0C, regs); err != nil {
		fmt.Println("Error writing ESN and Date. Error:", err)
	} else {
		fmt.Printf("ESN set to %s, Date set to %d/%d/%d\n", esn, year, month, day)
	}
}
