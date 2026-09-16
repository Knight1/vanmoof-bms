# BMS Register & Command Reference

Complete reference for the DynaPack SX3 BMS (VM13-147) used in VanMoof S3/S4 bicycles.

## Protocol Configuration

| Parameter | Value |
|-----------|-------|
| Protocol | Modbus RTU |
| Slave ID | 0xAA (170) |
| Baud Rate | 9600 |
| Data Bits | 8 |
| Parity | None |
| Stop Bits | 1 |

---

## Passive Registers (0x02 - 0x2C)

Read via Modbus holding registers. These reflect live BMS state.

| Hex | Dec | Name | Type | Formula | Unit |
|-----|-----|------|------|---------|------|
| 0x02 | 2 | Fault Status | H (16-bit flags) | - | hex |
| 0x03 | 3 | Battery Temperature | U16 | (x-2731)/10 | °C |
| 0x04 | 4 | Battery Voltage | U16 | - | mV |
| 0x05 | 5 | RSOC (Real State of Charge) | U16 | - | % |
| 0x06 | 6 | Current | I16 | x*10 | mA |
| 0x07 | 7 | Charging Status | H (16-bit flags) | - | hex |
| 0x08 | 8 | Discharging Status | H (16-bit flags) | - | hex |
| 0x09 | 9 | Test Mode | H | - | hex |
| 0x0A | 10 | Hardware Version | H | - | hex |
| 0x0B | 11 | Software Version | H | - | hex |
| 0x0C-0x12 | 12-18 | ESN (Electronic Serial Number) | ASCII | - | 14 chars (7 regs) |
| 0x13-0x14 | 19-20 | Manufacture Date | DATE | - | [0x00, Year, Month, Day] |
| 0x15 | 21 | Normal Capacity | U16 | - | mAh |
| 0x16 | 22 | Full Charge Capacity | U16 | - | mAh |
| 0x17 | 23 | Remaining Capacity | U16 | - | mAh |
| 0x18 | 24 | Absolute SOC | U16 | - | % |
| 0x19 | 25 | Cycle Count | U16 | - | count |
| 0x1A | 26 | CHG MOS Control | H (16-bit flags) | - | hex |
| 0x1B | 27 | Cell 1 Voltage | U16 | - | mV |
| 0x1C | 28 | Cell 2 Voltage | U16 | - | mV |
| 0x1D | 29 | Cell 3 Voltage | U16 | - | mV |
| 0x1E | 30 | Cell 4 Voltage | U16 | - | mV |
| 0x1F | 31 | Cell 5 Voltage | U16 | - | mV |
| 0x20 | 32 | Cell 6 Voltage | U16 | - | mV |
| 0x21 | 33 | Cell 7 Voltage | U16 | - | mV |
| 0x22 | 34 | Cell 8 Voltage | U16 | - | mV |
| 0x23 | 35 | Cell 9 Voltage | U16 | - | mV |
| 0x24 | 36 | Cell 10 Voltage | U16 | - | mV |
| 0x25 | 37 | Temperature Sensor 1 | U16 | (x-2731)/10 | °C |
| 0x26 | 38 | Temperature Sensor 2 | U16 | (x-2731)/10 | °C |
| 0x27 | 39 | Discharge MOSFET Temperature | U16 | (x-2731)/10 | °C |
| 0x28 | 40 | Warning Status | H (16-bit flags) | - | hex |
| 0x29 | 41 | Max Cell Voltage | U16 | - | mV |
| 0x2A | 42 | Min Cell Voltage | U16 | - | mV |
| 0x2B | 43 | Cell Balance | H | - | - |
| 0x2C | 44 | Bootloader Version | H | - | hex |

---

## DataFlash Registers (0x30 - 0x5B)

Stored BMS records. Read via Modbus holding registers after reading passive registers.

