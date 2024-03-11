/*
Experiment to control the Raspberry Pi through go-rpio.

Blinks a light.
*/
package main

import (
	"fmt"
	"github.com/mikew686/rpidemo/internal/rpitest"
	"github.com/stianeikeland/go-rpio/v4"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func multi(stopall chan int) {
	q1 := make(chan int)
	d1 := make(chan int)
	fmt.Println("Starting blink")
	go rpitest.Blink(26, 200*time.Millisecond, q1, d1)

	for {
		select {
		case <-stopall:
			fmt.Println("Shutting blink")
			q1 <- 0
			<-d1
			return
		default:
		}
		time.Sleep(300 * time.Millisecond)
	}
}

func main() {
	fmt.Println("Opening device")

	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		fmt.Println("Closing device")
		rpio.Close()
	}()

	quit := make(chan int)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigc
		fmt.Println("shutdown signal:", sig)
		quit <- 0
	}()

	multi(quit)
}
