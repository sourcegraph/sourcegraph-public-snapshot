// Package jsonrpc2 provides a client and server implementation of
// [JSON-RPC 2.0](http://www.jsonrpc.org/specification).
package jsonrpc2

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

// Request represents a JSON-RPC request or
// notification. See
// http://www.jsonrpc.org/specification#request_object and
// http://www.jsonrpc.org/specification#notification.
type Request struct {
	Method string           `json:"method"`
	Params *json.RawMessage `json:"params,omitempty"`
	ID     uint64           `json:"id"`
	Notif  bool             `json:"-"`
}

// MarshalJSON implements json.Marshaler and adds the "jsonrpc":"2.0"
// property.
func (r *Request) MarshalJSON() ([]byte, error) {
	if r == nil {
		return nil, errors.New("can't marshal nil *jsonrpc2.Request")
	}
	r2 := struct {
		Method  string           `json:"method"`
		Params  *json.RawMessage `json:"params,omitempty"`
		ID      *uint64          `json:"id,omitempty"`
		JSONRPC string           `json:"jsonrpc"`
	}{
		Method:  r.Method,
		Params:  r.Params,
		JSONRPC: "2.0",
	}
	if !r.Notif {
		r2.ID = &r.ID
	}
	return json.Marshal(r2)
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *Request) UnmarshalJSON(data []byte) error {
	var r2 struct {
		Method string           `json:"method"`
		Params *json.RawMessage `json:"params,omitempty"`
		ID     *uint64          `json:"id"`
	}
	if err := json.Unmarshal(data, &r2); err != nil {
		return err
	}
	r.Method = r2.Method
	r.Params = r2.Params
	if r2.ID == nil {
		r.ID = 0
		r.Notif = true
	} else {
		r.ID = *r2.ID
		r.Notif = false
	}
	return nil
}

// SetParams sets r.Params to the JSON representation of v. If JSON
// marshaling fails, it returns an error.
func (r *Request) SetParams(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	r.Params = (*json.RawMessage)(&b)
	return nil
}

// Response represents a JSON-RPC response. See
// http://www.jsonrpc.org/specification#response_object.
type Response struct {
	ID     uint64           `json:"id"`
	Result *json.RawMessage `json:"result,omitempty"`
	Error  *Error           `json:"error,omitempty"`

	// SPEC NOTE: The spec says "If there was an error in detecting
	// the id in the Request object (e.g. Parse error/Invalid
	// Request), it MUST be Null." If we made the ID field nullable,
	// then we'd have to make it a pointer type. For simplicity, we're
	// ignoring the case where there was an error in detecting the ID
	// in the Request object.
}

// MarshalJSON implements json.Marshaler and adds the "jsonrpc":"2.0"
// property.
func (r *Response) MarshalJSON() ([]byte, error) {
	if r == nil {
		return nil, errors.New("can't marshal nil *jsonrpc2.Response")
	}
	b, err := json.Marshal(*r)
	if err != nil {
		return nil, err
	}
	b = append(b[:len(b)-1], []byte(`,"jsonrpc":"2.0"}`)...)
	return b, nil
}

// SetResult sets r.Result to the JSON representation of v. If JSON
// marshaling fails, it returns an error.
func (r *Response) SetResult(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	r.Result = (*json.RawMessage)(&b)
	return nil
}

// Error represents a JSON-RPC response error.
type Error struct {
	Code    int64            `json:"code"`
	Message string           `json:"message"`
	Data    *json.RawMessage `json:"data"`
}

// SetError sets e.Error to the JSON representation of v. If JSON
// marshaling fails, it panics.
func (e *Error) SetError(v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		panic("Error.SetData: " + err.Error())
	}
	e.Data = (*json.RawMessage)(&b)
}

// Error implements the Go error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("jsonrpc2: code %v message: %s", e.Code, e.Message)
}

