package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/gorilla/websocket"
)

type Print struct {
	host       string
	port       string
	ws         *websocket.Conn
	job_path   string
	filament   string
	color      string
	recv_flag  bool // recv_flag will allow received messages to be printed. TODO
	print_flag bool
	done       chan struct{}
}

func NewPrinter(host string, port string) *Print {
	p := new(Print)
	p.host = host
	p.port = port
	p.recv_flag = true
	p.print_flag = false
	return p
}

func (p *Print) Connect() {
	u := url.URL{Scheme: "ws", Host: p.host + ":" + p.port, Path: "/websocket"}
	log.Printf("connecting to %s", u.String())
	var err error
	p.ws, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
}

func (p *Print) Start_receive_thread() {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {

			_, message, err := p.ws.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			data, err := JSON_Unmarshal(message)
			if err != nil {
				log.Print(err)
				continue
			}
			data.Print_jsonrpc_data()
		}
	}()
}

func (p *Print) Send_jsonrpc(data Jsonrpc) {
	err := p.ws.WriteJSON(data)
	if err != nil {
		log.Println("write:", err)
		return
	}
}

func (p *Print) Get_printer_status() {
	fmt.Print("\n ############## GET PRINTER STATUS ###############\n")
	msg := New_jsonrpc()
	msg.Add_method("printer.info")
	msg.Add_id(8888)
	p.Send_jsonrpc(msg)
}

func (p *Print) Send_print_notification() {
	fmt.Print("\n ############## FILE_PENDING_NOTIFICATION ###############\n")
	p.Send_gcode_script("FILE_PENDING_NOTIFICATION")
}

func (p *Print) Default_display() {
	fmt.Print("\n ############## DESKTOP CHANGE ###############\n")
	p.Send_gcode_script("DEFAULT_DISPLAY")
}

func (p *Print) Change_notification_string(s string) {
	s = "SEND_STRING STR=" + `"` + s + `"`
	fmt.Print("\n ############## NOTIFICATION ###############\n")
	p.Send_gcode_script(s)

}

func (p *Print) Send_gcode_script(s string) {
	msg := New_jsonrpc()
	msg.Add_method("printer.gcode.script")
	msg.Add_id(1111)
	msg.Add_params_script(s)
	p.Send_jsonrpc(msg)

}

func (p *Print) Start_filename_print(s string) {
	msg := New_jsonrpc()
	msg.Add_method("printer.print.start")
	msg.Add_id(1234)
	msg.Add_params_filename(s)
	p.Send_jsonrpc(msg)

}

func (p *Print) Upload_file(file_name string) {
	url := url.URL{Scheme: "http", Host: p.host + ":" + p.port, Path: "/server/files/upload"}
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, errFile1 := os.Open(fmt.Sprintf("./gcode/%s", file_name))
	defer file.Close()
	part1,
		errFile1 := writer.CreateFormFile("file", filepath.Base(fmt.Sprintf("./gcode/%s", file_name)))
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
	req, err := http.NewRequest("POST", url.String(), payload)

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
	p.print_flag = false
}
