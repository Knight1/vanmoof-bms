# VanMooof BMS

This tool is **ONLY** for the following VanMoof bikes
- Electrified S3 & X3 (VM01-{202, 203, 212, 213}-{EU, JP, UK, US}, VM01-201-2G EU, VM01-200-2G EU)
- Electrified S4

with the Battery Model Name / Product Code: VM13-147  
from DynaPack from Taiwan.  

It uses ModBus via UART RS232. The VanMoof Electrified S5 and later uses CANBus to talk to the Battery.

---

# VanMoof A5 / S5 - DynaPack BMS (CAN bus)

The A5/S5 battery uses a **CAN-only** DynaPack BMS. There is **no Modbus / UART interface** on this pack. All communication is via 29-bit extended CAN frames at 1 Mbps.

> **Warning:** The battery defaults to Protection Failure (PF) mode if the cell spread exceeds 250 mV. This is expected for an imbalanced pack. Do **not** attempt to clear PF without first ensuring the cells are safe.

## Hardware Setup

- **Raspberry Pi 4** (or similar Linux SBC)
- **CANable 2** USB CAN adapter in `gs_usb` mode -> `can0`, or via `slcand` on `/dev/ttyACM0`
- **GPIO 17** = INT/WAKE line - must be held **HIGH continuously** while the BMS is active (not just pulsed)
- CAN bus: **1 Mbps**, 29-bit extended frames

### External Connector Pin-out

The A5/S5 battery connector carries CAN H/L, 12 V supply, GPIO wake line (INT), and ground. GPIO 17 on the Pi connects to the INT pin and must stay HIGH for the BMS to remain awake.

### Bring up `can0`

```console
# gs_usb mode (CANable 2 default):
sudo ip link set can0 type can bitrate 1000000
sudo ip link set can0 up

# slcand fallback (if can0 is absent and adapter appears as /dev/ttyACM0):
sudo slcand -o -s8 -t hw /dev/ttyACM0 can0
sudo ip link set can0 up
```

The `can` action does this automatically if `can0` is absent and `/dev/ttyACM0` is present.

## Usage

```console
./bms --action can [--can-iface can0]
```

The tool will:
1. Assert GPIO 17 HIGH via the Linux GPIO character device (ABI v1)
2. Bring up `can0` via `slcand` if needed
3. Open a raw SocketCAN socket
4. Send the BMS wake sequence
5. Decode all broadcast frames and display a live-updating screen
6. Release GPIO 17 on Ctrl+C

## CAN Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--can-iface` | `can0` | SocketCAN interface name |

## Wake Sequence

The BMS must be woken by impersonating the `power_control` node:

| CAN ID | DLC | Data | Purpose |
|--------|-----|------|---------|
| `0x01111460` | 4 | `01 00 00 00` | PC heartbeat - repeat every 250 ms |
| `0x14901470` | 0 | (none) | AP version request - triggers BMS broadcast burst |
| `0x14803470` | 2 | `40 00` | SetStatus / RequestBMSInfo |

## BMS Broadcast Frames (BMS -> power_control, ~250 ms interval)

All frames use 29-bit extended IDs. BMS source address encodes to `0x460`, destination to `0x470`.

| CAN ID | Bytes | Description |
|--------|-------|-------------|
| `0x14807460` | 0: status flags (`0x80` = PF active), 1: mode code | Protection / mode status |
| `0x14809460` | 4: sub-state counter | Standby sub-state |
| `0x1480B460` | 0: SoC %, 4-5: signed LE int16 x 10 mA (current) | State of charge + current |
| `0x1480D460` | 0-1: pack mV LE, 2-3: cell min mV LE, 4-5: cell max mV LE | Voltages |
| `0x14811460` | 0-3: temp raw bytes, 4: rolling counter, 7: SoC cross-check | Temperatures |
| `0x14815460` | 0-1: design voltage Q8.8 / 256 V LE, 4: rated cap Ah | One-shot at startup |
| `0x14901460` | 0: hw_rev, 1-3: fw_ver BCD | FW version reply |
| `0x1490F460` | ASCII bytes 0-7 of serial number string | Serial number (part 1) |
| `0x1490F461` | ASCII bytes 8-15 of serial number string | Serial number (part 2) |

### Pack Specification

- **13S** pack, nominal **47.97 V** (13 x 3.69 V/cell)
- `0x1480D460` bytes 2-3 = **weakest cell** across all 13 cells (mV)
- `0x1480D460` bytes 4-5 = **strongest cell** across all 13 cells (mV)
- **PF shutdown** triggers at > 250 mV spread between min and max cell
- CAN only reports the minimum and maximum cell - individual cell voltages are not broadcast

### Temperature Decoding

The four raw bytes in `0x14811460` bytes 0-3 encode temperature as:

```
degC = raw_byte - 40
```

