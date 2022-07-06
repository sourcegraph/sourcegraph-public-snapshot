package printer

import (
	"fmt"
	"strconv"

	otlog "github.com/opentracing/opentracing-go/log"
)

type keyValueWriter interface {
	Write(key, value string)
}

type fieldEncoder struct {
	w keyValueWriter
}

func (e fieldEncoder) EmitString(key, value string) {
	e.w.Write(key, value)
}

func (e fieldEncoder) EmitBool(key string, value bool) {
	e.w.Write(key, strconv.FormatBool(value))
}
func (e fieldEncoder) EmitInt(key string, value int) {
	e.w.Write(key, strconv.Itoa(value))
}
func (e fieldEncoder) EmitInt32(key string, value int32) {
	e.w.Write(key, strconv.Itoa(int(value)))
}
func (e fieldEncoder) EmitInt64(key string, value int64) {
	e.w.Write(key, strconv.Itoa(int(value)))
}
func (e fieldEncoder) EmitUint32(key string, value uint32) {
	e.w.Write(key, strconv.Itoa(int(value)))
}
func (e fieldEncoder) EmitUint64(key string, value uint64) {
	e.w.Write(key, strconv.FormatUint(value, 10))
}
func (e fieldEncoder) EmitFloat32(key string, value float32) {
	e.w.Write(key, fmt.Sprintf("%g", value))
}
func (e fieldEncoder) EmitFloat64(key string, value float64) {
	e.w.Write(key, fmt.Sprintf("%g", value))
}
func (e fieldEncoder) EmitObject(key string, value interface{}) {
	e.w.Write(key, fmt.Sprintf("%+v", value))
}
func (e fieldEncoder) EmitLazyLogger(value otlog.LazyLogger) {
	value(e)
}
