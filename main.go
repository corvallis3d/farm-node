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

	var printerArray []Printer

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

	// Get firebase instance
	client, ctx, err := FirebaseInstance()
	if err != nil {
		panic(err)
	}

	// Spin-off snapshot worker
	go jobsSnapshot(ctx, client, printerArray)
	// Wait forever!
	for {

	}

}
