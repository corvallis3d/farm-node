package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func uploadAFile() {
	url := viper.GetString("moonraker.baseUrl") + "server/files/upload"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, errFile1 := os.Open(fmt.Sprintf("./gcode/%s", viper.GetString("temp.filename")))
	defer file.Close()
	part1,
		errFile1 := writer.CreateFormFile("file", filepath.Base(fmt.Sprintf("./gcode/%s", viper.GetString("temp.filename"))))
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

func printAFile() {
	// Note: This function works but must be adapted to our use case, as in get it to work with our Firebase system
	// TO DO: Job struct passed in as argument, have url param filename be set to filename from struct

	url := viper.GetString("moonraker.baseUrl") + "printer/print/start?filename=" + viper.GetString("temp.filename")
	method := "POST"

	var payload = []byte(fmt.Sprintf(
		`{
			"jsonrpc": "2.0",
			"method": "printer.print.start",
			"params": {
				"filename": %s
			}
		}`,
		viper.GetString("temp.filename"),
	))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

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

	uploadAFile()
	printAFile()

	// Get firebase instance
	// client, ctx, err := FirebaseInstance()
	// if err != nil {
	// 	panic(err)
	// }

	// Spin-off snapshot worker
	// go jobsSnapshot(ctx, client)

	// Wait forever!
	// for {

	// }

}
