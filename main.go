package main

import (
	"fmt"
	"time"
)

func main() {
	// p := NewPrinter("10.0.0.147", "7125")
	p := NewPrinter("192.168.1.187", "7125")
	p.Connect()
	p.Start_receive_thread()
	time.Sleep(1 * time.Second)
	fmt.Println("########STATUS")
	p.Change_printer_status("Test String")
	time.Sleep(1 * time.Second)
	// p.Send_print_file()
	// fmt.Println("########ENQUEUE")
	// p.Enqueue_file()
	// time.Sleep(1 * time.Second)
	// p.Resume_queue()
	// fmt.Println("########QUEUE STATUS")
	// p.Queue_state()
	// time.Sleep(3 * time.Second)
	// p.Pause_printer()

}
