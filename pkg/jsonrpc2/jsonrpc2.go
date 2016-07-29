// Package jsonrpc2 provides a client and server implementation of
// [JSON-RPC 2.0](http://www.jsonrpc.org/specification).
package jsonrpc2

import (
	"bufio"
	"bytes"
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

const jsonrpcVersion = "2.0"

// Request represents a JSON-RPC request or
// notification. See
// http://www.jsonrpc.org/specification#request_object and
// http://www.jsonrpc.org/specification#notification.
type Request struct {
	Method       string
	Params       *json.RawMessage
	ID           string
	Notification bool
	JSONRPC      string // "2.0"
}

func (r *Request) MarshalJSON() ([]byte, error) {
	// Override to omit ID if Notification.
	m := map[string]interface{}{
		"method":  r.Method,
		"jsonrpc": r.JSONRPC,
	}
	if r.Params != nil {
		m["params"] = r.Params
	}
	if !r.Notification {
		m["id"] = r.ID
	}
	return json.Marshal(m)
}

func (r *Request) UnmarshalJSON(data []byte) error {
	var m struct {
		Method  string           `json:"method"`
		Params  *json.RawMessage `json:"params"`
		ID      *string          `json:"id,omitempty"`
		JSONRPC string           `json:"jsonrpc"` // "2.0"
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*r = Request{
		Method:  m.Method,
		Params:  m.Params,
		JSONRPC: m.JSONRPC,
	}
	if m.ID != nil {
		(*r).ID = *m.ID
	} else {
		(*r).Notification = true
	}
	return nil
}

// SetParams sets r.Params to the JSON representation of v. If JSON
// marshaling fails, it panics.
func (r *Request) SetParams(v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		panic("Response.SetParams: " + err.Error())
	}
	r.Params = (*json.RawMessage)(&b)
}

// Response represents a JSON-RPC response. See
// http://www.jsonrpc.org/specification#response_object.
type Response struct {
	Result  *json.RawMessage `json:"result,omitempty"`
	Error   *Error           `json:"error,omitempty"`
	ID      string           `json:"id"`
	JSONRPC string           `json:"jsonrpc"` // "2.0"

	// SPEC NOTE: The spec says "If there was an error in detecting
	// the id in the Request object (e.g. Parse error/Invalid
	// Request), it MUST be Null." If we made the ID field nullable,
	// then we'd have to make it a pointer type. For simplicity, we're
	// ignoring the case where there was an error in detecting the ID
	// in the Request object.
}

// SetResult sets r.Result to the JSON representation of v. If JSON
// marshaling fails, it panics.
func (r *Response) SetResult(v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		panic("Response.SetResult: " + err.Error())
	}
	r.Result = (*json.RawMessage)(&b)
}

// Error represents a JSON-RPC response error.
type Error struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    *json.RawMessage
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

// Client is a JSON-RPC client.
type Client struct {
	conn io.ReadWriteCloser

	mu    sync.Mutex
	resps chan Response // only set if someone is in a RequestAndWaitForResponse call
}

// NewClient creates a new JSON-RPC client using the given ReadWriteCloser
// (typically a TCP connection or stdio).
//
// NewClient consumes conn, you should call Close on the returned client not
// on the given conn.
func NewClient(conn io.ReadWriteCloser) *Client {
	c := &Client{conn: conn}
	go c.readResponses()
	return c
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) readResponses() {
	r := bufio.NewReader(c.conn)
	for {
		var resp responseOrResponseBatch

		n, err := readHeaderContentLength(r)
		if err == nil {
			err = json.NewDecoder(io.LimitReader(r, int64(n))).Decode(&resp)
		}
		if err != nil {
			if isFatalConnError(err) {
				if err != io.EOF && !strings.Contains(err.Error(), "use of closed network connection") {
					log.Println("jsonrpc2 client:", err)
				}
				c.conn.Close()
				break
			}
			log.Println("jsonrpc2 client:", err)
			continue
		}

		c.mu.Lock()
		if c.resps != nil {
			if resp.single != nil {
				c.resps <- *resp.single
			} else {
				for _, resp := range resp.batch {
					c.resps <- resp
				}
			}
		}
		c.mu.Unlock()
	}
}

func (c *Client) send(body interface{}) error {
	w := bufio.NewWriter(c.conn)
	if err := marshalHeadersAndBody(w, body); err != nil {
		w.Flush()
		return err
	}
	return w.Flush()
}

// Request sends req to the server. It does not wait for the server to
// send the corresponding JSON-RPC response object. An error is
// returned if the request is unable to be sent. To wait for the
// corresponding response, use RequestAndWaitForResponse.
func (c *Client) Request(req Request) error {
	req.JSONRPC = jsonrpcVersion
	return c.send(req)
}

