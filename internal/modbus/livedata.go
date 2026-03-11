package modbus

import (
	"bms/v2/internal"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/simonvetter/modbus"
)

const (
	// ANSI escape codes
	ansiClearScreen = "\033[2J"
	ansiHome        = "\033[H"
	ansiReset       = "\033[0m"
	ansiBold        = "\033[1m"
	ansiRed         = "\033[31m"
	ansiGreen       = "\033[32m"
	ansiYellow      = "\033[33m"
	ansiCyan        = "\033[36m"
)

func LiveData(client *modbus.ModbusClient, debug bool) {
	if internal.Debug {
		fmt.Println("[DEBUG] LiveData: starting continuous read loop")
	}

	// Handle Ctrl+C gracefully — show cursor and reset terminal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Print("\033[?25h") // show cursor
		fmt.Print(ansiReset)
		fmt.Println("\nLive mode stopped.")
		os.Exit(0)
	}()

	// Hide cursor for cleaner display
	fmt.Print("\033[?25l")
	// Clear screen once at start
	fmt.Print(ansiClearScreen)

	readErrors := 0

	for {
		if internal.Debug {
			fmt.Println("[DEBUG] LiveData: reading registers 0-44")
		}

		if _, readErr := ReadRegisters(client, 0, 45); readErr != nil {
			readErrors++
			log.Printf("Failed to read registers (attempt %d): %v", readErrors, readErr)
			if internal.Debug {
				fmt.Println("[DEBUG] LiveData: reconnecting to BMS")
			}
			if _, err := ConnectToBMS(client, debug); err != nil {
				log.Printf("Failed to connect to BMS: %v", err)
			}
			time.Sleep(1 * time.Second)
			continue
		}
		readErrors = 0

		renderLiveDisplay()
		time.Sleep(1 * time.Second)
	}
}

