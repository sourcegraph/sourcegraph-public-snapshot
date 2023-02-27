package trace

import (
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

type attributesStringer []attribute.KeyValue

func (a attributesStringer) String() string {
	var b strings.Builder
	for i, attr := range a {
		if i > 0 {
			b.WriteString("\n")
		}
		var (
			key   = string(attr.Key)
			value = attr.Value.Emit()
		)
		b.Grow(len(key) + 1 + len(value))
		b.WriteString(key)
		b.WriteString(":")
		b.WriteString(value)
	}
	return b.String()
}

type stringerFunc func() string

func (s stringerFunc) String() string { return s() }
