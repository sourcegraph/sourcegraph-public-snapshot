// Copyright (c) 2015-2016 Marin Atanasov Nikolov <dnaeon@gmail.com>
// Copyright (c) 2016 David Jack <davars@gmail.com>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer
//    in this position and unchanged.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE AUTHOR(S) ``AS IS'' AND ANY EXPRESS OR
// IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
// OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
// IN NO EVENT SHALL THE AUTHOR(S) BE LIABLE FOR ANY DIRECT, INDIRECT,
// INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
// NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
// THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package recorder

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"time"

	"github.com/dnaeon/go-vcr/cassette"
)

// Mode represents recording/playback mode
type Mode int

// Recorder states
const (
	ModeRecording Mode = iota
	ModeReplaying
	ModeDisabled
	// Replay record from cassette or record a new one when a request is not
	// present in cassette instead of throwing ErrInteractionNotFound
	ModeReplayingOrRecording
)

// Recorder represents a type used to record and replay
// client and server interactions
type Recorder struct {
	// Operating mode of the recorder
	mode Mode

	// Cassette used by the recorder
	cassette *cassette.Cassette

	// realTransport is the underlying http.RoundTripper to make real requests
	realTransport http.RoundTripper

	// Pass through requests.
	Passthroughs []Passthrough
}

// Passthrough function allows ignoring certain requests.
type Passthrough func(*http.Request) bool

// SetTransport can be used to configure the behavior of the 'real' client used in record-mode
func (r *Recorder) SetTransport(t http.RoundTripper) {
	r.realTransport = t
}

// Proxies client requests to their original destination
func requestHandler(r *http.Request, c *cassette.Cassette, mode Mode, realTransport http.RoundTripper) (*cassette.Interaction, error) {
	// Return interaction from cassette if in replay mode or replay/record mode
	if mode == ModeReplaying || mode == ModeReplayingOrRecording {
		if err := r.Context().Err(); err != nil {
			return nil, err
		}

		if interaction, err := c.GetInteraction(r); mode == ModeReplaying {
			return interaction, err
		} else if mode == ModeReplayingOrRecording && err == nil {
			return interaction, err
		}
	}

	// Copy the original request, so we can read the form values
	reqBytes, err := httputil.DumpRequestOut(r, true)
	if err != nil {
		return nil, err
	}

	reqBuffer := bytes.NewBuffer(reqBytes)
	copiedReq, err := http.ReadRequest(bufio.NewReader(reqBuffer))
	if err != nil {
		return nil, err
	}

	err = copiedReq.ParseForm()
	if err != nil {
		return nil, err
	}

	reqBody := &bytes.Buffer{}
	if r.Body != nil && !isNoBody(r.Body) {
		// Record the request body so we can add it to the cassette
		r.Body = ioutil.NopCloser(io.TeeReader(r.Body, reqBody))
	}

	// Perform client request to it's original
	// destination and record interactions
	resp, err := realTransport.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Add interaction to cassette
	interaction := &cassette.Interaction{
		Request: cassette.Request{
			Body:    reqBody.String(),
			Form:    copiedReq.PostForm,
			Headers: r.Header,
			URL:     r.URL.String(),
			Method:  r.Method,
		},
		Response: cassette.Response{
			Body:    string(respBody),
			Headers: resp.Header,
			Status:  resp.Status,
			Code:    resp.StatusCode,
		},
	}
	for _, filter := range c.Filters {
		err = filter(interaction)
		if err != nil {
			return nil, err
		}
	}
	c.AddInteraction(interaction)

	return interaction, nil
}

// New creates a new recorder
func New(cassetteName string) (*Recorder, error) {
	// Default mode is "replay" if file exists
	return NewAsMode(cassetteName, ModeReplaying, nil)
}

// NewAsMode creates a new recorder in the specified mode
func NewAsMode(cassetteName string, mode Mode, realTransport http.RoundTripper) (*Recorder, error) {
	var c *cassette.Cassette
	cassetteFile := fmt.Sprintf("%s.yaml", cassetteName)

	if mode != ModeDisabled {
		// Depending on whether the cassette file exists or not we
		// either create a new empty cassette or load from file
		if _, err := os.Stat(cassetteFile); os.IsNotExist(err) || mode == ModeRecording {
			// Create new cassette and enter in recording mode
			c = cassette.New(cassetteName)
			mode = ModeRecording
		} else {
			// Load cassette from file and enter replay mode or replay/record mode
			c, err = cassette.Load(cassetteName)
			if err != nil {
				return nil, err
			}
		}
	}

	if realTransport == nil {
		realTransport = http.DefaultTransport
	}

	r := &Recorder{
		mode:          mode,
		cassette:      c,
		realTransport: realTransport,
	}

	return r, nil
}

