package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/spf13/viper"
)

// Parses the toml config for printer host and ports, creates printer objects,
// and stores Printer pointers in array
func instantiateAllPrinters() {
	printers := viper.GetStringMap("printers")

	if len(printers) == 0 {
		fmt.Println("No printers in config")
		os.Exit(1)
	}

	for i := range printers {
		printer_host := "printers." + i + ".host"
		printer_port := "printers." + i + ".port"

		// fmt.Printf("%s, %s\n", viper.GetString(printer_host), viper.GetString(printer_port))

		p := NewPrinter(viper.GetString(printer_host), viper.GetString(printer_port))

		printerArray = append(printerArray, *p)
	}
}

// Have printers call method to update their status
func updatePrinterStatus() {
	for i := range printerArray {
		printerArray[i].RequestPrintStatus()
	}
}

// Given a GcodeFile, return a printer to handle it
func findPrinterToHandleFile(gcode GcodeFile) Print {
	for {
		updatePrinterStatus()
		for i := range printerArray {
			printer := *printerArray[i]
			if (printer.LastUsedColor == gcode.Filament.Color &&
				printer.LastUsedMaterial == gcode.Filament.Material) &&
				printer.GetStatus() == Standby {
				return printer
			}
		}
		for i := range printerArray {
			printer := *printerArray[i]
			if printer.GetStatus() == Standby {
				return printer
			}
		}
	}
}

// Spins off a thread for a printer method to handle a file. Update that
// file's status in the database
func assignFileToPrinter(printer Print, gcode GcodeFile, ctx context.Context, client *firestore.Client) {
	go printer.HandlePrintRequest(gcode, ctx, client)
	gcode.SetStatus(GcodePrinting)
	UpdateFileStatus(gcode, ctx, client)
}

func managePrintJobs(ctx context.Context, client *firestore.Client) {

	// Just keep looping until a GF is in queue
	for range time.Tick(time.Second * 10) {

		// if gcodeQueue is empty
		if len(gcodeQueue) == 0 {
			continue
		}

		gcode := popFromGcodeQueue()
		printer := findPrinterToHandleFile(gcode)
		assignFileToPrinter(printer, gcode, ctx, client)

	}
}

// pops gcodeFile from global GcodeFile queue
func popFromGcodeQueue() GcodeFile {
	var m sync.Mutex

	m.Lock()
	gcode := gcodeQueue[0]
	gcodeQueue = gcodeQueue[1:]
	m.Unlock()

	return gcode
}