func renderLiveDisplay() {
	regs := internal.Registers
	if len(regs) < 45 {
		return
	}

	var b strings.Builder

	// Move cursor to top-left, don't clear (reduces flicker)
	b.WriteString(ansiHome)

	// Header
	b.WriteString(ansiBold + ansiCyan)
	b.WriteString("  VanMoof BMS Live Monitor")
	b.WriteString(ansiReset)
	b.WriteString(fmt.Sprintf("  [%s]\n", time.Now().Format("15:04:05")))
	b.WriteString(strings.Repeat("-", 58) + "\n")

	// Fault status
	fault := regs[2]
	if fault == 0 {
		b.WriteString(ansiGreen + ansiBold + "  STATUS: OK" + ansiReset + "                                          \n")
	} else {
		b.WriteString(ansiRed + ansiBold + fmt.Sprintf("  STATUS: SHUTDOWN (0x%04X)", fault) + ansiReset + "                        \n")
		flags := []string{"DOTP", "DUTP", "COTP", "CUTP", "DOCP1", "DOCP2", "COCP1", "COCP2", "OVP1", "OVP2", "UVP1", "UVP2", "PDOCP", "PDSCP", "MOTP", "SCP"}
		var active []string
		for i, f := range flags {
			if fault&(1<<i) != 0 {
				active = append(active, f)
			}
		}
		b.WriteString(ansiRed + "  Faults: " + strings.Join(active, ", ") + ansiReset + "                        \n")
	}

	b.WriteString(strings.Repeat("-", 58) + "\n")

	// Main values
	b.WriteString(fmt.Sprintf("  Battery Voltage: %s%-8d mV%s", ansiBold, regs[4], ansiReset))
	b.WriteString(fmt.Sprintf("  Current: %s%-10.0f mA%s\n", ansiBold, internal.CalculateAmperes(regs[6]), ansiReset))

	b.WriteString(fmt.Sprintf("  RSOC: %s%-4d %%%s", ansiBold, regs[5], ansiReset))
	b.WriteString(fmt.Sprintf("              Absolute SOC: %s%-4d %%%s\n", ansiBold, regs[24], ansiReset))

	b.WriteString(fmt.Sprintf("  Remaining: %-6d mAh", regs[23]))
	b.WriteString(fmt.Sprintf("        Full Charge: %-6d mAh\n", regs[22]))

	b.WriteString(fmt.Sprintf("  Cycle Count: %-6d\n", regs[25]))

	// Temperatures
	b.WriteString(strings.Repeat("-", 58) + "\n")
	b.WriteString(ansiBold + "  Temperatures" + ansiReset + "\n")

	batTemp := internal.CalculateCelsius(regs[3])
	temp1 := internal.CalculateCelsius(regs[37])
	temp2 := internal.CalculateCelsius(regs[38])
	mosTemp := internal.CalculateCelsius(regs[39])

	b.WriteString(fmt.Sprintf("  Battery: %s%-7.1f °C%s", colorTemp(batTemp), batTemp, ansiReset))
	b.WriteString(fmt.Sprintf("      MOS: %s%-7.1f °C%s\n", colorTemp(mosTemp), mosTemp, ansiReset))
	b.WriteString(fmt.Sprintf("  Sensor 1: %s%-7.1f °C%s", colorTemp(temp1), temp1, ansiReset))
	b.WriteString(fmt.Sprintf("    Sensor 2: %s%-7.1f °C%s\n", colorTemp(temp2), temp2, ansiReset))

	// Cell voltages
	b.WriteString(strings.Repeat("-", 58) + "\n")
	b.WriteString(ansiBold + "  Cell Voltages (mV)" + ansiReset + "\n")

	var minV, maxV uint16
	minV = 0xFFFF
	for i := 0; i < 10; i++ {
		v := regs[27+i]
		if v > 0 && v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}

	for i := 0; i < 10; i++ {
		v := regs[27+i]
		cellStr := fmt.Sprintf("  Cell %2d: %s%-5d%s", i+1, colorVoltage(v), v, ansiReset)
		if i%2 == 0 {
			b.WriteString(cellStr)
		} else {
			b.WriteString("        " + cellStr + "\n")
		}
	}

	imbalance := int(maxV) - int(minV)
	b.WriteString(strings.Repeat("-", 58) + "\n")
	b.WriteString(fmt.Sprintf("  Min: %-5d mV  Max: %-5d mV  ", minV, maxV))
	if imbalance > 20 {
		b.WriteString(ansiRed + ansiBold + fmt.Sprintf("Imbalance: %d mV!", imbalance) + ansiReset)
	} else {
		b.WriteString(ansiGreen + fmt.Sprintf("Imbalance: %d mV", imbalance) + ansiReset)
	}
	b.WriteString("          \n")

	// Warnings
	warnings := regs[40]
	if warnings != 0 {
		b.WriteString(ansiYellow + fmt.Sprintf("  Warnings: 0x%04X", warnings) + ansiReset + "                              \n")
	} else {
		b.WriteString("  Warnings: None                                          \n")
	}

	// Charging / Discharging status
	chgStatus := regs[7]
	dsgStatus := regs[8]
	b.WriteString(fmt.Sprintf("  Charging: 0x%04X  Discharging: 0x%04X", chgStatus, dsgStatus))
	b.WriteString("                  \n")

	b.WriteString(strings.Repeat("-", 58) + "\n")
	b.WriteString("  Press Ctrl+C to exit live mode\n")

	// Pad with blank lines to overwrite any leftover content from previous frame
	for i := 0; i < 3; i++ {
		b.WriteString("                                                          \n")
	}

	fmt.Print(b.String())
}

func colorTemp(temp float32) string {
	switch {
	case temp > 55:
		return ansiRed + ansiBold
	case temp > 45:
		return ansiYellow
	default:
		return ansiGreen
	}
}

func colorVoltage(v uint16) string {
	switch {
	case v < internal.CellVoltageLow:
		return ansiRed + ansiBold
	case v > internal.CellVoltageHigh:
		return ansiRed + ansiBold
	default:
		return ansiGreen
	}
}