This is a Kelvin-offset encoding with a -40 degC baseline (common in embedded BMS firmware). Empirically confirmed: raw value 58-59 at ~18 degC room temperature.

The four sensors correspond to: Cell Temp 1, Cell Temp 2, Charge Temp, Discharge MOS Temp.

## DynaPack Proprietary Commands (ID `0x1082FF00`)

Always send the ASCII string `"DynaPack"` (8 bytes) on the same ID first, wait 5 ms, then send the command frame. Replies come back on `0x1082FF00`.

| `byte[0]` | Function | Notes |
|-----------|----------|-------|
| `0x00` | SetRTC | bytes 1-6: ss, mm, hh, dd, MM, yy |
| `0x0A` | SetCellTemp1Offset | byte 1: signed offset (-128 to 127) |
| `0x0B` | SetCellTemp2Offset | byte 1: signed offset |
| `0x0C` | SetChargeTempOffset | byte 1: signed offset |
| `0x0E` + `byte[1]=1` | **UnlockPF** | Temporarily allows commands in PF mode |
| `0x0F` | ReadChargeCurrentOffset | reply: int16 LE |
| `0x10` | ReadDischargeCurrentOffset | reply: int16 LE |
| `0x12` | ReadChargeADCCurrentOffset | reply: int16 LE |
| `0x13` | ReadRTC | reply bytes 1-6: ss/mm/hh dd/MM/yy |
| `0x14` + `byte[1]=0/1` | TurnDischarge Off/On | |
| `0x15` + `byte[1]=0/1` | TurnCharge Off/On | |
| `0x21` | ReadChargeVoltageOffset | reply: int16 LE |
| `0x22` | ClearErrorLog | write, no reply |
| `0xFF` | ResetBMS | write, no reply |

## Feature Control (ID `0x000002FF`)

No `"DynaPack"` prefix needed.

| `byte[0]` | Function |
|-----------|----------|
| `2` | DisablePDSCP |
| `3` | EnablePDSCP |
| `4` | DisableCellOffLine |
| `5` | EnableCellOffLine |
| `6` | DisableCellBalance |
| `7` | EnableCellBalance |
| `254` | ResetSystemFunction |

---


For everything in here you need to remove the Battery from the Frame and be able to connect the battery to your PC. I suggest a Raspberry Pi.  

## Ports

### SWD Port

This Port is on the PCB. You need a + screwdriver to open the Battery. Break the glue to remove the external port. After that you can slide out the Cells with the BMS PCB on top of the cells.  
The SWD Port is right at the beginning of the PCB when you slide it out. No need to remove the cell package or the black plastic sheet protecting the PCB.  

- VCC (3.3Vdc)
- DIO
- CLK
- RST
- GND

### External Port

```aiignore
-----------------------------
\ TEST | DET | TX | KEY_IN /
 \  FAULT  |  GND  |  RX  /
  \     CHG+  |  CHG-    /
   \    DSG-  |  DSG+   /
    --------------------
```

## Connect to the BMS via UART

Open minicom or Putty with a Baudrate ot 9600 bits.

On the Raspberry Pi you have to enable the UART Pins via raspi-config, Interface Options, Serial Port, Enter, Enter, Reboot. 
Connect the UART Pins with Cables to the BMS. RX to TX and TX to RX, TEST to Ground, Ground to Ground. 

Click into the window. You must have TEST connected to Ground. You can check that the connection works by removing and reconnecting TEST to Ground. It will display a Message with "I am **G?** VanMoof **Version** **Build Date** **Build Time**". Only proceed if this happens.

## FAQ

### Can i update the BMS?

This is possible via modbus, the module does this. But it is fairly easy to update via SWD.

### Build

```console
go build -trimpath -buildmode=pie -mod=vendor -ldflags "-w -s" -v ./...
```

### Usage

```console
./bms --serial-port /dev/serial0 --action <action>
```

### Actions

