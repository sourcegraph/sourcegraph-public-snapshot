package localstore

import (
	"database/sql"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspext"
)

func TestParseIncludePattern(t *testing.T) {
	tests := map[string]struct {
		exact  []string
		like   []string
		regexp string
	}{
		`^$`:              {exact: []string{""}},
		`(^$)`:            {exact: []string{""}},
		`((^$))`:          {exact: []string{""}},
		`^((^$))$`:        {exact: []string{""}},
		`^()$`:            {exact: []string{""}},
		`^(())$`:          {exact: []string{""}},
		`^a$`:             {exact: []string{"a"}},
		`(^a$)`:           {exact: []string{"a"}},
		`((^a$))`:         {exact: []string{"a"}},
		`^((^a$))$`:       {exact: []string{"a"}},
		`^(a)$`:           {exact: []string{"a"}},
		`^((a))$`:         {exact: []string{"a"}},
		`^a|b$`:           {like: []string{"a%", "%b"}}, // "|" has higher precedence than "^" or "$"
		`^(a)|(b)$`:       {like: []string{"a%", "%b"}}, // "|" has higher precedence than "^" or "$"
		`^(a|b)$`:         {exact: []string{"a", "b"}},
		`(^a$)|(^b$)`:     {exact: []string{"a", "b"}},
		`((^a$)|(^b$))`:   {exact: []string{"a", "b"}},
		`^((^a$)|(^b$))$`: {exact: []string{"a", "b"}},
		`^((a)|(b))$`:     {exact: []string{"a", "b"}},
		`abc`:             {like: []string{"%abc%"}},
		`a|b`:             {like: []string{"%a%", "%b%"}},
		`^a(b|c)$`:        {exact: []string{"ab", "ac"}},

		`^github\.com/foo/bar`: {like: []string{"github.com/foo/bar%"}},

		`(^github\.com/Microsoft/vscode$)|(^github\.com/sourcegraph/go-langserver$)`: {exact: []string{"github.com/Microsoft/vscode", "github.com/sourcegraph/go-langserver"}},

		// Avoid DoS when there are too many possible matches to enumerate.
		`^(a|b)(c|d)(e|f)(g|h)(i|j)(k|l)(m|n)$`: {regexp: `^(a|b)(c|d)(e|f)(g|h)(i|j)(k|l)(m|n)$`},
		`^[0-a]$`: {regexp: `^[0-a]$`},
	}
	for pattern, want := range tests {
		t.Run(pattern, func(t *testing.T) {
			exact, like, regexp, err := parseIncludePattern(pattern)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(exact, want.exact) {
				t.Errorf("got exact %q, want %q", exact, want.exact)
			}
			if !reflect.DeepEqual(like, want.like) {
				t.Errorf("got like %q, want %q", like, want.like)
			}
			if regexp != want.regexp {
				t.Errorf("got regexp %q, want %q", regexp, want.regexp)
			}
		})
	}
}

func TestRepos_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	if err := Repos.TryInsertNew(ctx, "myrepo", "", false, false); err != nil {
		t.Fatal(err)
	}

	rp, err := Repos.GetByURI(ctx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}

	pks := []lspext.PackageInformation{{
		Package: map[string]interface{}{"name": "pkg"},
		Dependencies: []lspext.DependencyReference{{
			Attributes: map[string]interface{}{"name": "dep1"},
		}},
	}}
	if err := dbutil.Transaction(ctx, globalDB, func(tx *sql.Tx) error {
		if err := Pkgs.update(ctx, tx, rp.ID, "go", pks); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	inputRefs := []lspext.DependencyReference{{
		Attributes: map[string]interface{}{"name": "dep1", "vendor": true},
	}}
	if err := dbutil.Transaction(ctx, globalDB, func(tx *sql.Tx) error {
		return GlobalDeps.update(ctx, tx, "global_dep", "go", inputRefs, rp.ID)
	}); err != nil {
		t.Fatal(err)
	}

	if err := Repos.Delete(ctx, rp.ID); err != nil {
		t.Fatal(err)
	}

	gotPkgs, err := Pkgs.getAll(ctx, globalDB)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotPkgs) > 0 {
		t.Errorf("expected no more pkgs after delete, but got %+v", gotPkgs)
	}

	gotRefs, err := GlobalDeps.Dependencies(ctx, DependenciesOptions{
		Language: "go",
		DepData:  map[string]interface{}{"name": "dep1"},
		Limit:    20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(gotRefs) > 0 {
		t.Errorf("expected no more refs after delete, but got %+v", gotRefs)
	}

	rp2, err := Repos.Get(ctx, rp.ID)
	if err != ErrRepoNotFound {
		t.Errorf("expected error %q, but got error %q with repo %v", ErrRepoNotFound, err, rp2)
	}
}