// Stop is used to stop the recorder and save any recorded interactions
func (r *Recorder) Stop() error {
	if r.mode == ModeRecording || r.mode == ModeReplayingOrRecording {
		if err := r.cassette.Save(); err != nil {
			return err
		}
	}

	return nil
}

// RoundTrip implements the http.RoundTripper interface
func (r *Recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	if r.mode == ModeDisabled {
		return r.realTransport.RoundTrip(req)
	}
	for _, passthrough := range r.Passthroughs {
		if passthrough(req) {
			return r.realTransport.RoundTrip(req)
		}
	}

	// Pass cassette and mode to handler, so that interactions can be
	// retrieved or recorded depending on the current recorder mode
	interaction, err := requestHandler(req, r.cassette, r.mode, r.realTransport)

	if err != nil {
		return nil, err
	}

	select {
	case <-req.Context().Done():
		return nil, req.Context().Err()
	default:
		buf := bytes.NewBuffer([]byte(interaction.Response.Body))
		// apply the duration defined in the interaction
		if interaction.Response.Duration != "" {
			d, err := time.ParseDuration(interaction.Duration)
			if err != nil {
				return nil, err
			}
			// block for the configured 'duration' to simulate the network latency and server processing time.
			<-time.After(d)
		}

		contentLength := int64(buf.Len())
		// For HTTP HEAD requests, the ContentLength should be set to the size
		// of the body that would have been sent for a GET.
		// https://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.13
		if req.Method == "HEAD" {
			if hdr := interaction.Response.Headers.Get("Content-Length"); hdr != "" {
				cl, err := strconv.ParseInt(hdr, 10, 64)
				if err == nil {
					contentLength = cl
				}
			}
		}
		return &http.Response{
			Status:        interaction.Response.Status,
			StatusCode:    interaction.Response.Code,
			Proto:         "HTTP/1.0",
			ProtoMajor:    1,
			ProtoMinor:    0,
			Request:       req,
			Header:        interaction.Response.Headers,
			Close:         true,
			ContentLength: contentLength,
			Body:          ioutil.NopCloser(buf),
		}, nil
	}
}

// CancelRequest implements the github.com/coreos/etcd/client.CancelableTransport interface
func (r *Recorder) CancelRequest(req *http.Request) {
	type cancelableTransport interface {
		CancelRequest(req *http.Request)
	}
	if ct, ok := r.realTransport.(cancelableTransport); ok {
		ct.CancelRequest(req)
	}
}

// SetMatcher sets a function to match requests against recorded HTTP interactions.
func (r *Recorder) SetMatcher(matcher cassette.Matcher) {
	if r.cassette != nil {
		r.cassette.Matcher = matcher
	}
}

// SetReplayableInteractions defines whether to allow interactions to be replayed or not.
func (r *Recorder) SetReplayableInteractions(replayable bool) {
	if r.cassette != nil {
		r.cassette.ReplayableInteractions = replayable
	}
}

// AddPassthrough appends a hook to determine if a request should be ignored by the
// recorder.
func (r *Recorder) AddPassthrough(pass Passthrough) {
	r.Passthroughs = append(r.Passthroughs, pass)
}

// AddFilter appends a hook to modify a request before it is recorded.
//
// Filters are useful for filtering out sensitive parameters from the recorded data.
func (r *Recorder) AddFilter(filter cassette.Filter) {
	if r.cassette != nil {
		r.cassette.Filters = append(r.cassette.Filters, filter)
	}
}

// AddSaveFilter appends a hook to modify a request before it is saved.
//
// This filter is suitable for treating recorded responses to remove sensitive data. Altering responses using a regular
// AddFilter can have unintended consequences on code that is consuming responses.
func (r *Recorder) AddSaveFilter(filter cassette.Filter) {
	if r.cassette != nil {
		r.cassette.SaveFilters = append(r.cassette.SaveFilters, filter)
	}
}

// Mode returns recorder state
func (r *Recorder) Mode() Mode {
	return r.mode
}
