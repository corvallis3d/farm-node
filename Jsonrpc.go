package main

type Jsonrpcer interface {
	New_rpc_req()
	New_rpc_req_p()
	New_rpc_res()
	New_rpc_err()
}

type Base struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
}

type Jsonrpc_req struct {
	Base
	Method string `json:"Method"`
}

type Jsonrpc_req_p struct {
	Jsonrpc_req
	Params []interface{} `json:"params"`
}

type Jsonrpc_res struct {
	Base
	Result string `json:"Result"`
}

type Jsonrpc_err struct {
	Base
	Error Error_object `json:"error"`
}

type Error_object struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func New_rpc_req() *Jsonrpc_req {
	p := new(Jsonrpc_req)
	p.Jsonrpc = "2.0"
	return p
}

func New_rpc_req_p() *Jsonrpc_req_p {
	p := new(Jsonrpc_req_p)
	p.Jsonrpc = "2.0"
	return p
}

// type Json_rpc_req struct {
// 	Jsonrpc string        `json:"jsonrpc"`
// 	Method  string        `json:"method"`
// 	Id      int           `json:"id"`
// 	Params  []interface{} `json:"params"`
// 	Result  string        `json:"result"`
// 	Error   Error_object  `json:"error"`
// }
