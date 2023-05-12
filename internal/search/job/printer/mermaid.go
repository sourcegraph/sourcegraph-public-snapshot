package printer

import (
	"bytes"
	"io"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/search/job"
)

// Mermaid outputs a Mermaid flowchart. See https://mermaid-js.github.io.
func Mermaid(j job.Describer) string {
	return MermaidVerbose(j, job.VerbosityNone)
}

func MermaidVerbose(j job.Describer, verbosity job.Verbosity) string {
	depth := 0
	id := 0
	b := new(bytes.Buffer)
	b.WriteString("flowchart TB\n")

	var writeMermaid func(job.Describer)
	writeMermaid = func(j job.Describer) {
		if j == nil {
			return
		}
		srcID := id
		depth++
		writeNode(b, depth, DefaultStyle, &id, buildLabel(j, verbosity))
		for _, child := range j.Children() {
			writeEdge(b, depth, srcID, id)
			writeMermaid(child)
		}
		depth--
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

func buildLabel(j job.Describer, v job.Verbosity) string {
	b := new(strings.Builder)
	b.WriteRune('"')
	b.WriteString(trimmedUpperName(j.Name()))
	enc := mermaidKeyValueWriter{b}
	for _, field := range j.Attributes(v) {
		enc.Write(string(field.Key), field.Value.Emit())
	}
	b.WriteRune('"')
	return b.String()
}

type mermaidKeyValueWriter struct{ io.StringWriter }

func (w mermaidKeyValueWriter) Write(key, value string) {
	w.WriteString(" <br> ")
	w.WriteString(mermaidEscaper.Replace(key))
	w.WriteString(": ")
	w.WriteString(mermaidEscaper.Replace(value))
}

// Copied from the `html` package and modified for mermaid
var mermaidEscaper = strings.NewReplacer(
	`"`, "#quot;",
	`'`, "#apos;",
	`&`, "#amp;",
	`<`, "#lt;",
	`>`, "#gt;",
)
