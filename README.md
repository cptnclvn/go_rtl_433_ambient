go_rtl_433_ambient
================

Send your RTL_433 captured data to InfluxDB.
This captures the JSON output of the rtl_433 and sends it to InfluxDB.
Pass the IDs with an associated name.
```
go_rtl_r433_ambient --namedFields 150=LivingRoom,120=Kitchen

```

This can easily be modified to to use the other devices that the rtl_433 can listen to. 