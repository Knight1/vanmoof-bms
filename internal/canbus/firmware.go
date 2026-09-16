package canbus

import (
	"bms/v2/internal"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"log"
	"os"
	"time"
)

// UpdateFirmware flashes a firmware binary to the BMS via the SPIN bootloader over CAN bus.
//
// Protocol:
//  1. Enter bootloader mode
//  2. Wait for bootloader to be ready (heartbeat status = Waiting)
//  3. Send START_UPLOAD with firmware size
//  4. Loop: wait for REQUEST_CHUNK, respond with SEND_CHUNK (6-byte chunks)
//  5. Wait for validation (heartbeat status = Ready)
//  6. Exit bootloader
func (c *Client) UpdateFirmware(filename string) {
	if filename == "" {
		log.Fatal("No firmware file specified. Use --firmware-file to provide the .bin file.")
	}

	firmware, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read firmware file: %v", err)
	}

	fmt.Printf("Firmware file: %s (%d bytes)\n", filename, len(firmware))

	// Calculate CRC32 for verification
	fwCRC := crc32.ChecksumIEEE(firmware)
	if internal.Debug {
		fmt.Printf("[DEBUG] UpdateFirmware: CRC32=0x%08X\n", fwCRC)
	}

	// Step 1: Enter bootloader
	if err := c.EnterBootloader(); err != nil {
		log.Fatalf("Failed to enter bootloader: %v", err)
	}

	// Step 2: Send firmware size
	fmt.Println("Starting firmware upload...")
	if err := c.StartUpload(uint32(len(firmware))); err != nil {
		log.Fatalf("Failed to send START_UPLOAD: %v", err)
	}

	// Step 3: Send firmware chunks
	totalChunks := (len(firmware) + FWChunkSize - 1) / FWChunkSize
	fmt.Printf("Uploading %d chunks (%d bytes each)...\n", totalChunks, FWChunkSize)

	chunksSent := 0
	for chunksSent < totalChunks {
		// Determine timeout: longer for the last chunk
		timeout := TimeoutFWChunk
		if chunksSent >= totalChunks-1 {
			timeout = TimeoutFWLastChunk
		}

		// Wait for BMS to request a chunk
		address, _, err := c.WaitForChunkRequest(timeout)
		if err != nil {
			log.Fatalf("Failed waiting for chunk request at chunk %d/%d: %v", chunksSent+1, totalChunks, err)
		}

		// Calculate firmware offset from address
		offset := int(address) * FWChunkSize
		if offset >= len(firmware) {
			if internal.Debug {
				fmt.Printf("[DEBUG] UpdateFirmware: address 0x%X beyond firmware end, sending zeros\n", address)
			}
			if err := c.SendFirmwareChunk(uint32(chunksSent), make([]byte, FWChunkSize)); err != nil {
				log.Fatalf("Failed to send padding chunk: %v", err)
			}
		} else {
			end := offset + FWChunkSize
			if end > len(firmware) {
				end = len(firmware)
			}
			chunk := firmware[offset:end]

			if err := c.SendFirmwareChunk(uint32(chunksSent), chunk); err != nil {
				log.Fatalf("Failed to send chunk %d: %v", chunksSent, err)
			}
		}

		chunksSent++
		if chunksSent%100 == 0 || chunksSent == totalChunks {
			fmt.Printf("Progress: %d/%d chunks (%.1f%%)\n", chunksSent, totalChunks, float64(chunksSent)/float64(totalChunks)*100)
		}

		time.Sleep(DelayBetweenChunks)
	}

	fmt.Println("Upload complete. Waiting for BMS validation...")

	// Step 4: Wait for validation
	if err := c.waitForValidation(); err != nil {
		log.Fatalf("Firmware validation failed: %v", err)
	}

	// Step 5: Verify CRC
	fmt.Println("Verifying CRC...")
	if err := c.verifyCRC(fwCRC); err != nil {
		log.Fatalf("CRC verification failed: %v", err)
	}

	// Step 6: Exit bootloader
	if err := c.ExitBootloader(); err != nil {
		log.Fatalf("Failed to exit bootloader: %v", err)
	}

	fmt.Println("Firmware update complete!")
}

// waitForValidation waits for the bootloader to finish validating the firmware.
func (c *Client) waitForValidation() error {
	deadline := time.Now().Add(30 * time.Second)

	for time.Now().Before(deadline) {
		status, err := c.GetBootloaderStatus()
		if err != nil {
			continue
		}

		if internal.Debug {
			fmt.Printf("[DEBUG] waitForValidation: status=0x%02X\n", status)
		}

		switch status {
		case BLStatusReady:
			fmt.Println("Firmware validated. Ready to jump.")
			return nil
		case BLStatusError:
			return fmt.Errorf("bootloader reported error during validation")
		case BLStatusValidating:
			// Still validating, keep waiting
		}
	}

	return fmt.Errorf("timeout waiting for firmware validation")
}

