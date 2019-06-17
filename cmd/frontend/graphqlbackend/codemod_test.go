package graphqlbackend

import (
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
)

type Args struct {
	MatchTemplate   string
	RewriteTemplate string
	FileFilter      string
}

func TestCodemod_validateArgsNoRegex(t *testing.T) {
	q, _ := query.ParseAndCheck("re.*gex")
	_, _, _, err := validateQuery(q)
	if err == nil {
		t.Fatalf("Expected query %v to fail", q)
	}
	if !strings.HasPrefix(err.Error(), "This looks like a regex search pattern.") {
		t.Fatalf("%v expected complaint about regex pattern. Got %s", q, err)
	}
}

func TestCodemod_validateArgsOk(t *testing.T) {
	q, _ := query.ParseAndCheck(`"not regex"`)
	_, _, _, err := validateQuery(q)
	if err != nil {
		t.Fatalf("Expected query %v to to be OK", q)
	}
}

func TestCodemod_resolver(t *testing.T) {
	raw := &rawCodemodResult{
		URI:                  "",
		RewrittenSource:      "",
		InPlaceSubstitutions: []inPlaceSubstitution{},
		Diff:                 "Not a valid diff",
	}
	_, err := toMatchResolver("", raw)
	if err == nil {
		t.Fatalf("Expected invalid diff for %v", raw.Diff)
	}
	if !strings.HasPrefix(err.Error(), "Invalid diff") {
		t.Fatalf("Expected error %q", err)
	}
}
