package can

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"
)

// gpiohandleRequest matches struct gpiohandle_request from <linux/gpio.h> (ABI v1).
//   GPIOHANDLES_MAX   = 64
//   GPIO_MAX_NAME_SIZE = 32
// Layout: [64]u32 + u32 + [64]u8 + [32]byte + u32 + s32 = 364 bytes (no padding)
type gpiohandleRequest struct {
	LineOffsets   [64]uint32
	Flags         uint32
	DefaultValues [64]uint8
	ConsumerLabel [32]byte
	Lines         uint32
	Fd            int32
}

const (
	gpiohandleFlagOutput = uint32(0x02)

	// GPIO_GET_LINEHANDLE_IOCTL = _IOWR(0xB4, 0x03, struct gpiohandle_request)
	// _IOWR: dir=3(R|W) << 30 | size(364=0x16C) << 16 | type(0xB4) << 8 | nr(3)
	// = 0xC0000000 | 0x016C0000 | 0x0000B400 | 0x00000003 = 0xC16CB403
	gpioGetLineHandleIoctl = uintptr(0xC16CB403)
)

// HoldWakeLine asserts GPIO17 HIGH via the Linux GPIO character device (ABI v1).
// It opens /dev/gpiochip0, calls GPIO_GET_LINEHANDLE_IOCTL to claim line 17 as
// an output driven HIGH, then immediately closes the chip fd - only the returned
// line handle fd needs to stay open to hold the pin state.
// Returns a cleanup func that closes the line fd and releases the pin.
func HoldWakeLine() func() {
	chipFd, err := unix.Open("/dev/gpiochip0", unix.O_RDWR, 0)
	if err != nil {
		fmt.Printf("[GPIO] Warning: cannot open /dev/gpiochip0: %v\n", err)
		return func() {}
	}

	req := gpiohandleRequest{
		Lines: 1,
		Flags: gpiohandleFlagOutput,
	}
	req.LineOffsets[0] = 17
	req.DefaultValues[0] = 1 // HIGH
	copy(req.ConsumerLabel[:], "bms-wake")

	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(chipFd),
		gpioGetLineHandleIoctl,
		uintptr(unsafe.Pointer(&req)),
	)

	// Chip fd is not needed once the line handle is obtained.
	unix.Close(chipFd)

	if errno != 0 {
		if errno == unix.EBUSY {
			fmt.Println("[GPIO] GPIO17 already held by another process - continuing")
			return func() {}
		}
		fmt.Printf("[GPIO] Warning: GPIO ioctl failed: %v\n", errno)
		return func() {}
	}

	lineFd := int(req.Fd)
	fmt.Printf("[GPIO] GPIO17 HIGH (line fd %d)\n", lineFd)

	return func() {
		unix.Close(lineFd)
		fmt.Println("[GPIO] GPIO17 released")
	}
}
