package main

import (
	"time"
)

func main() {
	p := NewPrinter("10.0.0.147", "7125")
	p.Connect()
	p.Start_receive_thread()
	p.Change_printer_status("Test String")
	// p.Send_print_file()
	p.Enqueue_file()
	p.Resume_queue()
	time.Sleep(3 * time.Second)
	p.Pause_printer()

}
