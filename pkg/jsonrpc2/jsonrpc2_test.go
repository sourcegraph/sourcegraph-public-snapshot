package jsonrpc2

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRequest_MarshalJSON_jsonrpc(t *testing.T) {
	b, err := json.Marshal(&Request{})
	if err != nil {
		t.Fatal(err)
	}
	if want := `"jsonrpc":"2.0"`; !strings.Contains(string(b), want) {
		t.Errorf("got %s, want it to include the string %s", b, want)
	}
}

func TestResponse_MarshalJSON_jsonrpc(t *testing.T) {
	b, err := json.Marshal(&Response{})
	if err != nil {
		t.Fatal(err)
	}
	if want := `"jsonrpc":"2.0"`; !strings.Contains(string(b), want) {
		t.Errorf("got %s, want it to include the string %s", b, want)
	}
}

func TestResponseMarshalJSON_Notif(t *testing.T) {
	tests := map[*Request]bool{
		&Request{ID: 0}:       true,
		&Request{ID: 1}:       true,
		&Request{Notif: true}: false,
	}
	for r, wantIDKey := range tests {
		b, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		hasIDKey := bytes.Contains(b, []byte(`"id"`))
		if hasIDKey != wantIDKey {
			t.Errorf("got %s, want contain id key: %v", b, wantIDKey)
		}
	}
}

func TestResponseUnmarshalJSON_Notif(t *testing.T) {
	tests := map[string]bool{
		`{"method":"f","id":0}`: false,
		`{"method":"f","id":1}`: false,
		`{"method":"f"}`:        true,
	}
	for s, want := range tests {
		var r Request
		if err := json.Unmarshal([]byte(s), &r); err != nil {
			t.Fatal(err)
		}
		if r.Notif != want {
			t.Errorf("%s: got %v, want %v", s, r.Notif, want)
		}
	}
}

// testHandlerA is the "server" handler.
type testHandlerA struct{ t *testing.T }

func (h *testHandlerA) Handle(ctx context.Context, conn *Conn, req *Request) {
	if req.Notif {
		return // notification
	}
	if err := conn.Reply(ctx, req.ID, fmt.Sprintf("hello, #%d: %s", req.ID, *req.Params)); err != nil {
		h.t.Error(err)
	}

	if err := conn.Notify(ctx, "m", fmt.Sprintf("notif for #%d", req.ID)); err != nil {
		h.t.Error(err)
	}
}

// testHandlerB is the "client" handler.
type testHandlerB struct {
	t   *testing.T
	mu  sync.Mutex
	got []string
}

func (h *testHandlerB) Handle(ctx context.Context, conn *Conn, req *Request) {
	if req.Notif {
		h.mu.Lock()
		defer h.mu.Unlock()
		h.got = append(h.got, string(*req.Params))
		return
	}
	h.t.Errorf("testHandlerB got unexpected request %+v", req)
}