const (
	// Errors defined in the JSON-RPC spec. See http://www.jsonrpc.org/specification#error_object.
	CodeParseError       = -32700
	CodeInvalidRequest   = -32600
	CodeMethodNotFound   = -32601
	CodeInvalidParams    = -32602
	CodeInternalError    = -32603
	codeServerErrorStart = -32099
	codeServerErrorEnd   = -32000
)

// Handler handles JSON-RPC requests and notifications.
type Handler interface {
	// Handle is called to handle a request.
	Handle(context.Context, *Conn, *Request)
}

// Conn is a JSON-RPC client/server connection. The JSON-RPC protocol
// is symmetric, so a Conn runs on both ends of a client-server
// connection.
type Conn struct {
	conn io.Closer // all writes should go through w, all reads through readMessages
	w    *bufio.Writer

	h Handler

	mu       sync.Mutex
	shutdown bool
	closing  bool
	seq      uint64
	pending  map[uint64]*call

	sending sync.Mutex

	disconnect chan struct{}

	// Set by ConnOpt funcs.
	onRecv func(*Request, *Response)
	onSend func(*Request, *Response)
}

// ErrClosed indicates that the JSON-RPC connection is closed (or in
// the process of closing).
var ErrClosed = errors.New("jsonrpc2: connection is closed")

// NewConn creates a new JSON-RPC client/server connection using the
// given ReadWriteCloser (typically a TCP connection or stdio). The
// JSON-RPC protocol is symmetric, so a Conn runs on both ends of a
// client-server connection.
//
// NewClient consumes conn, so you should call Close on the returned
// client not on the given conn.
func NewConn(ctx context.Context, conn io.ReadWriteCloser, h Handler, opt ...ConnOpt) *Conn {
	c := &Conn{
		conn:       conn,
		w:          bufio.NewWriter(conn),
		h:          h,
		pending:    map[uint64]*call{},
		disconnect: make(chan struct{}),
	}
	for _, opt := range opt {
		opt(c)
	}
	go c.readMessages(ctx, bufio.NewReader(conn))
	return c
}

// Close closes the JSON-RPC connection. The connection may not be
// used after it has been closed.
func (c *Conn) Close() error {
	c.mu.Lock()
	if c.shutdown || c.closing {
		c.mu.Unlock()
		return ErrClosed
	}
	c.closing = true
	c.mu.Unlock()
	return c.conn.Close()
}

func (c *Conn) send(ctx context.Context, m *anyMessage, wait bool) (*call, error) {
	c.sending.Lock()
	defer c.sending.Unlock()

	c.mu.Lock()
	if c.shutdown || c.closing {
		c.mu.Unlock()
		return nil, ErrClosed
	}

	// Store requests so we can later associate them with incoming
	// responses.
	var cc *call
	if m.request != nil && wait {
		cc = &call{request: m.request, seq: c.seq, done: make(chan error)}
		c.pending[c.seq] = cc // use first seq as call ID for batch
		for _, req := range m.requests() {
			req.ID = c.seq
			c.seq++
		}
	}
	c.mu.Unlock()

	if c.onSend != nil {
		switch {
		case m.request != nil:
			if m.request.batch != nil {
				panic("batching not yet implemented")
			}
			c.onSend(m.request.single, nil)
		case m.response != nil:
			if m.response.batch != nil {
				panic("batching not yet implemented")
			}
			c.onSend(nil, m.response.single)
		}
	}

	err := marshalHeadersAndBody(c.w, m)
	if err != nil {
		c.w.Flush()
		if cc != nil {
			c.mu.Lock()
			delete(c.pending, cc.seq)
			c.mu.Unlock()
		}
		return nil, err
	}
	return cc, c.w.Flush()
}

