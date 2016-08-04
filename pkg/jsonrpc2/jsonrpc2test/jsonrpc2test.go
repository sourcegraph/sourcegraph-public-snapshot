package jsonrpc2test

import (
	"encoding/json"
	"fmt"
	"net"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
)

type Server struct {
	Addr     string
	Listener net.Listener

	// Response with the key ID is returned for a request with the
	// corresponding ID.
	Response map[string]*jsonrpc2.Response

	// SeenRequest is the latest request received with key ID.
	SeenRequest map[string]*jsonrpc2.Request

	// T if set will ensure that if a request with ID key is received, it
	// matches what is in WantRequest.
	T           testing.TB
	WantRequest map[string]*jsonrpc2.Request
}

// NewServer creates and starts a Server listening on Server.Addr. Callers
// must call Close() when done with the server.
func NewServer() *Server {
	conn := newLocalListener()
	s := &Server{
		Addr:     conn.Addr().String(),
		Listener: conn,

		Response:    map[string]*jsonrpc2.Response{},
		SeenRequest: map[string]*jsonrpc2.Request{},
		WantRequest: map[string]*jsonrpc2.Request{},
	}
	go jsonrpc2.Serve(conn, s)
	return s
}

func (h *Server) Handle(req *jsonrpc2.Request) *jsonrpc2.Response {
	h.SeenRequest[req.ID] = req
	want, ok := h.WantRequest[req.ID]
	if ok && h.T != nil {
		// We compare via the json versions, rather than deep
		// equal. This makes it easier to print when things go wrong
		got, _ := json.Marshal(req)
		w, _ := json.Marshal(want)
		if string(got) != string(w) {
			h.T.Errorf("got req %s, want %s", string(got), string(w))
		}
	}
	resp, ok := h.Response[req.ID]
	if !ok {
		return &jsonrpc2.Response{
			ID:    req.ID,
			Error: &jsonrpc2.Error{Message: "no mocked response for " + req.ID},
		}
	}
	return resp
}

func (h *Server) Close() error {
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
