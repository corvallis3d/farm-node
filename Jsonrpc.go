package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/mitchellh/mapstructure"
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
	Script   string `json:"script,omitempty"`
	Filename string `json:"filename,omitempty"`
}

type Error_object struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    string `json:"data,omitempty"`
}

type Result_object struct {
	State            string `json:"state,omitempty"`
	State_message    string `json:"state_message,omitempty"`
	Hostname         string `json:"hostname,omitempty"`
	Software_version string `json:"software_version,omitempty"`
	Cpu_info         string `json:"cpu_info,omitempty"`
	Klipper_path     string `json:"klipper_path,omitempty"`
	Python_path      string `json:"python_path,omitempty"`
	Config_file      string `json:"config_file,omitempty"`
}

func New_jsonrpc() Jsonrpc {
	new_jsonrpc := new(Jsonrpc)
	new_jsonrpc.Jsonrpc = "2.0"
	return *new_jsonrpc
}

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
		raw.Result = *ro
	}

	return raw, err
}

func (p *Jsonrpc) Print_jsonrpc_data() {
	log.Print(">>recv\n")
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
