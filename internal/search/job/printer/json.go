package printer

import (
	"encoding/json"
	"fmt"

	"go.opentelemetry.io/otel/attribute"

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
	tags     []attribute.KeyValue
	children []node
}

func (n node) params() map[string]interface{} {
	m := make(map[string]interface{})
	for _, field := range n.tags {
		m[string(field.Key)] = field.Value.AsInterface()
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
		tags: j.Attributes(v),
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
