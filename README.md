# VanMooof BMS

## Connect to the BMS via UART

Open minicom or Putty with a Baudrate ot 9600 bits.

Click into the window. You must have TEST connected to Ground. You can check that the connection works by removing and reconnecting TEST to Ground. It will display a Message with "I am G* VanMoof **Version** **Build Date**". Only proceed if this happens.


## Clear Power Failure

To clear *ANY* Power Failure if the Tool displays a BMS Shutdown.

Make sure that the Cells are in good shape, the Battery does not have *ANY* Leaks, deformation or burn marks on or IN it!
If the Battery is leaking or have burn markings make sure you give it to someone who knows how to recycle it properly.
If you clear the Power Failure the BMS seems to just reboot and not check for problems again. So be sure that the Battery is in a good shape!

Click into the Window and Write

```
PF=0
```

The BMS will give you a Message with "OK" and it will reboot. After you see the Startup Message, the Power Failure should now be reset. But only on the Software side. If the Fuse is burnt or something else is off then the Bike will still show Errors. You can also test if the BMS would output Electricity by shorting TEST to Ground and Shorting the Fuse. If you get the full pack Voltage the BMS Error is cleared. The bike might still show error 19 if the 0 Ohm ressistors or the capacitors near the LSI Chip are broken / burnt. 

If you fixed the BMS correctly. You MUST have the pack Voltage accross both Discharge Ports when you Short TEST to GND. If there is no Pack Voltage on the Discharge Port then the Battery is still not fixed.

You can also fix the BMS Error State by conencting to the SWD Port on the internal side of the 


## Clear Logs

```
Log Clear
```

## Clear Serial Number

```
Reset ESN
```

It will Display Done if success, Reset ESN fail if the command failed


## Calibrate Discharge Current

x in mAh
```
DSG CAL=x
```

## Calibrate Charge Current

x in mAh
```
CHG CAL=x
```

# Reset BMS (untested!)

This resets the BMS. This removes the Serial Number, any calibration and the Charge Cycles. As far as i know. 


```
Reset BMS V0106
```
