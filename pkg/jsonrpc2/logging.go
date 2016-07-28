package jsonrpc2

import (
	"encoding/json"
	"log"
)

// LoggingHandler wraps a Handler and logs requests, notifications,
// and responses (using the log.Printf).
type LoggingHandler struct {
	Handler
}

func (h *LoggingHandler) Handle(req *Request) *Response {
	logRequest(*req)
	resp := h.Handler.Handle(req)
	if resp != nil {
		logResponse(req.ID, req.Method, *resp)
	}
	return resp
}

func logRequest(req Request) {
	params, _ := json.Marshal(req.Params)
	if !req.Notification {
		log.Printf("REQUEST[%s]: %s: %s", req.ID, req.Method, params)
	} else {
		log.Printf("NOTIFY: %s: %s", req.Method, params)
	}
}

func logResponse(id string, method string, resp Response) {
	result, _ := json.Marshal(resp.Result)
	log.Printf("RESPONSE[%s]: %s: %s", id, method, result)
}
