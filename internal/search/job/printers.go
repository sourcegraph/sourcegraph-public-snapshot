package job

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/search/commit"
	"github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/structural"
	"github.com/sourcegraph/sourcegraph/internal/search/symbol"
	"github.com/sourcegraph/sourcegraph/internal/search/textsearch"
)

func writeSep(b *bytes.Buffer, sep, indent string, depth int) {
	b.WriteString(sep)
	if indent == "" {
		return
	}
	for i := 0; i < depth; i++ {
		b.WriteString(indent)
	}
}

// SexpFormat controls the s-expression format that represents a Job. `sep`
// specifies the separator between terms. If `indent` is not empty, `indent` is
// prefixed the number of times corresponding to depth of the term in the tree.
// See the `Sexp` and `PrettySexp` convenience functions to see how these
// options are used.
func SexpFormat(job Job, sep, indent string) string {
	b := new(bytes.Buffer)
	depth := 0
	var writeSexp func(Job)
	writeSexp = func(job Job) {
		switch j := job.(type) {
		case
			*run.RepoSearch,
			*textsearch.RepoSubsetTextSearch,
			*textsearch.RepoUniverseTextSearch,
			*structural.StructuralSearch,
			*commit.CommitSearch,
			*symbol.RepoSubsetSymbolSearch,
			*symbol.RepoUniverseSymbolSearch,
			*repos.ComputeExcludedRepos,
			*noopJob:
			b.WriteString(j.Name())
		case *AndJob:
			b.WriteString("(AND")
			depth++
			for _, child := range j.children {
				writeSep(b, sep, indent, depth)
				writeSexp(child)
			}
			b.WriteString(")")
			depth--
		case *OrJob:
			b.WriteString("(OR")
			depth++
			for _, child := range j.children {
				writeSep(b, sep, indent, depth)
				writeSexp(child)
			}
			b.WriteString(")")
			depth--
		case *PriorityJob:
			b.WriteString("(PRIORITY")
			depth++
			writeSep(b, sep, indent, depth)
			b.WriteString("(REQUIRED")
			depth++
			writeSep(b, sep, indent, depth)
			writeSexp(j.required)
			b.WriteString(")")
			depth--
			writeSep(b, sep, indent, depth)
			b.WriteString("(OPTIONAL")
			depth++
			writeSep(b, sep, indent, depth)
			writeSexp(j.optional)
			b.WriteString(")")
			depth--
			b.WriteString(")")
			depth--
		case *ParallelJob:
			b.WriteString("(PARALLEL")
			depth++
			for _, child := range j.children {
				writeSep(b, sep, indent, depth)
				writeSexp(child)
			}
			depth--
			b.WriteString(")")
		case *TimeoutJob:
			b.WriteString("(TIMEOUT")
			depth++
			writeSep(b, sep, indent, depth)
			b.WriteString(j.timeout.String())
			writeSep(b, sep, indent, depth)
			writeSexp(j.child)
			b.WriteString(")")
			depth--
		case *LimitJob:
			b.WriteString("(LIMIT")
			depth++
			writeSep(b, sep, indent, depth)
			b.WriteString(strconv.Itoa(j.limit))
			writeSep(b, sep, indent, depth)
			writeSexp(j.child)
			b.WriteString(")")
			depth--
		case *subRepoPermsFilterJob:
			b.WriteString("(FILTER")
			depth++
			writeSep(b, sep, indent, depth)
			b.WriteString("SubRepoPermissions")
			writeSep(b, sep, indent, depth)
			writeSexp(j.child)
			b.WriteString(")")
			depth--
		default:
			panic(fmt.Sprintf("unsupported job %T for SexpFormat printer", job))
		}
	}
	writeSexp(job)
	return b.String()
}

// Sexp outputs the s-expression on a single line.
func Sexp(job Job) string {
	return SexpFormat(job, " ", "")
}

// PrettySexp outputs a formatted s-expression with two spaces of indentation, potentially spanning multiple lines.
func PrettySexp(job Job) string {
	return SexpFormat(job, "\n", "  ")
}
