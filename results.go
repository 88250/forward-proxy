package main

// Result represents HTTP response body.
type Result struct {
	Code int         `json:"code"` // return code, 0 for succ
	Msg  string      `json:"msg"`  // message
	Data interface{} `json:"data"` // data object
}

// NewResult creates a result with Code=0, Msg="", Data=nil.
func NewResult() *Result {
	return &Result{
		Code: 0,
		Msg:  "",
		Data: nil,
	}
}

// Result codes.
const (
	CodeOk      = 0  // OK
	CodeErr     = -1 // general error
	CodeAuthErr = 2  // unauthenticated request
)