// Call initiates a JSON-RPC call using the specified method and
// params, and waits for the response. If the response is successful,
// its result is stored in result (a pointer to a value that can be
// JSON-unmarshaled into); otherwise, a non-nil error is returned.
func (c *Conn) Call(ctx context.Context, method string, params, result interface{}) error {
	req := &Request{Method: method}
	if err := req.SetParams(params); err != nil {
		return err
	}
	call, err := c.send(ctx, &anyMessage{request: &requestOrRequestBatch{single: req}}, true)
	if err != nil {
		return err
	}
	select {
	case err, ok := <-call.done:
		if !ok {
			err = ErrClosed
		}
		if err != nil {
			return err
		}
		if result != nil && call.response.single.Result != nil {
			// TODO(sqs): error handling
			if err := json.Unmarshal(*call.response.single.Result, result); err != nil {
				return err
			}
		}
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}

// Notify is like Call, but it returns when the notification request
// is sent (without waiting for a response, because JSON-RPC
// notifications do not have responses).
func (c *Conn) Notify(ctx context.Context, method string, params interface{}) error {
	req := &Request{Method: method, Notif: true}
	if err := req.SetParams(params); err != nil {
		return err
	}
	_, err := c.send(ctx, &anyMessage{request: &requestOrRequestBatch{single: req}}, false)
	return err
}

// Reply sends a successful response with a result.
func (c *Conn) Reply(ctx context.Context, id uint64, result interface{}) error {
	resp := &Response{ID: id}
	if err := resp.SetResult(result); err != nil {
		return err
	}
	_, err := c.send(ctx, &anyMessage{response: &responseOrResponseBatch{single: resp}}, false)
	return err
}

// ReplyWithError sends a response with an error.
func (c *Conn) ReplyWithError(ctx context.Context, id uint64, respErr *Error) error {
	_, err := c.send(ctx, &anyMessage{response: &responseOrResponseBatch{single: &Response{ID: id, Error: respErr}}}, false)
	return err
}

// SendResponse sends resp to the peer. It is lower level than (*Conn).Reply.
func (c *Conn) SendResponse(ctx context.Context, resp *Response) error {
	_, err := c.send(ctx, &anyMessage{response: &responseOrResponseBatch{single: resp}}, false)
	return err
}

// DisconnectNotify returns a channel that is closed when the
// underlying connection is disconnected.
func (c *Conn) DisconnectNotify() <-chan struct{} {
	return c.disconnect
}

func (c *Conn) readMessages(ctx context.Context, r *bufio.Reader) {
	var err error
	for err == nil {
		var m anyMessage

		var n uint32
		n, err = readHeaderContentLength(r)
		if err == nil {
			err = json.NewDecoder(io.LimitReader(r, int64(n))).Decode(&m)
		}
		if err != nil {
			break
		}

		switch {
		case m.request != nil:
			switch {
			case m.request.batch != nil:
				panic("batching not yet implemented")

			case m.request.single != nil:
				if c.onRecv != nil {
					c.onRecv(m.request.single, nil)
				}
				go c.h.Handle(ctx, c, m.request.single)
			}

		case m.response != nil:
			resp := *m.response
			if resp := resp.single; resp != nil {
				seq := resp.ID
				c.mu.Lock()
				call := c.pending[seq]
				delete(c.pending, seq)
				c.mu.Unlock()

				if call != nil {
					call.response = &responseOrResponseBatch{single: resp}
				}

				if c.onRecv != nil {
					var req *Request

					if call != nil {
						if call.request.batch != nil {
							panic("batching not yet implemented")
						}
						req = call.request.single
					}
					c.onRecv(req, resp)
				}

				switch {
				case call == nil:
					log.Printf("jsonrpc2: ignoring response %d with no corresponding request", seq)

				case resp.Error != nil:
					call.done <- resp.Error
					close(call.done)

				default:
					call.done <- nil
					close(call.done)
				}
			} else {
				panic("batches are not yet implemented") // TODO(sqs): support batches
			}
		}
	}

	c.sending.Lock()
	c.mu.Lock()
	c.shutdown = true
	closing := c.closing
	if err == io.EOF {
		if closing {
			err = ErrClosed
		} else {
			err = io.ErrUnexpectedEOF
		}
	}
	for _, call := range c.pending {
		call.done <- err
		close(call.done)
	}
	c.mu.Unlock()
	c.sending.Unlock()
	if err != io.ErrUnexpectedEOF && !closing {
		log.Println("jsonrpc2: protocol error:", err)
	}
	close(c.disconnect)
}

// Server is a JSON-RPC server.
type Server struct{}

// Serve starts a new JSON-RPC server.
func Serve(ctx context.Context, lis net.Listener, h Handler, opt ...ConnOpt) error {
	for {
		conn, err := lis.Accept()
		if err != nil {
			return err
		}
		NewConn(ctx, conn, h, opt...)
	}
}

// mapRespsToReq returns a slice whose i'th element reports the index
// in reqs of the i'th response in resps.
//
// It returns an error if a response's ID does not refer to that of a
// request in reqs, or if two responses have the same ID, or if there
// is a request (non-notification) that does not have a corresponding
// response.
func mapRespsToReq(reqs []*Request, resps []*Response) ([]int, error) {
	reqIndexByID := make(map[uint64]int, len(reqs))
	for i, req := range reqs {
		if !req.Notif {
			reqIndexByID[req.ID] = i
		}
	}

	if len(resps) != len(reqIndexByID) {
		return nil, fmt.Errorf("jsonrpc2: response batch too small: %d responses for %d non-notification requests", len(resps), len(reqIndexByID))
	}

	m := make([]int, len(resps))
	seenIDs := make(map[uint64]struct{}, len(resps))
	for i, resp := range resps {
		reqIndex, present := reqIndexByID[resp.ID]
		if !present {
			return nil, fmt.Errorf("jsonrpc2: response batch contains response with ID %d that doesn't match any IDs in request batch", resp.ID)
		}
		m[i] = reqIndex

		if _, seen := seenIDs[resp.ID]; seen {
			return nil, fmt.Errorf("jsonrpc2: response batch contains multiple responses with same ID %d", resp.ID)
		}
		seenIDs[resp.ID] = struct{}{}
	}

	return m, nil
}

// call represents a JSON-RPC call over its entire lifecycle.
type call struct {
	request  *requestOrRequestBatch
	response *responseOrResponseBatch
	seq      uint64 // the seq of the request (or first request for a batch)
	done     chan error
}

// anyMessage represents either a JSON Request or Response, or a batch
// thereof.
type anyMessage struct {
	request  *requestOrRequestBatch
	response *responseOrResponseBatch
}

func (m *anyMessage) requests() []*Request {
	if m.request.single != nil {
		return []*Request{m.request.single}
	}
	return m.request.batch
}

func (m *anyMessage) responses() []*Response {
	if m.response.single != nil {
		return []*Response{m.response.single}
	}
	return m.response.batch
}

func (m *anyMessage) MarshalJSON() ([]byte, error) {
	var v interface{}
	switch {
	case m.request != nil && m.response == nil:
		v = m.request
	case m.request == nil && m.response != nil:
		v = m.response
	}
	if v != nil {
		return json.Marshal(v)
	}
	return nil, errors.New("jsonrpc2: message (or each message in batch) must have exactly one of the request or response fields set")
}

func (m *anyMessage) UnmarshalJSON(data []byte) error {
	// The presence of these fields distinguishes between the 2
	// message types.
	type msg struct {
		Method *string     `json:"method"`
		Result interface{} `json:"result"`
		Error  interface{} `json:"error"`
	}

	var isRequest, isResponse bool
	checkType := func(m *msg) error {
		mIsRequest := m.Method != nil
		mIsResponse := m.Result != nil || m.Error != nil
		if (!mIsRequest && !mIsResponse) || (mIsRequest && mIsResponse) {
			return errors.New("jsonrpc2: unable to determine message type (request or response)")
		}
		if (mIsRequest && isResponse) || (mIsResponse && isRequest) {
			return errors.New("jsonrpc2: batch message type mismatch (must be all requests or all responses)")
		}
		isRequest = mIsRequest
		isResponse = mIsResponse
		return nil
	}

	if isArray := len(data) > 0 && data[0] == '['; isArray {
		var msgs []msg
		if err := json.Unmarshal(data, &msgs); err != nil {
			return err
		}
		if len(msgs) == 0 {
			return errors.New("jsonrpc2: invalid empty batch")
		}
		for _, msg := range msgs {
			if err := checkType(&msg); err != nil {
				return err
			}
		}
	} else {
		var msg msg
		if err := json.Unmarshal(data, &msg); err != nil {
			return err
		}
		if err := checkType(&msg); err != nil {
			return err
		}
	}

	var v interface{}
	switch {
	case isRequest && !isResponse:
		v = &m.request
	case !isRequest && isResponse:
		v = &m.response
	}
	return json.Unmarshal(data, v)
}

type requestOrRequestBatch struct {
	batch  []*Request
	single *Request
}

func (v *requestOrRequestBatch) MarshalJSON() ([]byte, error) {
	if v.single != nil {
		return json.Marshal(v.single)
	}
	return json.Marshal(v.batch)
}

func (v *requestOrRequestBatch) UnmarshalJSON(data []byte) error {
	data = bytes.TrimLeft(data, " \t\n\r")
	if len(data) == 0 {
		return errInvalidRequestJSON
	}
	switch data[0] {
	case '[':
		*v = requestOrRequestBatch{}
		if err := json.Unmarshal(data, &v.batch); err != nil {
			return err
		}
		return nil

	case '{':
		*v = requestOrRequestBatch{}
		if err := json.Unmarshal(data, &v.single); err != nil {
			return err
		}
		return nil

	default:
		return errInvalidRequestJSON
	}
}

type responseOrResponseBatch struct {
	batch  []*Response
	single *Response
}

func (v *responseOrResponseBatch) MarshalJSON() ([]byte, error) {
	if v.single != nil {
		return json.Marshal(v.single)
	}
	return json.Marshal(v.batch)
}

func (v *responseOrResponseBatch) UnmarshalJSON(data []byte) error {
	data = bytes.TrimLeft(data, " \t\n\r")
	if len(data) == 0 {
		return errInvalidResponseJSON
	}
	switch data[0] {
	case '[':
		*v = responseOrResponseBatch{}
		if err := json.Unmarshal(data, &v.batch); err != nil {
			return err
		}
		return nil

	case '{':
		*v = responseOrResponseBatch{}
		if err := json.Unmarshal(data, &v.single); err != nil {
			return err
		}
		return nil

	default:
		return errInvalidResponseJSON
	}
}

func readHeaderContentLength(r *bufio.Reader) (contentLength uint32, err error) {
	for {
		line, err := r.ReadString('\r')
		if err != nil {
			return 0, err
		}
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		if b != '\n' {
			return 0, fmt.Errorf(`jsonrpc2: line endings must be \r\n`)
		}
		if line == "\r" {
			break
		}
		if strings.HasPrefix(line, "Content-Length: ") {
			line = strings.TrimPrefix(line, "Content-Length: ")
			line = strings.TrimSpace(line)
			n, err := strconv.ParseUint(line, 10, 32)
			if err != nil {
				return 0, err
			}
			contentLength = uint32(n)
		}
	}
	if contentLength == 0 {
		err = fmt.Errorf("jsonrpc2: no Content-Length header found")
	}
	return
}

func marshalHeadersAndBody(w io.Writer, v interface{}) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "Content-Length: %d\r\n", len(body))
	fmt.Fprint(w, "Content-Type: application/vscode-jsonrpc; charset=utf8\r\n\r\n")
	_, err = w.Write(body)
	return err
}

var (
	errInvalidRequestJSON  = errors.New("jsonrpc2: request must be either a JSON object or JSON array")
	errInvalidResponseJSON = errors.New("jsonrpc2: response must be either a JSON object or JSON array")
)
