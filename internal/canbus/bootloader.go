package canbus

import (
	"bms/v2/internal"
	"encoding/binary"
	"fmt"
	"time"
)

// SPIN Bootloader commands.
const (
	BLCmdHeartbeat       = 0
	BLCmdVersions        = 1
	BLCmdEnterBootloader = 2
	BLCmdExitBootloader  = 3
	BLCmdStartUpload     = 4
	BLCmdRequestChunk    = 5
	BLCmdSendChunk       = 6
	BLCmdSystemShutdown  = 7
)

// Bootloader heartbeat status values.
const (
	BLStatusWaiting    byte = 0x00
	BLStatusErasing    byte = 0x01
	BLStatusUploading  byte = 0x02
	BLStatusValidating byte = 0x03
	BLStatusReady      byte = 0x04
	BLStatusError      byte = 0x15
)

// exitBootloaderMagic is sent to exit the bootloader.
var exitBootloaderMagic = []byte{0xDE, 0xAD, 0xC0, 0xDE, 0xCA, 0xFE, 0x55, 0xAA}

// MCU bootloader command bytes and their checksums.
const (
	mcuBLCmd1     byte = 0x11
	mcuBLCmd1Chk  byte = 0xEE
	mcuBLCmd2     byte = 0x21
	mcuBLCmd2Chk  byte = 0xDE
	mcuBLCmd3     byte = 0x31
	mcuBLCmd3Chk  byte = 0xCE
)

// EnterBootloader sends the ENTER_BOOTLOADER command and waits for the BMS
// to enter bootloader mode (heartbeat status = Waiting).
func (c *Client) EnterBootloader() error {
	if internal.Debug {
		fmt.Println("[DEBUG] EnterBootloader: sending enter bootloader command")
	}

	fmt.Println("Entering bootloader mode...")

	// Send enter bootloader command
	data := padTo8([]byte{BLCmdEnterBootloader})
	if err := c.SendFrame(IDBootloaderData, data); err != nil {
		return fmt.Errorf("failed to send enter bootloader: %w", err)
	}

	// Wait for heartbeat with status = Waiting
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		frame, err := c.WaitForID(IDBootloaderACK, TimeoutBLWait)
		if err != nil {
			continue
		}

		status := frame.Data[0]
		if internal.Debug {
			fmt.Printf("[DEBUG] EnterBootloader: heartbeat status=0x%02X\n", status)
		}

		if status == BLStatusWaiting {
			fmt.Println("Bootloader ready (waiting for upload).")
			return nil
		}
	}

	return fmt.Errorf("timeout waiting for bootloader to enter waiting state")
}

// ExitBootloader sends the EXIT_BOOTLOADER command with the magic bytes.
func (c *Client) ExitBootloader() error {
	if internal.Debug {
		fmt.Printf("[DEBUG] ExitBootloader: sending magic bytes %X\n", exitBootloaderMagic)
	}

	fmt.Println("Exiting bootloader mode...")

	if err := c.SendFrame(IDBootloaderData, exitBootloaderMagic); err != nil {
		return fmt.Errorf("failed to send exit bootloader: %w", err)
	}

	time.Sleep(DelayAfterExit)
	fmt.Println("Bootloader exit sent.")
	return nil
}

// StartUpload sends the START_UPLOAD command with the firmware size in little-endian.
func (c *Client) StartUpload(firmwareSize uint32) error {
	if internal.Debug {
		fmt.Printf("[DEBUG] StartUpload: firmwareSize=%d (0x%X)\n", firmwareSize, firmwareSize)
	}

	data := make([]byte, MaxCANDataLen)
	data[0] = BLCmdStartUpload
	binary.LittleEndian.PutUint32(data[1:5], firmwareSize)

	return c.SendFrame(IDBootloaderData, data)
}

// SendFirmwareChunk sends a single firmware chunk (up to 6 bytes) with the chunk number
// encoded in the CAN ID offset.
func (c *Client) SendFirmwareChunk(chunkNumber uint32, chunk []byte) error {
	if internal.Debug {
		fmt.Printf("[DEBUG] SendFirmwareChunk: chunk=%d len=%d\n", chunkNumber, len(chunk))
	}

	// The CAN ID for SEND_CHUNK is offset by the chunk number
	id := IDBootloaderData + chunkNumber

	data := make([]byte, MaxCANDataLen)
	data[0] = BLCmdSendChunk
	data[1] = byte(len(chunk))
	copy(data[2:], chunk)

	return c.SendFrame(id, data)
}

// WaitForChunkRequest waits for a REQUEST_CHUNK command from the BMS.
// Returns the requested address and chunk ID.
func (c *Client) WaitForChunkRequest(timeout time.Duration) (address uint32, chunkID byte, err error) {
	frame, err := c.WaitForID(IDBootloaderACK, timeout)
	if err != nil {
		return 0, 0, err
	}

	if frame.Data[0] != BLCmdRequestChunk {
		return 0, 0, fmt.Errorf("expected REQUEST_CHUNK (0x%02X), got 0x%02X", BLCmdRequestChunk, frame.Data[0])
	}

	address = binary.LittleEndian.Uint32([]byte{frame.Data[1], frame.Data[2], frame.Data[3], 0})
	chunkID = frame.Data[2]

	if internal.Debug {
		fmt.Printf("[DEBUG] WaitForChunkRequest: address=0x%X chunkID=%d\n", address, chunkID)
	}

	return address, chunkID, nil
}

// SendMCUBootloaderCommand sends an MCU bootloader command with its checksum byte.
func (c *Client) SendMCUBootloaderCommand(cmd byte, chk byte) error {
	if internal.Debug {
		fmt.Printf("[DEBUG] SendMCUBootloaderCommand: cmd=0x%02X chk=0x%02X\n", cmd, chk)
	}

	data := padTo8([]byte{cmd, chk})
	return c.SendFrame(IDBootloaderData, data)
}

// SendBLAddress sends a 4-byte address with XOR checksum to the bootloader.
// Format: [addr0, addr1, addr2, addr3, xor_checksum]
func (c *Client) SendBLAddress(address uint32) error {
	addrBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(addrBytes, address)

	data := make([]byte, MaxCANDataLen)
	copy(data[:4], addrBytes)
	data[4] = xorChecksum(addrBytes)

	if internal.Debug {
		fmt.Printf("[DEBUG] SendBLAddress: address=0x%08X checksum=0x%02X\n", address, data[4])
	}

	return c.SendFrame(IDBootloaderData, data)
}

// SendBLData sends bootloader data bytes with XOR checksum.
// Format: [length-1, data..., xor_checksum]
func (c *Client) SendBLData(payload []byte) error {
	if len(payload) == 0 || len(payload) > 6 {
		return fmt.Errorf("bootloader data must be 1-6 bytes, got %d", len(payload))
	}

	data := make([]byte, MaxCANDataLen)
	data[0] = byte(len(payload) - 1)
	copy(data[1:], payload)

	// XOR checksum of length byte + data bytes
	chkData := data[:1+len(payload)]
	data[1+len(payload)] = xorChecksum(chkData)

	if internal.Debug {
		fmt.Printf("[DEBUG] SendBLData: len=%d checksum=0x%02X\n", len(payload), data[1+len(payload)])
	}

	return c.SendFrame(IDBootloaderData, data)
}

// GetBootloaderStatus reads the current bootloader heartbeat status.
func (c *Client) GetBootloaderStatus() (byte, error) {
	frame, err := c.WaitForID(IDBootloaderACK, TimeoutBLWait)
	if err != nil {
		return 0, err
	}
	return frame.Data[0], nil
}
