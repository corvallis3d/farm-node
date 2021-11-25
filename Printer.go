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

type Printer interface {
	NewPrinter()
	Connect()
	Start_receive_thread()
	Send_msg()
	Send_print_notification()
	Change_notification_string(s string)
	Default_display()
	Upload_file(file_name string)
	Start_print()
	// Print_json_rpc(data json_rpc_data)
	Get_printer_status()
}

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

			// data := New_rpc_req_p()
			// err = json.Unmarshal(message, &data)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			fmt.Print("\n")
			log.Print("\n", string(message))

			// p.Print_json_rpc(data)
		}
	}()
}

func (p *Print) Process_message() {

}

// func (p *Print) Print_json_rpc(data Json_rpc_req) {
// 	if data.Method != "" {
// 		fmt.Printf("Method: %s\n", data.Method)
// 	}
// 	if data.Method != "notify_proc_stat_update" && data.Params != nil {
// 		fmt.Printf("Params: %s\n", data.Params)
// 	}
// 	if data.Result != "" {
// 		fmt.Printf("Resut: %s\n", data.Result)
// 	}
// 	if data.Error.Message != "" {
// 		fmt.Printf("Error: %s\n", data.Error.Message)
// 	}
// 	if data.Id != 0 {
// 		fmt.Printf("Id: %d\n", data.Id)
// 	}
// }

// func (p *Print) Is_printer_ready(msg json_rpc_data) {
// 	if compare, ok := msg.Params.([]interface{}); ok {
// 		if compare[len(compare)-1] == "// printer_ready" {
// 			p.print_flag = true
// 		}
// 	}
// }

func (p *Print) Send_msg(msg string) {
	var payload = []byte(msg)
	err := p.ws.WriteMessage(websocket.TextMessage, payload)
	if err != nil {
		log.Println("write:", err)
		return
	}
}

func (p *Print) Send_print_notification() {
	msg :=
		`{
			"jsonrpc": "2.0",
			"method": "printer.gcode.script",
			"params": {
				"script": "FILE_PENDING_NOTIFICATION"
			},
			"id": 7466
			}`
	p.Send_msg(msg)
}

func (p *Print) Default_display() {
	msg :=
		`{
			"jsonrpc": "2.0",
			"method": "printer.gcode.script",
			"params": {
				"script": "DEFAULT_DISPLAY"
			},
			"id": 7466
			}`
	p.Send_msg(msg)
}

func (p *Print) Change_notification_string(s string) {
	s = "'" + s + "'"
	msg :=
		`{
			"jsonrpc": "2.0",
			"method": "printer.gcode.script",
			"params": {
				"script": "SEND_STRING STR=` + s + `"
			},
			"id": 7466
			}`
	p.Send_msg(msg)
}

func (p *Print) Start_print() {
	msg := `{
			"jsonrpc": "2.0",
			"method": "printer.print.start",
			"params": {
				"filename": "testing.gcode"
			},
			"id": 4654
		}`
	p.Send_msg(msg)
	p.print_flag = false
}

func (p *Print) Get_printer_status() {
	fmt.Print("\nAAAAAAAAAAAAAAAAAAAAAAa")
	msg := New_rpc_req()
	msg.Method = "printer.info"
	msg.Id = 1988
	err := p.ws.WriteJSON(msg)
	if err != nil {
		log.Println("write:", err)
		return
	}
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
