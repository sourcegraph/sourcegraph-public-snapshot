package shared

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestEnsurePostgresVersion(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "../dockerfile.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	install := []string{}
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			if strings.HasPrefix(c.Text, "//docker:") {
				parts := strings.SplitN(c.Text[9:], " ", 2)
				switch parts[0] {
				case "install":
					install = append(install, strings.Fields(parts[1])...)
				}
			}
		}
	}

	for _, pkg := range install {
		if pkg == "'postgresql<9.7'" {
			return
		}
	}
	t.Fatal("Could not find postgres 9.6 specified in docker:install. We have to stay on postgres 9.6 since changing versions would cause existing customers to not run due to postgres data files only working on 9.6. Got:", install)
}
