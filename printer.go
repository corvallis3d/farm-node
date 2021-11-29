package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/websocket"
)

type Print struct {
	Host             string
	Port             string
	ws               *websocket.Conn
	JobPath          string
	LastUsedMaterial string
	LastUsedColor    string
	Status           int
	IdleFlag         bool
	done             chan struct{}
}

func NewPrinter(host string, port string) *Print {
	p := new(Print)
	p.Host = host
	p.Port = port
	p.Status = Standby
	p.Connect()
	p.StartReceiveThread()
	return p
}

func (p *Print) Connect() {
	u := url.URL{Scheme: "ws", Host: p.Host + ":" + p.Port, Path: "/websocket"}
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
			data, err := JsonUnmarshal(message)
			if err != nil {
				log.Print(err)
				continue
			}
			p.ProcessReceivedData(*data)
		}
	}()
}

func (p *Print) ProcessReceivedData(data Jsonrpc) {
	// Process data according to Id number
	switch data.Id {
	case IdPrintStatus:
		result_object := data.Result.(Result_object)
		p.Status = result_object.get_status_code()
		return
	}

	// Process data according to method information
	switch data.Method {
	case "notify_proc_stat_update":
		return
	case "notify_gcode_response":
		fmt.Print(p.IdleFlag)
		p.ProcessGcodeResponse(data.Params.([]interface{})[0].(string))
		return
	}

	// Print all unprocessed data
	data.Print_jsonrpc_data()
}

func (p *Print) ProcessGcodeResponse(res string) {
	if strings.Contains(res, "IdleFlag:1.0") {
		p.IdleFlag = true
	} else if strings.Contains(res, "IdleFlag:0.0") {
		p.IdleFlag = false
	}
}

func (p *Print) SendJsonrpc(data Jsonrpc) {
	err := p.ws.WriteJSON(data)
	if err != nil {
		log.Println("write:", err)
		return
	}
}

func (p *Print) SetDisplayNotification(GF GcodeFile) {
	Script := "DISPLAY_NOTIFICATION "
	Script += `NAME="` + GF.Filename + `"`
	Script += ` COLOR="` + GF.Color + `"`
	Script += ` MATERIAL="` + GF.Material + `"`
	p.RequestGcodeScript(IdDisplayNotification, Script)
}

func (p *Print) SetDefaultDisplay() {
	p.RequestGcodeScript(IdDefaultDisplay, "DISPLAY_DEFAULT")
}

func (p *Print) RequestGcodeScript(Id int, Script string) {
	Jsonrpc_req := NewJsonrpc()
	Jsonrpc_req.Add_method("printer.gcode.script")
	Jsonrpc_req.Add_id(Id)
	Jsonrpc_req.Add_params_script(Script)
	p.SendJsonrpc(Jsonrpc_req)
}

func (p *Print) RequestPrintStatus() {
	Jsonrpc_req := NewJsonrpc()
	Jsonrpc_req.Add_method("printer.objects.query")
	Jsonrpc_req.Add_id(IdPrintStatus)
	Jsonrpc_req.Adds_params_objects()
	p.SendJsonrpc(Jsonrpc_req)
}

func (p *Print) StartFilenamePrint(FileName string) {
	Jsonrpc_req := NewJsonrpc()
	Jsonrpc_req.Add_method("printer.print.start")
	Jsonrpc_req.Add_id(IdStartFileNamePrint)
	Jsonrpc_req.Add_params_filename(FileName)
	p.SendJsonrpc(Jsonrpc_req)
}

func (p *Print) UploadFile(GF GcodeFile) {
	url := url.URL{Scheme: "http", Host: p.Host + ":" + p.Port, Path: "/server/files/upload"}
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, errFile1 := os.Open(fmt.Sprintf("C:/Models/Processed Orders/Order #%s - First Last/Upload-Gcode/%s", GF.JobId, GF.Filename))
	defer file.Close()
	part1, errFile1 := writer.CreateFormFile("file", filepath.Base(fmt.Sprintf("C:/Models/Processed Orders/Order #%s - First Last/Upload-Gcode/%s", GF.JobId, GF.Filename)))
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
}

func (p *Print) SetStatus(status uint) {
	p.Status = int(status)
}

func (p *Print) GetStatus() int {
	return p.Status
}

func (p *Print) GetIdleFlag() bool {
	return p.IdleFlag
}

// pass off gcode file for printer to handle
func (p *Print) HandlePrintRequest(GF GcodeFile, ctx context.Context, client *firestore.Client) {

	p.SetStatus(Setup)
	p.UploadFile(GF)

	p.SetDisplayNotification(GF)

	// If printer is idle, GetIdleFlag==True, stay in for loop
	for p.GetIdleFlag() {
		time.Sleep(time.Second)
	}

	p.StartFilenamePrint(GF.Filename)

	// Check on the print status
	for range time.Tick(time.Second * 30) {
		p.RequestPrintStatus()
		printStatus := p.Status

		if printStatus == Completed {
			p.SetStatus(Resetting)
			GF.SetStatus(GcodePrintSuccess)
			UpdateFileStatus(GF, ctx, client)

			// Wait until technician removes print, reset printer status to standby
			// Send notification to release printer back to the queue

			//-----------------------------------------------------------------------------
			/* While printing, GetIdleFlag evaluates to false.
			When technician is ready, LCD status is changed to Idle and
			GetIdleFlag evaluates to true
			*/
			for p.GetIdleFlag() == false {
				time.Sleep(time.Second)
			}
			p.LastUsedColor = GF.Color
			p.LastUsedMaterial = GF.Material
			p.SetStatus(Standby)
			runtime.Goexit()

			// IF PAUSED
		} else if printStatus == Paused {
			fmt.Println("Paused")
		} else if printStatus == Canceled {
			p.SetStatus(Resetting)
			GF.SetStatus(GcodeCanceled)

			UpdateFileStatus(GF, ctx, client)
			// Send notification to release printer back to the queue

			for p.GetIdleFlag() == false {
				time.Sleep(time.Second)
			}
			p.LastUsedColor = GF.Color
			p.LastUsedMaterial = GF.Material

			p.SetStatus(Standby)
			runtime.Goexit()

		} else if printStatus == E {

		}
	}
}
