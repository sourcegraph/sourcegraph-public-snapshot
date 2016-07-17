package lsp

import "fmt"

// Refer to https://github.com/Microsoft/language-server-protocol/blob/master/protocol.md for documentation.

type ResponseError struct {
	Code    int
	Message string
	Data    interface{}
}

type ErrorCode int

const (
	ParseError       ErrorCode = -32700
	InvalidRequest             = -32600
	MethodNotFound             = -32601
	InvalidParams              = -32602
	InternalError              = -32603
	serverErrorStart           = -3209
	serverErrorEnd             = -32000
)

func (e *ResponseError) Error() string {
	return fmt.Sprintf("ResponseError{Code: %v, Message: %v, Data: %v}", e.Code, e.Message, e.Data)
}