func TestClientServer(t *testing.T) {
	ctx := context.Background()
	done := make(chan struct{})
	lis, err := net.Listen("tcp", "127.0.0.1:0") // any available address
	if err != nil {
		t.Fatal("Listen:", err)
	}
	defer func() {
		if lis == nil {
			return // already closed
		}
		if err := lis.Close(); err != nil {
			if !strings.HasSuffix(err.Error(), "use of closed network connection") {
				t.Fatal(err)
			}
		}
	}()

	ha := testHandlerA{t: t}
	go func() {
		if err := Serve(ctx, lis, &ha); err != nil {
			if !strings.HasSuffix(err.Error(), "use of closed network connection") {
				t.Error(err)
			}
		}
		close(done)
	}()

	conn, err := net.Dial("tcp", lis.Addr().String())
	if err != nil {
		t.Fatal("Dial:", err)
	}
	hb := testHandlerB{t: t}
	cc := NewConn(ctx, conn, &hb)
	defer func() {
		if err := cc.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	// Simple
	const n = 100
	for i := 0; i < n; i++ {
		var got string
		if err := cc.Call(ctx, "f", []int32{1, 2, 3}, &got); err != nil {
			t.Fatal(err)
		}
		if want := fmt.Sprintf("hello, #%d: [1,2,3]", i); got != want {
			t.Errorf("got result %q, want %q", got, want)
		}
	}
	time.Sleep(100 * time.Millisecond)
	hb.mu.Lock()
	if len(hb.got) != n {
		t.Errorf("testHandlerB got %d notifications, want %d", len(hb.got), n)
	}
	hb.mu.Unlock()

	lis.Close()
	<-done // ensure Serve's error return (if any) is caught by this test
}

type noopHandler struct{}

func (noopHandler) Handle(ctx context.Context, conn *Conn, req *Request) {}

type readWriteCloser struct {
	read, write func(p []byte) (n int, err error)
}

func (x readWriteCloser) Read(p []byte) (n int, err error) {
	return x.read(p)
}

func (x readWriteCloser) Write(p []byte) (n int, err error) {
	return x.write(p)
}

func (readWriteCloser) Close() error { return nil }

func eof(p []byte) (n int, err error) {
	return 0, io.EOF
}

func TestConn_DisconnectNotify_EOF(t *testing.T) {
	c := NewConn(context.Background(), &readWriteCloser{eof, eof}, nil)
	select {
	case <-c.DisconnectNotify():
	case <-time.After(200 * time.Millisecond):
		t.Fatal("no disconnect notification")
	}
}

func TestConn_DisconnectNotify_Close(t *testing.T) {
	c := NewConn(context.Background(), &readWriteCloser{eof, eof}, nil)
	if err := c.Close(); err != nil {
		t.Error(err)
	}
	select {
	case <-c.DisconnectNotify():
	case <-time.After(200 * time.Millisecond):
		t.Fatal("no disconnect notification")
	}
}

func TestConn_DisconnectNotify_Close_async(t *testing.T) {
	done := make(chan struct{})
	c := NewConn(context.Background(), &readWriteCloser{eof, eof}, nil)
	go func() {
		if err := c.Close(); err != nil && err != ErrClosed {
			t.Error(err)
		}
		close(done)
	}()
	select {
	case <-c.DisconnectNotify():
	case <-time.After(200 * time.Millisecond):
		t.Fatal("no disconnect notification")
	}
	<-done
}

func TestConn_Close_waitingForResponse(t *testing.T) {
	c := NewConn(context.Background(), &readWriteCloser{eof, eof}, noopHandler{})
	done := make(chan struct{})
	go func() {
		if err := c.Call(context.Background(), "m", nil, nil); err != ErrClosed {
			t.Errorf("got error %v, want %v", err, ErrClosed)
		}
		close(done)
	}()
	if err := c.Close(); err != nil && err != ErrClosed {
		t.Error(err)
	}
	select {
	case <-c.DisconnectNotify():
	case <-time.After(200 * time.Millisecond):
		t.Fatal("no disconnect notification")
	}
	<-done
}

func TestAnyMessage(t *testing.T) {
	tests := map[string]struct {
		request, response bool
	}{
		// Single messages
		`{}`:                                   {},
		`{"foo":"bar"}`:                        {},
		`{"method":"m"}`:                       {request: true},
		`{"result":123}`:                       {response: true},
		`{"error":{"code":456,"message":"m"}}`: {response: true},

		// Batches
		`[{"method":"m"}]`:                                      {request: true},
		`[{"method":"m"},{"foo":"bar"}]`:                        {},
		`[{"method":"m"},{"result":123}]`:                       {},
		`[{"result":123},{"method":"foo"}]`:                     {},
		`[{"result":123}]`:                                      {response: true},
		`[{"error":{"code":456,"message":"m"}}]`:                {response: true},
		`[{"result":123},{"error":{"code":456,"message":"m"}}]`: {response: true},
	}
	for s, want := range tests {
		var m anyMessage
		json.Unmarshal([]byte(s), &m)
		if (m.request != nil) != want.request {
			t.Errorf("%s: got request %v, want %v", s, m.request != nil, want.request)
		}
		if (m.response != nil) != want.response {
			t.Errorf("%s: got response %v, want %v", s, m.response != nil, want.response)
		}
	}
}

func TestMessageCodec(t *testing.T) {
	tests := []struct {
		v, vempty interface{}
	}{
		{
			v:      &requestOrRequestBatch{single: &Request{ID: 123}},
			vempty: &requestOrRequestBatch{single: &Request{ID: 123}},
		},
		{
			v:      &responseOrResponseBatch{single: &Response{ID: 123}},
			vempty: &responseOrResponseBatch{},
		},
	}
	for _, test := range tests {
		b, err := json.Marshal(test.v)
		if err != nil {
			t.Fatal(err)
		}

		if err := json.Unmarshal(b, test.vempty); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(test.vempty, test.v) {
			t.Errorf("got %+v, want %+v", test.vempty, test.v)
		}
	}
}

func TestMapRespsToReq(t *testing.T) {
	tests := []struct {
		reqs      []*Request
		resps     []*Response
		want      []int
		wantError bool
	}{
		{
			reqs: nil, resps: nil, want: []int{}, wantError: false,
		},
		{
			reqs: []*Request{{ID: 1}}, resps: []*Response{{ID: 1}}, want: []int{0},
		},
		{
			reqs: []*Request{{ID: 2}}, resps: []*Response{}, wantError: true,
		},
		{
			reqs: []*Request{}, resps: []*Response{{ID: 3}}, wantError: true,
		},
		{
			reqs: []*Request{{ID: 4}}, resps: []*Response{{ID: 4}, {ID: 4}}, wantError: true,
		},
	}
	for _, test := range tests {
		m, err := mapRespsToReq(test.reqs, test.resps)
		if (err != nil) != test.wantError {
			t.Errorf("got error %v, wantError %v", err, test.wantError)
			continue
		}
		if test.wantError {
			continue
		}
		if !reflect.DeepEqual(m, test.want) {
			t.Errorf("got %v, want %v", m, test.want)
		}
	}
}

func TestReadHeaderContentLength(t *testing.T) {
	s := "Content-Type: foo\r\nContent-Length: 123\r\n\r\n{}"
	n, err := readHeaderContentLength(bufio.NewReader(strings.NewReader(s)))
	if err != nil {
		t.Fatal(err)
	}
	if want := uint32(123); n != want {
		t.Errorf("got %d, want %d", n, want)
	}
}
