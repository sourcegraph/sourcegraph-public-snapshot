package jsonrpc2

import (
	"context"
	"log"
	"reflect"
)

// HandlerWithError implements Handler by calling the func for each
// request and handling returned errors and results.
type HandlerWithError func(context.Context, *Conn, *Request) (result interface{}, err error)

// Handle implements Handler.
func (h HandlerWithError) Handle(ctx context.Context, conn *Conn, req *Request) {
	result, err := h(ctx, conn, req)
	if req.Notif {
		if err != nil {
			log.Printf("jsonrpc2 handler: notification %q handling error: %s", req.Method, err)
		}
		return
	}

	resp := &Response{ID: req.ID}
	if err == nil {
		if isNilValue(result) {
			result = struct{}{}
		}
		err = resp.SetResult(result)
	}
	if err != nil {
		if e, ok := err.(*Error); ok {
			resp.Error = e
		} else {
			resp.Error = &Error{Message: err.Error()}
		}
	}

	if !req.Notif {
		if err := conn.SendResponse(ctx, resp); err != nil {
			log.Printf("jsonrpc2 handler: sending response %d: %s", resp.ID, err)
		}
	}
}

// isNilValue tests if an interface is empty, because an empty interface does
// not encode any information, we can't encode it in JSON so that the proxy
// knows it's a response, not a request.
func isNilValue(resp interface{}) bool {
	if resp == nil {
		return true
	}
	kind := reflect.TypeOf(resp).Kind()
	value := reflect.ValueOf(resp)
	nilPtr := kind == reflect.Ptr && value.IsNil()
	nilSlice := kind == reflect.Slice && value.IsNil()
	return nilPtr || nilSlice
}
