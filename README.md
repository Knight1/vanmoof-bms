# VanMoof BMS Toolkit

Diagnostic and control tool for VanMoof DynaPack BMS batteries.
Supports two completely different hardware interfaces depending on the bike model.

```console
go build -trimpath -buildmode=pie -mod=vendor -ldflags "-w -s" -v ./...
```

## Platform support

The intended target is a Linux SBC (Raspberry Pi, `linux/arm64`); build/deploy with `./build.sh`.

- **Linux** - everything: Modbus/UART (S3/S4) and CAN (A5/S5).
- **macOS** - Modbus/UART (S3/S4) only, for local development with a USB-serial adapter. **CAN is not supported on macOS**: it relies on Linux SocketCAN and the GPIO character device, which don't exist on macOS. CAN actions build (via no-op stubs in `internal/can/stub_other.go`) but print a notice and do nothing.

CAN-on-macOS is *feasible* - the CANable adapter speaks SLCAN over `/dev/tty.usbmodem...` directly (no `slcand`, which is Linux-only), and the GPIO wake line can be tied to an external 3.3V source instead of GPIO 17 - but it is deliberately unimplemented until someone needs it. Ask if you do.

---

# S3 / S4 - Modbus / UART BMS

Supported bikes:
- Electrified S3 & X3 (VM01-{202, 203, 212, 213}-{EU, JP, UK, US}, VM01-201-2G EU, VM01-200-2G EU)
- Electrified S4

Battery model: **VM13-147** from DynaPack, Taiwan.
Interface: **Modbus RTU over UART RS232** at 9600 baud.

For everything in this section you need to remove the battery from the frame and connect it directly to your PC. A Raspberry Pi works well.

## Ports

### External Port

```
-----------------------------
\ TEST | DET | TX | KEY_IN /
 \  FAULT  |  GND  |  RX  /
  \     CHG+  |  CHG-    /
   \    DSG-  |  DSG+   /
    --------------------
```

### SWD Port

Located on the PCB. Open the battery with a + screwdriver, break the glue on the external port, then slide out the cell pack with the BMS PCB on top. The SWD port is at the near end of the PCB - no need to fully disassemble.

- VCC (3.3V)
- DIO
- CLK
- RST
- GND

## Connect via UART

Open minicom or PuTTY at 9600 baud, 8N1.

On Raspberry Pi: enable UART via `raspi-config` -> Interface Options -> Serial Port, then reboot.
Connect RX->TX, TX->RX, TEST->GND, GND->GND.

With TEST grounded the BMS prints a startup message:
```
I am G5 VanMoof BL V004 2019-11-19
I am VanMoof BL V007 2022-11-04 09:32:30
```
Only proceed if you see this message.

## Usage

```console
./bms --serial-port /dev/serial0 --action <action>
```

## Actions

