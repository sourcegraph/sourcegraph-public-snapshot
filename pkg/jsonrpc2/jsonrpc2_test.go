package jsonrpc2

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestResponseMarshalJSON_JSONRPC(t *testing.T) {
	var r Response
	r.JSONRPC = jsonrpcVersion
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}

	if want := `"jsonrpc":"2.0"`; !strings.Contains(string(b), want) {
		t.Errorf("got %s, want it to include the string %s", b, want)
	}
}

func TestResponseMarshalJSON_Notification(t *testing.T) {
	tests := map[string]bool{
		`{"id":"0"}`:        false,
		`{"id":"1"}`:        false,
		`{"id":"foo"}`:      false,
		`{}`:                true,
		`{"jsonrpc":"2.0"}`: true,
	}
	for s, want := range tests {
		var r Request
		if err := json.Unmarshal([]byte(s), &r); err != nil {
			t.Fatal(err)
		}
		if r.Notification != want {
			t.Errorf("%s: got %v, want %v", s, r.Notification, want)
		}
	}
}

type testHandler struct{}

func (testHandler) Handle(req *Request) *Response {
	if req.Notification {
		return nil // notification
	}
	return &Response{ID: req.ID, Result: toRaw(fmt.Sprintf("hello, #%s: %s", req.ID, *req.Params))}
}

func TestClientServer(t *testing.T) {
	var h testHandler

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
	go func() {
		if err := Serve(lis, h); err != nil {
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
	cl := NewClient(conn)
	defer func() {
		if err := cl.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	// Simple
	for i := 1; i <= 100; i++ {
		resp, err := cl.RequestAndWaitForResponse(Request{ID: strconv.Itoa(i), Notification: false, Params: toRaw([]int32{1, 2, 3})})
		if err != nil {
			t.Fatal(err)
		}
		var got string
		if err := json.Unmarshal(*resp.Result, &got); err != nil {
			t.Fatal(err)
		}
		if want := fmt.Sprintf("hello, #%d: [1,2,3]", i); !reflect.DeepEqual(got, want) {
			t.Errorf("got result %q, want %q", got, want)
		}
	}

	// Batch
	resps, err := cl.RequestBatchAndWaitForAllResponses(
		Request{ID: "1", Notification: false, Params: toRaw([]int32{1})},
		Request{ID: "0", Notification: true, Params: toRaw([]string{"x"})}, // notification
		Request{ID: "2", Notification: false, Params: toRaw([]int32{2})},
		Request{ID: "foo", Notification: false, Params: toRaw([]int32{3})},
	)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]*Response{
		"1":   &Response{ID: "1", Result: toRaw("hello, #1: [1]"), JSONRPC: jsonrpcVersion},
		"2":   &Response{ID: "2", Result: toRaw("hello, #2: [2]"), JSONRPC: jsonrpcVersion},
		"foo": &Response{ID: "foo", Result: toRaw("hello, #foo: [3]"), JSONRPC: jsonrpcVersion},
	}
	if !reflect.DeepEqual(resps, want) {
		t.Errorf("got responses %s, want %s", toString(resps), toString(want))
	}

	lis.Close()
	<-done // ensure Serve's error return (if any) is caught by this test
}

func TestMessageCodec(t *testing.T) {
	tests := []struct {
		v, vempty interface{}
	}{
		{
			v:      &requestOrRequestBatch{single: &Request{ID: "123", Notification: false}},
			vempty: &requestOrRequestBatch{single: &Request{ID: "123", Notification: false}},
		},
		{
			v:      &responseOrResponseBatch{single: &Response{ID: "123"}},
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
			reqs: []*Request{{ID: "1"}}, resps: []*Response{{ID: "1"}}, want: []int{0},
		},
		{
			reqs: []*Request{{ID: "foo"}}, resps: []*Response{{ID: "foo"}}, want: []int{0},
		},
		{
			reqs: []*Request{{ID: "2"}}, resps: []*Response{}, wantError: true,
		},
		{
			reqs: []*Request{}, resps: []*Response{{ID: "3"}}, wantError: true,
		},
		{
			reqs: []*Request{{ID: "4"}}, resps: []*Response{{ID: "4"}, {ID: "4"}}, wantError: true,
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

func toString(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func toRaw(v interface{}) *json.RawMessage {
	b, _ := json.Marshal(v)
	m := json.RawMessage(b)
	return &m
}
