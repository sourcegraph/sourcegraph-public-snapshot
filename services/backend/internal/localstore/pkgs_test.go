package localstore

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspext"
)

func TestPkgs_update(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx, done := testContext()
	defer done()

	p := dbPkgs{}

	pkgs := []lspext.PackageInformation{{
		Package: map[string]interface{}{"name": "pkg"},
		Dependencies: []lspext.DependencyReference{{
			Attributes: map[string]interface{}{"name": "dep1"},
		}},
	}}

	if err := dbutil.Transaction(ctx, globalGraphDBH.Db, func(tx *sql.Tx) error {
		if err := p.update(ctx, tx, 1, "go", pkgs); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	expPkgs := []PackageInfo{{
		RepoID: 1,
		Lang:   "go",
		Pkg:    map[string]interface{}{"name": "pkg"},
	}}
	gotPkgs, err := p.get(ctx, globalGraphDBH.Db, "")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotPkgs, expPkgs) {
		t.Errorf("got %+v, expected %+v", gotPkgs, expPkgs)
	}
}

func TestPkgs_UnsafeRefreshIndex(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx, done := testContext()
	defer done()
	xlangDone := mockXlang(func(ctx context.Context, mode, rootPath, method string, params, results interface{}) error {
		switch method {
		case "workspace/packages":
			res, ok := results.(*[]lspext.PackageInformation)
			if !ok {
				t.Fatalf("attempted to call workspace/packages with invalid return type %T", results)
			}
			if rootPath != "git://github.com/my/repo?aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
				t.Fatalf("unexpected rootPath: %q", rootPath)
			}
			switch mode {
			case "go_bg":
				*res = []lspext.PackageInformation{{
					Package: map[string]interface{}{
						"name":    "gopkg",
						"version": "1.1.1",
					},
					Dependencies: []lspext.DependencyReference{},
				}}
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

	p := dbPkgs{}

	op := &sourcegraph.DefsRefreshIndexOp{RepoURI: "github.com/my/repo", RepoID: 1, CommitID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
	langs := []*inventory.Lang{{Name: "Go"}, {Name: "TypeScript"}}
	if err := p.UnsafeRefreshIndex(ctx, op, langs); err != nil {
		t.Fatal(err)
	}

	expPkgs := []PackageInfo{{
		RepoID: 1,
		Lang:   "go",
		Pkg: map[string]interface{}{
			"name":    "gopkg",
			"version": "1.1.1",
		},
	}, {
		RepoID: 1,
		Lang:   "typescript",
		Pkg: map[string]interface{}{
			"name":    "tspkg",
			"version": "2.2.2",
		},
	}}
	gotPkgs, err := p.get(ctx, globalGraphDBH.Db, "ORDER BY lang ASC")
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
	ctx, done := testContext()
	defer done()

	p := dbPkgs{}

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
	}

	for repo, pkgs := range repoToPkgs {
		if err := dbutil.Transaction(ctx, globalGraphDBH.Db, func(tx *sql.Tx) error {
			if err := p.update(ctx, tx, repo, "go", pkgs); err != nil {
				return err
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}

	{ // Test case 1
		expPkgInfo := []PackageInfo{{
			RepoID: 1,
			Lang:   "go",
			Pkg:    map[string]interface{}{"name": "pkg1", "version": "1.1.1"},
		}}
		op := &ListPackagesOp{
			Lang:     "go",
			PkgQuery: map[string]interface{}{"name": "pkg1"},
			Limit:    10,
		}
		gotPkgInfo, err := p.ListPackages(ctx, op)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(gotPkgInfo, expPkgInfo) {
			t.Errorf("got %+v, expected %+v", gotPkgInfo, expPkgInfo)
		}
	}
	{ // Test case 2
		expPkgInfo := []PackageInfo{{
			RepoID: 1,
			Lang:   "go",
			Pkg:    map[string]interface{}{"name": "pkg1", "version": "1.1.1"},
		}}
		op := &ListPackagesOp{
			Lang:     "go",
			PkgQuery: map[string]interface{}{"name": "pkg1", "version": "1.1.1"},
			Limit:    10,
		}
		gotPkgInfo, err := p.ListPackages(ctx, op)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(gotPkgInfo, expPkgInfo) {
			t.Errorf("got %+v, expected %+v", gotPkgInfo, expPkgInfo)
		}
	}
	{ // Test case 3
		var expPkgInfo []PackageInfo
		op := &ListPackagesOp{
			Lang:     "go",
			PkgQuery: map[string]interface{}{"name": "pkg1", "version": "2"},
			Limit:    10,
		}
		gotPkgInfo, err := p.ListPackages(ctx, op)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(gotPkgInfo, expPkgInfo) {
			t.Errorf("got %+v, expected %+v", gotPkgInfo, expPkgInfo)
		}
	}
}
