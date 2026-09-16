package canbus

import (
	"bms/v2/internal"
	"fmt"
	"os"
	"time"
)

// ConnectToBMS opens the CAN bus, starts the read loop, and waits for the BMS
// command heartbeat to confirm an A5/S5/S6 battery is present. Only the A5/S5/S6
// battery speaks this CAN protocol; the S3/S4 battery is Modbus/UART only and
// never appears on the CAN bus, so requiring the heartbeat keeps the CAN command
// path from acting on anything but an A5/S5/S6 battery.
func ConnectToBMS(client *Client) error {
	if internal.Debug {
		fmt.Printf("[DEBUG] ConnectToBMS: interface=%s loop=%v retries=%d\n",
			client.iface, internal.Loop, internal.ConnectionRetries)
	}

	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to start CAN bus: %w", err)
	}

	// Wait for the A5/S5/S6 BMS command heartbeat to confirm the battery is present.
	fmt.Println("Waiting for A5/S5/S6 BMS heartbeat on CAN bus...")

	for attempt := 0; internal.Loop || attempt < internal.ConnectionRetries; attempt++ {
		if internal.Debug {
			fmt.Printf("[DEBUG] ConnectToBMS: attempt %d - waiting for heartbeat\n", attempt+1)
		}

		_, err := client.WaitForID(IDCommand, TimeoutHeartbeat)
		if err != nil {
			if internal.Debug {
				fmt.Printf("[DEBUG] ConnectToBMS: no heartbeat: %v\n", err)
			}
			time.Sleep(internal.ConnectionRetryDelay)
			continue
		}

		fmt.Println("A5/S5/S6 BMS heartbeat received. Connection established.")
		if internal.Debug {
			fmt.Printf("[DEBUG] ConnectToBMS: connected after %d attempts\n", attempt+1)
		}
		return nil
	}

	fmt.Println("Retry counter exceeded. Giving up. Retry counter:", internal.ConnectionRetries)
	fmt.Println("Failed to connect to an A5/S5/S6 BMS via CAN bus. Check if")
	fmt.Println("-> The CAN interface is up (ip link set can0 up type can bitrate 500000)")
	fmt.Println("-> The CAN transceiver is wired correctly (CANH/CANL)")
	fmt.Println("-> The BMS is powered on")
	fmt.Println("-> This is an A5/S5/S6 battery (the S3/S4 battery is Modbus/UART only — use --protocol modbus)")
	os.Exit(1)
	return nil
}

// MonitorHeartbeat runs in a goroutine and tracks BMS heartbeats.
// If too many heartbeats are missed, it logs a warning.
func MonitorHeartbeat(client *Client) {
	const maxMissed = 10
	ticker := time.NewTicker(TimeoutHeartbeat)
	defer ticker.Stop()

	for range ticker.C {
		client.heartbeatMu.Lock()
		elapsed := time.Since(client.lastHeartbeat)
		if elapsed > TimeoutHeartbeat {
			client.missedHeartbeats++
			if client.missedHeartbeats >= maxMissed {
				fmt.Printf("WARNING: BMS heartbeat not received for %v (%d missed)\n", elapsed, client.missedHeartbeats)
			}
		}
		client.heartbeatMu.Unlock()
	}
}
