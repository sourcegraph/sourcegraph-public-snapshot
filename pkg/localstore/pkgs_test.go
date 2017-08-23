package localstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-langserver/pkg/lsp"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspext"
)

func TestPkgs_update(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	pks := []lspext.PackageInformation{{
		Package: map[string]interface{}{"name": "pkg"},
		Dependencies: []lspext.DependencyReference{{
			Attributes: map[string]interface{}{"name": "dep1"},
		}},
	}}

	if err := dbutil.Transaction(ctx, appDBH(ctx), func(tx *sql.Tx) error {
		if err := Pkgs.update(ctx, tx, 1, "go", pks); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	expPkgs := []sourcegraph.PackageInfo{{
		RepoID: 1,
		Lang:   "go",
		Pkg:    map[string]interface{}{"name": "pkg"},
	}}
	gotPkgs, err := Pkgs.getAll(ctx, appDBH(ctx))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotPkgs, expPkgs) {
		t.Errorf("got %+v, expected %+v", gotPkgs, expPkgs)
	}
}

// 🚨 SECURITY: This test is critical for testing security 🚨
func TestPkgs_RefreshIndex(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()
	xlangDone := mockXLang(func(ctx context.Context, mode string, rootPath lsp.DocumentURI, method string, params, results interface{}) error {
		switch method {
		case "workspace/xpackages":
			res, ok := results.(*[]lspext.PackageInformation)
			if !ok {
				t.Fatalf("attempted to call workspace/xpackages with invalid return type %T", results)
			}
			if rootPath != "git://github.com/my/repo?aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
				t.Fatalf("unexpected rootPath: %q", rootPath)
			}
			switch mode {
			case "typescript_bg":
				*res = []lspext.PackageInformation{{
					Package: map[string]interface{}{
						"name":    "tspkg",
						"version": "2.2.2",
					},
					Dependencies: []lspext.DependencyReference{},
				}}
			case "python_bg":
				*res = []lspext.PackageInformation{{
					Package: map[string]interface{}{
						"name":    "pypkg",
						"version": "3.3.3",
					},
					Dependencies: []lspext.DependencyReference{},
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
	Mocks.Repos.GetByURI = func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
		calledReposGetByURI = true
		switch repo {
		case "github.com/my/repo":
			return &sourcegraph.Repo{ID: 1, URI: repo}, nil
		default:
			return nil, errors.New("not found")
		}
	}

	reposGetInventory := func(context.Context, *sourcegraph.RepoRevSpec) (*inventory.Inventory, error) {
		return &inventory.Inventory{Languages: []*inventory.Lang{{Name: "TypeScript"}}}, nil
	}

	commitID := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if err := Pkgs.RefreshIndex(ctx, "github.com/my/repo", commitID, reposGetInventory); err != nil {
		t.Fatal(err)
	}
	if !calledReposGetByURI {
		t.Fatalf("!calledReposGetByURI")
	}

	expPkgs := []sourcegraph.PackageInfo{{
		RepoID: 1,
		Lang:   "typescript",
		Pkg: map[string]interface{}{
			"name":    "tspkg",
			"version": "2.2.2",
		},
	}}
	gotPkgs, err := Pkgs.getAll(ctx, appDBH(ctx))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotPkgs, expPkgs) {
		t.Errorf("got %+v, expected %+v", gotPkgs, expPkgs)
	}
}

// 🚨 SECURITY: This test is critical for testing security 🚨
func TestPkgs_ListPackages(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

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

	repoToPkgs := map[int32][]lspext.PackageInformation{
		1: []lspext.PackageInformation{{
			Package: map[string]interface{}{"name": "pkg1", "version": "1.1.1"},
			Dependencies: []lspext.DependencyReference{{
				Attributes: map[string]interface{}{"name": "pkg1-dep", "version": "1.1.2"},
			}},
		}},
		2: []lspext.PackageInformation{{
			Package: map[string]interface{}{"name": "pkg2", "version": "2.2.1"},
			Dependencies: []lspext.DependencyReference{{
				Attributes: map[string]interface{}{"name": "pkg2-dep", "version": "2.2.2"},
			}},
		}},
		3: []lspext.PackageInformation{{Package: map[string]interface{}{"name": "pkg3", "version": "3.3.1"}}},
		4: []lspext.PackageInformation{{Package: map[string]interface{}{"name": "pkg3", "version": "3.3.1"}}},
		5: []lspext.PackageInformation{{Package: map[string]interface{}{"name": "pkg3", "version": "3.3.1"}}},
	}

	for repo, pks := range repoToPkgs {
		if err := dbutil.Transaction(ctx, appDBH(ctx), func(tx *sql.Tx) error {
			if _, err := tx.Exec(`INSERT INTO repo(id, vcs, default_branch) VALUES ($1, '', 'master')`, repo); err != nil {
				return err
			}
			if err := Pkgs.update(ctx, tx, repo, "go", pks); err != nil {
				return err
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}

	{ // Test case 1
		calledReposGet = false
		expPkgInfo := []sourcegraph.PackageInfo{{
			RepoID: 1,
			Lang:   "go",
			Pkg:    map[string]interface{}{"name": "pkg1", "version": "1.1.1"},
		}}
		op := &sourcegraph.ListPackagesOp{
			Lang:     "go",
			PkgQuery: map[string]interface{}{"name": "pkg1"},
			Limit:    10,
		}
		gotPkgInfo, err := Pkgs.ListPackages(ctx, op)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(gotPkgInfo, expPkgInfo) {
			t.Errorf("got %+v, expected %+v", gotPkgInfo, expPkgInfo)
		}
		if !calledReposGet {
			t.Fatalf("!calledReposGet")
		}
	}
	{ // Test case 2
		calledReposGet = false
		expPkgInfo := []sourcegraph.PackageInfo{{
			RepoID: 1,
			Lang:   "go",
			Pkg:    map[string]interface{}{"name": "pkg1", "version": "1.1.1"},
		}}
		op := &sourcegraph.ListPackagesOp{
			Lang:     "go",
			PkgQuery: map[string]interface{}{"name": "pkg1", "version": "1.1.1"},
			Limit:    10,
		}
		gotPkgInfo, err := Pkgs.ListPackages(ctx, op)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(gotPkgInfo, expPkgInfo) {
			t.Errorf("got %+v, expected %+v", gotPkgInfo, expPkgInfo)
		}
		if !calledReposGet {
			t.Fatalf("!calledReposGet")
		}
	}
	{ // Test case 3
		var expPkgInfo []sourcegraph.PackageInfo
		op := &sourcegraph.ListPackagesOp{
			Lang:     "go",
			PkgQuery: map[string]interface{}{"name": "pkg1", "version": "2"},
			Limit:    10,
		}
		gotPkgInfo, err := Pkgs.ListPackages(ctx, op)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(gotPkgInfo, expPkgInfo) {
			t.Errorf("got %+v, expected %+v", gotPkgInfo, expPkgInfo)
		}
	}
	{ // Test case 4, permissions: filter out unauthorized repository from results
		calledReposGet = false
		expPkgInfo := []sourcegraph.PackageInfo{{
			RepoID: 3,
			Lang:   "go",
			Pkg:    map[string]interface{}{"name": "pkg3", "version": "3.3.1"},
		}, {
			RepoID: 4,
			Lang:   "go",
			Pkg:    map[string]interface{}{"name": "pkg3", "version": "3.3.1"},
		}}
		op := &sourcegraph.ListPackagesOp{
			Lang:     "go",
			PkgQuery: map[string]interface{}{"name": "pkg3"},
			Limit:    10,
		}
		gotPkgInfo, err := Pkgs.ListPackages(ctx, op)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(gotPkgInfo, expPkgInfo) {
			t.Errorf("got %+v, expected %+v", gotPkgInfo, expPkgInfo)
		}
		if !calledReposGet {
			t.Fatalf("!calledReposGet")
		}
	}
	{ // Test case 5, filter by repo ID
		calledReposGet = false
		expPkgInfo := []sourcegraph.PackageInfo{{
			RepoID: 3,
			Lang:   "go",
			Pkg:    map[string]interface{}{"name": "pkg3", "version": "3.3.1"},
		}}
		op := &sourcegraph.ListPackagesOp{
			Lang:   "go",
			RepoID: 3,
			Limit:  10,
		}
		gotPkgInfo, err := Pkgs.ListPackages(ctx, op)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(gotPkgInfo, expPkgInfo) {
			t.Errorf("got %+v, expected %+v", gotPkgInfo, expPkgInfo)
		}
		if !calledReposGet {
			t.Fatalf("!calledReposGet")
		}
	}
}

func (p *pkgs) getAll(ctx context.Context, db dbQueryer) (packages []sourcegraph.PackageInfo, err error) {
	rows, err := db.Query("SELECT * FROM pkgs ORDER BY language ASC")
	if err != nil {
		return nil, errors.Wrap(err, "query")
	}
	defer rows.Close()

	for rows.Next() {
		var (
			repoID   int32
			language string
			pkg      string
		)
		if err := rows.Scan(&repoID, &language, &pkg); err != nil {
			return nil, errors.Wrap(err, "Scan")
		}
		p := sourcegraph.PackageInfo{
			RepoID: repoID,
			Lang:   language,
		}
		if err := json.Unmarshal([]byte(pkg), &p.Pkg); err != nil {
			return nil, errors.Wrap(err, "unmarshaling package metadata from SQL scan")
		}
		packages = append(packages, p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows error")
	}
	return packages, nil
}
