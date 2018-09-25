package jsonrpc2

import (
	"context"
	"log"
)

// HandlerWithError implements Handler by calling the func for each
// request and handling returned errors and results.
func HandlerWithError(handleFunc func(context.Context, *Conn, *Request) (result interface{}, err error)) *HandlerWithErrorConfigurer {
	return &HandlerWithErrorConfigurer{handleFunc: handleFunc}
}

type HandlerWithErrorConfigurer struct {
	handleFunc        func(context.Context, *Conn, *Request) (result interface{}, err error)
	suppressErrClosed bool
}

// Handle implements Handler.
func (h *HandlerWithErrorConfigurer) Handle(ctx context.Context, conn *Conn, req *Request) {
	result, err := h.handleFunc(ctx, conn, req)
	if req.Notif {
		if err != nil {
			log.Printf("jsonrpc2 handler: notification %q handling error: %s", req.Method, err)
		}
		return
	}

	resp := &Response{ID: req.ID}
	if err == nil {
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
			if err != ErrClosed || !h.suppressErrClosed {
				log.Printf("jsonrpc2 handler: sending response %s: %s", resp.ID, err)
			}
		}
	}
}

// SuppressErrClosed makes the handler suppress jsonrpc2.ErrClosed errors from
// being logged. The original handler `h` is returned.
//
// This is optional because only in some cases is this behavior desired. For
// example, a handler that serves end-user connections may not want to log
// ErrClosed because it just indicates the end-user connection has gone away
// for any reason (they could have lost wifi connection, are no longer
// interested in the request and closed the connection, etc) and as such it
// would be log spam, whereas a handler that serves internal connections would
// never expect connections to go away unexpectedly (which could indicate
// service degradation, etc) and as such ErrClosed should always be logged.
func (h *HandlerWithErrorConfigurer) SuppressErrClosed() Handler {
	h.suppressErrClosed = true
	return h
}
