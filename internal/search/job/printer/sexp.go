package printer

import (
	"bytes"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/search/job"
)

// Sexp outputs the s-expression on a single line.
func Sexp(j job.DescriptiveJob) string {
	return SexpFormat(j, job.VerbosityBasic, " ", "")
}

// PrettySexp outputs a formatted s-expression with two spaces of indentation, potentially spanning multiple lines.
func PrettySexp(j job.DescriptiveJob) string {
	return SexpFormat(j, job.VerbosityBasic, "\n", "  ")
}

func SexpFormat(j job.DescriptiveJob, verbosity job.Verbosity, sep, indent string) string {
	b := new(bytes.Buffer)
	depth := 0

	var writeSexp func(job.DescriptiveJob)
	writeSexp = func(j job.DescriptiveJob) {
		if j == nil {
			return
		}
		tags := j.Tags(verbosity)
		children := j.Children()
		if len(tags) == 0 && len(children) == 0 {
			b.WriteString(j.Name())
			return
		}

		b.WriteByte('(')
		b.WriteString(trimmedUpperName(j.Name()))
		depth++
		if len(tags) > 0 {
			enc := fieldStringEncoder{sexpKeyValueWriter{b}}
			for _, field := range tags {
				writeSep(b, sep, indent, depth)
				field.Marshal(enc)
			}
		}
		if len(children) > 0 {
			for _, child := range children {
				writeSep(b, sep, indent, depth)
				writeSexp(child)
			}
		}
		b.WriteByte(')')
		depth--
	}
	writeSexp(j)
	return b.String()
}

func writeSep(b *bytes.Buffer, sep, indent string, depth int) {
	b.WriteString(sep)
	if indent == "" {
		return
	}
	for i := 0; i < depth; i++ {
		b.WriteString(indent)
	}
}

type sexpKeyValueWriter struct{ io.StringWriter }

func (w sexpKeyValueWriter) Write(key, value string) {
	w.WriteString("(")
	w.WriteString(key)
	w.WriteString(" . ")
	w.WriteString(value)
	w.WriteString(")")
}
