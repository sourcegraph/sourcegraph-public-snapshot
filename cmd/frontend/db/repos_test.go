package db

import (
	"context"
	"reflect"
	"testing"

	dbtesting "github.com/sourcegraph/sourcegraph/cmd/frontend/db/testing"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
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
		`^[0-a]$`:                               {regexp: `^[0-a]$`},
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
	ctx := dbtesting.TestContext(t)

	pkgsDeletedCalls := make(map[api.RepoID]struct{})
	Mocks.Pkgs.Delete = func(ctx context.Context, repo api.RepoID) error {
		pkgsDeletedCalls[repo] = struct{}{}
		return nil
	}

	if err := Repos.Upsert(ctx, api.InsertRepoOp{URI: "myrepo", Description: "", Fork: false, Enabled: true}); err != nil {
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
	if err := Pkgs.UpdateIndexForLanguage(ctx, "go", rp.ID, pks); err != nil {
		t.Fatal(err)
	}

	inputRefs := []lspext.DependencyReference{{
		Attributes: map[string]interface{}{"name": "dep1", "vendor": true},
	}}
	if err := GlobalDeps.UpdateIndexForLanguage(ctx, "go", rp.ID, inputRefs); err != nil {
		t.Fatal(err)
	}

	if err := Repos.Delete(ctx, rp.ID); err != nil {
		t.Fatal(err)
	}

	if _, wasDeleted := pkgsDeletedCalls[rp.ID]; !wasDeleted {
		t.Error("expected Pkgs.Delete to be called, but it wasn't")
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
	if !errcode.IsNotFound(err) {
		t.Errorf("expected repo not found, but got error %q with repo %v", err, rp2)
	}
}

func TestRepos_Count(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	if count, err := Repos.Count(ctx, ReposListOptions{Enabled: true}); err != nil {
		t.Fatal(err)
	} else if want := 0; count != want {
		t.Errorf("got %d, want %d", count, want)
	}

	if err := Repos.Upsert(ctx, api.InsertRepoOp{URI: "myrepo", Description: "", Fork: false, Enabled: true}); err != nil {
		t.Fatal(err)
	}

	if count, err := Repos.Count(ctx, ReposListOptions{Enabled: true}); err != nil {
		t.Fatal(err)
	} else if want := 1; count != want {
		t.Errorf("got %d, want %d", count, want)
	}

	repos, err := Repos.List(ctx, ReposListOptions{Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := Repos.Delete(ctx, repos[0].ID); err != nil {
		t.Fatal(err)
	}

	if count, err := Repos.Count(ctx, ReposListOptions{Enabled: true}); err != nil {
		t.Fatal(err)
	} else if want := 0; count != want {
		t.Errorf("got %d, want %d", count, want)
	}
}

func TestRepos_Upsert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	if _, err := Repos.GetByURI(ctx, "myrepo"); !errcode.IsNotFound(err) {
		if err == nil {
			t.Fatal("myrepo already present")
		} else {
			t.Fatal(err)
		}
	}

	if err := Repos.Upsert(ctx, api.InsertRepoOp{URI: "myrepo", Description: "", Fork: false, Enabled: true}); err != nil {
		t.Fatal(err)
	}

	rp, err := Repos.GetByURI(ctx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}

	if rp.URI != "myrepo" {
		t.Fatalf("rp.URI: %s != %s", rp.URI, "myrepo")
	}

	if err := Repos.Upsert(ctx, api.InsertRepoOp{URI: "myrepo", Description: "asdfasdf", Fork: false, Enabled: true}); err != nil {
		t.Fatal(err)
	}

	rp, err = Repos.GetByURI(ctx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}

	if rp.URI != "myrepo" {
		t.Fatalf("rp.URI: %s != %s", rp.URI, "myrepo")
	}
	if rp.Description != "asdfasdf" {
		t.Fatalf("rp.URI: %q != %q", rp.Description, "asdfasdf")
	}
}
