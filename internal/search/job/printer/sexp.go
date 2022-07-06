package printer

import (
	"bytes"
	"io"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/search/job"
)

// Sexp outputs the s-expression on a single line.
func Sexp(j job.DescriptiveJob) string {
	return SexpFormat(j, job.VerbosityBasic, "", " ")
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
		b.WriteByte('(')
		b.WriteString(strings.ToUpper(strings.TrimSuffix(j.Name(), "Job")))
		depth++
		writeSep(b, sep, indent, depth)
		if verbosity > job.VerbosityNone {
			enc := fieldEncoder{sexpKeyValueWriter{b}}
			for _, field := range j.Tags(verbosity) {
				field.Marshal(enc)
				writeSep(b, sep, indent, depth)
			}
		}
		for _, child := range j.Children() {
			writeSexp(child)
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
