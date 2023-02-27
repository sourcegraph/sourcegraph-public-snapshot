package printer

import (
	"fmt"
	"strconv"
	"strings"

	otlog "github.com/opentracing/opentracing-go/log"
)

func trimmedUpperName(name string) string {
	return strings.ToUpper(strings.TrimSuffix(name, "Job"))
}

type keyValueWriter interface {
	Write(key, value string)
}

type fieldStringEncoder struct {
	w keyValueWriter
}

func (e fieldStringEncoder) EmitString(key, value string) {
	e.w.Write(key, value)
}
func (e fieldStringEncoder) EmitBool(key string, value bool) {
	e.w.Write(key, strconv.FormatBool(value))
}
func (e fieldStringEncoder) EmitInt(key string, value int) {
	e.w.Write(key, strconv.Itoa(value))
}
func (e fieldStringEncoder) EmitInt32(key string, value int32) {
	e.w.Write(key, strconv.Itoa(int(value)))
}
func (e fieldStringEncoder) EmitInt64(key string, value int64) {
	e.w.Write(key, strconv.Itoa(int(value)))
}
func (e fieldStringEncoder) EmitUint32(key string, value uint32) {
	e.w.Write(key, strconv.Itoa(int(value)))
}
func (e fieldStringEncoder) EmitUint64(key string, value uint64) {
	e.w.Write(key, strconv.FormatUint(value, 10))
}
func (e fieldStringEncoder) EmitFloat32(key string, value float32) {
	e.w.Write(key, fmt.Sprintf("%g", value))
}
func (e fieldStringEncoder) EmitFloat64(key string, value float64) {
	e.w.Write(key, fmt.Sprintf("%g", value))
}
func (e fieldStringEncoder) EmitObject(key string, value interface{}) {
	e.w.Write(key, fmt.Sprintf("%+v", value))
}
func (e fieldStringEncoder) EmitLazyLogger(value otlog.LazyLogger) {
	value(e)
}
