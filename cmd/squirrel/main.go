// Command symbols is a service that serves code symbols (functions, variables, etc.) from a repository at a
// specific commit.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/inconshreveable/log15"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/profiler"
	"github.com/sourcegraph/sourcegraph/internal/sentry"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func main() {
	routines := []goroutine.BackgroundRoutine{}

	// Set up Google Cloud Profiler when running in Cloud
	if err := profiler.Init(); err != nil {
		log.Fatalf("Failed to start profiler: %v", err)
	}

	// Initialization
	env.HandleHelpFlag()
	conf.Init()
	logging.Init()
	tracer.Init(conf.DefaultClient())
	sentry.Init(conf.DefaultClient())
	trace.Init()

	// Start debug server
	ready := make(chan struct{})
	go debugserver.NewServerRoutine(ready).Start()

	// Create HTTP server
	server := httpserver.NewFromAddr(":8984", &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      actor.HTTPMiddleware(ot.HTTPMiddleware(trace.HTTPMiddleware(NewHandler(), conf.DefaultClient()))),
	})
	routines = append(routines, server)

	// Mark health server as ready and go!
	close(ready)
	goroutine.MonitorBackgroundRoutines(context.Background(), routines...)
}

func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/definition", definitionHandler)
	mux.HandleFunc("/healthz", handleHealthCheck)
	return mux
}

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte("OK")); err != nil {
		log15.Error("failed to write response to health check, err: %s", err)
	}
}

func definitionHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	repo := q.Get("repo")
	if repo == "" {
		http.Error(w, "missing repo", http.StatusBadRequest)
		return
	}
	commit := q.Get("commit")
	if commit == "" {
		http.Error(w, "missing commit", http.StatusBadRequest)
		return
	}
	path := q.Get("path")
	if path == "" {
		http.Error(w, "missing path", http.StatusBadRequest)
		return
	}
	row64, err := strconv.ParseInt(q.Get("row"), 10, 32)
	if err != nil {
		http.Error(w, "missing or invalid int row", http.StatusBadRequest)
		return
	}
	row := uint32(row64)
	column64, err := strconv.ParseInt(q.Get("column"), 10, 32)
	if err != nil {
		http.Error(w, "missing or invalid int column", http.StatusBadRequest)
		return
	}
	column := uint32(column64)
	fmt.Println("repo:", repo, "commit:", commit, "path:", path, "row:", row, "column:", column)

	// get file extension
	ext := filepath.Ext(path)
	if ext != ".go" {
		http.Error(w, "only .go files are supported", http.StatusBadRequest)
		return
	}

	readFile := func(RepoCommitPath) ([]byte, error) {
		cmd := gitserver.DefaultClient.Command("git", "cat-file", "blob", commit+":"+path)
		cmd.Repo = api.RepoName(repo)
		stdout, stderr, err := cmd.DividedOutput(r.Context())
		if err != nil {
			return nil, fmt.Errorf("failed to get file contents: %s\n\nstdout:\n\n%s\n\nstderr:\n\n%s", err, stdout, stderr)
		}
		return stdout, nil
	}

	squirrel := NewSquirrel(readFile)

	result, err := squirrel.definition(Location{
		RepoCommitPath: RepoCommitPath{
			Repo:   repo,
			Commit: commit,
			Path:   path},
		Row:    row,
		Column: column,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get definition: %s", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, result)
}

type RepoCommitPath struct {
	Repo   string `json:"repo"`
	Commit string `json:"commit"`
	Path   string `json:"path"`
}

type Location struct {
	RepoCommitPath
	Row    uint32 `json:"row"`
	Column uint32 `json:"column"`
}

type ReadFileFunc func(RepoCommitPath) ([]byte, error)

type Squirrel struct {
	readFile ReadFileFunc
}

func NewSquirrel(readFile ReadFileFunc) *Squirrel {
	return &Squirrel{readFile: readFile}
}

func (s *Squirrel) definition(location Location) (*Location, error) {
	parser := sitter.NewParser()

	var bindingsExpr expr
	var lang *sitter.Language
	ext := filepath.Ext(location.Path)
	switch ext {
	case ".go":
		bindingsExpr = goBindings
		lang = golang.GetLanguage()
	default:
		return nil, fmt.Errorf("unrecognized file extension %s", ext)
	}

	parser.SetLanguage(lang)

	input, err := s.readFile(location.RepoCommitPath)
	if err != nil {
		return nil, err
	}

	tree, err := parser.ParseCtx(context.Background(), nil, input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file contents: %s", err)
	}
	root := tree.RootNode()

	startNode := root.NamedDescendantForPointRange(
		sitter.Point{Row: location.Row, Column: location.Column},
		sitter.Point{Row: location.Row, Column: location.Column},
	)

	if startNode == nil {
		return nil, errors.New("node is nil")
	}

	if startNode.Type() != "identifier" {
		return nil, errors.Newf("can't find definition of %s", startNode.Type(), location)
	}

	for currentNode := startNode; currentNode != nil; currentNode = currentNode.Parent() {
		captures := bindingsExpr.captures(currentNode)
		for _, capture := range captures {
			if capture.Content(input) == startNode.Content(input) {
				return &Location{
					RepoCommitPath: location.RepoCommitPath,
					Row:            capture.StartPoint().Row,
					Column:         capture.StartPoint().Column,
				}, nil
			}
		}
	}

	return nil, errors.New("could not find definition")
}

var goBindings expr = eOr(
	eAnd(eType("function_declaration"), eField("name", eCapture(eType("identifier"))), eField("parameters", goBindingsParameters)),
	eAnd(eType("func_literal"), eField("parameters", goBindingsParameters)),
)

var goBindingsParameters expr = eAnd(
	eType("parameter_list"),
	eChildren(
		eOr(
			eAnd(
				eType("parameter_declaration"),
				eField("name", eCapture(eType("identifier"))),
			),
			eAnd(
				eType("variadic_parameter_declaration"),
				eField("name", eCapture(eType("identifier"))),
			),
		),
	),
)

type expr interface {
	captures(node *sitter.Node) []*sitter.Node
}

// Match node type
type exprType struct{ typeName string }

func eType(typeName string) expr {
	return exprType{typeName: typeName}
}

func (e exprType) captures(node *sitter.Node) []*sitter.Node {
	if node.Type() != e.typeName {
		return nil
	}
	return []*sitter.Node{}
}

// Capture nodes
type exprCapture struct {
	inner expr
}

func eCapture(inner expr) exprCapture {
	return exprCapture{inner: inner}
}

func (e exprCapture) captures(node *sitter.Node) []*sitter.Node {
	innerCaptures := e.inner.captures(node)
	if innerCaptures == nil {
		return nil
	}
	return append(innerCaptures, node)
}

// Match field
type exprField struct {
	field string
	expr  expr
}

func eField(field string, expr expr) exprField {
	return exprField{field: field, expr: expr}
}

func (e exprField) captures(node *sitter.Node) []*sitter.Node {
	child := node.ChildByFieldName(e.field)
	if child == nil {
		return nil
	}
	return e.expr.captures(child)
}

// Conjunction of expressions
type exprAnd struct {
	conjuncts []expr
}

func eAnd(conjuncts ...expr) exprAnd {
	return exprAnd{conjuncts: conjuncts}
}

func (e exprAnd) captures(node *sitter.Node) []*sitter.Node {
	captures := []*sitter.Node{}
	for _, conjunct := range e.conjuncts {
		innerCaptures := conjunct.captures(node)
		if innerCaptures == nil {
			return nil
		}
		captures = append(captures, innerCaptures...)
	}
	return captures
}

// Disjunction of expressions
type exprOr struct {
	disjuncts []expr
}

func eOr(disjuncts ...expr) exprOr {
	return exprOr{disjuncts: disjuncts}
}

func (e exprOr) captures(node *sitter.Node) []*sitter.Node {
	var captures []*sitter.Node
	for _, disjunct := range e.disjuncts {
		innerCaptures := disjunct.captures(node)
		if innerCaptures == nil {
			continue
		}
		if captures == nil {
			captures = []*sitter.Node{}
		}
		captures = append(captures, innerCaptures...)
	}
	return captures
}

// Match children
type exprChildren struct {
	inner expr
}

func eChildren(inner expr) exprChildren {
	return exprChildren{inner: inner}
}

func (e exprChildren) captures(node *sitter.Node) []*sitter.Node {
	var captures []*sitter.Node
	for i := 0; i < int(node.ChildCount()); i++ {
		innerCaptures := e.inner.captures(node.Child(i))
		if innerCaptures == nil {
			continue
		}
		if captures == nil {
			captures = []*sitter.Node{}
		}
		captures = append(captures, innerCaptures...)
	}
	return captures
}
