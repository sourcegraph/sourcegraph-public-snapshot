package localstore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/sourcegraph/go-langserver/pkg/lsp"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspext"
)

func TestGlobalDeps_TotalRefsExpansion(t *testing.T) {
	tests := map[string][]string{
		// azul3d.org
		"github.com/azul3d/engine": []string{"azul3d.org/engine"},

		// dasa.cc
		"github.com/dskinner/ztext": []string{"dasa.cc/ztext"},

		// k8s.io
		"github.com/kubernetes/kubernetes":   []string{"k8s.io/kubernetes"},
		"github.com/kubernetes/apimachinery": []string{"k8s.io/apimachinery"},
		"github.com/kubernetes/client-go":    []string{"k8s.io/client-go"},
		"github.com/kubernetes/heapster":     []string{"k8s.io/heapster"},

		// golang.org/x
		"github.com/golang/net":    []string{"golang.org/x/net"},
		"github.com/golang/tools":  []string{"golang.org/x/tools"},
		"github.com/golang/oauth2": []string{"golang.org/x/oauth2"},
		"github.com/golang/crypto": []string{"golang.org/x/crypto"},
		"github.com/golang/sys":    []string{"golang.org/x/sys"},
		"github.com/golang/text":   []string{"golang.org/x/text"},
		"github.com/golang/image":  []string{"golang.org/x/image"},
		"github.com/golang/mobile": []string{"golang.org/x/mobile"},

		// google.golang.org
		"github.com/grpc/grpc-go":                []string{"google.golang.org/grpc"},
		"github.com/google/google-api-go-client": []string{"google.golang.org/api"},
		"github.com/golang/appengine":            []string{"google.golang.org/appengine"},

		// go.uber.org
		"github.com/uber-go/yarpc":    []string{"github.com/uber-go/yarpc", "go.uber.org/yarpc"},
		"github.com/uber-go/thriftrw": []string{"github.com/uber-go/thriftrw", "go.uber.org/thriftrw"},
		"github.com/uber-go/zap":      []string{"github.com/uber-go/zap", "go.uber.org/zap"},
		"github.com/uber-go/atomic":   []string{"github.com/uber-go/atomic", "go.uber.org/atomic"},
		"github.com/uber-go/fx":       []string{"github.com/uber-go/fx", "go.uber.org/fx"},

		// go4.org
		"github.com/camlistore/go4": []string{"go4.org"},

		// honnef.co
		"github.com/dominikh/go-staticcheck": []string{"honnef.co/go/staticcheck"},
		"github.com/dominikh/go-js-dom":      []string{"honnef.co/go/js/dom"},
		"github.com/dominikh/go-ssa":         []string{"honnef.co/go/ssa"},

		// gopkg.in
		"github.com/go-mgo/mgo":         []string{"github.com/go-mgo/mgo", "gopkg.in/mgo", "labix.org/v1/mgo", "labix.org/v2/mgo"},
		"github.com/go-yaml/yaml":       []string{"github.com/go-yaml/yaml", "gopkg.in/yaml", "labix.org/v1/yaml", "labix.org/v2/yaml"},
		"github.com/fatih/set":          []string{"github.com/fatih/set", "gopkg.in/fatih/set"},
		"github.com/juju/environschema": []string{"github.com/juju/environschema", "gopkg.in/juju/environschema"},
	}
	for input, want := range tests {
		got := repoURIToGoPathPrefixes(input)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %q want %q", got, want)
		}
	}

}