| Hex | Dec | Name | Type | Formula | Unit |
|-----|-----|------|------|---------|------|
| 0x30 | 48 | Fault Status Record | H (16-bit flags) | - | hex |
| 0x31 | 49 | Battery Temp Sensor 1 Record | U16 | (x-2731)/10 | °C |
| 0x32 | 50 | Battery Temp Sensor 2 Record | U16 | (x-2731)/10 | °C |
| 0x33 | 51 | MOSFET Temperature Record | U16 | (x-2731)/10 | °C |
| 0x34 | 52 | Battery Voltage Record | U16 | - | mV |
| 0x35 | 53 | Current Record | I16 | x*10 | mA |
| 0x36 | 54 | Full Charge Capacity Record | U16 | - | mAh |
| 0x37 | 55 | Remaining Capacity Record | U16 | - | mAh |
| 0x38 | 56 | RSOC Record | U16 | - | % |
| 0x39 | 57 | Absolute SOC Record | U16 | - | % |
| 0x3A | 58 | Cycle Count Record | U16 | - | count |
| 0x3B | 59 | Cell 1 Voltage Record | U16 | - | mV |
| 0x3C | 60 | Cell 2 Voltage Record | U16 | - | mV |
| 0x3D | 61 | Cell 3 Voltage Record | U16 | - | mV |
| 0x3E | 62 | Cell 4 Voltage Record | U16 | - | mV |
| 0x3F | 63 | Cell 5 Voltage Record | U16 | - | mV |
| 0x40 | 64 | Cell 6 Voltage Record | U16 | - | mV |
| 0x41 | 65 | Cell 7 Voltage Record | U16 | - | mV |
| 0x42 | 66 | Cell 8 Voltage Record | U16 | - | mV |
| 0x43 | 67 | Cell 9 Voltage Record | U16 | - | mV |
| 0x44 | 68 | Cell 10 Voltage Record | U16 | - | mV |
| 0x45 | 69 | Max Battery Voltage Record | U16 | - | mV |
| 0x46 | 70 | Min Battery Voltage Record | U16 | - | mV |
| 0x47 | 71 | DOTP Trigger Count | U16 | - | count |
| 0x48 | 72 | DUTP Trigger Count | U16 | - | count |
| 0x49 | 73 | COTP Trigger Count | U16 | - | count |
| 0x4A | 74 | CUTP Trigger Count | U16 | - | count |
| 0x4B | 75 | DOCP1 Trigger Count | U16 | - | count |
| 0x4C | 76 | DOCP2 Trigger Count | U16 | - | count |
| 0x4D | 77 | COCP1 Trigger Count | U16 | - | count |
| 0x4E | 78 | COCP2 Trigger Count | U16 | - | count |
| 0x4F | 79 | OVP1 Trigger Count | U16 | - | count |
| 0x50 | 80 | OVP2 Trigger Count | U16 | - | count |
| 0x51 | 81 | UVP1 Trigger Count | U16 | - | count |
| 0x52 | 82 | UVP2 Trigger Count | U16 | - | count |
| 0x53 | 83 | PDOCP Trigger Count | U16 | - | count |
| 0x54 | 84 | PDSCP Trigger Count | U16 | - | count |
| 0x55 | 85 | MOTP Trigger Count | U16 | - | count |
| 0x56 | 86 | SCP Trigger Count | U16 | - | count |
| 0x57 | 87 | Max Charging Current Record | I16 | x*10 | mA |
| 0x58 | 88 | Max Discharging Current Record | I16 | x*10 | mA |
| 0x59 | 89 | Max Cell Temperature Record | I16 | (x-2731)/10 | °C |
| 0x5A | 90 | Min Cell Temperature Record | I16 | (x-2731)/10 | °C |
| 0x5B | 91 | Max MOSFET Temperature Record | I16 | (x-2731)/10 | °C |

---

## Firmware Update Registers

