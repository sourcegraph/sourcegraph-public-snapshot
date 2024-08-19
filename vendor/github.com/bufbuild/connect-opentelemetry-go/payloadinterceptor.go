// Copyright 2022-2023 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package otelconnect

import (
	"fmt"
	"sync"

	connect "github.com/bufbuild/connect-go"
)

type streamingClientInterceptor struct {
	connect.StreamingClientConn

	receive func(any, connect.StreamingClientConn) error
	send    func(any, connect.StreamingClientConn) error
	onClose func()

	mu             sync.Mutex
	requestClosed  bool
	responseClosed bool
	onCloseCalled  bool
}

func (s *streamingClientInterceptor) Receive(msg any) error {
	return s.receive(msg, s.StreamingClientConn)
}

func (s *streamingClientInterceptor) Send(msg any) error {
	return s.send(msg, s.StreamingClientConn)
}

func (s *streamingClientInterceptor) CloseRequest() error {
	err := s.StreamingClientConn.CloseRequest()
	s.mu.Lock()
	s.requestClosed = true
	shouldCall := s.responseClosed && !s.onCloseCalled
	if shouldCall {
		s.onCloseCalled = true
	}
	s.mu.Unlock()
	if shouldCall {
		s.onClose()
	}
	return err
}

func (s *streamingClientInterceptor) CloseResponse() error {
	err := s.StreamingClientConn.CloseResponse()
	s.mu.Lock()
	s.responseClosed = true
	shouldCall := s.requestClosed && !s.onCloseCalled
	if shouldCall {
		s.onCloseCalled = true
	}
	s.mu.Unlock()
	if shouldCall {
		s.onClose()
	}
	return err
}

type errorStreamingClientInterceptor struct {
	connect.StreamingClientConn

	err error
}

func (e *errorStreamingClientInterceptor) Send(any) error {
	return e.err
}

func (e *errorStreamingClientInterceptor) CloseRequest() error {
	if err := e.StreamingClientConn.CloseRequest(); err != nil {
		return fmt.Errorf("%w %s", err, e.err.Error())
	}
	return e.err
}

func (e *errorStreamingClientInterceptor) Receive(any) error {
	return e.err
}

func (e *errorStreamingClientInterceptor) CloseResponse() error {
	if err := e.StreamingClientConn.CloseResponse(); err != nil {
		return fmt.Errorf("%w %s", err, e.err.Error())
	}
	return e.err
}

type streamingHandlerInterceptor struct {
	connect.StreamingHandlerConn

	receive func(any, connect.StreamingHandlerConn) error
	send    func(any, connect.StreamingHandlerConn) error
}

func (p *streamingHandlerInterceptor) Receive(msg any) error {
	return p.receive(msg, p.StreamingHandlerConn)
}

func (p *streamingHandlerInterceptor) Send(msg any) error {
	return p.send(msg, p.StreamingHandlerConn)
}
