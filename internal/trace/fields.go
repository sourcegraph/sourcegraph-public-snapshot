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
