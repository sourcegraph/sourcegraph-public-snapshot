// Input/Output operations for streams produced by DevTools. (experimental)
package io

import (
	"github.com/neelance/cdp-go/rpc"
)

// Input/Output operations for streams produced by DevTools. (experimental)
type Client struct {
	*rpc.Client
}

type StreamHandle string

type ReadRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Read a chunk of the stream
func (d *Client) Read() *ReadRequest {
	return &ReadRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Handle of the stream to read.
func (r *ReadRequest) Handle(v StreamHandle) *ReadRequest {
	r.opts["handle"] = v
	return r
}

// Seek to the specified offset before reading (if not specificed, proceed with offset following the last read). (optional)
func (r *ReadRequest) Offset(v int) *ReadRequest {
	r.opts["offset"] = v
	return r
}

// Maximum number of bytes to read (left upon the agent discretion if not specified). (optional)
func (r *ReadRequest) Size(v int) *ReadRequest {
	r.opts["size"] = v
	return r
}

type ReadResult struct {
	// Data that were read.
	Data string `json:"data"`

	// Set if the end-of-file condition occured while reading.
	Eof bool `json:"eof"`
}

func (r *ReadRequest) Do() (*ReadResult, error) {
	var result ReadResult
	err := r.client.Call("IO.read", r.opts, &result)
	return &result, err
}

type CloseRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Close the stream, discard any temporary backing storage.
func (d *Client) Close() *CloseRequest {
	return &CloseRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Handle of the stream to close.
func (r *CloseRequest) Handle(v StreamHandle) *CloseRequest {
	r.opts["handle"] = v
	return r
}

func (r *CloseRequest) Do() error {
	return r.client.Call("IO.close", r.opts, nil)
}

func init() {
}
