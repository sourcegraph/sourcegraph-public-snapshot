package trace

import (
	"fmt"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"
)

// Printf is an opentracing log.Field which is a LazyLogger. So the format
// string will only be evaluated if the trace is collected. In the case of
// net/trace, it will only be evaluated on page load.
func Printf(key, f string, args ...any) log.Field {
	return log.Lazy(func(fv log.Encoder) {
		fv.EmitString(key, fmt.Sprintf(f, args...))
	})
}

// Stringer is an opentracing log.Field which is a LazyLogger. So the String()
// will only be called if the trace is collected. In the case of net/trace, it
// will only be evaluated on page load.
func Stringer(key string, v fmt.Stringer) log.Field {
	return log.Lazy(func(fv log.Encoder) {
		fv.EmitString(key, v.String())
	})
}

// Scoped is an opentracing log.Field which is a LazyLogger. It will log each
// field with a key scoped by the given scope like `scope.key`.
func Scoped(scope string, fields ...log.Field) log.Field {
	return log.Lazy(func(fv log.Encoder) {
		enc := &scopedEncoder{scope: scope, enc: fv}
		for _, field := range fields {
			field.Marshal(enc)
		}
	})
}

// Strings is an opentracing log.Field which is a LazyLogger. It will log each
// string as key.$i.
func Strings(key string, values []string) log.Field {
	return log.Lazy(func(fv log.Encoder) {
		for i, v := range values {
			fv.EmitString(fmt.Sprintf("%s.%d", key, i), v)
		}
	})
}

// LazyFields is an opentracing log.Field that will only call the field-generating
// function if the trace is collected.
func LazyFields(lazyFields func() []log.Field) log.Field {
	return log.Lazy(func(fv log.Encoder) {
		for _, field := range lazyFields() {
			field.Marshal(fv)
		}
	})
}

// SQL is an opentracing log.Field which is a LazyLogger. It will log the
// query as well as each argument.
func SQL(q *sqlf.Query) log.Field {
	return log.Lazy(func(fv log.Encoder) {
		fv.EmitString("sql", q.Query(sqlf.PostgresBindVar))
		for i, arg := range q.Args() {
			fv.EmitObject(fmt.Sprintf("arg%d", i+1), arg)
		}
	})
}

// scopedEncoder encodes each field with a scoped key, like `scope.key`
type scopedEncoder struct {
	scope string
	enc   log.Encoder
}

func (e *scopedEncoder) scoped(key string) string {
	return fmt.Sprintf("%s.%s", e.scope, key)
}

func (e *scopedEncoder) EmitString(k, v string)              { e.enc.EmitString(e.scoped(k), v) }
func (e *scopedEncoder) EmitBool(k string, v bool)           { e.enc.EmitBool(e.scoped(k), v) }
func (e *scopedEncoder) EmitInt(k string, v int)             { e.enc.EmitInt(e.scoped(k), v) }
func (e *scopedEncoder) EmitInt32(k string, v int32)         { e.enc.EmitInt32(e.scoped(k), v) }
func (e *scopedEncoder) EmitInt64(k string, v int64)         { e.enc.EmitInt64(e.scoped(k), v) }
func (e *scopedEncoder) EmitUint32(k string, v uint32)       { e.enc.EmitUint32(e.scoped(k), v) }
func (e *scopedEncoder) EmitUint64(k string, v uint64)       { e.enc.EmitUint64(e.scoped(k), v) }
func (e *scopedEncoder) EmitFloat32(k string, v float32)     { e.enc.EmitFloat32(e.scoped(k), v) }
func (e *scopedEncoder) EmitFloat64(k string, v float64)     { e.enc.EmitFloat64(e.scoped(k), v) }
func (e *scopedEncoder) EmitObject(k string, v any)          { e.enc.EmitObject(e.scoped(k), v) }
func (e *scopedEncoder) EmitLazyLogger(value log.LazyLogger) { value(e) }
