package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/mitchellh/mapstructure"
)

// Global variables for Json RPC request IDs
const (
	ID_DEFAULT_DISPLAY           = 1000
	ID_CUSTOM_NOTIFICATION       = 2222
	ID_FILE_PENDING_NOTIFICATION = 2223
	ID_GET_KLIPPER_STATUS        = 3330
	ID_GET_PRINTER_STATUS        = 3331
	ID_GET_PRINT_JOB_STATUS      = 7777
	ID_START_FILENAME_PRINT      = 5555

	Standby   = 0
	Printing  = 1
	Completed = 2
	Paused    = 3
	Canceled  = 4
	Setup     = 5
	Resetting = 6
	E         = 9
)

type Jsonrpc struct {
	Jsonrpc string       `json:"jsonrpc,omitempty"`
	Method  string       `json:"method,omitempty"`
	Id      int          `json:"id,omitempty"`
	Params  interface{}  `json:"params,omitempty"`
	Result  interface{}  `json:"result,omitempty"`
	Error   Error_object `json:"error,omitempty"`
}

type Params_object struct {
	Script   string      `json:"script,omitempty"`
	Filename string      `json:"filename,omitempty"`
	Objects  interface{} `json:"objects,omitempty"`
}

type Error_object struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    string `json:"data,omitempty"`
}

type Result_object struct {
	State            string         `json:"state,omitempty"`
	State_message    string         `json:"state_message,omitempty"`
	Hostname         string         `json:"hostname,omitempty"`
	Software_version string         `json:"software_version,omitempty"`
	Cpu_info         string         `json:"cpu_info,omitempty"`
	Klipper_path     string         `json:"klipper_path,omitempty"`
	Python_path      string         `json:"python_path,omitempty"`
	Config_file      string         `json:"config_file,omitempty"`
	Status           Objects_object `json:"status,omitempty"`
	Eventtime        float32        `json:"eventtime,omitempty"`
}

type Objects_object struct {
	Webhooks       *Webhooks_object       `json:"webhooks"`
	Virtual_sdcard *Virtual_sdcard_object `json:"virtual_sdcard"`
	Print_stats    *Print_stats_object    `json:"print_stats"`
}

type Virtual_sdcard_object struct {
	Progress      float32     `json:"progress,omitempty"`
	File_position int         `json:"file_position,omitempty"`
	Is_active     bool        `json:"is_active,omitempty"`
	File_path     interface{} `json:"file_path,omitempty"`
	File_size     int         `json:"file_size,omitempty"`
}

type Webhooks_object struct {
	State         string `json:"state,omitempty"`
	State_message string `json:"state_message,omitempty"`
}

type Print_stats_object struct {
	Print_duration float32 `json:"print_duration,omitempty"`
	Total_duration float32 `json:"total_duration,omitempty"`
	Filament_used  float32 `json:"filament_used,omitempty"`
	Filename       string  `json:"filename,omitempty"`
	State          string  `json:"state,omitempty"`
	Message        string  `json:"message,omitempty"`
}

func New_jsonrpc() Jsonrpc {
	new_jsonrpc := new(Jsonrpc)
	new_jsonrpc.Jsonrpc = "2.0"
	return *new_jsonrpc
}

/*
Unmarshals bytes to appropriate data structure
*/
func JSON_Unmarshal(bytes []byte) (*Jsonrpc, error) {
	raw := new(Jsonrpc)
	err := json.Unmarshal(bytes, &raw)
	if err != nil {
		return nil, err
	}
	// Map Result to Result_object
	switch v := raw.Result.(type) {
	case map[string]interface{}:
		ro := new(Result_object)
		mapstructure.Decode(v, &ro)
		// Create fields for Status under Results_object
		if raw.Id == ID_GET_PRINT_JOB_STATUS {
			ro.Create_status_object()
		}
		raw.Result = *ro
	}
	return raw, err
}

/*
Debug function to print contents of JSON RPC to console
*/
func (p *Jsonrpc) Print_jsonrpc_data() {
	log.Print(">> recv\n")

	if p.Method != "" {
		fmt.Printf("Method: %s\n", p.Method)
	}

	if p.Params != nil {
		switch v := p.Params.(type) {
		case []string:
			fmt.Printf("Params: %s\n", v)
		case interface{}:
			fmt.Printf("Params:")
			indented_v, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				log.Fatalf(err.Error())
			}
			fmt.Println(string(indented_v))
		}
	}

	switch v := p.Result.(type) {
	case string:
		fmt.Printf("Result: %s\n", v)
	case Result_object:
		fmt.Printf("Result_object:")
		indented_v, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			log.Fatalf(err.Error())
		}
		fmt.Println(string(indented_v))
	}

	if p.Error.Message != "" {
		fmt.Printf("Error: %s\n", p.Error.Message)
	}

	if p.Id != 0 {
		fmt.Printf("Id: %d\n", p.Id)
	}
	fmt.Print("\n")
}

func (p *Jsonrpc) Add_method(method string) {
	p.Method = method
}

func (p *Jsonrpc) Add_id(id int) {
	p.Id = id
}

func (p *Jsonrpc) Add_params_script(script string) {
	Params_object := new(Params_object)
	Params_object.Script = script
	p.Params = *Params_object
}

func (p *Jsonrpc) Add_params_filename(filename string) {
	Params_object := new(Params_object)
	Params_object.Filename = filename
	p.Params = *Params_object
}

/* Creates the following struct for the Params field of a Jsonrpc object
{
	params: {
		objects: {
			webhooks = nil, virtual_sdcard = nil, print_stats = nil,
		}
	}
}
*/
func (p *Jsonrpc) Adds_params_objects() {
	obj := new(Objects_object)
	obj.Webhooks = nil
	obj.Virtual_sdcard = nil
	obj.Print_stats = nil
	po := new(Params_object)
	po.Objects = obj
	p.Params = po
}

/*
Adds fields for Status under Result_object when checking printer status
*/
func (ro *Result_object) Create_status_object() {
	so := new(Objects_object)
	mapstructure.Decode(ro.Status, &so)

	wo := new(Webhooks_object)
	mapstructure.Decode(so.Webhooks, &wo)
	so.Webhooks = wo

	vso := new(Virtual_sdcard_object)
	mapstructure.Decode(so.Virtual_sdcard, &vso)
	so.Virtual_sdcard = vso

	pso := new(Print_stats_object)
	mapstructure.Decode(so.Print_stats, &pso)
	so.Print_stats = pso
	ro.Status = *so
}

func (ro *Result_object) get_status_code() int {
	switch ro.Status.Print_stats.State {
	case "standby":
		return Standby
	case "printing":
		return Printing
	case "paused":
		return Paused
	case "completed":
		return Completed
	case "canceled":
		return Canceled
	case "error":
		return E
	default:
		return -1
	}
}
