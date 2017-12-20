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

func TestPkgs_update_delete(t *testing.T) {
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

	t.Log("update")
	if err := dbutil.Transaction(ctx, globalDB, func(tx *sql.Tx) error {
		if err := Pkgs.update(ctx, tx, rp.ID, "go", pks); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	expPkgs := []sourcegraph.PackageInfo{{
		RepoID: rp.ID,
		Lang:   "go",
		Pkg:    map[string]interface{}{"name": "pkg"},
	}}
	gotPkgs, err := Pkgs.getAll(ctx, globalDB)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotPkgs, expPkgs) {
		t.Errorf("got %+v, expected %+v", gotPkgs, expPkgs)
	}

	t.Log("delete nothing")
	if err := Pkgs.Delete(ctx, 0); err != nil {
		t.Fatal(err)
	}
	gotPkgs, err = Pkgs.getAll(ctx, globalDB)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotPkgs, expPkgs) {
		t.Errorf("got %+v, expected %+v", gotPkgs, expPkgs)
	}

	t.Log("delete")
	if err := Pkgs.Delete(ctx, 1); err != nil {
		t.Fatal(err)
	}
	gotPkgs, err = Pkgs.getAll(ctx, globalDB)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotPkgs) > 0 {
		t.Errorf("expected all pkgs corresponding to repo %d deleted, but got %+v", rp.ID, gotPkgs)
	}
}

func TestPkgs_RefreshIndex(t *testing.T) {
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

	calledReposGetByURI := false
	Mocks.Repos.GetByURI = func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
		calledReposGetByURI = true
		switch repo {
		case "github.com/my/repo":
			return &sourcegraph.Repo{ID: rp.ID, URI: repo}, nil
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
		RepoID: rp.ID,
		Lang:   "typescript",
		Pkg: map[string]interface{}{
			"name":    "tspkg",
			"version": "2.2.2",
		},
	}}
	gotPkgs, err := Pkgs.getAll(ctx, globalDB)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotPkgs, expPkgs) {
		t.Errorf("got %+v, expected %+v", gotPkgs, expPkgs)
	}
}

func TestPkgs_ListPackages(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

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
		if err := dbutil.Transaction(ctx, globalDB, func(tx *sql.Tx) error {
			if _, err := tx.ExecContext(ctx, `INSERT INTO repo(id) VALUES ($1)`, repo); err != nil {
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
	}
	{ // Test case 2
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
	{ // Test case 4
		expPkgInfo := []sourcegraph.PackageInfo{{
			RepoID: 3,
			Lang:   "go",
			Pkg:    map[string]interface{}{"name": "pkg3", "version": "3.3.1"},
		}, {
			RepoID: 4,
			Lang:   "go",
			Pkg:    map[string]interface{}{"name": "pkg3", "version": "3.3.1"},
		},
			{
				RepoID: 5,
				Lang:   "go",
				Pkg:    map[string]interface{}{"name": "pkg3", "version": "3.3.1"},
			},
		}
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
	}
	{ // Test case 5, filter by repo ID
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
	}
}

func (p *pkgs) getAll(ctx context.Context, db dbQueryer) (packages []sourcegraph.PackageInfo, err error) {
	rows, err := db.QueryContext(ctx, "SELECT * FROM pkgs ORDER BY language ASC")
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
