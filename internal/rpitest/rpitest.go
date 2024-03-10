package rpitest

import (
	"github.com/stianeikeland/go-rpio/v4"
	"time"
)

func Blink(pin_num int, dur time.Duration, quit, done chan int) {
	pin := rpio.Pin(pin_num)

	pin.Output()

	for {
		select {
		case <-quit:
			pin.Write(rpio.Low)
			done <- 0
			return
		default:
		}
		pin.Toggle()
		time.Sleep(dur)
	}
}
