package main

import (
	"time"
)

func main() {
	p := NewPrinter("10.0.0.147", "7125")
	p.Connect()
	p.Start_receive_thread()
	p.Change_printer_status("Test String")
	p.Send_print_file()
	for {
		time.Sleep(time.Second)
	}
}