| Action | Description |
|--------|-------------|
| `show` | Read and display all BMS registers (default) |
| `live` | Continuously read and display passive registers |
| `calibrateCHG` | Calibrate charge current (requires `--calibrate-current`) |
| `calibrateDSG` | Calibrate discharge current (requires `--calibrate-current`) |
| `chargeOn` | Enable charge MOSFET (register 0x1A=1) |
| `chargeOff` | Disable charge MOSFET (register 0x1A=0) |
| `clearLog` | Clear the BMS log via serial command |
| `clearPF` | Clear Power Failure via serial command |
| `convertLog` | Convert a BMS customer log text file to CSV (requires `--log-input`) |
| `detectOn` | Enable detect pin (IO2=1) |
| `detectOff` | Disable detect pin (IO2=0) |
| `discharge` | Enable discharging (register 0x08=1) |
| `dischargeoff` | Disable discharging (register 0x08=0) |
| `exportLog` | Export 100 BMS log entries to CSV file |
| `gpioOn` | Enable charge port GPIO (PF2=1) |
| `gpioOff` | Disable charge port GPIO (PF2=0) |
| `keyInOn` | Enable key input pin (IO1=1) |
| `keyInOff` | Disable key input pin (IO1=0) |
| `debug` | Enable BMS debug mode (register 0x09=1) |
| `debugoff` | Disable BMS debug mode (register 0x09=0) |
| `resetBMS` | Factory reset the BMS (removes ESN, calibration, cycles) |
| `resetESN` | Clear the Electronic Serial Number via serial command |
| `resetESNModbus` | Clear the Electronic Serial Number via Modbus (register 0x0A=0) |
| `resetMCU` | Reset the BMS microcontroller (register 0x80=0) |
| `ship` | Ship mode: disable battery output and discharge |
| `shipMode` | Ship mode only: disable battery output (register 0x01=0) |
| `showPorts` | List available serial ports |
| `updateFirmware` | Flash firmware .bin file to BMS via Modbus (requires `--firmware-file`) |
| `writeESN` | Write ESN and manufacture date (registers 0x0C-0x14) |

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--serial-port` | `/dev/serial0` | Serial device path |
| `--debug` | `false` | Enable debug output |
| `--loop` | `false` | Retry connection indefinitely |
| `--overview` | `false` | Show essentials only and exit |
| `--calibrate-current` | | Current in mA for `calibrateDSG` / `calibrateCHG` |
| `--log-file` | | Output CSV file path for `exportLog` |
| `--log-input` | | Input text file path for `convertLog` |
| `--esn` | | ESN (14 characters) for `writeESN` |
| `--esn-date` | | Manufacture date as YYYYMMDD for `writeESN` |
| `--firmware-file` | | Firmware .bin file path for `updateFirmware` |

## UART Commands

These are typed directly into the UART shell after the BMS has started (TEST pin must be grounded).

### Identify Firmware Version

```console
I am G5 VanMoof BL V004 2019-11-19
```

### Read Logs

Logs are read via Modbus using the `exportLog` action, which reads up to 100 entries
by writing register 0x0F45 with the log index and reading back the data.

### Clear Logs

```console
Log Clear
```

### Clear Power Failure

> **Warning:** Only do this if the cells are in good physical condition with no leaks, deformation, or burn marks. Clearing PF does not fix the underlying cause - it only resets the software flag.

<details>
    <summary><b>I'll be careful, i promise! </b></summary>
    Click into the Window and Write

```console
PF=0
```
</details>

The BMS replies with "OK" and reboots. After the startup message reappears, PF is cleared. Verify by checking for full pack voltage across the discharge ports when TEST is grounded.

You can also clear PF via SWD by setting address `0x08080001` to `0x03` in EEPROM.

### Clear Serial Number

```console
Reset ESN
```

Returns "Done" on success or "Reset ESN fail" on failure.

### Calibrate Discharge Current

```console
DSG CAL=<mA>
```

### Calibrate Charge Current

```console
CHG CAL=<mA>
```

### Reset BMS (untested)

Removes the serial number, calibration, and charge cycle count.

```console
Reset BMS V0106
```

### GPIO / Pin Control

| Command | Effect |
|---------|--------|
| `GPIO.IO2=1.` | DetectPin ON |
| `GPIO.IO2=0.` | DetectPin OFF |
| `GPIO.IO1=1.` | KeyIn ON |
| `GPIO.IO1=0.` | KeyIn OFF |
| `GPIO.PF2=1.` | Charge port GPIO ON |
| `GPIO.PF2=0.` | Charge port GPIO OFF |

### Firmware Update

Firmware can be updated via Modbus (`updateFirmware` action) or via SWD.

---

# A5 / S5 - CAN Bus BMS

Supported bikes:
- VanMoof A5 / S5 and later

Interface: **CAN bus only** at 1 Mbps, 29-bit extended frames.
There is **no Modbus or UART interface** on this pack.

> **Warning:** The battery defaults to Protection Failure (PF) mode if any cell spread exceeds 250 mV. This is expected for an imbalanced pack. Do **not** attempt to clear PF without first confirming the cells are safe.

> **Linux only.** The CAN path uses SocketCAN + the Linux GPIO character device and does not run on macOS (see [Platform support](#platform-support)).

## Hardware Setup

- **Raspberry Pi 4** (or similar Linux SBC)
- **CANable 2** USB CAN adapter in `gs_usb` mode -> `can0`, or `slcand` on `/dev/ttyACM0`
- **GPIO 17** = INT/WAKE line - must be held **HIGH continuously** while the BMS is active (not just pulsed)

The A5/S5 connector carries CAN H/L, 12V supply, GPIO wake (INT), and GND. GPIO 17 on the Pi connects to the INT pin.

### Bring up can0

```console
# gs_usb mode (CANable 2 default):
sudo ip link set can0 type can bitrate 1000000
sudo ip link set can0 up

