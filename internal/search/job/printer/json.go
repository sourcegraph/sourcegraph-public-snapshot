package printer

import (
	"encoding/json"
	"fmt"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/search/job"
)

// JSON returns a summary of a job in formatted JSON.
func JSON(j job.Describer) string {
	return JSONVerbose(j, job.VerbosityNone)
}

// JSONVerbose returns the full fidelity of values that comprise a job in formatted JSON.
func JSONVerbose(j job.Describer, verbosity job.Verbosity) string {
	result, err := json.MarshalIndent(toNode(j, verbosity), "", "  ")
	if err != nil {
		panic(err)
	}
	return string(result)
}

type node struct {
	name     string
	tags     []otlog.Field
	children []node
}

func (n node) params() map[string]interface{} {
	m := make(map[string]interface{})
	enc := jsonFieldEncoder{&m}
	for _, field := range n.tags {
		field.Marshal(enc)
	}
	seenJobNames := map[string]int{}
	for _, child := range n.children {
		key := child.name
		if seenCount, ok := seenJobNames[key]; ok {
			if seenCount == 1 {
				m[fmt.Sprintf("%s.%d", key, 0)] = m[key]
				delete(m, key)
			}
			key = fmt.Sprintf("%s.%d", key, seenCount)
		}
		m[key] = child.params()
		seenJobNames[key]++
	}
	return m
}

func (n node) MarshalJSON() ([]byte, error) {
	if len(n.tags) == 0 && len(n.children) == 0 {
		return json.Marshal(n.name)
	}
	m := make(map[string]interface{})
	m[n.name] = n.params()
	return json.Marshal(m)
}

func toNode(j job.Describer, v job.Verbosity) node {
	return node{
		name: j.Name(),
		tags: j.Fields(v),
		children: func() []node {
			childJobs := j.Children()
			res := make([]node, 0, len(childJobs))
			for _, childJob := range childJobs {
				res = append(res, toNode(childJob, v))
			}
			return res
		}(),
	}
}

type jsonFieldEncoder struct {
	m *map[string]interface{}
}

func (e jsonFieldEncoder) EmitString(key, value string)             { (*e.m)[key] = value }
func (e jsonFieldEncoder) EmitBool(key string, value bool)          { (*e.m)[key] = value }
func (e jsonFieldEncoder) EmitInt(key string, value int)            { (*e.m)[key] = value }
func (e jsonFieldEncoder) EmitInt32(key string, value int32)        { (*e.m)[key] = value }
func (e jsonFieldEncoder) EmitInt64(key string, value int64)        { (*e.m)[key] = value }
func (e jsonFieldEncoder) EmitUint32(key string, value uint32)      { (*e.m)[key] = value }
func (e jsonFieldEncoder) EmitUint64(key string, value uint64)      { (*e.m)[key] = value }
func (e jsonFieldEncoder) EmitFloat32(key string, value float32)    { (*e.m)[key] = value }
func (e jsonFieldEncoder) EmitFloat64(key string, value float64)    { (*e.m)[key] = value }
func (e jsonFieldEncoder) EmitObject(key string, value interface{}) { (*e.m)[key] = value }
func (e jsonFieldEncoder) EmitLazyLogger(value otlog.LazyLogger)    { value(e) }
