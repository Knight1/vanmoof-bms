# VanMooof BMS

This tool is **ONLY** for the following VanMoof bikes
- Electrified S3 & X3 (VM01-{202, 203, 212, 213}-{EU, JP, UK, US}, VM01-201-2G EU, VM01-200-2G EU)
- Electrified S4

with the Battery Model Name / Product Code: VM13-147  
from DynaPack from Taiwan.  

It uses ModBus via UART RS232. The VanMoof Electrified S5 and later uses CANBus to talk to the Battery. 

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
| `writeESN` | Write ESN and manufacture date (registers 0x0C-0x14) |
| `showPorts` | List available serial ports |

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--serial-port` | `/dev/serial0` | Serial device path |
| `--debug` | `false` | Enable debug output |
| `--loop` | `false` | Retry connection indefinitely |
| `--overview` | `false` | Show essentials only |
| `--calibrate-current` | | Current in mA for calibrateDSG / calibrateCHG |
| `--log-file` | | Output CSV file path for exportLog |
| `--log-input` | | Input text file path for convertLog |
| `--esn` | | ESN (14 characters) for writeESN |
| `--esn-date` | | Manufacture date as YYYYMMDD for writeESN |

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