# slcand fallback (adapter appears as /dev/ttyACM0):
sudo slcand -o -s8 -t hw /dev/ttyACM0 can0
sudo ip link set can0 up
```

The `can` action does this automatically if `can0` is absent and `/dev/ttyACM0` is present.

## Usage

```console
# Live monitor (Ctrl+C to stop)
./bms --action can [--can-iface can0]

# Read values once and exit
./bms --action can --can-snapshot [--can-iface can0]

# Send a command
./bms --action canUnlockPF [--can-iface can0]
```

The monitor will:
1. Assert GPIO 17 HIGH via the Linux GPIO character device (ABI v1)
2. Bring up `can0` via `slcand` if needed
3. Open a raw SocketCAN socket
4. Send the BMS wake sequence
5. Decode all broadcast frames and display a live-updating screen
6. Release GPIO 17 on Ctrl+C

## Actions

### Monitoring

| Action | Description |
|--------|-------------|
| `can` | Live CAN monitor - refreshes every 500 ms until Ctrl+C |
| `can` + `--can-snapshot` | Read all core values once and exit |

### DynaPack Commands

All DynaPack commands wake the BMS first, send the command, then release GPIO and exit.

| Action | Description |
|--------|-------------|
| `canDischarge` | Enable discharge MOSFET |
| `canDischargeOff` | Disable discharge MOSFET |
| `canChargeOn` | Enable charge MOSFET |
| `canChargeOff` | Disable charge MOSFET |
| `canUnlockPF` | Temporarily unlock Protection Failure mode |
| `canClearLog` | Erase the BMS error log |
| `canResetBMS` | Reset BMS controller (removes ESN, calibration, cycles) |
| `canCoulombCheck` | Trigger coulomb counter verification |
| `canSetRTC` | Set BMS real-time clock from system time |
| `canReadRTC` | Read and print BMS real-time clock |
| `canSetSN` | Write 13-character serial number (requires `--esn`) |
| `canReadChargeOffset` | Read charge current calibration offset |
| `canReadDischargeOffset` | Read discharge current calibration offset |
| `canReadADCOffset` | Read charge ADC current calibration offset |
| `canReadVoltageOffset` | Read charge voltage calibration offset |

### SetStatus Commands

| Action | Description |
|--------|-------------|
| `canNormalMode` | Set normal operating mode |
| `canChargeMode` | Set charge mode |
| `canSleep` | Set sleep / shipping mode |
| `canStandby` | Set standby mode |
| `canStatusChargeOn` | Normal mode + charge MOSFET open |
| `canStatusChargeOff` | Normal mode + charge MOSFET closed |

### Feature Control Commands

| Action | Description |
|--------|-------------|
| `canEnableBalance` | Enable cell balancing |
| `canDisableBalance` | Disable cell balancing |
| `canEnablePDSCP` | Enable power-down SCP protection |
| `canDisablePDSCP` | Disable power-down SCP protection |
| `canEnableCellOffline` | Enable cell offline detection |
| `canDisableCellOffline` | Disable cell offline detection |
| `canResetSystem` | Reset BMS system functions |

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--can-iface` | `can0` | SocketCAN interface name |
| `--can-snapshot` | `false` | Read values once and exit (with `can` action) |
| `--esn` | | 13-character serial number for `canSetSN` |

---

## CAN Protocol Reference

### Wake Sequence

The BMS must be woken by impersonating the `power_control` node (PFSA=0xA4):

| CAN ID | DLC | Data | Purpose |
|--------|-----|------|---------|
| `0x01111460` | 4 | `01 00 00 00` | PC heartbeat - repeat every 250 ms |
| `0x14901470` | 0 | (none) | AP version request - triggers BMS broadcast burst |
| `0x14803470` | 2 | `40 00` | SetStatus / RequestBMSInfo |

### Broadcast Frames (BMS -> power_control, ~250 ms)

