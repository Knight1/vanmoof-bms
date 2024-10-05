# VanMooof BMS

## Connect to the BMS via UART

Open minicom or Putty with a Baudrate ot 9600 bits.

click into the window. You must have TEST connected to Ground. You can check that the connection works by removing and reconnecting TEST to Ground. It will display a Message with "I am VanMoof G* Version". Only proceed if this happens.


## Clear Power Failure

to clear *ANY* Power Failure if the Tool displays a BMS Shutdown

Click into the Window and Write

```
PF=0
```

The BMS will give you a Message with "OK" and it will reboot. The PowerFailure is now reset.

If you fixed the BMS correctly. You MUST have the pack Voltage accross both Discharge Ports when you Short TEST to GND. If there is no Pack Voltage on the Discharge Port then the Battery is still not fixed.


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
