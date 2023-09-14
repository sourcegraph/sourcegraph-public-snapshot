package nolint

// This file is copied from:
// https://sourcegraph.com/github.com/pingcap/tidb@a0f2405981960ec705bf44f12b96ca8aec1506c4/-/blob/build/linter/util/util.go
// We copy it because we do not want to pull in the entire dependency.
//
// A few changes have been made:
// * nolint directives on a line is more accurate. We use the Slash position instead of the Node position
// * we do extra pre-processing when parsing directives
import (
	"go/ast"
	"go/token"
	"reflect"
	"strings"

	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/analysis/report"
)

type skipType int

const (
	skipNone skipType = iota
	skipLinter
	skipGroup
	skipFile
)

// Directive is a comment of the form '//lint:<command> [arguments...]' and `//nolint:<command>`.
// It represents instructions to the static analysis tool.
type Directive struct {
	Command   skipType
	Linters   []string
	Directive *ast.Comment
	Node      ast.Node
}

func parseDirective(s string) (cmd skipType, args []string) {
	if strings.HasPrefix(s, "//lint:") {
		s = strings.TrimPrefix(s, "//lint:")
		fields := strings.Split(s, " ")
		switch fields[0] {
		case "ignore":
			return skipLinter, fields[1:]
		case "file-ignore":
			return skipFile, fields[1:]
		}
		return skipNone, nil
	}
	// our comment directive can look like this:
	// //nolint:staticcheck // other things
	// or
	// //nolint:SA10101,SA999,SA19191 // other stuff
	s = strings.TrimPrefix(s, "//nolint:")
	// We first split on spaces and take the first part
	parts := strings.Split(s, " ")
	s = parts[0]
	// our directive can specify a range of linters seperated by comma
	parts = strings.Split(s, ",")
	return skipLinter, parts
}

// parseDirectives extracts all directives from a list of Go files.
func parseDirectives(files []*ast.File, fset *token.FileSet) []Directive {
	var dirs []Directive
	for _, f := range files {
		cm := ast.NewCommentMap(fset, f, f.Comments)
		for node, cgs := range cm {
			for _, cg := range cgs {
				for _, c := range cg.List {
					if !strings.HasPrefix(c.Text, "//lint:") && !strings.HasPrefix(c.Text, "//nolint:") {
						continue
					}
					cmd, args := parseDirective(c.Text)
					d := Directive{
						Command:   cmd,
						Linters:   args,
						Directive: c,
						Node:      node,
					}
					dirs = append(dirs, d)
				}
			}
		}
	}
	return dirs
}

func doDirectives(pass *analysis.Pass) (interface{}, error) {
	return parseDirectives(pass.Files, pass.Fset), nil
}

func skipAnalyzer(linters []string, analyzerName string) bool {
	for _, l := range linters {
		switch l {
		case "staticcheck":
			return strings.HasPrefix(analyzerName, "SA")
		case analyzerName:
			return true
		}
	}
	return false
}

// Directives is a fact that contains a list of directives.
var Directives = &analysis.Analyzer{
	Name:             "directives",
	Doc:              "extracts linter directives",
	Run:              doDirectives,
	RunDespiteErrors: true,
	ResultType:       reflect.TypeOf([]Directive{}),
}

// Wrap wraps a Analyzer and so that it will respect nolint directives
//
// It does this by replacing the run method with a method that first inspects
// whether there is a comment directive to skip the analyzer for this particular
// issue.
func Wrap(analyzer *analysis.Analyzer) *analysis.Analyzer {
	respectNolintDirectives(analyzer)
	return analyzer
}

// respectNolintDirectives updates an analyzer to make it work on nogo.
// They have "lint:ignore" or "nolint" to make the analyzer ignore the code.
func respectNolintDirectives(analyzer *analysis.Analyzer) {
	analyzer.Requires = append(analyzer.Requires, Directives)
	oldRun := analyzer.Run
	analyzer.Run = func(p *analysis.Pass) (interface{}, error) {
		pass := *p
		oldReport := p.Report
		pass.Report = func(diag analysis.Diagnostic) {
			dirs := pass.ResultOf[Directives].([]Directive)
			for _, dir := range dirs {
				cmd := dir.Command
				linters := dir.Linters
				switch cmd {
				case skipLinter:
					// we use the Directive Slash position, since that gives us the exact line where the directive is used
					ignorePos := report.DisplayPosition(pass.Fset, dir.Directive.Slash)
					nodePos := report.DisplayPosition(pass.Fset, diag.Pos)
					if ignorePos.Filename != nodePos.Filename || (ignorePos.Line != nodePos.Line && ignorePos.Line+1 != nodePos.Line) {
						// we're either in the wrong file for where this directive applies
						// OR
						// the line we're currently looking at does not match where this directive is defined
						continue
					}
					// we've found a offending line that has a directive ... let's see whether we should ignore it
					if skipAnalyzer(linters, analyzer.Name) {
						return
					}
				case skipFile:
					ignorePos := report.DisplayPosition(pass.Fset, dir.Node.Pos())
					nodePos := report.DisplayPosition(pass.Fset, diag.Pos)
					if ignorePos.Filename == nodePos.Filename {
						return
					}
				default:
					continue
				}
			}
			oldReport(diag)
		}
		return oldRun(&pass)
	}
}
