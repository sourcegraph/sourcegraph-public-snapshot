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
	"context"
	"errors"
	"io"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

type streamingState struct {
	mu              sync.Mutex
	protocol        string
	req             *Request
	attributeFilter AttributeFilter
	omitTraceEvents bool
	attributes      []attribute.KeyValue
	error           error
	sentCounter     int
	receivedCounter int
	receiveSize     metric.Int64Histogram
	receivesPerRPC  metric.Int64Histogram
	sendSize        metric.Int64Histogram
	sendsPerRPC     metric.Int64Histogram
}

func newStreamingState(
	req *Request,
	attributeFilter AttributeFilter,
	omitTraceEvents bool,
	attributes []attribute.KeyValue,
	receiveSize, receivesPerRPC, sendSize, sendsPerRPC metric.Int64Histogram,
) *streamingState {
	attributes = attributeFilter.filter(req, attributes...)
	return &streamingState{
		protocol:        protocolToSemConv(req.Peer.Protocol),
		attributeFilter: attributeFilter,
		omitTraceEvents: omitTraceEvents,
		req:             req,
		attributes:      attributes,
		receiveSize:     receiveSize,
		receivesPerRPC:  receivesPerRPC,
		sendSize:        sendSize,
		sendsPerRPC:     sendsPerRPC,
	}
}

type sendReceiver interface {
	Receive(any) error
	Send(any) error
}

func (s *streamingState) addAttributes(attributes ...attribute.KeyValue) {
	s.attributes = append(s.attributes, s.attributeFilter.filter(s.req, attributes...)...)
}

func (s *streamingState) receive(ctx context.Context, msg any, conn sendReceiver) error {
	err := conn.Receive(msg)
	s.mu.Lock()
	defer s.mu.Unlock()
	if errors.Is(err, io.EOF) {
		return err
	}
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
		s.receivedCounter++
		s.event(ctx, semconv.MessageTypeReceived, s.receivedCounter, ok, size)
	}
	s.receiveSize.Record(ctx, int64(size), metric.WithAttributes(s.attributes...))
	s.receivesPerRPC.Record(ctx, 1, metric.WithAttributes(s.attributes...))
	return err
}

func (s *streamingState) send(ctx context.Context, msg any, conn sendReceiver) error {
	err := conn.Send(msg)
	s.mu.Lock()
	defer s.mu.Unlock()
	if errors.Is(err, io.EOF) {
		return err
	}
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
		s.sentCounter++
		s.event(ctx, semconv.MessageTypeSent, s.sentCounter, ok, size)
	}
	s.sendSize.Record(ctx, int64(size), metric.WithAttributes(s.attributes...))
	s.sendsPerRPC.Record(ctx, 1, metric.WithAttributes(s.attributes...))
	return err
}

func (s *streamingState) event(ctx context.Context, messageType attribute.KeyValue, messageID int, msgOk bool, size int) {
	span := trace.SpanFromContext(ctx)
	if msgOk {
		span.AddEvent("message", trace.WithAttributes(s.attributeFilter.filter(
			s.req,
			messageType,
			semconv.MessageUncompressedSizeKey.Int(size),
			semconv.MessageIDKey.Int(messageID),
		)...))
	} else {
		span.AddEvent("message", trace.WithAttributes(s.attributeFilter.filter(
			s.req,
			messageType,
			semconv.MessageIDKey.Int(messageID),
		)...))
	}
}
