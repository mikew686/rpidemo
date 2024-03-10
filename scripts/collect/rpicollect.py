#!/bin/env python

import time
import signal
import sys
from datetime import datetime
import redis
import RPi.GPIO as GPIO
from freenove.dht11.Freenove_DHT import DHT
from waveshare.ADS1256 import ADS1256
from freenove.Adafruit_LCD1602.PCF8574 import PCF8574_GPIO
from freenove.Adafruit_LCD1602.Adafruit_LCD1602 import Adafruit_CharLCD

"""
Utility to collect stats on the demo platform. Every second:
* Toggles an led
* Collects temperature and humidity data from DHT11
* Collects light level from ADS1256 photoresistor
* Prints out some info on the LCD display
* If a redis server is available, writes out the timestamped sample to a stream "samples:current"
"""

class Blinker:
    def __init__(self):
        self._pin = 26 # gpio 26 board 37
        self._state = GPIO.LOW
    
    def init(self):
        print("Blinker init")
        GPIO.setup(self._pin, GPIO.OUT)
        GPIO.output(self._pin, self._state)

    def toggle(self):
        print("Blinker toggle")
        if self._state == GPIO.LOW:
            self._state = GPIO.HIGH
        else:
            self._state = GPIO.LOW
        GPIO.output(self._pin, self._state)

    def shut(self):
        print("Blinker shut")
        GPIO.output(self._pin, GPIO.LOW)


class Temperature:
    def __init__(self):
        self._pin = 13 # gpio 13 board 33
    
    def init(self):
        print("Temperature init")
        self.dht = DHT(self._pin)   #create a DHT class object

    def sample(self):
        print("Temperature sample")
        res = self.dht.readDHT11()
        if res != self.dht.DHTLIB_OK:
            print("DHT11 read failure")
        print("humidity=%.2f temperature=%.2f" % (self.dht.humidity, self.dht.temperature))
        return self.dht.humidity, self.dht.temperature

    def shut(self):
        print("Temperature shut")


class LightLevel:
    def __init__(self):
        self.adc = ADS1256() # board uses spi interface
    
    def init(self):
        print("LightLevel init")
        self.adc.ADS1256_init()

    def sample(self):
        print("LightLevel sample")
        voltage = self.adc.ADS1256_GetChannalValue(1)*5.0/0x7fffff # returns a range 0 to 5 volts, device 2
        level = (3.3-voltage)*100.0/3.3 # convert to percentage (0 is dark, 100 is maximum)
        print("voltage=%.2f level=%.2f" % (voltage, level))
        return level

    def shut(self):
        print("LightLevel shut")


class Display:
    def __init__(self):
        self.PCF8574_address = 0x27
        self.PCF8574A_address = 0x3F
    
    def init(self):
        print("Display init")
        try:
            self.mcp = PCF8574_GPIO(self.PCF8574_address)
        except:
            self.mcp = PCF8574_GPIO(self.PCF8574A_address)
        self.lcd = Adafruit_CharLCD(pin_rs=0, pin_e=2, pins_db=[4,5,6,7], GPIO=self.mcp)
        self.lcd.begin(16,2)     # set number of LCD lines and columns
        self.mcp.output(3,1)     # turn on LCD backlight

    def update(self, level, temp):
        print("Display update")
        self.lcd.clear()
        tstr = datetime.now().strftime("%m/%d %H:%M:%S")
        off = int((16-len(tstr))/2)
        self.lcd.setCursor(off, 0)
        self.lcd.message(tstr)
        line2 = "%.1f C  %d %%" % (temp, int(level))
        off = int((16-len(line2))/2)
        self.lcd.setCursor(off, 1)
        self.lcd.message(line2)

    def shut(self):
        print("Display shut")
        self.lcd.clear()
        self.mcp.output(3,0)


class Datastore:
    def __init__(self, host="localhost", port=6379):
        self.host = host
        self.port = port
        self.enabled = False
    
    def init(self):
        print("Datastore init")
        self.enabled = False
        self.redis = redis.Redis(self.host, self.port, decode_responses=True)
        try:
            test = self.redis.ping()
            if test:
                self.enabled = True
                print("Connected to redis on", self.host, self.port)
        except:
            print("Could not connect to redis on", self.host, self.port)

    def update(self, level, temp, humid):
        print("Datastore update")
        id = self.redis.xadd("samples:current", {"level":level,"temp":temp,"humid":humid})
        print(f"Added entry with id", id)

    def shut(self):
        print("Datastore shut")


blinker = Blinker()
temp = Temperature()
light = LightLevel()
display = Display()
datastore = Datastore()

def initialize():
    print("initialize")
    # global GPIO setup
    #GPIO.setmode(GPIO.BOARD) # this is pin mode
    GPIO.setmode(GPIO.BCM) # required by the Waveshare library
    blinker.init()
    temp.init()
    light.init()
    display.init()
    datastore.init()


def sample():
    blinker.toggle()
    h, t = temp.sample()
    l = light.sample()
    display.update(l, t)
    datastore.update(l, t, h)


def cleanup():
    print ("cleanup")
    blinker.shut()
    temp.shut()
    light.shut()
    display.shut()
    datastore.shut()
    # global GPIO shutdown
    time.sleep(0.5)
    GPIO.cleanup()

def sig_handler(_signo, _stack_frame):
    sys.exit(0)


if __name__ == "__main__":
    signal.signal(signal.SIGTERM, sig_handler)
    signal.signal(signal.SIGINT, sig_handler)
    try:
        initialize()
        t1 = int(time.time())
        while True:
            t2 = int(time.time())
            if t2 > t1:
                print("sample at", datetime.now().strftime("%H:%M:%S"))
                sample()
                t1 = int(time.time())
            time.sleep(0.1)
    except SystemExit:
        cleanup()
    except:
        ex_type, ex_value, traceback = sys.exc_info()
        print(ex_type, ex_value)
        cleanup()
