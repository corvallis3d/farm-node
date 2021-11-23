package main

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
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

	instantiateAllPrinters()

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
