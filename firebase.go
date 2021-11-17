package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
)

var jobs = []Job{}
var stored_index = 0

// jobsSnapshot keeps our local global jobs in sync with firebase
func jobsSnapshot(ctx context.Context, client *firestore.Client, printerArray []Printer) {

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
			err = errors.New("Order Document Issue reading from the database")
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
				err = errors.New("Job Document Issue casting data to object")
				fmt.Println(err.Error())
			}

			// Check each possible case of the changes that could occur
			switch change := docChange.Kind; change {
			case firestore.DocumentAdded:
				// Document has been added to our array
				fmt.Println("Job Document has been added")
				// Append our current document in our array
				jobDocs = append(jobDocs, orderDocument)
				stored_index = index
				// Do something when document is added?
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
		fileName := jobs[0].GcodeFiles[0].Filename
		fmt.Println(fileName)

		//uploadAFile()
		// call printer selector function
		p.startPrintJob(fileName)
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

func uploadAFile() {
	fileName := jobs[len(jobs)-1].GcodeFiles[0].Filename
	fmt.Println(fileName)
	url := viper.GetString("moonraker.baseUrl") + "server/files/upload"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, errFile1 := os.Open(fmt.Sprintf("./gcode/%s", fileName))
	defer file.Close()
	part1,
		errFile1 := writer.CreateFormFile("file", filepath.Base(fmt.Sprintf("./gcode/%s", fileName)))
	_, errFile1 = io.Copy(part1, file)
	if errFile1 != nil {
		fmt.Println(errFile1)
		return
	}
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
