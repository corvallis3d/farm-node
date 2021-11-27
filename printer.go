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
	"time"

	"github.com/gorilla/websocket"
)

const Standby = 0
const Printing = 1
const Paused = 3
const Completed = 2
const Canceled = 4
const E = 9

const Setup = 5

type Print struct {
	host       string
	port       string
	ws         *websocket.Conn
	job_path   string
	filament   string
	color      string
	status     int
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
	p.status = 0
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

func (p *Print) StartReceiveThread() {
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

			p.ProcessReceivedData(*data)
			fmt.Print("\nprinter status:", p.status)
			fmt.Print("\n")

			data.Print_jsonrpc_data()
		}
	}()
}

func (p *Print) ProcessReceivedData(data Jsonrpc) {
	switch data.Id {
	case ID_GET_PRINT_JOB_STATUS:
		result_object := data.Result.(Result_object)
		p.status = result_object.get_status_code()
	default:
	}
}

func (p *Print) SendJsonrpc(data Jsonrpc) {
	err := p.ws.WriteJSON(data)
	if err != nil {
		log.Println("write:", err)
		return
	}
}

func (p *Print) SetPendingNotification() {
	fmt.Print("\n ############## FILE_PENDING_NOTIFICATION ###############\n")
	p.RequestGcodeScript(ID_FILE_PENDING_NOTIFICATION, "FILE_PENDING_NOTIFICATION")
}

func (p *Print) SetDefaultDisplay() {
	fmt.Print("\n ############## DESKTOP CHANGE ###############\n")
	p.RequestGcodeScript(ID_DEFAULT_DISPLAY, "DEFAULT_DISPLAY")
}

func (p *Print) SetNotificationString(s string) {
	fmt.Print("\n ############## NOTIFICATION ###############\n")
	s = "SEND_STRING STR=" + `"` + s + `"`
	p.RequestGcodeScript(ID_CUSTOM_NOTIFICATION, s)
}

func (p *Print) RequestGcodeScript(id int, s string) {
	msg := New_jsonrpc()
	msg.Add_method("printer.gcode.script")
	msg.Add_id(id)
	msg.Add_params_script(s)
	p.SendJsonrpc(msg)
}

func (p *Print) RequestPrintStatus() {
	fmt.Print("\n ############## PRINT STATSSSSSSS ###############\n")
	msg := New_jsonrpc()
	msg.Add_method("printer.objects.query")
	msg.Add_id(ID_GET_PRINT_JOB_STATUS)
	msg.Adds_params_objects()
	p.SendJsonrpc(msg)
}

func (p *Print) RequestKlipperStatus() {
	fmt.Print("\n ############## GET PRINTER STATUS ###############\n")
	msg := New_jsonrpc()
	msg.Add_method("printer.info")
	msg.Add_id(ID_GET_KLIPPER_STATUS)
	p.SendJsonrpc(msg)
}

func (p *Print) StartFilenamePrint(s string) {
	msg := New_jsonrpc()
	msg.Add_method("printer.print.start")
	msg.Add_id(ID_START_FILENAME_PRINT)
	msg.Add_params_filename(s)
	p.SendJsonrpc(msg)
}

func (p *Print) UploadFile(gcodeFile GcodeFile) {
	url := url.URL{Scheme: "http", Host: p.host + ":" + p.port, Path: "/server/files/upload"}
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, errFile1 := os.Open(fmt.Sprintf("C:/Models/Processed Orders/Order #%s - First Last/Upload-Gcode/%s", gcodeFile.JobId, gcodeFile.Filename))
	defer file.Close()
	part1,
		errFile1 := writer.CreateFormFile("file", filepath.Base(fmt.Sprintf("C:/Models/Processed Orders/Order #%s - First Last/Upload-Gcode/%s", gcodeFile.JobId, gcodeFile.Filename)))
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

func (p *Print) SetStatus(status uint) {
	p.status = int(status)
}

func (p *Print) GetStatus() int {
	return p.status
}

// pass off gcode file for printer to handle
func (p *Print) HandlePrintRequest(gcodeFile GcodeFile) {

	// set this printer to busy
	p.SetStatus(Setup)

	// upload the file ASYNC
	p.UploadFile(gcodeFile)

	for {
		// Once file is uploaded break
		break
	}

	// prompt the printer technician, and wait for print start confirmation
	p.SetPendingNotification()
	// placeholder
	proceedWithPrint := true

	if proceedWithPrint {
		// start the print
		p.StartFilenamePrint(gcodeFile.Filename)

		for range time.Tick(time.Second * 30) {

			// poll for print status per 30 seconds
			// placeholder
			printStatus := p.GetStatus()

			//-----------------------------------------------------------------------------
			// only have to handle Completed, Paused, Canceled, Error
			if printStatus == Completed {
				// Wait until technician removes print, resets printer status to standby

			} else if printStatus == Paused {

			} else if printStatus == Canceled {

			} else if printStatus == E {

			} else {
				continue
			}
		}
	}
}
