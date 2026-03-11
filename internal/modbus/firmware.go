package modbus

import (
	"bms/v2/internal"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/simonvetter/modbus"
)

const (
	// BMS firmware base address in flash memory
	firmwareBaseAddress = 0x08005000

	// Modbus registers used during firmware update
	regResetMCU    = 0x80 // Write 0 to reset the MCU after update
	regFWCRCCheck  = 0x81 // Read 2 bytes: byte[1]=0 OK, 1=CRC error, 2=no file
	regFWWriteData = 0x82 // Write firmware data chunks (4-byte address + up to 32 bytes data)
	regEraseShadow = 0x95 // Write 0 to erase shadow flash before programming

	// Firmware is written in 32-byte chunks
	fwChunkSize = 32
)

// UpdateFirmware flashes a .bin firmware file to the BMS via Modbus.
//
// Protocol:
//  1. Erase shadow flash (register 0x95=0)
//  2. Write firmware in 32-byte chunks to register 0x82
//     Each chunk: [4-byte offset (big-endian)] + [up to 32 bytes of data]
//  3. Verify CRC via register 0x81
//  4. Reset MCU via register 0x80=0
//  5. Wait for BMS to reboot and confirm via register 0x00
func UpdateFirmware(client *modbus.ModbusClient, filename string) {
	if filename == "" {
		log.Fatal("updateFirmware requires --firmware-file <path to .bin file>")
	}

	binData, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read firmware file: %v", err)
	}

	fileLength := len(binData)
	fmt.Printf("Firmware file: %s (%d bytes)\n", filename, fileLength)
	fmt.Printf("Flash base address: 0x%08X\n", firmwareBaseAddress)
	fmt.Printf("Flash end address:  0x%08X\n", firmwareBaseAddress+fileLength)

	if fileLength == 0 {
		log.Fatal("Firmware file is empty")
	}

	if internal.Debug {
		fmt.Printf("[DEBUG] UpdateFirmware: file=%s size=%d bytes\n", filename, fileLength)
	}

	// Step 1: Erase shadow flash
	fmt.Println("Step 1/4: Erasing shadow flash (register 0x95=0)...")
	if internal.Debug {
		fmt.Println("[DEBUG] UpdateFirmware: writing register 0x95=0 to erase shadow flash")
	}
	if err := client.WriteRegister(regEraseShadow, 0); err != nil {
		log.Fatalf("Failed to erase shadow flash: %v", err)
	}
	fmt.Println("Shadow flash erased.")

	// Step 2: Write firmware data in 32-byte chunks
	fmt.Println("Step 2/4: Writing firmware data...")
	written := 0
	for written < fileLength {
		chunkSize := fwChunkSize
		remaining := fileLength - written
		if remaining < chunkSize {
			chunkSize = remaining
		}

		// Build the data payload: 4-byte address (big-endian) + chunk data
		// The address is the offset within the firmware file (not the absolute flash address)
		// The BMS adds the base address internally
		payload := make([]byte, 4+chunkSize)
		payload[0] = byte((written >> 24) & 0xFF)
		payload[1] = byte((written >> 16) & 0xFF)
		payload[2] = byte((written >> 8) & 0xFF)
		payload[3] = byte(written & 0xFF)
		copy(payload[4:], binData[written:written+chunkSize])

		if internal.Debug {
			fmt.Printf("[DEBUG] UpdateFirmware: writing chunk at offset 0x%08X (flash 0x%08X), %d bytes\n",
				written, firmwareBaseAddress+written, chunkSize)
		}

		if err := client.WriteRawBytes(regFWWriteData, payload); err != nil {
			log.Fatalf("Failed to write firmware chunk at offset 0x%08X: %v", written, err)
		}

		written += chunkSize

		// Progress
		progress := float64(written) / float64(fileLength) * 100
		fmt.Printf("\rProgress: %d/%d bytes (%.1f%%)", written, fileLength, progress)
	}
	fmt.Println()
	fmt.Println("Firmware data written.")

	// Step 3: Verify CRC
	fmt.Println("Step 3/4: Verifying CRC...")
	if internal.Debug {
		fmt.Println("[DEBUG] UpdateFirmware: reading register 0x81 (2 bytes) for CRC check")
	}

	crcRegs, err := client.ReadRegisters(regFWCRCCheck, 1, modbus.HOLDING_REGISTER)
	if err != nil {
		log.Fatalf("Failed to read CRC check status: %v", err)
	}

	// The CRC status is in the low byte of the register
	crcStatus := byte(crcRegs[0] & 0xFF)
	if internal.Debug {
		fmt.Printf("[DEBUG] UpdateFirmware: CRC register raw=0x%04X status=%d\n", crcRegs[0], crcStatus)
	}

	switch crcStatus {
	case 0:
		fmt.Println("CRC check passed!")
	case 1:
		log.Fatal("CRC check failed! Firmware data may be corrupted.")
	case 2:
		log.Fatal("CRC check returned 'No File'. The BMS did not receive firmware data.")
	default:
		log.Fatalf("CRC check returned unknown status: %d", crcStatus)
	}

	// Step 4: Reset MCU
	fmt.Println("Step 4/4: Resetting MCU (register 0x80=0)...")
	if internal.Debug {
		fmt.Println("[DEBUG] UpdateFirmware: writing register 0x80=0 to reset MCU")
	}

	time.Sleep(50 * time.Millisecond)

	if err := client.WriteRegister(regResetMCU, 0); err != nil {
		fmt.Printf("Warning: MCU reset write returned error (expected during reboot): %v\n", err)
	}

	// Wait for BMS to reboot and confirm
	fmt.Println("Waiting for BMS to reboot...")
	waitForBMSReboot(client)
}

// waitForBMSReboot polls the BMS after a firmware update reset.
// It tries to read register 0x00 and expects value [1, 0] (0x0100) to confirm the BMS is running.
// Also listens on the serial port for "DP Mode" or "I am VM-BATT AP" messages.
func waitForBMSReboot(client *modbus.ModbusClient) {
	timeout := time.After(180 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	if internal.Debug {
		fmt.Println("[DEBUG] waitForBMSReboot: polling register 0x00 for up to 180s")
	}

	// Give the BMS time to start rebooting
	time.Sleep(2 * time.Second)

	for {
		select {
		case <-timeout:
			log.Fatal("Timeout waiting for BMS to reboot after firmware update (180s)")
		case <-ticker.C:
			// Try to reconnect and read register 0x00
			if err := client.Open(); err != nil {
				if internal.Debug {
					fmt.Printf("[DEBUG] waitForBMSReboot: open failed: %v\n", err)
				}
				continue
			}

			if err := client.SetUnitId(internal.DynaPackVanMoofSlaveID); err != nil {
				if internal.Debug {
					fmt.Printf("[DEBUG] waitForBMSReboot: set unit ID failed: %v\n", err)
				}
				continue
			}

			regs, err := client.ReadRegisters(0x00, 1, modbus.HOLDING_REGISTER)
			if err != nil {
				if internal.Debug {
					fmt.Printf("[DEBUG] waitForBMSReboot: read register 0x00 failed: %v\n", err)
				}
				continue
			}

			if internal.Debug {
				fmt.Printf("[DEBUG] waitForBMSReboot: register 0x00=0x%04X\n", regs[0])
			}

			// The BMS reports 0x0100 (high byte=1, low byte=0) when running
			if regs[0] == 0x0100 {
				fmt.Println("BMS rebooted successfully! Firmware update complete.")
				return
			}

			fmt.Printf("BMS register 0x00=0x%04X, waiting...\n", regs[0])
		}
	}
}
