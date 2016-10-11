package jsonrpc2test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync/atomic"
	"testing"

	"github.com/sourcegraph/jsonrpc2"
)

type Server struct {
	Addr     string
	Listener net.Listener

	// Response with the key ID is returned for a request with the
	// corresponding ID.
	Response map[uint64]interface{}

	// SeenRequest is the latest request received with key ID.
	SeenRequest map[uint64]*jsonrpc2.Request

	// T if set will ensure that if a request with ID key is received, it
	// matches what is in WantRequest.
	T           testing.TB
	WantRequest map[uint64]*jsonrpc2.Request

	closed uint32
}

// NewServer creates and starts a Server listening on Server.Addr. Callers
// must call Close() when done with the server.
func NewServer() *Server {
	conn := newLocalListener()
	s := &Server{
		Addr:     conn.Addr().String(),
		Listener: conn,

		Response:    map[uint64]interface{}{},
		SeenRequest: map[uint64]*jsonrpc2.Request{},
		WantRequest: map[uint64]*jsonrpc2.Request{},
	}
	go func() {
		if err := jsonrpc2.Serve(context.Background(), conn, jsonrpc2.HandlerWithError(s.Handle)); err != nil && atomic.LoadUint32(&s.closed) == 0 {
			log.Printf("jsonrpc2test serve: %s", err)
		}
	}()
	return s
}

func (h *Server) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	h.SeenRequest[req.ID] = req
	want, ok := h.WantRequest[req.ID]
	if ok && h.T != nil {
		// We compare via the json versions, rather than deep
		// equal. This makes it easier to print when things go wrong
		got, _ := json.Marshal(req)
		w, _ := json.Marshal(want)
		if string(got) != string(w) {
			h.T.Errorf("got req\n%s, want\n%s", string(got), string(w))
		}
	}
	resp, ok := h.Response[req.ID]
	if !ok {
		return nil, &jsonrpc2.Error{Message: fmt.Sprintf("no mocked response for %d", req.ID)}
	}
	return resp, nil
}

func (h *Server) Close() error {
	atomic.StoreUint32(&h.closed, 1)
	return h.Listener.Close()
}

// modified from net/http/httptest
func newLocalListener() net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			panic(fmt.Sprintf("httptest: failed to listen on a port: %v", err))
		}
	}
	return l
}
