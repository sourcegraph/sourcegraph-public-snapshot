package trace

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go"
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

// fieldsStringer lazily marshals a slice of log.Field into a string for
// printing in net/trace.
type fieldsStringer []log.Field

func (fs fieldsStringer) String() string {
	var e encoder
	for _, f := range fs {
		f.Marshal(&e)
	}
	return e.Builder.String()
}

// encoder is a log.Encoder used by fieldsStringer.
type encoder struct {
	strings.Builder
	prefixNewline bool
}

func (e *encoder) EmitString(key, value string) {
	if e.prefixNewline {
		// most times encoder is used is for one field
		e.Builder.WriteString("\n")
	}
	if !e.prefixNewline {
		e.prefixNewline = true
	}

	e.Builder.Grow(len(key) + 1 + len(value))
	e.Builder.WriteString(key)
	e.Builder.WriteString(":")
	e.Builder.WriteString(value)
}

func (e *encoder) EmitBool(key string, value bool) {
	e.EmitString(key, strconv.FormatBool(value))
}

func (e *encoder) EmitInt(key string, value int) {
	e.EmitString(key, strconv.Itoa(value))
}

func (e *encoder) EmitInt32(key string, value int32) {
	e.EmitString(key, strconv.FormatInt(int64(value), 10))
}

func (e *encoder) EmitInt64(key string, value int64) {
	e.EmitString(key, strconv.FormatInt(value, 10))
}

func (e *encoder) EmitUint32(key string, value uint32) {
	e.EmitString(key, strconv.FormatUint(uint64(value), 10))
}

func (e *encoder) EmitUint64(key string, value uint64) {
	e.EmitString(key, strconv.FormatUint(value, 10))
}

func (e *encoder) EmitFloat32(key string, value float32) {
	e.EmitString(key, strconv.FormatFloat(float64(value), 'E', -1, 64))
}

func (e *encoder) EmitFloat64(key string, value float64) {
	e.EmitString(key, strconv.FormatFloat(value, 'E', -1, 64))
}

func (e *encoder) EmitObject(key string, value any) {
	e.EmitString(key, fmt.Sprintf("%+v", value))
}

func (e *encoder) EmitLazyLogger(value log.LazyLogger) {
	value(e)
}

// spanTagEncoder wraps the opentracing.Span.SetTags to write values
// of type log.Field to span tags. The doc string of SetTags notes
// that it only accepts strings, numeric types, and bools, so these
// encoder methods convert to those types before writing the tag.
type spanTagEncoder struct {
	opentracing.Span
}

func (e *spanTagEncoder) EmitString(key, value string) {
	e.SetTag(key, value)
}

func (e *spanTagEncoder) EmitBool(key string, value bool) {
	e.SetTag(key, value)
}

func (e *spanTagEncoder) EmitInt(key string, value int) {
	e.SetTag(key, value)
}

func (e *spanTagEncoder) EmitInt32(key string, value int32) {
	e.SetTag(key, value)
}

func (e *spanTagEncoder) EmitInt64(key string, value int64) {
	e.SetTag(key, value)
}

func (e *spanTagEncoder) EmitUint32(key string, value uint32) {
	e.SetTag(key, value)
}

func (e *spanTagEncoder) EmitUint64(key string, value uint64) {
	e.SetTag(key, value)
}

func (e *spanTagEncoder) EmitFloat32(key string, value float32) {
	e.SetTag(key, value)
}

func (e *spanTagEncoder) EmitFloat64(key string, value float64) {
	e.SetTag(key, value)
}

func (e *spanTagEncoder) EmitObject(key string, value any) {
	s := fmt.Sprintf("%#+v", value)
	e.EmitString(key, s)
}

func (e *spanTagEncoder) EmitLazyLogger(value log.LazyLogger) {
	value(e)
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