| Hex | Dec | Name | R/W | Value | Description |
|-----|-----|------|-----|-------|-------------|
| 0x80 | 128 | Reset MCU | W | 0 | Resets the BMS microcontroller |
| 0x81 | 129 | CRC Check Status | R | - | 0=OK, 1=CRC Error, 2=No File |
| 0x82 | 130 | Firmware Write Data | W | bytes | 4-byte address (big-endian) + up to 32 bytes data |
| 0x95 | 149 | Erase Shadow Flash | W | 0 | Erases shadow flash before programming |

### Firmware Update Protocol

1. **Erase shadow flash** — Write 0 to register 0x95
2. **Write firmware chunks** — Send 32-byte chunks to register 0x82, each prefixed with a 4-byte big-endian offset from base address 0x08005000
3. **Verify CRC** — Read register 0x81 (0=OK, 1=CRC Error, 2=No File)
4. **Reset MCU** — Write 0 to register 0x80
5. **Wait for reboot** — Poll register 0x00 until it reads 0x0100 (timeout: 180s)

---

## Log Export Registers

| Hex | Dec | Name | R/W | Description |
|-----|-----|------|-----|-------------|
| 0x0F45 | 3909 | Log Entry Select | W | Write log ID (0-99) to select which entry to read |
| 0x0F30-0x0F44 | - | Log Data | R | Read from 0x0F00 + register address (0x30-0x44) |

The BMS stores 100 log entries. Write the log ID to 0x0F45, then read each DataFlash register from the 0x0F00+ address space.

---

## Modbus Write Commands

| Action | Register | Value | Description |
|--------|----------|-------|-------------|
| TurnDebugOn | 0x09 (9) | 1 | Enable debug/test mode |
| TurnDebugOff | 0x09 (9) | 0 | Disable debug/test mode |
| TurnDischargingOn | 0x08 (8) | 1 | Enable discharging |
| TurnDischargingOff | 0x08 (8) | 0 | Disable discharging |
| TurnChargeMOSOn | 0x1A (26) | 1 | Enable charge MOSFET |
| TurnChargeMOSOff | 0x1A (26) | 0 | Disable charge MOSFET |
| ResetESNModbus | 0x0A (10) | 0 | Clear Electronic Serial Number |
| ResetMCU | 0x80 (128) | 0 | Reset BMS microcontroller |
| ShipMode | 0x01 (1) | 0 | Enter ship mode (lowest power state) |
| ShipAndDischargeTurnOff | 0x01 (1) then 0x08 (8) | 0, then 0 | Ship mode + disable discharge (20ms delay between) |
| WriteESN | 0x0C-0x12 (12-18) | ASCII bytes | Write 14-char ESN packed into 7 registers |
| WriteDate | 0x13-0x14 (19-20) | DATE bytes | Write manufacture date [0x00, Year, Month, Day] |

---

## UART Serial Commands

Sent as ASCII strings over serial (9600 baud, 8N1). These do not use Modbus protocol.

### GPIO Control

| Command | Expected Response | Delay | Description |
|---------|-------------------|-------|-------------|
| `GPIO.PF2=1.` | - | - | Enable charge port GPIO (ON) |
| `GPIO.PF2=0.` | - | - | Disable charge port GPIO (OFF) |
| `GPIO.IO2=1.` | - | - | Enable detect pin IO2 (ON) |
| `GPIO.IO2=0.` | - | - | Disable detect pin IO2 (OFF) |
| `GPIO.IO1=1.` | - | - | Enable key input IO1 (ON) |
| `GPIO.IO1=0.` | - | - | Disable key input IO1 (OFF) |

### Calibration

| Command | Expected Response | Delay | Description |
|---------|-------------------|-------|-------------|
| `DSG CAL=<mA>` | - | 7s | Calibrate discharge current (mA positive integer) |
| `CHG CAL=<mA>` | - | 7s | Calibrate charge current (mA positive integer) |

### Log & System

| Command | Expected Response | Delay | Description |
|---------|-------------------|-------|-------------|
| `Log Clear` | `OK` | 3s | Clear BMS event log |
| `PF=0` | - | - | Clear power failure flags (may need retries) |
| `Reset BMS V0106` | - | 1s | Factory reset BMS (clears ESN, calibration, cycles) |
| `Reset ESN` | `Done` or `Reset ESN fail` | - | Clear Electronic Serial Number |

