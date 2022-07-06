package printer

import (
	"io"
	"bytes"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/search/job"
)

// PrettyMermaid outputs a Mermaid flowchart. See https://mermaid-js.github.io.
func PrettyMermaid(j job.DescriptiveJob) string {
	depth := 0
	id := 0
	b := new(bytes.Buffer)
	b.WriteString("flowchart TB\n")

	var writeMermaid func(job.Job)
	writeMermaid = func(j job.Job) {
		if j == nil {
			return
		}
		srcID := id
		writeNode(b, depth, DefaultStyle, &id, )
	}
	writeMermaid(j)
	return b.String()
}


type NodeStyle int

const (
	DefaultStyle NodeStyle = iota
	RoundedStyle
)

func writeEdge(b *bytes.Buffer, depth, src, dst int) {
	b.WriteString(strconv.Itoa(src))
	b.WriteString("---")
	b.WriteString(strconv.Itoa(dst))
	writeSep(b, "\n", "  ", depth)
}

func writeNode(b *bytes.Buffer, depth int, style NodeStyle, id *int, label string) {
	open := "["
	close := "]"
	if style == RoundedStyle {
		open = "(["
		close = "])"
	}
	b.WriteString(strconv.Itoa(*id))
	b.WriteString(open)
	b.WriteString(label)
	b.WriteString(close)
	writeSep(b, "\n", "  ", depth)
	*id++
}

func buildLabel(j job.DescriptiveJob, v job.Verbosity) string {
	var b strings.Builder
	b.WriteString(strings.ToUpper(strings.TrimSuffix(j.Name(), "Job")))
	enc := fieldEncoder{sexpKeyValueWriter{b}}
	for _, field := range j.Tags(v) {
		b.WriteString("<br/>")
		field.Marshal(enc)
	}
	return b.String()
}

type mermaidKeyValueWriter struct { io.StringWriter }

func (w mermaidKeyValueWriter) Write(key, value string) {
	w.WriteString(key)
	w.WriteString(": ")
	w.WriteString(value)
}

