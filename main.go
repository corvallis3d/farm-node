package main

import (
	"fmt"
	"time"
)

func main() {
	p := NewPrinter("10.0.0.147", "7125")
	// p := NewPrinter("192.168.1.187", "7125")
	p.Connect()
	p.Start_receive_thread()
	time.Sleep(time.Second)

	// file name received from Golang application:
	file_name := "testing.gcode"

	// Send file name notification
	p.Change_notification_string(file_name)
	p.Send_print_notification()

	for {
		// Check if user accepted print order throug LCD user input
		if p.print_flag {
			fmt.Print("############# Print order Accepted") // terminal print
			p.Upload_file(file_name)
			p.Start_print()
			//GCode macro removes notification and set display to default
		} else {
			p.Get_printer_status()
			time.Sleep(time.Second)
		}
	}
}