func TestGlobalDeps_update(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	mockListUserPrivateRepoIDs = func(ctx context.Context) ([]int32, error) {
		return []int32{1}, nil
	}
	Mocks.Repos.Get = func(ctx context.Context, repo int32) (*sourcegraph.Repo, error) {
		switch repo {
		case 1:
			return &sourcegraph.Repo{ID: repo}, nil
		default:
			return nil, errors.New("not found")
		}
	}

	repoID := int32(1)
	inputRefs := []lspext.DependencyReference{{
		Attributes: map[string]interface{}{"name": "dep1", "vendor": true},
	}}
	if err := dbutil.Transaction(ctx, appDBH(ctx), func(tx *sql.Tx) error {
		return GlobalDeps.update(ctx, tx, "global_dep", "go", inputRefs, repoID)
	}); err != nil {
		t.Fatal(err)
	}

	wantRefs := []*sourcegraph.DependencyReference{{
		DepData: map[string]interface{}{"name": "dep1", "vendor": true},
		RepoID:  repoID,
	}}
	gotRefs, err := GlobalDeps.Dependencies(ctx, DependenciesOptions{
		Language: "go",
		DepData:  map[string]interface{}{"name": "dep1"},
		Limit:    20,
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Sort(sortDepRefs(wantRefs))
	sort.Sort(sortDepRefs(gotRefs))
	if !reflect.DeepEqual(gotRefs, wantRefs) {
		t.Errorf("got %+v, expected %+v", gotRefs, wantRefs)
	}
}

// 🚨 SECURITY: This test is critical for testing security 🚨
func TestGlobalDeps_RefreshIndex(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	mockListUserPrivateRepoIDs = func(ctx context.Context) ([]int32, error) {
		return []int32{3}, nil
	}
	Mocks.Repos.Get = func(ctx context.Context, repo int32) (*sourcegraph.Repo, error) {
		switch repo {
		case 3:
			return &sourcegraph.Repo{ID: repo}, nil
		default:
			return nil, errors.New("not found")
		}
	}

	xlangDone := mockXLang(func(ctx context.Context, mode string, rootPath lsp.DocumentURI, method string, params, results interface{}) error {
		switch method {
		case "workspace/xdependencies":
			res, ok := results.(*[]lspext.DependencyReference)
			if !ok {
				t.Fatalf("attempted to call workspace/xpackages with invalid return type %T", results)
			}
			if rootPath != "git://github.com/my/repo?aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
				t.Fatalf("unexpected rootPath: %q", rootPath)
			}
			switch mode {
			case "go_bg":
				*res = []lspext.DependencyReference{{
					Attributes: map[string]interface{}{
						"name":   "github.com/gorilla/dep",
						"vendor": true,
					},
				}}
			default:
				t.Fatalf("unexpected mode: %q", mode)
			}
		}
		return nil
	})
	defer xlangDone()

	// 🚨 SECURITY: This is critical for testing security 🚨
	calledReposGetByURI := false
	repoID := int32(3)
	Mocks.Repos.GetByURI = func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
		calledReposGetByURI = true
		switch repo {
		case "github.com/my/repo":
			return &sourcegraph.Repo{ID: repoID, URI: repo}, nil
		default:
			return nil, errors.New("not found")
		}
	}

	reposGetInventory := func(context.Context, *sourcegraph.RepoRevSpec) (*inventory.Inventory, error) {
		return &inventory.Inventory{Languages: []*inventory.Lang{{Name: "Go"}}}, nil
	}

	commitID := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if err := GlobalDeps.RefreshIndex(ctx, "github.com/my/repo", commitID, reposGetInventory); err != nil {
		t.Fatal(err)
	}
	if !calledReposGetByURI {
		t.Fatalf("!calledReposGetByURI")
	}

	wantRefs := []*sourcegraph.DependencyReference{{
		DepData: map[string]interface{}{"name": "github.com/gorilla/dep", "vendor": true},
		RepoID:  repoID,
	}}
	gotRefs, err := GlobalDeps.Dependencies(ctx, DependenciesOptions{
		Language: "go",
		DepData:  map[string]interface{}{"name": "github.com/gorilla/dep"},
		Limit:    20,
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Sort(sortDepRefs(wantRefs))
	sort.Sort(sortDepRefs(gotRefs))
	if !reflect.DeepEqual(gotRefs, wantRefs) {
		t.Errorf("got %+v, expected %+v", gotRefs, wantRefs)
	}
}

// 🚨 SECURITY: This test is critical for testing security 🚨
func TestGlobalDeps_Dependencies(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	mockListUserPrivateRepoIDs = func(ctx context.Context) ([]int32, error) {
		return []int32{1, 2, 3, 4}, nil
	}

	// 🚨 SECURITY: This is critical for testing security 🚨
	calledReposGet := false
	Mocks.Repos.Get = func(ctx context.Context, repo int32) (*sourcegraph.Repo, error) {
		calledReposGet = true
		switch repo {
		case 1, 2, 3, 4:
			return &sourcegraph.Repo{ID: repo}, nil
		case 5:
			return nil, errors.New("unauthorized")
		default:
			return nil, errors.New("not found")
		}
	}

	inputRefs := map[int32][]lspext.DependencyReference{
		1: []lspext.DependencyReference{{Attributes: map[string]interface{}{"name": "github.com/gorilla/dep2", "vendor": true}}},
		2: []lspext.DependencyReference{{Attributes: map[string]interface{}{"name": "github.com/gorilla/dep3", "vendor": true}}},
		3: []lspext.DependencyReference{{Attributes: map[string]interface{}{"name": "github.com/gorilla/dep4", "vendor": true}}},
		4: []lspext.DependencyReference{{Attributes: map[string]interface{}{"name": "github.com/gorilla/dep4", "vendor": true}}},
		5: []lspext.DependencyReference{{Attributes: map[string]interface{}{"name": "github.com/gorilla/dep4", "vendor": true}}},
	}

	for repoID, inputRefs := range inputRefs {
		if err := dbutil.Transaction(ctx, appDBH(ctx), func(tx *sql.Tx) error {
			return GlobalDeps.update(ctx, tx, "global_dep", "go", inputRefs, repoID)
		}); err != nil {
			t.Fatal(err)
		}
	}

	{ // Test case 1
		calledReposGet = false
		wantRefs := []*sourcegraph.DependencyReference{{
			DepData: map[string]interface{}{"name": "github.com/gorilla/dep2", "vendor": true},
			RepoID:  1,
		}}
		gotRefs, err := GlobalDeps.Dependencies(ctx, DependenciesOptions{
			Language: "go",
			DepData:  map[string]interface{}{"name": "github.com/gorilla/dep2"},
			Limit:    20,
		})
		if err != nil {
			t.Fatal(err)
		}
		sort.Sort(sortDepRefs(wantRefs))
		sort.Sort(sortDepRefs(gotRefs))
		if !reflect.DeepEqual(gotRefs, wantRefs) {
			t.Errorf("got %+v, expected %+v", gotRefs, wantRefs)
		}
		if !calledReposGet {
			t.Fatalf("!calledReposGet")
		}
	}
	{ // Test case 2
		calledReposGet = false
		wantRefs := []*sourcegraph.DependencyReference{{
			DepData: map[string]interface{}{"name": "github.com/gorilla/dep3", "vendor": true},
			RepoID:  2,
		}}
		gotRefs, err := GlobalDeps.Dependencies(ctx, DependenciesOptions{
			Language: "go",
			DepData:  map[string]interface{}{"name": "github.com/gorilla/dep3"},
			Limit:    20,
		})
		if err != nil {
			t.Fatal(err)
		}
		sort.Sort(sortDepRefs(wantRefs))
		sort.Sort(sortDepRefs(gotRefs))
		if !reflect.DeepEqual(gotRefs, wantRefs) {
			t.Errorf("got %+v, expected %+v", gotRefs, wantRefs)
		}
		if !calledReposGet {
			t.Fatalf("!calledReposGet")
		}
	}
	{ // Test case 3, permissions: filter out unauthorized repository from results
		calledReposGet = false
		wantRefs := []*sourcegraph.DependencyReference{{
			DepData: map[string]interface{}{"name": "github.com/gorilla/dep4", "vendor": true},
			RepoID:  3,
		}, {
			DepData: map[string]interface{}{"name": "github.com/gorilla/dep4", "vendor": true},
			RepoID:  4,
		}}
		gotRefs, err := GlobalDeps.Dependencies(ctx, DependenciesOptions{
			Language: "go",
			DepData:  map[string]interface{}{"name": "github.com/gorilla/dep4"},
			Limit:    20,
		})
		if err != nil {
			t.Fatal(err)
		}
		sort.Sort(sortDepRefs(wantRefs))
		sort.Sort(sortDepRefs(gotRefs))
		if !reflect.DeepEqual(gotRefs, wantRefs) {
			t.Errorf("got %+v, expected %+v", gotRefs, wantRefs)
		}
		if !calledReposGet {
			t.Fatalf("!calledReposGet")
		}
	}
}

type sortDepRefs []*sourcegraph.DependencyReference

func (s sortDepRefs) Len() int { return len(s) }

func (s sortDepRefs) Swap(a, b int) { s[a], s[b] = s[b], s[a] }

func (s sortDepRefs) Less(a, b int) bool {
	if s[a].RepoID != s[b].RepoID {
		return s[a].RepoID < s[b].RepoID
	}
	if !reflect.DeepEqual(s[a].DepData, s[b].DepData) {
		return stringMapLess(s[a].DepData, s[b].DepData)
	}
	return stringMapLess(s[a].Hints, s[b].Hints)
}

func stringMapLess(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return len(a) < len(b)
	}
	ak := make([]string, 0, len(a))
	for k := range a {
		ak = append(ak, k)
	}
	bk := make([]string, 0, len(b))
	for k := range b {
		bk = append(bk, k)
	}
	sort.Strings(ak)
	sort.Strings(bk)
	for i := range ak {
		if ak[i] != bk[i] {
			return ak[i] < bk[i]
		}
		// This does not consistentlbk order the output, but in the
		// cases we use this it will since it is just a simple value
		// like bool or string
		av, _ := json.Marshal(a[ak[i]])
		bv, _ := json.Marshal(b[bk[i]])
		if bytes.Equal(av, bv) {
			return string(av) < string(bv)
		}
	}
	return false
}
