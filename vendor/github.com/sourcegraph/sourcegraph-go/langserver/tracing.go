package langserver

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	basictracer "github.com/opentracing/basictracer-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sourcegraph/jsonrpc2"
)

// InitTracer initializes the tracer for the connection if it has not
// already been initialized.
//
// It assumes that h is only ever called for this conn.
func (h *HandlerCommon) InitTracer(conn *jsonrpc2.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.tracerOK {
		return
	}

	t := tracer{conn: conn}
	opt := basictracer.DefaultOptions()
	opt.Recorder = &t
	h.tracer = basictracer.NewWithOptions(opt)
	h.tracerOK = true
	go func() {
		<-conn.DisconnectNotify()
		t.mu.Lock()
		t.conn = nil
		t.mu.Unlock()
	}()
}

func (h *HandlerCommon) SpanForRequest(ctx context.Context, buildOrLang string, req *jsonrpc2.Request, tags opentracing.Tags) (opentracing.Span, context.Context, error) {
	opName := "LSP " + buildOrLang + " server: " + req.Method
	var span opentracing.Span

	// The parent span context can come from a few sources, depending
	// on how we're running this server and whether we are a build or
	// (wrapped) lang server.

	if span == nil && req.Meta != nil {
		// Try to get our parent span context from the JSON-RPC request from the LSP proxy.
		var carrier opentracing.TextMapCarrier
		if req.Meta != nil {
			if err := json.Unmarshal(*req.Meta, &carrier); err != nil {
				return nil, nil, err
			}
		}
		if clientCtx, err := h.tracer.Extract(opentracing.TextMap, carrier); err == nil {
			span = h.tracer.StartSpan(opName, ext.RPCServerOption(clientCtx), tags)
		} else if err != opentracing.ErrSpanContextNotFound {
			return nil, nil, err
		}
	}

	// Try to get our parent span context from the ctx. If this
	// succeeds, it means we're a language server being wrapped by a
	// build server, and the parent span is the build server's.
	if span == nil {
		if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
			span = parentSpan.Tracer().StartSpan(opName, tags, opentracing.ChildOf(parentSpan.Context()))
		}
	}

	if span == nil {
		// No opentracing context from our JSON-RPC peer, so we need to create our own.
		span = opentracing.StartSpan(opName, tags)
	}

	if !IsFileSystemRequest(req.Method) && req.Params != nil {
		span.SetTag("params", string(*req.Params))
	}

	return span, opentracing.ContextWithSpan(ctx, span), nil
}

type tracer struct {
	mu   sync.Mutex
	conn *jsonrpc2.Conn
}

func (t *tracer) RecordSpan(span basictracer.RawSpan) {
	t.mu.Lock()
	if t.conn == nil {
		t.mu.Unlock()
		return
	}
	t.mu.Unlock()

	ctx := context.Background()
	if err := t.conn.Notify(ctx, "telemetry/event", span); err != nil {
		log.Println("Error sending LSP telemetry/event notification:", err)
	}
}
