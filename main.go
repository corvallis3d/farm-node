package main

import (
	"time"
)

func main() {
	p := NewPrinter("10.0.0.147", "7125")
	p.Connect()
	p.Start_receive_thread()
	for {
		time.Sleep(time.Second)
	}
}
