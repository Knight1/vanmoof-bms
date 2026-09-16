package canbus

import (
	"bms/v2/internal"
	"fmt"
	"time"
)

// SendCommand sends a multi-frame command. Data is padded to 8-byte boundaries
// and sent in 8-byte chunks. The last chunk is marked as final.
func (c *Client) SendCommand(id uint32, data []byte) error {
	if internal.Debug {
		fmt.Printf("[DEBUG] SendCommand: ID=0x%08X dataLen=%d\n", id, len(data))
	}

	// Pad to 8-byte boundary
	padded := data
	if rem := len(data) % MaxCANDataLen; rem != 0 {
		padded = make([]byte, len(data)+MaxCANDataLen-rem)
		copy(padded, data)
	}
	if len(padded) == 0 {
		padded = make([]byte, MaxCANDataLen)
	}

	for offset := 0; offset < len(padded); offset += MaxCANDataLen {
		chunk := padded[offset : offset+MaxCANDataLen]
		if err := c.SendFrame(id, chunk); err != nil {
			return fmt.Errorf("send chunk at offset %d: %w", offset, err)
		}
	}

	return nil
}

// SetStatus sends a status command to the BMS. The command ID depends on the
// detected firmware version.
func (c *Client) SetStatus(status byte) error {
	id, ok := setStatusIDs[c.Version]
	if !ok {
		// Default to v015+ ID
		id = setStatusIDs["015"]
	}

	if internal.Debug {
		fmt.Printf("[DEBUG] SetStatus: version=%s ID=0x%08X status=0x%02X\n", c.Version, id, status)
	}

	data := padTo8([]byte{status, 0x00})
	if err := c.SendFrame(id, data); err != nil {
		return fmt.Errorf("SetStatus failed: %w", err)
	}

	time.Sleep(DelayAfterCommand)
	return nil
}

// Wakeup sends the wakeup/normal mode status to the BMS.
func (c *Client) Wakeup() error {
	if internal.Debug {
		fmt.Println("[DEBUG] Wakeup: sending StatusNormal")
	}
	return c.SetStatus(StatusNormal)
}

// SetChargeOn enables charging.
func (c *Client) SetChargeOn() error {
	if internal.Debug {
		fmt.Println("[DEBUG] SetChargeOn: sending StatusSetChargeOn")
	}
	if err := c.SetStatus(StatusSetChargeOn); err != nil {
		fmt.Println("Error enabling charge:", err)
		return err
	}
	fmt.Println("Charge enabled!")
	return nil
}

// SetChargeOff disables charging.
func (c *Client) SetChargeOff() error {
	if internal.Debug {
		fmt.Println("[DEBUG] SetChargeOff: sending StatusChargeOff")
	}
	if err := c.SetStatus(StatusChargeOff); err != nil {
		fmt.Println("Error disabling charge:", err)
		return err
	}
	fmt.Println("Charge disabled!")
	return nil
}

// SetShipMode puts the BMS into shipping/sleep mode.
func (c *Client) SetShipMode() error {
	if internal.Debug {
		fmt.Println("[DEBUG] SetShipMode: sending StatusSleepShipping")
	}
	if err := c.SetStatus(StatusSleepShipping); err != nil {
		fmt.Println("Error setting ship mode:", err)
		return err
	}
	fmt.Println("Ship mode set!")
	return nil
}

// SetStandby puts the BMS into standby mode.
func (c *Client) SetStandby() error {
	if internal.Debug {
		fmt.Println("[DEBUG] SetStandby: sending StatusStandby")
	}
	if err := c.SetStatus(StatusStandby); err != nil {
		fmt.Println("Error setting standby:", err)
		return err
	}
	fmt.Println("Standby mode set!")
	return nil
}

// RequestBMSInfo requests BMS info broadcast.
func (c *Client) RequestBMSInfo() error {
	if internal.Debug {
		fmt.Println("[DEBUG] RequestBMSInfo: sending StatusRequestBMSInfo")
	}
	return c.SetStatus(StatusRequestBMSInfo)
}

// RequestCellBalancing requests cell balancing status.
func (c *Client) RequestCellBalancing() error {
	if internal.Debug {
		fmt.Println("[DEBUG] RequestCellBalancing: sending StatusCellBalancing")
	}
	return c.SetStatus(StatusCellBalancing)
}

// SetChargerState sends a charger state command (version 004+ only).
func (c *Client) SetChargerState(state byte) error {
	if internal.Debug {
		fmt.Printf("[DEBUG] SetChargerState: state=0x%02X\n", state)
	}

	data := padTo8([]byte{state})
	if err := c.SendFrame(IDSetChargerState, data); err != nil {
		return fmt.Errorf("SetChargerState failed: %w", err)
	}

	time.Sleep(DelayAfterCommand)
	return nil
}

// SetUSBPDState sends a USB-PD state command (version 004+ only).
func (c *Client) SetUSBPDState(state byte) error {
	if internal.Debug {
		fmt.Printf("[DEBUG] SetUSBPDState: state=0x%02X\n", state)
	}

	data := padTo8([]byte{state})
	if err := c.SendFrame(IDSetUSBPDState, data); err != nil {
		return fmt.Errorf("SetUSBPDState failed: %w", err)
	}

	time.Sleep(DelayAfterCommand)
	return nil
}

// SystemShutdown sends the BMS system shutdown command via the bootloader protocol.
func (c *Client) SystemShutdown() error {
	if internal.Debug {
		fmt.Println("[DEBUG] SystemShutdown: sending BMS_SYSTEM_SHUT_DOWN")
	}
	fmt.Println("Sending BMS system shutdown...")
	return c.SetStatus(StatusSleepShipping)
}
