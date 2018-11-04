package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	dbtesting "github.com/sourcegraph/sourcegraph/cmd/frontend/db/testing"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
)

func TestGlobalDeps_TotalRefsExpansion(t *testing.T) {
	tests := map[api.RepoName][]string{
		// azul3d.org
		"github.com/azul3d/engine": {"azul3d.org/engine"},

		// dasa.cc
		"github.com/dskinner/ztext": {"dasa.cc/ztext"},

		// k8s.io
		"github.com/kubernetes/kubernetes":   {"k8s.io/kubernetes"},
		"github.com/kubernetes/apimachinery": {"k8s.io/apimachinery"},
		"github.com/kubernetes/client-go":    {"k8s.io/client-go"},
		"github.com/kubernetes/heapster":     {"k8s.io/heapster"},

		// golang.org/x
		"github.com/golang/net":    {"golang.org/x/net"},
		"github.com/golang/tools":  {"golang.org/x/tools"},
		"github.com/golang/oauth2": {"golang.org/x/oauth2"},
		"github.com/golang/crypto": {"golang.org/x/crypto"},
		"github.com/golang/sys":    {"golang.org/x/sys"},
		"github.com/golang/text":   {"golang.org/x/text"},
		"github.com/golang/image":  {"golang.org/x/image"},
		"github.com/golang/mobile": {"golang.org/x/mobile"},

		// google.golang.org
		"github.com/grpc/grpc-go":                {"google.golang.org/grpc"},
		"github.com/google/google-api-go-client": {"google.golang.org/api"},
		"github.com/golang/appengine":            {"google.golang.org/appengine"},

		// go.uber.org
		"github.com/uber-go/yarpc":    {"github.com/uber-go/yarpc", "go.uber.org/yarpc"},
		"github.com/uber-go/thriftrw": {"github.com/uber-go/thriftrw", "go.uber.org/thriftrw"},
		"github.com/uber-go/zap":      {"github.com/uber-go/zap", "go.uber.org/zap"},
		"github.com/uber-go/atomic":   {"github.com/uber-go/atomic", "go.uber.org/atomic"},
		"github.com/uber-go/fx":       {"github.com/uber-go/fx", "go.uber.org/fx"},

		// go4.org
		"github.com/camlistore/go4": {"go4.org"},

		// honnef.co
		"github.com/dominikh/go-staticcheck": {"honnef.co/go/staticcheck"},
		"github.com/dominikh/go-js-dom":      {"honnef.co/go/js/dom"},
		"github.com/dominikh/go-ssa":         {"honnef.co/go/ssa"},

		// gopkg.in
		"github.com/go-mgo/mgo":         {"github.com/go-mgo/mgo", "gopkg.in/mgo", "labix.org/v1/mgo", "labix.org/v2/mgo"},
		"github.com/go-yaml/yaml":       {"github.com/go-yaml/yaml", "gopkg.in/yaml", "labix.org/v1/yaml", "labix.org/v2/yaml"},
		"github.com/fatih/set":          {"github.com/fatih/set", "gopkg.in/fatih/set"},
		"github.com/juju/environschema": {"github.com/juju/environschema", "gopkg.in/juju/environschema"},
	}
	for input, want := range tests {
		got := repoNameToGoPathPrefixes(input)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %q want %q", got, want)
		}
	}

}

