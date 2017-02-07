package localstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/pkg/errors"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspext"
)

func TestPkgs_update(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	p := pkgs{}

	pks := []lspext.PackageInformation{{
		Package: map[string]interface{}{"name": "pkg"},
		Dependencies: []lspext.DependencyReference{{
			Attributes: map[string]interface{}{"name": "dep1"},
		}},
	}}

	if err := dbutil.Transaction(ctx, appDBH(ctx).Db, func(tx *sql.Tx) error {
		if err := p.update(ctx, tx, 1, "go", pks); err != nil {
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
	gotPkgs, err := p.getAll(ctx, appDBH(ctx).Db)
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
	ctx := testContext()
	xlangDone := mockXLang(func(ctx context.Context, mode, rootPath, method string, params, results interface{}) error {
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

	p := pkgs{}

	op := &sourcegraph.DefsRefreshIndexOp{RepoURI: "github.com/my/repo", RepoID: 1, CommitID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
	langs := []*inventory.Lang{{Name: "TypeScript"}}
	if err := p.UnsafeRefreshIndex(ctx, op, langs); err != nil {
		t.Fatal(err)
	}

	expPkgs := []sourcegraph.PackageInfo{{
		RepoID: 1,
		Lang:   "typescript",
		Pkg: map[string]interface{}{
			"name":    "tspkg",
			"version": "2.2.2",
		},
	}}
	gotPkgs, err := p.getAll(ctx, appDBH(ctx).Db)
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

	p := pkgs{}

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
		if err := dbutil.Transaction(ctx, appDBH(ctx).Db, func(tx *sql.Tx) error {
			if err := p.update(ctx, tx, repo, "go", pks); err != nil {
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
		gotPkgInfo, err := p.ListPackages(ctx, op)
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
		gotPkgInfo, err := p.ListPackages(ctx, op)
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
		gotPkgInfo, err := p.ListPackages(ctx, op)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(gotPkgInfo, expPkgInfo) {
			t.Errorf("got %+v, expected %+v", gotPkgInfo, expPkgInfo)
		}
	}
	{ // Test case 4, permissions: filter out unauthorized repository from results
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
		gotPkgInfo, err := p.ListPackages(ctx, op)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(gotPkgInfo, expPkgInfo) {
			t.Errorf("got %+v, expected %+v", gotPkgInfo, expPkgInfo)
		}
	}

	if !calledReposGet {
		t.Fatalf("!calledReposGet")
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
