package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

type Printer_Interface interface {
	NewPrinter()
	Connect()
	Start_receive_thread()
	Change_printer_status(s string)
	Send_print_file()
	Pause_printer()
	Enqueue_file()
	Send_msg()
}

type Printer struct {
	host      string
	port      string
	ws        *websocket.Conn
	job_path  string
	filament  string
	color     string
	recv_flag bool // recv_flag will allow received messages to be printed. TODO
	done      chan struct{}
}

func NewPrinter(host string, port string) *Printer {
	p := new(Printer)
	p.host = host
	p.port = port
	p.recv_flag = true
	return p
}

func (p *Printer) Connect() {
	u := url.URL{Scheme: "ws", Host: p.host + ":" + p.port, Path: "/websocket"}
	log.Printf("connecting to %s", u.String())
	var err error
	p.ws, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
}

func (p *Printer) Start_receive_thread() {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := p.ws.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			var b map[string]interface{}
			json.Unmarshal([]byte(message), &b)
			if p.recv_flag == true {
				log.Println("recv:")
				for k, v := range b {
					fmt.Printf("%s: %s\n", k, v)
				}
				fmt.Printf("\n")
			}
		}
	}()
}

func (p *Printer) Send_msg(msg string) {
	var payload = []byte(fmt.Sprintf(msg))
	err := p.ws.WriteMessage(websocket.TextMessage, payload)
	if err != nil {
		log.Println("write:", err)
		return
	}
}

func (p *Printer) Change_printer_status(s string) {
	msg := `{
		"jsonrpc": "2.0",
		"method": "printer.gcode.script",
		"params": {
			"script": "M117_` + s + `"
		},
		"id": 7466}`
	p.Send_msg(msg)
}

func (p *Printer) Send_print_file() {
	msg := `{
			"jsonrpc": "2.0",
			"method": "printer.print.start",
			"params": {
				"filename": "testing.gcode"
			},
			"id": 4654
		}`
	p.Send_msg(msg)
}

func (p *Printer) Enqueue_file() {
	msg := `{
		"jsonrpc": "2.0",
		"method": "server.job_queue.post_job",
		"params": {
			"filenames": [
				"testing.gcode",
			]
		},
		"id": 4654
	}`
	p.Send_msg(msg)
}

func (p *Printer) Pause_printer() {
	msg := `{
		"jsonrpc": "2.0",
		"method": "printer.print.pause",
		"id": 4564
	}`
	p.Send_msg(msg)
}
