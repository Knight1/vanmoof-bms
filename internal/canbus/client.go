package canbus

import (
	"bms/v2/internal"
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/brutella/can"
)

// SocketCAN flags for extended CAN frames (29-bit IDs).
const (
	canEFFFlag uint32 = 0x80000000 // Extended Frame Format flag
	canRTRFlag uint32 = 0x40000000 // Remote Transmission Request flag
	canEFFMask uint32 = 0x1FFFFFFF // 29-bit extended ID mask
)

// Client wraps a SocketCAN bus connection for communicating with the BMS.
type Client struct {
	bus     *can.Bus
	iface   string
	Version string // Detected BMS firmware version: "001", "004", "015", "016", "019", "020"

	// publish sends a frame onto the bus. Defaults to bus.Publish; overridable
	// in tests to capture outgoing frames without a live interface.
	publish func(can.Frame) error

	// Frame routing
	waiters map[uint32]chan can.Frame
	mu      sync.Mutex

	// Heartbeat tracking
	lastHeartbeat     time.Time
	heartbeatMu       sync.RWMutex
	missedHeartbeats  int
	HeartbeatCallback func(data [8]uint8)
}

// NewClient creates a CAN bus client bound to the given SocketCAN interface (e.g. "can0").
func NewClient(iface string) (*Client, error) {
	if internal.Debug {
		fmt.Printf("[DEBUG] NewClient: opening CAN interface %s\n", iface)
	}

	bus, err := can.NewBusForInterfaceWithName(iface)
	if err != nil {
		return nil, fmt.Errorf("failed to open CAN interface %s: %w", iface, err)
	}

	c := &Client{
		bus:     bus,
		iface:   iface,
		waiters: make(map[uint32]chan can.Frame),
	}
	c.publish = bus.Publish

	bus.SubscribeFunc(c.handleFrame)

	return c, nil
}

// Connect starts the CAN bus read loop. Must be called before sending/receiving frames.
func (c *Client) Connect() error {
	if internal.Debug {
		fmt.Println("[DEBUG] Client.Connect: starting CAN bus read loop")
	}
	return c.bus.ConnectAndPublish()
}

// Close disconnects from the CAN bus.
func (c *Client) Close() error {
	if internal.Debug {
		fmt.Println("[DEBUG] Client.Close: disconnecting CAN bus")
	}
	return c.bus.Disconnect()
}

// handleFrame is called for every received CAN frame. It routes frames to
// any goroutine waiting for a specific CAN ID.
func (c *Client) handleFrame(frame can.Frame) {
	id := frame.ID & canEFFMask

	if internal.Debug {
		fmt.Printf("[DEBUG] handleFrame: received CAN ID=0x%08X len=%d data=%X\n", id, frame.Length, frame.Data[:frame.Length])
	}

	// Route heartbeats
	if id == IDHeartbeat || id == IDBootloaderACK {
		c.heartbeatMu.Lock()
		c.lastHeartbeat = time.Now()
		c.missedHeartbeats = 0
		c.heartbeatMu.Unlock()
		if c.HeartbeatCallback != nil {
			c.HeartbeatCallback(frame.Data)
		}
	}

	// Route to any waiting goroutine
	c.mu.Lock()
	if ch, ok := c.waiters[id]; ok {
		select {
		case ch <- frame:
		default:
		}
	}
	c.mu.Unlock()
}

// SendFrame sends a single CAN frame with an extended 29-bit ID.
func (c *Client) SendFrame(id uint32, data []byte) error {
	frame := can.Frame{
		ID:     id | canEFFFlag,
		Length: uint8(len(data)),
	}
	copy(frame.Data[:], data)

	if internal.Debug {
		fmt.Printf("[DEBUG] SendFrame: CAN ID=0x%08X len=%d data=%X\n", id, frame.Length, data)
	}

	return c.publish(frame)
}

// WaitForID blocks until a frame with the given extended CAN ID is received or the timeout expires.
func (c *Client) WaitForID(id uint32, timeout time.Duration) (can.Frame, error) {
	ch := make(chan can.Frame, 1)

	c.mu.Lock()
	c.waiters[id] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.waiters, id)
		c.mu.Unlock()
	}()

	select {
	case frame := <-ch:
		return frame, nil
	case <-time.After(timeout):
		return can.Frame{}, fmt.Errorf("timeout waiting for CAN ID 0x%08X after %v", id, timeout)
	}
}

// WaitForAnyID blocks until a frame with any of the given CAN IDs arrives, or
// the timeout expires. It shares one channel across all IDs and cleanly
// deregisters them on return. Returns the first matching frame.
func (c *Client) WaitForAnyID(ids []uint32, timeout time.Duration) (can.Frame, error) {
	ch := make(chan can.Frame, len(ids))

	c.mu.Lock()
	for _, id := range ids {
		c.waiters[id] = ch
	}
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		for _, id := range ids {
			if c.waiters[id] == ch { // only remove our own waiter
				delete(c.waiters, id)
			}
		}
		c.mu.Unlock()
	}()

	select {
	case frame := <-ch:
		return frame, nil
	case <-time.After(timeout):
		return can.Frame{}, fmt.Errorf("timeout waiting for any of %d CAN IDs after %v", len(ids), timeout)
	}
}

// CollectBurst gathers a burst of frames that all share the given CAN ID, as
// used for multi-frame telemetry records (the payload of one logical value
// split across several back-to-back frames). It waits up to `first` for the
// first frame, then keeps collecting while frames keep arriving within `idle`
// of each other, stopping at maxFrames, on an idle gap, or when a payload
// repeats (the broadcast has looped back to the start of the record).
func (c *Client) CollectBurst(id uint32, maxFrames int, first, idle time.Duration) []can.Frame {
	ch := make(chan can.Frame, maxFrames+4)

	c.mu.Lock()
	c.waiters[id] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.waiters, id)
		c.mu.Unlock()
	}()

	var frames []can.Frame
	var firstPayload []byte
	timeout := first

	for len(frames) < maxFrames {
		select {
		case frame := <-ch:
			payload := append([]byte(nil), frame.Data[:frame.Length]...)
			// Stop once the record loops back to its first frame.
			if len(frames) > 0 && bytes.Equal(payload, firstPayload) {
				return frames
			}
			if len(frames) == 0 {
				firstPayload = payload
			}
			frames = append(frames, frame)
			timeout = idle
		case <-time.After(timeout):
			return frames
		}
	}

	return frames
}

// SendAndWait sends a frame and waits for a response on the given response ID.
func (c *Client) SendAndWait(sendID uint32, data []byte, responseID uint32, timeout time.Duration) (can.Frame, error) {
	if err := c.SendFrame(sendID, data); err != nil {
		return can.Frame{}, fmt.Errorf("send failed: %w", err)
	}
	return c.WaitForID(responseID, timeout)
}

// LastHeartbeat returns the time of the last received heartbeat.
func (c *Client) LastHeartbeat() time.Time {
	c.heartbeatMu.RLock()
	defer c.heartbeatMu.RUnlock()
	return c.lastHeartbeat
}
