package canbus

// CommandStyle identifies how a register is accessed.
type CommandStyle int

const (
	StylePassive    CommandStyle = iota // Polled registers (被動)
	StyleInitiative                     // Broadcast/active registers (主動)
	StyleBootLoader                     // Bootloader commands
	StyleDataFlash                      // Stored data
	StyleCustomLog                      // Custom log entries
	StyleFlag                           // Bitfield flags
)

// Register defines a single BMS register field within a CAN response.
type Register struct {
	CANID   uint32       // CAN ID this register is received on
	Offset  int          // Byte offset within the CAN data
	Length  int          // Number of bytes (1 or 2)
	Name    string       // Human-readable name
	Formula string       // Conversion: "temp" = (x-2731)/10, "current" = x*10, "" = raw
	Unit    string       // Display unit (°C, mV, mA, %, count, hex)
	Flags   string       // Comma-separated flag names for bitfield registers
	Format  string       // Display format override
	Style   CommandStyle // How this register is accessed
}

// RegisterMap holds all known registers grouped by command style.
// Key: CAN ID -> sub-index -> Register
//
// TODO: Populate from VMS5Manager.xml when available.
// The S5/S6 register XML has not been extracted yet.
// Add registers here once the XML is obtained.
var RegisterMap = map[CommandStyle]map[uint32]map[int]Register{
	StylePassive:    {},
	StyleInitiative: {},
	StyleDataFlash:  {},
	StyleCustomLog:  {},
	StyleFlag:       {},
}

// DataFlashMap holds DataFlash register definitions.
// Key: CAN ID -> sub-index -> Register
//
// TODO: Populate from VMS5Manager.xml when available.
var DataFlashMap = map[uint32]map[int]Register{}

// AddRegister adds a register definition to the register map.
func AddRegister(style CommandStyle, canID uint32, subIndex int, reg Register) {
	reg.Style = style
	if _, ok := RegisterMap[style]; !ok {
		RegisterMap[style] = make(map[uint32]map[int]Register)
	}
	if _, ok := RegisterMap[style][canID]; !ok {
		RegisterMap[style][canID] = make(map[int]Register)
	}
	RegisterMap[style][canID][subIndex] = reg
}

// AddDataFlash adds a DataFlash register definition.
func AddDataFlash(canID uint32, subIndex int, reg Register) {
	if _, ok := DataFlashMap[canID]; !ok {
		DataFlashMap[canID] = make(map[int]Register)
	}
	DataFlashMap[canID][subIndex] = reg
}

// GetPassiveRegisters returns all passive register definitions.
func GetPassiveRegisters() map[uint32]map[int]Register {
	return RegisterMap[StylePassive]
}

// GetInitiativeRegisters returns all initiative (broadcast) register definitions.
func GetInitiativeRegisters() map[uint32]map[int]Register {
	return RegisterMap[StyleInitiative]
}
