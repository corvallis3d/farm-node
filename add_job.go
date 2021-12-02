package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"cloud.google.com/go/firestore"
)

var letters = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func addFalseDocumentToJobsCollection(ctx context.Context, client *firestore.Client) {
	// Properly generate random string
	rand.Seed(time.Now().UnixNano())

	// Refer to "jobs" collection
	jobCollection := client.Collection("jobs")
	// Generate 20 character jobId string to write to
	newDocument := jobCollection.Doc(generateString(20))

	// Write default values to doc
	doc := make(map[string]interface{})
	doc["gcode"] = []interface{}{map[string]interface{}{
		"filament": map[string]interface{}{
			"color":    "black",
			"material": "PLA",
			"process":  "FDM",
		},
		"status":   0,
		"time":     45,
		"filename": "testing.gcode",
		"max_dim": map[string]interface{}{
			"height": 435,
			"length": 13,
			"width":  4234,
		},
	}}
	doc["status"] = 0

	// Push new dummy order to our "jobs" collection
	wr, err := newDocument.Create(ctx, doc)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(wr.UpdateTime)
}

// Generates a random string to be used as a collection id within "jobs" collection
func generateString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
