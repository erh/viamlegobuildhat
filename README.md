

Links
=====
* https://www.raspberrypi.com/documentation/accessories/build-hat.html
* serial spec - https://datasheets.raspberrypi.com/build-hat/build-hat-serial-protocol.pdf?


Test Via Python
=====
```
sudo pip3 install buildhat
python3
>>> import buildhat
>>> motor = buildhat.Motor("D")
>>> motor.run_for_seconds(5)

```
