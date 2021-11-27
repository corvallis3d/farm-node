package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
)

const GcodePrintSuccess = 2
const GcodePrinting = 1
const GcodeIdle = 0
const GcodeError = 9

const JobIdle = 0
const JobInProgress = 1
const JobCompleted = 2

var jobs = []Job{}
var printerArray []Print

var gcodeQueue []GcodeFile

// jobsSnapshot keeps our local global jobs in sync with firebase
func jobsSnapshot(ctx context.Context, client *firestore.Client) {

	// Get all our order documents
	snapIter := client.Collection("jobs").Snapshots(ctx)
	defer snapIter.Stop()

	// Block our thread to never return
	for {
		// Get our current array
		jobDocs := jobs

		// Prepare our snapshot changes
		snap, err := snapIter.Next()
		if err != nil {
			fmt.Println(err)
		}

		// Grab what document changes have occurred
		docChanges := snap.Changes
		for index := 0; index < len(docChanges); index++ {
			// Get the current document change
			docChange := docChanges[index]
			var orderDocument Job

			// Cast our object into something we can work with
			err = docChange.Doc.DataTo(&orderDocument)
			if err != nil {
				fmt.Println(err)
			}

			// Give the Job its document ID from the database
			orderDocument.JobId = docChange.Doc.Ref.ID
			// Give each of the Gcode files the ID of its Job
			// Give each of the Gcode files their index
			for i := range orderDocument.GcodeFiles {
				orderDocument.GcodeFiles[i].JobId = docChange.Doc.Ref.ID
				orderDocument.GcodeFiles[i].FileIndex = i
			}

			// Check each possible case of the changes that could occur
			switch change := docChange.Kind; change {
			case firestore.DocumentAdded:
				// Document has been added to our array
				fmt.Println("Job Document has been added")
				// Append our current document in our array
				jobDocs = append(jobDocs, orderDocument)
				// Do something when document is added?

				// fmt.Println(orderDocument)

				// Put all the Gcode files into gcodeQueue
				for i := range orderDocument.GcodeFiles {
					gcodeQueue = append(gcodeQueue, orderDocument.GcodeFiles[i])
				}
			case firestore.DocumentModified:
				// Document has been modified
				fmt.Println("Job Document has been modified")
				// Modify that element in our local array, orderDocument = document that was just modified
				jobDocs[docChange.NewIndex] = orderDocument
				fmt.Println(orderDocument)
				// Do something when document is modified?
			case firestore.DocumentRemoved:
				// Document has been removed
				fmt.Println("Job Document has been removed")
				// Remove the document from our local array
				jobDocs = append(jobDocs[:docChange.OldIndex], jobDocs[docChange.OldIndex+1:]...)
				// Do something when document is removed?
			default:
				fmt.Println("Default Job Document changes called")
			}
		}
		// Save to our global variable
		jobs = jobDocs
	}
}

// FirebaseInstance obtains the client and ctx when needed
func FirebaseInstance() (*firestore.Client, context.Context, error) {

	// Get static variables for setting up the firestore
	var opt = option.WithCredentialsFile(viper.GetString("database.path"))
	var config = &firebase.Config{ProjectID: viper.GetString("database.projectId")}

	// Setup the FireStore data
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, ctx, err
	}

	return client, ctx, nil
}

// Parses the toml config for printer host and ports, creates printer objects,
// and stores Printer pointers in array
func instantiateAllPrinters() {
	printers := viper.GetStringMap("printers")

	for i := range printers {
		printer_host := "printers." + i + ".host"
		printer_port := "printers." + i + ".port"

		// fmt.Printf("%s, %s\n", viper.GetString(printer_host), viper.GetString(printer_port))

		p := NewPrinter(viper.GetString(printer_host), viper.GetString(printer_port))
		p.Connect()
		p.Start_receive_thread()

		printerArray = append(printerArray, *p)
	}
}

// Have printers call method to update their status
func updatePrinterStatus() {
	for i, p := range printerArray {
		fmt.Println(i, p)
	}
}

// Update status for a single Gcode file in database
// Copies whole doc, modifes, then replaces
func updateFileStatus(gcode GcodeFile, ctx context.Context, client *firestore.Client) {
	jobId := gcode.JobId
	fileIndex := gcode.FileIndex

	job := client.Doc(fmt.Sprintf("jobs/%s", jobId))

	var jobDocument Job

	docsnap, err := job.Get(ctx)
	if err != nil {
		fmt.Println(err)
	}

	err = docsnap.DataTo(&jobDocument)
	if err != nil {
		fmt.Println(err)
	}

	// fmt.Println(jobDocument)

	// Modify the job doc
	jobDocument.JobId = jobId
	jobDocument.GcodeFiles[fileIndex].FileIndex = fileIndex
	jobDocument.GcodeFiles[fileIndex].Status = gcode.Status
	jobDocument.GcodeFiles[fileIndex].JobId = jobId

	wr, err := job.Set(ctx, jobDocument)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(wr.UpdateTime)
}

func managePrintJobs(ctx context.Context, client *firestore.Client) {
	// Just keep looping
	for range time.Tick(time.Second * 10) {
		// if gcodeQueue is empty
		if len(gcodeQueue) == 0 {
			continue
		}

		// First Gcode file popped from queue
		gcode := gcodeQueue[0]
		gcodeQueue = gcodeQueue[1:]

		updatePrinterStatus()

		// We need to loop through the printers array to see if any can handle the file
		for i := range printerArray {
			// A printer can handle the file if the printer's status is idle,
			// the file's max_dim does not exceed printer dimensions, and the
			// filament matches the filament in the gcodeFile
			if printerArray[i].color == gcode.Filament.Color &&
				printerArray[i].filament == gcode.Filament.Material {

				printerArray[i].HandlePrintRequest(gcode)

				// Set gcode file status to 2 locally
				gcode.Status = 2

				// Set file status code to 2 in database, we need the JobId,
				// and the FileIndex. Pass as argument local gcode file
				updateFileStatus(gcode, ctx, client)
			}

			// If none of the printers can handle the file, but there is an idle,
			// printer, call sendFilamentRequest() for that printer and wait for
			// response

			// If none can handle, just reloop again.
		}
		// FIFO approach. Blocks at last popped file until it can be handled.
	}
}
