# VanMooof BMS

This is **ONLY** for the VanMoof SX3 / S4 
Electrified S3 (VM13-144), Electrified X3, Electrified S4

with the Battery Model Name / Product Code: VM13-147 from DynaPack from Taiwan.

It uses ModBus via UART RS232. The VanMoof Electrified S5 uses CANBus. If you want to integrate that, have fun. You can contact me if you have a spare battery to play with. 

For everything in here you need to remove the Battery from the Frame and be able to connect it to your PC. I suggest a raspberry pi.

## Connect to the BMS via UART

Open minicom or Putty with a Baudrate ot 9600 bits.

On the Raspberry Pi you have to enable the UART Pins via raspi-config, Interface Options, Serial Port, Enter, Enter, Reboot. 
Connect the UART Pins with Cables to the BMS. RX to TX and TX to RX, TEST to Ground, Ground to Ground. 

Click into the window. You must have TEST connected to Ground. You can check that the connection works by removing and reconnecting TEST to Ground. It will display a Message with "I am **G?** VanMoof **Version** **Build Date** **Build Time**". Only proceed if this happens.

## FAQ

### Can i update the BMS?

In theory yes, but this is another Level. 

## UART Commands

### Clear Power Failure

To clear *ANY* Power Failure if the Tool displays a BMS Shutdown.

- Make sure that the Cells are in good shape, the Battery does not have *ANY* Leaks, deformation or burn marks on or IN it!
- If the Battery is leaking or have burn markings make sure you give it to someone who knows how to recycle it properly.
- If the BMS triggered the Heating Element in the Fuse. This has a reason. **UNDER NO CIRCUMSTANCES YOU SHOULD SHORT ANY FUSE**. If you do you shall not have the privilage to use electicity!
- Fuses are good and they have a Reason to exist!
- If you clear the Power Failure the BMS seems to just reboot and not check for problems again. So be sure that the Battery is in a good shape!

<details>
    <summary><b>I'll be careful, i promise! </b></summary> 
    Click into the Window and Write

```console
PF=0
```
</details>


The BMS will give you a Message with "OK" and it will reboot. After you see the Startup Message, the Power Failure should now be reset. But only on the Software side. If the Fuse is burnt or something else is off then the Bike will still show Errors. You can also test if the BMS would output Electricity by shorting TEST to Ground and shorting the Fuse with a wire. Yes, this is fine. There is no Load connected, just your Multimeter. If you get the full pack Voltage the BMS Error is cleared. The bike might still show error 19 if the 0 Ohm ressistors or the capacitors near the LSI Chip are broken / burnt. 

If you fixed the BMS correctly. You *MUST* have the full pack Voltage across both Discharge Ports when you Short TEST to GND. If there is no Pack Voltage on the Discharge Port then the Battery is still not fixed.

You can also fix the BMS Error State by conencting to the SWD Port on the internal side of the PCB and set the Value 0x08080001 to "03" in the EEPROM. This is the same as setting it via the UART Console. 


### Clear Logs

```console
Log Clear
```

### Clear Serial Number

```console
Reset ESN
```

It will Display Done if success, Reset ESN fail if the command failed


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
