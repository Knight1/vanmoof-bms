package canbus

import "time"

// CAN IDs for BMS communication (29-bit extended IDs).
const (
	// Main register read/write command
	IDCommand uint32 = 0x10801000

	// Write commands
	IDWrite        uint32 = 0x1077E400
	IDWriteVariant uint32 = 0x1077E800

	// Bootloader (MCU)
	IDBootloaderData    uint32 = 0x1077F200
	IDBootloaderACK     uint32 = 0x1077F400
	IDBootloaderVersion uint32 = 0x1077F600

	// Heartbeat
	IDHeartbeat uint32 = 0x10801000
)

// AP (Application) firmware update CAN IDs.
const (
	IDAPVersionReply   uint32 = 0x14901460
	IDAPVersion        uint32 = 0x14901470
	IDAPFlashSizeReply uint32 = 0x14903460
	IDAPFlashSize      uint32 = 0x14903470
	IDAPErase          uint32 = 0x14905470
	IDAPWrite          uint32 = 0x14907470
	IDAPCRCReply       uint32 = 0x14909460
	IDAPCRC            uint32 = 0x14909470
	IDAPUpdateReply    uint32 = 0x1490B460
	IDAPUpdate         uint32 = 0x1490B470
	IDAPReboot         uint32 = 0x1490D470
)

// Version-specific SetStatus command IDs.
var setStatusIDs = map[string]uint32{
	"001": 0x30,       // Standard 11-bit for v001
	"004": 0x149D9060, // Extended for v004
	"015": 0x1489BB70, // Extended for v015+
	"016": 0x1489BB70,
	"019": 0x1489BB70,
	"020": 0x1489BB70,
}

// Version 004 additional command IDs.
const (
	IDSetChargerState uint32 = 0x149DE9A0
	IDSetUSBPDState   uint32 = 0x149E4160
)

// Status codes for SetStatus command (version 001).
const (
	StatusNormal          byte = 0x01
	StatusCharge          byte = 0x02
	StatusSleepShipping   byte = 0x03
	StatusChargeOff       byte = 0x05
	StatusStandby         byte = 0x08
	StatusSetChargeOn     byte = 0x0C
	StatusChargeOpen      byte = 0x0D
	StatusSpecialControl  byte = 0x20
	StatusSetSN           byte = 0x31
	StatusRequestBMSInfo  byte = 0x40
	StatusCellBalancing   byte = 0x80
)

// Timeouts
const (
	TimeoutCommand     = 2000 * time.Millisecond
	TimeoutHeartbeat   = 2000 * time.Millisecond
	TimeoutFWChunk     = 1000 * time.Millisecond
	TimeoutFWLastChunk = 20000 * time.Millisecond
	TimeoutBLWait      = 2000 * time.Millisecond

	DelayAfterCommand = 50 * time.Millisecond
	DelayBetweenChunks = 2 * time.Millisecond
	DelayAfterExit     = 100 * time.Millisecond
)

// MaxCANDataLen is the maximum payload size per CAN frame.
const MaxCANDataLen = 8

// FWChunkSize is the firmware chunk size for the SPIN bootloader upload.
// Each chunk is 6 bytes (CAN frame data minus header bytes).
const FWChunkSize = 6

// padTo8 pads data to 8 bytes with zeros.
func padTo8(data []byte) []byte {
	if len(data) >= MaxCANDataLen {
		return data[:MaxCANDataLen]
	}
	padded := make([]byte, MaxCANDataLen)
	copy(padded, data)
	return padded
}

// xorChecksum computes the XOR checksum of all bytes.
func xorChecksum(data []byte) byte {
	var chk byte
	for _, b := range data {
		chk ^= b
	}
	return chk
}