| CAN ID | Bytes | Description |
|--------|-------|-------------|
| `0x14807460` | 0: status flags (`0x80` = PF active), 1: mode | Protection / mode status |
| `0x14809460` | 4: sub-state counter | Standby sub-state |
| `0x1480B460` | 0: SoC %, 4-5: signed LE int16 x 10 mA | State of charge + current |
| `0x1480D460` | 0-1: pack mV LE, 2-3: cell min mV LE, 4-5: cell max mV LE | Voltages |
| `0x14811460` | 0-3: temp raw bytes, 4: rolling counter, 7: SoC cross-check | Temperatures |
| `0x14815460` | 0-1: design voltage Q8.8 / 256 V LE, 4: rated cap Ah | One-shot at startup |
| `0x14901460` | 0: hw_rev, 1-3: fw_ver BCD | FW version reply |
| `0x1490F460` | ASCII bytes 0-7 of serial number | Serial number (part 1) |
| `0x1490F461` | ASCII bytes 8-15 of serial number | Serial number (part 2) |

### Pack Specification

- **13S** pack, nominal **47.97 V** (13 x 3.69 V/cell)
- Bytes 2-3 of `0x1480D460` = weakest cell voltage across all 13 cells (mV)
- Bytes 4-5 of `0x1480D460` = strongest cell voltage across all 13 cells (mV)
- PF shutdown triggers at > 250 mV spread between min and max cell
- CAN only reports min and max - individual cell voltages are not broadcast

### Temperature Decoding

The four bytes in `0x14811460` bytes 0-3 encode temperature with a Kelvin offset:

```
degC = raw_byte - 40
```

Empirically confirmed: raw 58-59 at ~18 degC room temperature.

Sensor mapping: [0] Cell Temp 1, [1] Cell Temp 2, [2] Charge Temp, [3] Discharge MOS Temp.

### DynaPack Proprietary Commands (ID `0x1082FF00`)

Always send the ASCII string `"DynaPack"` (8 bytes) on the same ID first, wait 5 ms, then send the command. Replies arrive on `0x1082FF00`.

| `byte[0]` | Function | Notes |
|-----------|----------|-------|
| `0x00` | SetRTC | bytes 1-6: ss, mm, hh, dd, MM, yy |
| `0x0A` | SetCellTemp1Offset | byte 1: signed offset (-128 to 127) |
| `0x0B` | SetCellTemp2Offset | byte 1: signed offset |
| `0x0C` | SetChargeTempOffset | byte 1: signed offset |
| `0x0E` + `byte[1]=1` | **UnlockPF** | Temporarily allows commands in PF mode |
| `0x0F` | ReadChargeCurrentOffset | reply bytes 0-1: int16 LE |
| `0x10` | ReadDischargeCurrentOffset | reply bytes 0-1: int16 LE |
| `0x12` | ReadChargeADCCurrentOffset | reply bytes 0-1: uint16 LE |
| `0x13` | ReadRTC | reply bytes 1-6: ss/mm/hh/dd/MM/yy |
| `0x14` + `byte[1]=0/1` | TurnDischarge Off/On | |
| `0x15` + `byte[1]=0/1` | TurnCharge Off/On | |
| `0x1A` | ReadLogCount | reply bytes 0-1: uint16 entry count |
| `0x1B` + bytes 1-7 | SetSN part 1 | first 7 chars of 13-char serial number |
| `0x1C` + bytes 1-6 | SetSN part 2 | last 6 chars of serial number |
| `0x21` | ReadChargeVoltageOffset | reply bytes 0-1: int16 LE |
| `0x22` | ClearErrorLog | write, no reply |
| `0x23` | CoulombCounterCheck | write, no reply |
| `0xFF` | ResetBMS | write, no reply |

### Feature Control (ID `0x000002FF`)

No `"DynaPack"` prefix needed.

| `byte[0]` | Function |
|-----------|----------|
| `0x02` | DisablePDSCP |
| `0x03` | EnablePDSCP |
| `0x04` | DisableCellOffline |
| `0x05` | EnableCellOffline |
| `0x06` | DisableCellBalance |
| `0x07` | EnableCellBalance |
| `0xFE` | ResetSystemFunction |

### Log Reading

The PBU tool supports reading the BMS event log over CAN. The protocol is:

1. Send DynaPack prefix + cmd `0x1A` on `0x1082FF00` to request the log entry count.
   Reply: uint16 LE at bytes 0-1 of the response frame.
2. For each entry (index 1 to count), send DynaPack prefix + cmd `0x19` + LE uint16 index.
   The log entry response arrives on a device-specific CAN ID defined in the encrypted
   device configuration file - this ID has not yet been decoded for the A5/S5.

Reading logs is not yet implemented in this tool.

For the S3/S4 Modbus BMS, use `--action exportLog` which reads up to 100 entries via
Modbus register 0x0F45.