| Action | Description |
|--------|-------------|
| `can` | Live CAN bus monitor for A5/S5 DynaPack BMS (no Modbus needed) |
| `show` | Read and display all BMS registers (default) |
| `live` | Continuously read and display passive registers |
| `calibrateCHG` | Calibrate charge current (requires `--calibrate-current`) |
| `calibrateDSG` | Calibrate discharge current (requires `--calibrate-current`) |
| `chargeOn` | Enable charge MOSFET (register 0x1A=1) |
| `chargeOff` | Disable charge MOSFET (register 0x1A=0) |
| `clearLog` | Clear the BMS log via serial command |
| `convertLog` | Convert a BMS customer log text file to CSV (requires `--log-input`) |
| `clearPF` | Clear Power Failure via serial command |
| `detectOn` | Enable detect pin (IO2=1) |
| `exportLog` | Export 100 BMS log entries to CSV file |
| `detectOff` | Disable detect pin (IO2=0) |
| `gpioOn` | Enable charge port GPIO (PF2=1) |
| `keyInOn` | Enable key input pin (IO1=1) |
| `keyInOff` | Disable key input pin (IO1=0) |
| `gpioOff` | Disable charge port GPIO (PF2=0) |
| `debug` | Enable BMS debug mode (register 0x09=1) |
| `debugoff` | Disable BMS debug mode (register 0x09=0) |
| `discharge` | Enable discharging (register 0x08=1) |
| `dischargeoff` | Disable discharging (register 0x08=0) |
| `resetBMS` | Factory reset the BMS (removes ESN, calibration, cycles) |
| `ship` | Ship mode: disable battery output and discharge |
| `shipMode` | Ship mode only: disable battery output (register 0x01=0) |
| `resetESN` | Clear the Electronic Serial Number via serial command |
| `resetESNModbus` | Clear the Electronic Serial Number via Modbus (register 0x0A=0) |
| `resetMCU` | Reset the BMS microcontroller (register 0x80=0) |
| `writeESN` | Write ESN and manufacture date (registers 0x0C-0x14) |
| `updateFirmware` | Flash firmware .bin file to BMS via Modbus (requires `--firmware-file`) |
| `showPorts` | List available serial ports |

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--can-iface` | `can0` | SocketCAN interface for `-action can` |
| `--serial-port` | `/dev/serial0` | Serial device path |
| `--debug` | `false` | Enable debug output |
| `--loop` | `false` | Retry connection indefinitely |
| `--overview` | `false` | Show essentials only |
| `--calibrate-current` | | Current in mA for calibrateDSG / calibrateCHG |
| `--log-file` | | Output CSV file path for exportLog |
| `--log-input` | | Input text file path for convertLog |
| `--esn` | | ESN (14 characters) for writeESN |
| `--esn-date` | | Manufacture date as YYYYMMDD for writeESN |
| `--firmware-file` | | Firmware .bin file path for updateFirmware |

## UART Commands

These Commands are written directly in to the UART Port (Shell) after the BMS started.

### Identify Battery Software Version without reading ModBus

```console
I am G5 VanMoof BL V004 2019-11-19
```
Version: V004  
Date: 2019-11-19

```console
I am VanMoof BL V007 2022-11-04 09:32:30
```
Version: V007  
Date: 2022-11-04 09:32:30

### Clear Power Failure

To clear *ANY* Power Failure if the Tool displays a BMS Shutdown.

- Make sure that the Cells are in good shape, the Battery does not have *ANY* Leaks, deformation or burn marks on or IN it!
- If the Battery is leaking or have burn markings make sure you give it to someone who knows how to recycle it properly.
- If the BMS triggered the Heating Element in the Fuse. This has a reason. **UNDER NO CIRCUMSTANCES SHOULD YOU SHORT ANY FUSE**. If you do you shall not have the privilege to use electricity!
- Fuses are good and they have a Reason to exist!
- If you clear the Power Failure the BMS seems to just reboot and not check for problems again. So be sure that the Battery is in a good shape!

<details>
    <summary><b>I'll be careful, i promise! </b></summary> 
    Click into the Window and Write

```console
PF=0
```
</details>


The BMS will give you a Message with "OK" and it will reboot. After you see the Startup Message, the Power Failure should now be reset. But only on the Software side. If the Fuse is burnt or something else is off then the Bike will still show Errors. You can also test if the BMS would output Electricity by shorting TEST to Ground and shorting the Fuse with a wire. Yes, this is fine. There is no Load connected, just your Multimeter. If you get the full pack Voltage the BMS Error is cleared. The bike might still show error 19 if the 0 Ohm resistors or the capacitors near the LSI Chip are broken / burnt. 

If you fixed the BMS correctly. You *MUST* have the full pack Voltage across both Discharge Ports when you Short TEST to GND. If there is no Pack Voltage on the Discharge Port then the Battery is still not fixed.

You can also fix the BMS Error State by connecting to the SWD Port on the internal side of the PCB and set the Value 0x08080001 to "03" in the EEPROM. This is the same as setting it via the UART Console.

### Clear Logs

```console
Log Clear
```

### Clear Serial Number

```console
Reset ESN
```

It will Display "Done" if success, "Reset ESN fail" if the command failed


### Calibrate Discharge Current

x in mAh

```console
DSG CAL=x
```

### Calibrate Charge Current

x in mAh

```console
CHG CAL=x
```

### Reset BMS (untested!)

This resets the BMS. This removes the Serial Number, any calibration and the Charge Cycles. As far as i know.

```console
Reset BMS V0106
```

### DetectPin (On/Off)

On
```console
GPIO.IO2=1.
```

Off
```console
GPIO.IO2=0.
```

### KeyIn (On/Off)

On
```console
GPIO.IO1=1.
```

Off
```console
GPIO.IO1=0.
```

### GPIO (On/Off)

On  
```console
GPIO.PF2=1.
```

Off  
```console
GPIO.PF2=0.
```