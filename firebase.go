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

// Update status for a single Gcode file in database
// Copies whole doc, modifes, then replaces
func UpdateFileStatus(gcode GcodeFile, ctx context.Context, client *firestore.Client) {
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

//-----------------------------------------------------------------------------
// check local jobs array, scanning for gcode statuses if all gcode statuses
// are completed, or canceled then update firestore database by removing job
// document from jobs collection and into completed_jobs collection
// keep looping
func maintainFirestore(ctx context.Context, client *firestore.Client) {
	for range time.Tick(time.Minute * 1) {
		for i := range jobs {
			job := jobs[i]
			jobId := job.JobId
			fmt.Print(jobId)
			gcodeFiles := job.GcodeFiles
			count := 0
			for j := range gcodeFiles {
				gcodeFile := gcodeFiles[j]
				if gcodeFile.Status == GcodePrintSuccess || gcodeFile.Status == GcodeCanceled {
					count += 1
				}
			}
			// Move job from jobs collection to completed_jobs collection
			// Grabs job document pertaining to jobId
			document := client.Doc(fmt.Sprintf("jobs/%s", jobId))
			var jobDocument Job
			docsnap, err := document.Get(ctx)

			if err != nil {
				fmt.Println(err)
			}

			err = docsnap.DataTo(&jobDocument)
			if err != nil {
				fmt.Println(err)
			}

			//fmt.Println(jobDocument)
			// references completed_jobs collection to be copied into
			newDocument := client.Doc(fmt.Sprintf("completed_jobs/%s", jobId))
			wr, err := newDocument.Set(ctx, jobDocument)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(wr.UpdateTime)

			// wr, err = document.Delete(ctx)
			// if err != nil {
			// 	fmt.Println(err)
			// }
			// fmt.Println(wr.UpdateTime)
		}
	}
}
