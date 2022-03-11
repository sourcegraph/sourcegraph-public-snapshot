package job

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/search/commit"
	"github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/search/structural"
	"github.com/sourcegraph/sourcegraph/internal/search/symbol"
	"github.com/sourcegraph/sourcegraph/internal/search/textsearch"
	"github.com/sourcegraph/sourcegraph/internal/search/zoekt"
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
		if job == nil {
			return
		}
		switch j := job.(type) {
		case
			*zoekt.ZoektRepoSubsetSearch,
			*searcher.Searcher,
			*run.RepoSearch,
			*textsearch.RepoUniverseTextSearch,
			*structural.StructuralSearch,
			*commit.CommitSearch,
			*symbol.RepoSubsetSymbolSearch,
			*symbol.RepoUniverseSymbolSearch,
			*repos.ComputeExcludedRepos,
			*noopJob:
			b.WriteString(j.Name())

		case *repoPagerJob:
			b.WriteString("REPOPAGER")
			depth++
			writeSep(b, sep, indent, depth)
			writeSexp(j.child)
			b.WriteString(")")
			depth--

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
		case *selectJob:
			b.WriteString("(SELECT")
			depth++
			writeSep(b, sep, indent, depth)
			b.WriteString(j.path.String())
			writeSep(b, sep, indent, depth)
			writeSexp(j.child)
			b.WriteString(")")
			depth--
		case *alertJob:
			b.WriteString("(ALERT")
			depth++
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

// PrettyMermaid outputs a Mermaid flowchart. See https://mermaid-js.github.io.
func PrettyMermaid(job Job) string {
	depth := 0
	id := 0
	b := new(bytes.Buffer)
	b.WriteString("flowchart TB\n")
	var writeMermaid func(Job)
	writeMermaid = func(job Job) {
		if job == nil {
			return
		}
		switch j := job.(type) {
		case
			*zoekt.ZoektRepoSubsetSearch,
			*searcher.Searcher,
			*run.RepoSearch,
			*textsearch.RepoUniverseTextSearch,
			*structural.StructuralSearch,
			*commit.CommitSearch,
			*symbol.RepoSubsetSymbolSearch,
			*symbol.RepoUniverseSymbolSearch,
			*repos.ComputeExcludedRepos,
			*noopJob:
			writeNode(b, depth, RoundedStyle, &id, j.Name())

		case *repoPagerJob:
			srcId := id
			depth++
			writeNode(b, depth, RoundedStyle, &id, "REPOPAGER")
			writeEdge(b, depth, srcId, id)
			writeMermaid(j.child)
			depth--
		case *AndJob:
			srcId := id
			depth++
			writeNode(b, depth, RoundedStyle, &id, "AND")
			for _, child := range j.children {
				writeEdge(b, depth, srcId, id)
				writeMermaid(child)
			}
			depth--
		case *OrJob:
			srcId := id
			depth++
			writeNode(b, depth, RoundedStyle, &id, "OR")
			for _, child := range j.children {
				writeEdge(b, depth, srcId, id)
				writeMermaid(child)
			}
			depth--
		case *PriorityJob:
			srcId := id
			depth++
			writeNode(b, depth, RoundedStyle, &id, "PRIORITY")

			requiredId := id
			writeEdge(b, depth, srcId, requiredId)
			writeNode(b, depth, RoundedStyle, &id, "REQUIRED")
			writeEdge(b, depth, requiredId, id)
			writeMermaid(j.required)

			optionalId := id
			writeEdge(b, depth, srcId, optionalId)
			writeNode(b, depth, RoundedStyle, &id, "OPTIONAL")
			writeEdge(b, depth, optionalId, id)
			writeMermaid(j.optional)
			depth--
		case *ParallelJob:
			srcId := id
			depth++
			writeNode(b, depth, RoundedStyle, &id, "PARALLEL")
			for _, child := range j.children {
				writeEdge(b, depth, srcId, id)
				writeMermaid(child)
			}
			depth--
		case *TimeoutJob:
			srcId := id
			depth++
			writeNode(b, depth, RoundedStyle, &id, "TIMEOUT")
			writeEdge(b, depth, srcId, id)
			writeNode(b, depth, DefaultStyle, &id, j.timeout.String())
			writeEdge(b, depth, srcId, id)
			writeMermaid(j.child)
			depth--
		case *LimitJob:
			srcId := id
			depth++
			writeNode(b, depth, RoundedStyle, &id, "LIMIT")
			writeEdge(b, depth, srcId, id)
			writeNode(b, depth, DefaultStyle, &id, strconv.Itoa(j.limit))
			writeEdge(b, depth, srcId, id)
			writeMermaid(j.child)
			depth--
		case *subRepoPermsFilterJob:
			srcId := id
			depth++
			writeNode(b, depth, RoundedStyle, &id, "FILTER")
			writeEdge(b, depth, srcId, id)
			writeNode(b, depth, DefaultStyle, &id, "SubRepoPermissions")
			writeEdge(b, depth, srcId, id)
			writeMermaid(j.child)
			depth--
		case *selectJob:
			srcId := id
			depth++
			writeNode(b, depth, RoundedStyle, &id, "SELECT")
			writeEdge(b, depth, srcId, id)
			writeNode(b, depth, DefaultStyle, &id, j.path.String())
			writeEdge(b, depth, srcId, id)
			writeMermaid(j.child)
			depth--
		case *alertJob:
			srcId := id
			depth++
			writeNode(b, depth, RoundedStyle, &id, "ALERT")
			writeEdge(b, depth, srcId, id)
			writeMermaid(j.child)
			depth--
		default:
			panic(fmt.Sprintf("unsupported job %T for PrettyMermaid printer", job))
		}
	}
	writeMermaid(job)
	return b.String()
}

// toJSON returns a JSON object representing a job. If `verbose` is true, values
// for all leaf jobs are emitted; if false, only the names of leaf nodes are
// emitted.
func toJSON(job Job, verbose bool) interface{} {
	var emitJSON func(Job) interface{}
	emitJSON = func(job Job) interface{} {
		if job == nil {
			return struct{}{}
		}
		switch j := job.(type) {
		case
			*zoekt.ZoektRepoSubsetSearch,
			*searcher.Searcher,
			*run.RepoSearch,
			*textsearch.RepoUniverseTextSearch,
			*structural.StructuralSearch,
			*commit.CommitSearch,
			*symbol.RepoSubsetSymbolSearch,
			*symbol.RepoUniverseSymbolSearch,
			*repos.ComputeExcludedRepos,
			*noopJob:
			if verbose {
				return map[string]interface{}{j.Name(): j}
			}
			return j.Name()

		case *repoPagerJob:
			return struct {
				Repopager interface{} `json:"REPOPAGER"`
			}{
				Repopager: emitJSON(j.child),
			}

		case *AndJob:
			children := make([]interface{}, 0, len(j.children))
			for _, child := range j.children {
				children = append(children, emitJSON(child))
			}
			return struct {
				And []interface{} `json:"AND"`
			}{
				And: children,
			}

		case *OrJob:
			children := make([]interface{}, 0, len(j.children))
			for _, child := range j.children {
				children = append(children, emitJSON(child))
			}
			return struct {
				Or []interface{} `json:"OR"`
			}{
				Or: children,
			}

		case *PriorityJob:
			priority := struct {
				Required interface{} `json:"REQUIRED"`
				Optional interface{} `json:"OPTIONAL"`
			}{
				Required: emitJSON(j.required),
				Optional: emitJSON(j.optional),
			}
			return struct {
				Priority interface{} `json:"PRIORITY"`
			}{
				Priority: priority,
			}

		case *ParallelJob:
			children := make([]interface{}, 0, len(j.children))
			for _, child := range j.children {
				children = append(children, emitJSON(child))
			}
			return struct {
				Parallel interface{} `json:"PARALLEL"`
			}{
				Parallel: children,
			}

		case *TimeoutJob:
			return struct {
				Timeout interface{} `json:"TIMEOUT"`
				Value   string      `json:"value"`
			}{
				Timeout: emitJSON(j.child),
				Value:   j.timeout.String(),
			}

		case *LimitJob:
			return struct {
				Limit interface{} `json:"LIMIT"`
				Value int         `json:"value"`
			}{
				Limit: emitJSON(j.child),
				Value: j.limit,
			}

		case *subRepoPermsFilterJob:
			return struct {
				Filter interface{} `json:"FILTER"`
				Value  string      `json:"value"`
			}{
				Filter: emitJSON(j.child),
				Value:  "SubRepoPermissions",
			}
		case *selectJob:
			return struct {
				Select interface{} `json:"SELECT"`
				Value  string      `json:"value"`
			}{
				Select: emitJSON(j.child),
				Value:  j.path.String(),
			}
		case *alertJob:
			return struct {
				Alert interface{} `json:"ALERT"`
			}{
				Alert: emitJSON(j.child),
			}
		default:
			panic(fmt.Sprintf("unsupported job %T for toJSON converter", job))
		}
	}
	return emitJSON(job)
}

// PrettyJSON returns a summary of a job in formatted JSON.
func PrettyJSON(job Job) string {
	result, _ := json.MarshalIndent(toJSON(job, false), "", "  ")
	return string(result)
}

// PrettyJSON returns the full fidelity of values that comprise a job in formatted JSON.
func PrettyJSONVerbose(job Job) string {
	result, _ := json.MarshalIndent(toJSON(job, true), "", "  ")
	return string(result)
}