// RequestAndWaitForResponse sends req to the server and waits for the
// corresponding response (where resp.ID == req.ID). If req is a
// notification, call Request instead, since the server will never
// respond to notifications, and this method would block forever.
//
// It is NOT safe to call this method from more than one goroutine
// concurrently.
func (c *Client) RequestAndWaitForResponse(req Request) (*Response, error) {
	if req.Notification {
		panic("jsonrpc2: Request ID must be set (if req is a notification, call Request to avoid a futile wait for a response)")
	}

	// Open this channel until we receive the response we're looking
	// for. readResponses will send us the responses on the channel if
	// it's non-nil.
	c.mu.Lock()
	c.resps = make(chan Response)
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		close(c.resps)
		c.resps = nil
		c.mu.Unlock()
	}()

	if err := c.Request(req); err != nil {
		return nil, err
	}

	for resp := range c.resps {
		if resp.ID == req.ID {
			return &resp, nil
		}
	}
	panic("unreachable") // above loop is infinite until condition is met
}

// RequestBatch is the batched version of Request.
func (c *Client) RequestBatch(reqs ...Request) error {
	for _, req := range reqs {
		req.JSONRPC = jsonrpcVersion
	}
	return c.send(reqs)
}

// RequestBatchAndWaitForAllResponses is the batched version of
// RequestAndWaitForResponse.
func (c *Client) RequestBatchAndWaitForAllResponses(reqs ...Request) (map[string]*Response, error) {
	wantResps := make(map[string]*Response, len(reqs)) // overallocation if reqs has notifs, but that's fine
	for _, req := range reqs {
		if !req.Notification {
			wantResps[req.ID] = nil
		}
	}

	// Open this channel until we receive the response we're looking
	// for. readResponses will send us the responses on the channel if
	// it's non-nil.
	c.mu.Lock()
	c.resps = make(chan Response)
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		close(c.resps)
		c.resps = nil
		c.mu.Unlock()
	}()

	if err := c.RequestBatch(reqs...); err != nil {
		return nil, err
	}

	remaining := len(wantResps)
	for resp := range c.resps {
		if existing, present := wantResps[resp.ID]; present {
			if existing != nil {
				return nil, fmt.Errorf("jsonrpc2: duplicate response received with ID %s", resp.ID)
			}
			tmp := resp
			wantResps[resp.ID] = &tmp
			remaining--
		}
		if remaining == 0 {
			return wantResps, nil
		}
	}
	panic("unreachable") // above loop is infinite until
}

// Handler handles JSON-RPC requests and notifications.
type Handler interface {
	// Handle handles a request (if !Request.Notification) or notification
	// (if Request.Notification). If the argument is a request, it must
	// return a non-nil response. If the argument is a notification,
	// it must return a nil response.
	//
	// If the client sent a batch request, and the handler doesn't
	// also implement BatchHandler, Handle will be called once for
	// each element of the batch, in the same order the requests
	// appeared in the batch.
	Handle(*Request) *Response
}

// BatchHandler is like Handler, but it also operates on entire
// batches at once.  Like Handler, it returns a nil response for
// notifications.
type BatchHandler interface {
	Handler
	HandleBatch([]*Request) []*Response
}

// Server is a JSON-RPC server.
type Server struct{}

// Serve starts a new JSON-RPC server.
func Serve(lis net.Listener, h Handler) error {
	for {
		conn, err := lis.Accept()
		if err != nil {
			return err
		}
		NewServerConn(conn, conn, h)
	}
}

// NewServerConn creates a server for a single client. It can also be
// used to create an interactive program that uses standard
// input/output, if os.Stdin and os.Stdout are passed.
func NewServerConn(r io.ReadCloser, w io.WriteCloser, h Handler) {
	sc := &serverConn{r: r, w: w, h: h}
	sc.serveConn()
}

const serverConnBufSize = 20

type serverConn struct {
	r io.ReadCloser
	w io.WriteCloser
	h Handler

	reqs chan requestOrRequestBatch // batches or 1-element slices of single requests
}

func (sc *serverConn) serveConn() {
	sc.reqs = make(chan requestOrRequestBatch, serverConnBufSize)

	go sc.readRequests()
	go sc.handleRequests()

	// TODO(sqs) clean up reqs, conn on exit
}

func (sc *serverConn) readRequests() {
	r := bufio.NewReader(sc.r)
	for {
		var req requestOrRequestBatch

		n, err := readHeaderContentLength(r)
		if err == nil {
			err = json.NewDecoder(io.LimitReader(r, int64(n))).Decode(&req)
		}
		if err != nil {
			if isFatalConnError(err) {
				if err != io.EOF {
					log.Println("jsonrpc2 server:", err)
				}
				close(sc.reqs)
				sc.r.Close()
				sc.w.Close()
				break
			}
			log.Println("jsonrpc2 server:", err)
			// TODO(sqs): return server error to client
			continue
		}

		sc.reqs <- req
	}
}

