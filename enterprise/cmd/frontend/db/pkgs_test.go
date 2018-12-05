package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/sourcegraph/go-lsp/lspext"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	dbtesting "github.com/sourcegraph/sourcegraph/cmd/frontend/db/testing"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/dbconn"
)

func TestPkgs_update_delete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	if err := db.Repos.Upsert(ctx, api.InsertRepoOp{Name: "myrepo", Description: "", Fork: false, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	rp, err := db.Repos.GetByName(ctx, "myrepo")
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
	if err := db.Transaction(ctx, dbconn.Global, func(tx *sql.Tx) error {
		if err := Pkgs.update(ctx, tx, rp.ID, "go", pks); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	expPkgs := []*api.PackageInfo{{
		RepoID: rp.ID,
		Lang:   "go",
		Pkg:    map[string]interface{}{"name": "pkg"},
	}}
	gotPkgs, err := Pkgs.getAll(ctx, dbconn.Global)
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
	gotPkgs, err = Pkgs.getAll(ctx, dbconn.Global)
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
	gotPkgs, err = Pkgs.getAll(ctx, dbconn.Global)
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
	ctx := dbtesting.TestContext(t)

	if err := db.Repos.Upsert(ctx, api.InsertRepoOp{Name: "myrepo", Description: "", Fork: false, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	rp, err := db.Repos.GetByName(ctx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}

	if err := Pkgs.UpdateIndexForLanguage(ctx, "typescript", rp.ID, []lspext.PackageInformation{
		{
			Package: map[string]interface{}{
				"name":    "tspkg",
				"version": "2.2.2",
			},
			Dependencies: []lspext.DependencyReference{},
		},
	}); err != nil {
		t.Fatal(err)
	}

	expPkgs := []*api.PackageInfo{{
		RepoID: rp.ID,
		Lang:   "typescript",
		Pkg: map[string]interface{}{
			"name":    "tspkg",
			"version": "2.2.2",
		},
	}}
	gotPkgs, err := Pkgs.getAll(ctx, dbconn.Global)
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
	ctx := dbtesting.TestContext(t)

	repoToPkgs := map[api.RepoID][]lspext.PackageInformation{
		1: {{
			Package: map[string]interface{}{"name": "pkg1", "version": "1.1.1"},
			Dependencies: []lspext.DependencyReference{{
				Attributes: map[string]interface{}{"name": "pkg1-dep", "version": "1.1.2"},
			}},
		}},
		2: {{
			Package: map[string]interface{}{"name": "pkg2", "version": "2.2.1"},
			Dependencies: []lspext.DependencyReference{{
				Attributes: map[string]interface{}{"name": "pkg2-dep", "version": "2.2.2"},
			}},
		}},
		3: {{Package: map[string]interface{}{"name": "pkg3", "version": "3.3.1"}}},
		4: {{Package: map[string]interface{}{"name": "pkg3", "version": "3.3.1"}}},
		5: {{Package: map[string]interface{}{"name": "pkg3", "version": "3.3.1"}}},
	}

	createdAt := time.Now()
	for repo, pks := range repoToPkgs {
		if err := db.Transaction(ctx, dbconn.Global, func(tx *sql.Tx) error {
			if _, err := tx.ExecContext(ctx, `INSERT INTO repo(id, uri, created_at) VALUES ($1, $2, $3)`, repo, strconv.Itoa(int(repo)), createdAt); err != nil {
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
		expPkgInfo := []*api.PackageInfo{{
			RepoID: 1,
			Lang:   "go",
			Pkg:    map[string]interface{}{"name": "pkg1", "version": "1.1.1"},
		}}
		op := &api.ListPackagesOp{
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
		expPkgInfo := []*api.PackageInfo{{
			RepoID: 1,
			Lang:   "go",
			Pkg:    map[string]interface{}{"name": "pkg1", "version": "1.1.1"},
		}}
		op := &api.ListPackagesOp{
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
		var expPkgInfo []*api.PackageInfo
		op := &api.ListPackagesOp{
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
		expPkgInfo := []*api.PackageInfo{{
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
		op := &api.ListPackagesOp{
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
		expPkgInfo := []*api.PackageInfo{{
			RepoID: 3,
			Lang:   "go",
			Pkg:    map[string]interface{}{"name": "pkg3", "version": "3.3.1"},
		}}
		op := &api.ListPackagesOp{
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

func (p *pkgs) getAll(ctx context.Context, db dbQueryer) (packages []*api.PackageInfo, err error) {
	rows, err := db.QueryContext(ctx, "SELECT * FROM pkgs ORDER BY language ASC")
	if err != nil {
		return nil, errors.Wrap(err, "query")
	}
	defer rows.Close()

	for rows.Next() {
		var (
			repo     api.RepoID
			language string
			pkg      string
		)
		if err := rows.Scan(&repo, &language, &pkg); err != nil {
			return nil, errors.Wrap(err, "Scan")
		}
		p := api.PackageInfo{
			RepoID: repo,
			Lang:   language,
		}
		if err := json.Unmarshal([]byte(pkg), &p.Pkg); err != nil {
			return nil, errors.Wrap(err, "unmarshaling package metadata from SQL scan")
		}
		packages = append(packages, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows error")
	}
	return packages, nil
}
