package main

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

const (
	GcodeIdle         = 0
	GcodePrinting     = 1
	GcodePrintSuccess = 2
	GcodeCanceled     = 3
	GcodeError        = 9

	JobIdle       = 0
	JobInProgress = 1
	JobCompleted  = 2
)

var (
	jobs         = []Job{}
	printerArray []*Print
	gcodeQueue   []GcodeFile
)

func main() {
	viper.SetConfigName("development")
	viper.SetConfigType("toml")
	viper.AddConfigPath("./config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
	})
	viper.WatchConfig()

	// Will need error handling
	instantiateAllPrinters()

	// Get firebase instance
	client, ctx, err := FirebaseInstance()
	if err != nil {
		panic(err)
	}

	// Spin-off snapshot worker
	go jobsSnapshot(ctx, client)

	go managePrintJobs(ctx, client)

	go maintainFirestore(ctx, client)

	//go addFalseDocumentToJobsCollection(ctx, client)

	// Wait forever!
	for {

	}

}
