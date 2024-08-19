// Copyright 2022-2023 The Connect Authors
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
	"context"
	"errors"
	"io"
	"sync"

	connect "connectrpc.com/connect"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

type streamingState struct {
	mu              sync.Mutex
	spec            connect.Spec
	protocol        string
	attributeFilter AttributeFilter
	omitTraceEvents bool
	attributes      []attribute.KeyValue
	error           error
	sentCounter     int64
	receivedCounter int64
	receiveSize     metric.Int64Histogram
	sendSize        metric.Int64Histogram
}

func newStreamingState(
	spec connect.Spec,
	peer connect.Peer,
	attributeFilter AttributeFilter,
	omitTraceEvents bool,
	receiveSize, sendSize metric.Int64Histogram,
) *streamingState {
	protocol := protocolToSemConv(peer.Protocol)
	attributes := attributeFilter.filter(spec,
		requestAttributes(spec, peer)...,
	)
	return &streamingState{
		spec:            spec,
		protocol:        protocol,
		attributeFilter: attributeFilter,
		omitTraceEvents: omitTraceEvents,
		attributes:      attributes,
		receiveSize:     receiveSize,
		sendSize:        sendSize,
	}
}

type sendReceiver interface {
	Receive(any) error
	Send(any) error
}

func (s *streamingState) addAttributes(attributes ...attribute.KeyValue) {
	s.attributes = append(s.attributes, s.attributeFilter.filter(s.spec, attributes...)...)
}

func (s *streamingState) receive(ctx context.Context, msg any, conn sendReceiver) error {
	err := conn.Receive(msg)
	s.mu.Lock()
	defer s.mu.Unlock()
	if errors.Is(err, io.EOF) {
		return err
	}
	s.receivedCounter++
	if err != nil {
		s.error = err
		// If error add it to the attributes because the stream is about to terminate.
		// If no error don't add anything because status only exists at end of stream.
		if statusCode, ok := statusCodeAttribute(s.protocol, err); ok {
			s.addAttributes(statusCode)
		}
	}
	protomsg, ok := msg.(proto.Message)
	size := proto.Size(protomsg)
	if !s.omitTraceEvents {
		s.emitEvent(ctx, semconv.MessageTypeReceived, s.receivedCounter, size, ok)
	}
	s.receiveSize.Record(ctx, int64(size), metric.WithAttributes(s.attributes...))
	return err
}

func (s *streamingState) send(ctx context.Context, msg any, conn sendReceiver) error {
	err := conn.Send(msg)
	s.mu.Lock()
	defer s.mu.Unlock()
	if errors.Is(err, io.EOF) {
		return err
	}
	s.sentCounter++
	if err != nil {
		s.error = err
		// If error add it to the attributes because the stream is about to terminate.
		// If no error don't add anything because status only exists at end of stream.
		if statusCode, ok := statusCodeAttribute(s.protocol, err); ok {
			s.addAttributes(statusCode)
		}
	}
	protomsg, ok := msg.(proto.Message)
	size := proto.Size(protomsg)
	if !s.omitTraceEvents {
		s.emitEvent(ctx, semconv.MessageTypeSent, s.sentCounter, size, ok)
	}
	s.sendSize.Record(ctx, int64(size), metric.WithAttributes(s.attributes...))
	return err
}

func (s *streamingState) emitEvent(ctx context.Context, msgType attribute.KeyValue, msgID int64, msgSize int, hasSize bool) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}
	attrs := []attribute.KeyValue{
		msgType, semconv.MessageIDKey.Int64(msgID),
	}
	if hasSize {
		attrs = append(attrs, semconv.MessageUncompressedSizeKey.Int(msgSize))
	}
	span.AddEvent(messageKey, trace.WithAttributes(
		s.attributeFilter.filter(s.spec, attrs...)...,
	))
}