---

## Flag Definitions

### Fault Status (Register 0x02) — 16-bit

| Bit | Flag | Description |
|-----|------|-------------|
| 0 | DOTP | Discharge Over Temperature Protection |
| 1 | DUTP | Discharge Under Temperature Protection |
| 2 | COTP | Charging Over Temperature Protection |
| 3 | CUTP | Charging Under Temperature Protection |
| 4 | DOCP1 | Discharge Over Current Protection Level 1 |
| 5 | DOCP2 | Discharge Over Current Protection Level 2 |
| 6 | COCP1 | Charging Over Current Protection Level 1 |
| 7 | COCP2 | Charging Over Current Protection Level 2 |
| 8 | OVP1 | Over Voltage Protection Level 1 |
| 9 | OVP2 | Over Voltage Protection Level 2 |
| 10 | UVP1 | Under Voltage Protection Level 1 |
| 11 | UVP2 | Under Voltage Protection Level 2 |
| 12 | PDOCP | Peak Discharge Over Current Protection |
| 13 | PDSCP | Peak Discharge Short Circuit Protection |
| 14 | MOTP | MOSFET Over Temperature Protection |
| 15 | SCP | Short Circuit Protection |

### Warning Status (Register 0x28) — 16-bit

| Bit | Flag | Description |
|-----|------|-------------|
| 0 | DOTPW | Discharge Over Temperature Warning |
| 1 | DUTPW | Discharge Under Temperature Warning |
| 2 | COTPW | Charging Over Temperature Warning |
| 3 | CUTPW | Charging Under Temperature Warning |
| 4 | DOCPW | Discharge Over Current Warning |
| 5 | - | Reserved |
| 6 | COCPW | Charging Over Current Warning |
| 7 | - | Reserved |
| 8 | OVP1W | Over Voltage Warning Level 1 |
| 9 | - | Reserved |
| 10 | UVP1W | Under Voltage Warning Level 1 |
| 11 | SOC | State of Charge Low Warning |
| 12 | PDOCPW | Peak Discharge Over Current Warning |
| 13 | - | Reserved |
| 14 | MOTPW | MOSFET Over Temperature Warning |
| 15 | - | Reserved |

### Charging Status (Register 0x07) — 16-bit

| Bit | Flag | Description |
|-----|------|-------------|
| 0 | CHG | Charging active |
| 1 | Fault | Charging fault detected |
| 2 | CHG_IN | Charger input detected |
| 3-15 | - | Reserved |

### Discharging Status (Register 0x08) — 16-bit

| Bit | Flag | Description |
|-----|------|-------------|
| 0 | DSG | Discharging active |
| 1-15 | - | Reserved |

### CHG MOS Control (Register 0x1A) — 16-bit

| Bit | Flag | Description |
|-----|------|-------------|
| 0 | CHG | Charge MOSFET enabled |
| 1-15 | - | Reserved |

---

## Conversion Formulas

| Parameter | Formula | Example |
|-----------|---------|---------|
| Temperature | (raw - 2731) / 10 | 3000 -> 26.9°C |
| Current | raw * 10 | 100 -> 1000 mA |
| Voltage | raw (direct) | 36500 -> 36500 mV |
| SOC / Capacity | raw (direct) | 85 -> 85% |

Valid temperature range: raw 2431-3731 (approx. -30°C to 100°C)

---

## Voltage Thresholds (Tool-defined)

| Parameter | Value |
|-----------|-------|
| Cell Voltage Low | 2500 mV |
| Cell Voltage High | 4300 mV |
| Pack Voltage Low | 25000 mV |
| Pack Voltage High | 43000 mV |
| Cell Voltage Imbalance Warning | 5 mV (between adjacent cells) |
| Cell Voltage Imbalance (Overview) | 20 mV (max-min) |