func TestGlobalDeps_update_delete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	if err := db.Repos.Upsert(ctx, api.InsertRepoOp{URI: "myrepo", Description: "", Fork: false, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	rp, err := db.Repos.GetByName(ctx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}
	repo := rp.ID

	inputRefs := []lspext.DependencyReference{{
		Attributes: map[string]interface{}{"name": "dep1", "vendor": true},
	}}
	if err := GlobalDeps.UpdateIndexForLanguage(ctx, "go", repo, inputRefs); err != nil {
		t.Fatal(err)
	}

	t.Log("update")
	wantRefs := []*api.DependencyReference{{
		Language: "go",
		DepData:  map[string]interface{}{"name": "dep1", "vendor": true},
		RepoID:   repo,
	}}
	gotRefs, err := GlobalDeps.Dependencies(ctx, db.DependenciesOptions{
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

	t.Log("delete other")
	if err := GlobalDeps.Delete(ctx, 345345345); err != nil {
		t.Fatal(err)
	}
	gotRefs, err = GlobalDeps.Dependencies(ctx, db.DependenciesOptions{
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

	t.Log("delete")
	if err := GlobalDeps.Delete(ctx, repo); err != nil {
		t.Fatal(err)
	}
	gotRefs, err = GlobalDeps.Dependencies(ctx, db.DependenciesOptions{
		Language: "go",
		DepData:  map[string]interface{}{"name": "dep1"},
		Limit:    20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(gotRefs) > 0 {
		t.Errorf("expected no matching refs, got %+v", gotRefs)
	}
}

func TestGlobalDeps_RefreshIndex(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	if err := db.Repos.Upsert(ctx, api.InsertRepoOp{URI: "myrepo", Description: "", Fork: false, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	repo, err := db.Repos.GetByName(ctx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}

	if err := GlobalDeps.UpdateIndexForLanguage(ctx, "go", repo.ID, []lspext.DependencyReference{{
		Attributes: map[string]interface{}{
			"name":   "github.com/gorilla/dep",
			"vendor": true,
		},
	}}); err != nil {
		t.Fatal(err)
	}

	wantRefs := []*api.DependencyReference{{
		Language: "go",
		DepData:  map[string]interface{}{"name": "github.com/gorilla/dep", "vendor": true},
		RepoID:   repo.ID,
	}}
	gotRefs, err := GlobalDeps.Dependencies(ctx, db.DependenciesOptions{
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

func TestGlobalDeps_Dependencies(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	repos := make([]api.RepoID, 5)
	for i := 0; i < 5; i++ {
		repoName := api.RepoName(fmt.Sprintf("myrepo-%d", i))
		if err := db.Repos.Upsert(ctx, api.InsertRepoOp{URI: repoName, Description: "", Fork: false, Enabled: true}); err != nil {
			t.Fatal(err)
		}
		rp, err := db.Repos.GetByName(ctx, repoName)
		if err != nil {
			t.Fatal(err)
		}
		repos[i] = rp.ID
	}

	inputRefs := map[api.RepoID][]lspext.DependencyReference{
		repos[0]: {{Attributes: map[string]interface{}{"name": "github.com/gorilla/dep2", "vendor": true}}},
		repos[1]: {{Attributes: map[string]interface{}{"name": "github.com/gorilla/dep3", "vendor": true}}},
		repos[2]: {{Attributes: map[string]interface{}{"name": "github.com/gorilla/dep4", "vendor": true}}},
		repos[3]: {{Attributes: map[string]interface{}{"name": "github.com/gorilla/dep4", "vendor": true}}},
		repos[4]: {{Attributes: map[string]interface{}{"name": "github.com/gorilla/dep4", "vendor": true}}},
	}
	for rp, deps := range inputRefs {
		err := GlobalDeps.UpdateIndexForLanguage(ctx, "go", rp, deps)
		if err != nil {
			t.Fatal(err)
		}
	}

	{ // Test case 1
		wantRefs := []*api.DependencyReference{{
			Language: "go",
			DepData:  map[string]interface{}{"name": "github.com/gorilla/dep2", "vendor": true},
			RepoID:   repos[0],
		}}
		gotRefs, err := GlobalDeps.Dependencies(ctx, db.DependenciesOptions{
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
	}
	{ // Test case 2
		wantRefs := []*api.DependencyReference{{
			Language: "go",
			DepData:  map[string]interface{}{"name": "github.com/gorilla/dep3", "vendor": true},
			RepoID:   repos[1],
		}}
		gotRefs, err := GlobalDeps.Dependencies(ctx, db.DependenciesOptions{
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
	}
	{ // Test case 3
		wantRefs := []*api.DependencyReference{{
			Language: "go",
			DepData:  map[string]interface{}{"name": "github.com/gorilla/dep4", "vendor": true},
			RepoID:   repos[2],
		}, {
			Language: "go",
			DepData:  map[string]interface{}{"name": "github.com/gorilla/dep4", "vendor": true},
			RepoID:   repos[3],
		},
			{
				Language: "go",
				DepData:  map[string]interface{}{"name": "github.com/gorilla/dep4", "vendor": true},
				RepoID:   repos[4],
			},
		}
		gotRefs, err := GlobalDeps.Dependencies(ctx, db.DependenciesOptions{
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
	}
}

type sortDepRefs []*api.DependencyReference

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