func checkResponse(req *Request, resp *Response) {
	if req.Notification {
		if resp != nil {
			panic("jsonrpc2: Handle returned non-nil response for notification (must be nil)")
		}
	} else {
		if resp.ID != req.ID {
			panic(fmt.Sprintf("jsonrpc2: Handle set an invalid response ID %s", resp.ID))
		}
		if resp.Result != nil && resp.Error != nil {
			panic("jsonrpc2: Handle returned non-nil result and error (exactly 1 must be non-nil)")
		}
		if resp.Result == nil && resp.Error == nil {
			panic("jsonrpc2: Handle returned nil result and error (exactly 1 must be non-nil)")
		}

		resp.JSONRPC = jsonrpcVersion
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
	reqIndexByID := make(map[string]int, len(reqs))
	for i, req := range reqs {
		if !req.Notification {
			reqIndexByID[req.ID] = i
		}
	}

	if len(resps) != len(reqIndexByID) {
		return nil, fmt.Errorf("jsonrpc2: response batch too small: %d responses for %d non-notification requests", len(resps), len(reqIndexByID))
	}

	m := make([]int, len(resps))
	seenIDs := make(map[string]struct{}, len(resps))
	for i, resp := range resps {
		reqIndex, present := reqIndexByID[resp.ID]
		if !present {
			return nil, fmt.Errorf("jsonrpc2: response batch contains response with ID %s that doesn't match any IDs in request batch", resp.ID)
		}
		m[i] = reqIndex

		if _, seen := seenIDs[resp.ID]; seen {
			return nil, fmt.Errorf("jsonrpc2: response batch contains multiple responses with same ID %s", resp.ID)
		}
		seenIDs[resp.ID] = struct{}{}
	}

	return m, nil
}

func (sc *serverConn) handleRequests() {
	w := bufio.NewWriter(sc.w)
reqLoop:
	for reqSingleOrBatch := range sc.reqs {
		var body interface{}
		if req := reqSingleOrBatch.single; req != nil {
			// Single (non-batch)
			resp := sc.h.Handle(req)
			checkResponse(req, resp)
			if req.Notification {
				continue
			}
			body = resp
		} else {
			// Batch
			reqs := reqSingleOrBatch.batch

			{
				// Validate that there aren't any duplicate request IDs.
				reqID := make(map[string]struct{}, len(reqs))
				for _, req := range reqs {
					if req.Notification {
						continue
					}
					if _, present := reqID[req.ID]; present {
						// TODO(sqs): return error Response object(s) instead?
						log.Printf("jsonrpc2: ignoring invalid request batch containing multiple requests with same ID %s", req.ID)
						continue reqLoop
					}
					reqID[req.ID] = struct{}{}
				}
			}

			bh, ok := sc.h.(BatchHandler)
			if !ok {
				bh = batchHandlerWrapper{sc.h}
			}
			resps := bh.HandleBatch(reqs)

			respIndexToReqIndex, err := mapRespsToReq(reqs, resps)
			if err != nil {
				// All errors are due to certain bugs in the Handler
				// and are not able to be triggered by clients when
				// the Handler is correct, so panicking is acceptable.
				panic(err)
			}
			for i, resp := range resps {
				req := reqs[respIndexToReqIndex[i]]
				checkResponse(req, resp)
			}

			body = resps
		}
		if err := marshalHeadersAndBody(w, body); err != nil {
			w.Flush()
			// TODO(sqs): return server error
			log.Println("jsonrpc2 server:", err)
			continue
		}
		if err := w.Flush(); err != nil {
			// TODO(sqs): return server error
			log.Println("jsonrpc2 server:", err)
			continue
		}
	}
}

// batchHandlerWrapper makes any Handler implement BatchHandler by
// just calling its Handle method for each request in the
// batch. Handlers can implement BatchHandler on their own to provide
// optimized batch handling or batch handling with different behavior
// than this wrapper's serial handling.
type batchHandlerWrapper struct{ Handler }

func (h batchHandlerWrapper) HandleBatch(reqs []*Request) []*Response {
	resps := make([]*Response, 0, len(reqs))
	for _, req := range reqs {
		resp := h.Handler.Handle(req)
		if resp != nil {
			checkResponse(req, resp)
			resps = append(resps, resp)
		}
	}

	return resps
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
	batch  []Response
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

func isFatalConnError(err error) bool {
	if err, ok := err.(net.Error); ok && !err.Temporary() {
		return true
	}
	return err == io.EOF
}
