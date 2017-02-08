package localstore

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspext"
)

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
	if err := dbutil.Transaction(ctx, appDBH(ctx).Db, func(tx *sql.Tx) error {
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

	xlangDone := mockXLang(func(ctx context.Context, mode, rootPath, method string, params, results interface{}) error {
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
		if err := dbutil.Transaction(ctx, appDBH(ctx).Db, func(tx *sql.Tx) error {
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
		if !reflect.DeepEqual(gotRefs, wantRefs) {
			t.Errorf("got %+v, expected %+v", gotRefs, wantRefs)
		}
		if !calledReposGet {
			t.Fatalf("!calledReposGet")
		}
	}
}
