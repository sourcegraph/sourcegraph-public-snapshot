package encoders

import (
	"go.uber.org/zap/zapcore"

	"github.com/sourcegraph/sourcegraph/lib/log/otfields"
)

type ResourceEncoder struct{ otfields.Resource }

func (r *ResourceEncoder) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", r.Name)
	if len(r.Namespace) > 0 {
		enc.AddString("namespace", r.Namespace)
	}
	if len(r.InstanceID) > 0 {
		enc.AddString("instance.id", r.InstanceID)
	}
	if len(r.Version) > 0 {
		enc.AddString("version", r.Version)
	}
	return nil
}

type TraceContextEncoder struct{ otfields.TraceContext }

func (t *TraceContextEncoder) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if len(t.TraceID) > 0 {
		enc.AddString("TraceId", t.TraceID)
	}
	if len(t.SpanID) > 0 {
		enc.AddString("SpanId", t.SpanID)
	}
	return nil
}
