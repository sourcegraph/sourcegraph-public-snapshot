package nolocalhost

import (
	"go/ast"
	"go/token"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/gobwas/glob"

	"github.com/sourcegraph/sourcegraph/dev/linters/nolint"
)

var excludePatterns = []string{
	"**_test.go",
	"**cmd/server/shared/nginx.go",
	"**dev/sg/sg_setup.go",
	"**pkg/conf/confdefaults/**",
	"**schema**",
	"**vendor**",
}

var compiledExcludePatterns = make([]glob.Glob, len(excludePatterns))

func init() {
	for i, p := range excludePatterns {
		compiledExcludePatterns[i] = glob.MustCompile(p, '/')
	}
}

// We generally prefer to use "127.0.0.1" instead of "localhost", because
// the Go DNS resolver fails to resolve "localhost" correctly in some
// situations (see https://github.com/sourcegraph/issues/issues/34 and
// https://github.com/sourcegraph/sourcegraph/issues/9129).

// If your usage of "localhost" is valid, then either
// 1) add the comment "CI:LOCALHOST_OK" to the line where "localhost" occurs, or
// 2) add an exclusion clause in the "git grep" command here and in no-localhost-guard.sh
//
// Ideally we would use nolint instead of CI:LOCALHOST_OK, but git grep (in the original script)
// can't handle checking the line before for nolint.
var Analyzer = nolint.Wrap(&analysis.Analyzer{
	Name: "nolocalhost",
	Doc:  "Disallow using 'localhost', preferring '127.0.0.1'. Adapted from ./dev/check/no-localhost-guard.sh",
	Run:  run,
})

func run(pass *analysis.Pass) (interface{}, error) {
	for _, f := range pass.Files {
		if !slices.ContainsFunc(compiledExcludePatterns, func(g glob.Glob) bool {
			return g.Match(pass.Fset.Position(f.Package).Filename)
		}) {
			v := &visitor{pass: pass}
			ast.Walk(v, f)
			// Not every comment is available when walking the AST, so instead we search in (*ast.File).Comments.
			for _, candidate := range v.candidates {
				hasOkComment := slices.ContainsFunc(f.Comments, func(cg *ast.CommentGroup) bool {
					return pass.Fset.Position(cg.Pos()).Line == pass.Fset.Position(candidate.Pos()).Line && strings.Contains(cg.Text(), "CI:LOCALHOST_OK")
				})
				if !hasOkComment {
					pass.ReportRangef(candidate, "disallowed instance of 'localhost', please use '127.0.0.1' instead or suffix the line with '// CI:LOCALHOST_OK'")
				}
			}
		}
	}

	return nil, nil
}

type visitor struct {
	pass       *analysis.Pass
	candidates []*ast.BasicLit
}

var _ (ast.Visitor) = &visitor{}

func (v *visitor) Visit(node ast.Node) (w ast.Visitor) {
	switch n := node.(type) {
	case *ast.BasicLit:
		if n.Kind == token.STRING && strings.Contains(n.Value, "localhost") {
			v.candidates = append(v.candidates, n)
		}
	}
	return v
}