// verifyCRC sends a CRC check command and waits for the reply.
func (c *Client) verifyCRC(expectedCRC uint32) error {
	data := make([]byte, MaxCANDataLen)
	binary.LittleEndian.PutUint32(data[:4], expectedCRC)

	if internal.Debug {
		fmt.Printf("[DEBUG] verifyCRC: sending CRC=0x%08X to ID 0x%08X\n", expectedCRC, IDAPCRC)
	}

	resp, err := c.SendAndWait(IDAPCRC, data, IDAPCRCReply, TimeoutFWLastChunk)
	if err != nil {
		return fmt.Errorf("CRC check failed: %w", err)
	}

	result := resp.Data[0]
	if result != 0x00 {
		return fmt.Errorf("CRC mismatch: BMS returned status 0x%02X", result)
	}

	fmt.Println("CRC verified OK.")
	return nil
}

// UpdateFirmwareAP updates the application firmware using the AP update protocol.
// This is an alternative to the SPIN bootloader for some firmware versions.
//
// Protocol:
//  1. Request version (IDAPVersion -> IDAPVersionReply)
//  2. Request flash size (IDAPFlashSize -> IDAPFlashSizeReply)
//  3. Erase flash (IDAPErase)
//  4. Write firmware chunks (IDAPWrite)
//  5. Verify CRC (IDAPCRC -> IDAPCRCReply)
//  6. Send update command (IDAPUpdate -> IDAPUpdateReply)
//  7. Reboot (IDAPReboot)
func (c *Client) UpdateFirmwareAP(filename string) {
	if filename == "" {
		log.Fatal("No firmware file specified. Use --firmware-file to provide the .bin file.")
	}

	firmware, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read firmware file: %v", err)
	}

	fmt.Printf("AP Firmware file: %s (%d bytes)\n", filename, len(firmware))
	fwCRC := crc32.ChecksumIEEE(firmware)

	// Step 1: Get version
	fmt.Println("Requesting AP version...")
	resp, err := c.SendAndWait(IDAPVersion, padTo8(nil), IDAPVersionReply, TimeoutCommand)
	if err != nil {
		log.Fatalf("Failed to get AP version: %v", err)
	}
	if internal.Debug {
		fmt.Printf("[DEBUG] UpdateFirmwareAP: version reply data=%X\n", resp.Data)
	}

	// Step 2: Get flash size
	fmt.Println("Requesting flash size...")
	resp, err = c.SendAndWait(IDAPFlashSize, padTo8(nil), IDAPFlashSizeReply, TimeoutCommand)
	if err != nil {
		log.Fatalf("Failed to get flash size: %v", err)
	}
	flashSize := binary.LittleEndian.Uint32(resp.Data[:4])
	fmt.Printf("Flash size: %d bytes\n", flashSize)

	if uint32(len(firmware)) > flashSize {
		log.Fatalf("Firmware (%d bytes) exceeds flash size (%d bytes)", len(firmware), flashSize)
	}

	// Step 3: Erase flash
	fmt.Println("Erasing flash...")
	if err := c.SendFrame(IDAPErase, padTo8(nil)); err != nil {
		log.Fatalf("Failed to send erase command: %v", err)
	}
	// Wait for erase to complete
	time.Sleep(5 * time.Second)

	// Step 4: Write firmware
	fmt.Println("Writing firmware...")
	chunkSize := FWChunkSize
	totalChunks := (len(firmware) + chunkSize - 1) / chunkSize

	for i := 0; i < totalChunks; i++ {
		offset := i * chunkSize
		end := offset + chunkSize
		if end > len(firmware) {
			end = len(firmware)
		}
		chunk := firmware[offset:end]

		// Build write frame: [offset_lo, offset_hi, data...]
		writeData := make([]byte, MaxCANDataLen)
		binary.LittleEndian.PutUint16(writeData[:2], uint16(i))
		copy(writeData[2:], chunk)

		if err := c.SendFrame(IDAPWrite, writeData); err != nil {
			log.Fatalf("Failed to write chunk %d: %v", i, err)
		}

		if (i+1)%100 == 0 || i+1 == totalChunks {
			fmt.Printf("Progress: %d/%d chunks (%.1f%%)\n", i+1, totalChunks, float64(i+1)/float64(totalChunks)*100)
		}

		time.Sleep(DelayBetweenChunks)
	}

	// Step 5: Verify CRC
	fmt.Println("Verifying CRC...")
	if err := c.verifyCRC(fwCRC); err != nil {
		log.Fatalf("CRC verification failed: %v", err)
	}

	// Step 6: Send update command
	fmt.Println("Sending update command...")
	resp, err = c.SendAndWait(IDAPUpdate, padTo8(nil), IDAPUpdateReply, TimeoutFWLastChunk)
	if err != nil {
		log.Fatalf("Update command failed: %v", err)
	}
	if internal.Debug {
		fmt.Printf("[DEBUG] UpdateFirmwareAP: update reply data=%X\n", resp.Data)
	}

	// Step 7: Reboot
	fmt.Println("Rebooting BMS...")
	if err := c.SendFrame(IDAPReboot, padTo8(nil)); err != nil {
		log.Fatalf("Failed to send reboot: %v", err)
	}

	fmt.Println("AP firmware update complete! BMS is rebooting.")
}
