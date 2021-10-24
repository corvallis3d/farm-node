package main

import "fmt"

func main() {

	// Example Println
	fmt.Println("Hello world")

	// Get firebase instance
	client, ctx, err := FirebaseInstance()
	if err != nil {
		panic(err)
	}

	// Spin-off snapshot worker
	go jobsSnapshot(ctx, client)

	// Wait forever!
	for {

	}
}
