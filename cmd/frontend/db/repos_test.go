package db

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
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

	if err := Repos.Upsert(ctx, api.InsertRepoOp{Name: "myrepo", Description: "", Fork: false, Enabled: true}); err != nil {
		t.Fatal(err)
	}

	rp, err := Repos.GetByName(ctx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}

	if err := Repos.Delete(ctx, rp.ID); err != nil {
		t.Fatal(err)
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

	if err := Repos.Upsert(ctx, api.InsertRepoOp{Name: "myrepo", Description: "", Fork: false, Enabled: true}); err != nil {
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

	if _, err := Repos.GetByName(ctx, "myrepo"); !errcode.IsNotFound(err) {
		if err == nil {
			t.Fatal("myrepo already present")
		} else {
			t.Fatal(err)
		}
	}

	if err := Repos.Upsert(ctx, api.InsertRepoOp{Name: "myrepo", Description: "", Fork: false, Enabled: true}); err != nil {
		t.Fatal(err)
	}

	rp, err := Repos.GetByName(ctx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}

	if rp.Name != "myrepo" {
		t.Fatalf("rp.Name: %s != %s", rp.Name, "myrepo")
	}
	if rp.ExternalRepo != nil {
		t.Fatalf("rp.ExternalRepo: %s != %s", rp.ExternalRepo, "<nil>")
	}

	ext := &api.ExternalRepoSpec{
		ID:          "ext:id",
		ServiceType: "test",
		ServiceID:   "ext:test",
	}
	if err := Repos.Upsert(ctx, api.InsertRepoOp{Name: "myrepo", Description: "asdfasdf", Fork: false, Enabled: true, ExternalRepo: ext}); err != nil {
		t.Fatal(err)
	}

	rp, err = Repos.GetByName(ctx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}

	if rp.Name != "myrepo" {
		t.Fatalf("rp.Name: %s != %s", rp.Name, "myrepo")
	}
	if rp.Description != "asdfasdf" {
		t.Fatalf("rp.Name: %q != %q", rp.Description, "asdfasdf")
	}
	if !reflect.DeepEqual(rp.ExternalRepo, ext) {
		t.Fatalf("rp.ExternalRepo: %s != %s", rp.ExternalRepo, ext)
	}

	// Rename. Detected by external repo
	if err := Repos.Upsert(ctx, api.InsertRepoOp{Name: "myrepo/renamed", Description: "asdfasdf", Fork: false, Enabled: true, ExternalRepo: ext}); err != nil {
		t.Fatal(err)
	}

	if _, err := Repos.GetByName(ctx, "myrepo"); !errcode.IsNotFound(err) {
		if err == nil {
			t.Fatal("myrepo should be renamed, but still present as myrepo")
		} else {
			t.Fatal(err)
		}
	}

	rp, err = Repos.GetByName(ctx, "myrepo/renamed")
	if err != nil {
		t.Fatal(err)
	}
	if rp.Name != "myrepo/renamed" {
		t.Fatalf("rp.Name: %s != %s", rp.Name, "myrepo/renamed")
	}
	if rp.Description != "asdfasdf" {
		t.Fatalf("rp.Name: %q != %q", rp.Description, "asdfasdf")
	}
	if !reflect.DeepEqual(rp.ExternalRepo, ext) {
		t.Fatalf("rp.ExternalRepo: %s != %s", rp.ExternalRepo, ext)
	}
}
