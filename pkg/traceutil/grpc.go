package traceutil

import (
	"time"

	"sourcegraph.com/sourcegraph/appdash"
)

func init() { appdash.RegisterEvent(GRPCCall{}) }

// GRPCCall records a GRPC method invocation.
type GRPCCall struct {
	Server string
	Method string

	Arg     string
	ArgType string
	Err     string

	ServerRecv time.Time
	ServerSend time.Time
}

// Schema returns the constant "GRPCCall".
func (GRPCCall) Schema() string { return "GRPCCall" }

func (e GRPCCall) Start() time.Time { return e.ServerRecv }
func (e GRPCCall) End() time.Time   { return e.ServerSend }
