package jsonrpc2

import "fmt"

type (
	Request struct {
		Version string      `json:"jsonrpc"`
		Method  string      `json:"method"`
		Params  interface{} `json:"params"`
		ID      interface{} `json:"id"`
	}

	Response struct {
		Version string      `json:"jsonrpc"`
		Result  interface{} `json:"result,omitempty"`
		Error   *Error      `json:"error,omitempty"`
		ID      interface{} `json:"id"`
	}

	Error struct {
		Code    int         `json:"code"`
		Message string      `json:"string"`
		Data    interface{} `json:"data,omitempty"`
		Err     error       `json:"-"`
	}

	ServerError struct {
		Code int
		Data interface{}
	}
)

func (se ServerError) Error() string {
	return fmt.Sprintf("ServerError: %v", se.Data)
}

func (e Error) Error() string {
	return fmt.Sprintf("%d %s Data: %v Err: %v", e.Code, e.Message, e.Data, e.Err)
}
