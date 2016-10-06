package ui

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/htmlutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func TestDefLanding_OK(t *testing.T) {
	c := newTest()

	tests := []struct {
		rev string

		wantCanonURL    string
		wantTitlePrefix string
		wantIndex       bool
		wantFollow      bool
	}{
		{"@v", "/r@c/-/info/t/u/-/p", "imp.scope.name", false, false},
		{"@b", "/r/-/info/t/u/-/p", "imp.scope.name", true, false},
		{"", "/r/-/info/t/u/-/p", "imp.scope.name", true, false},
	}

	for _, test := range tests {
		calledReposResolve := backend.Mocks.Repos.MockResolve_Local(t, "r", 1)
		var calledGet bool
		backend.Mocks.Repos.Get = func(ctx context.Context, op *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
			calledGet = true
			return &sourcegraph.Repo{
				ID:            1,
				URI:           "r",
				Description:   "desc",
				DefaultBranch: "b",
			}, nil
		}
		calledReposResolveRev := backend.Mocks.Repos.MockResolveRev_NoCheck(t, "c")
		calledReposGetSrclibDataVersionForPath := backend.Mocks.Repos.MockGetSrclibDataVersionForPath_Current(t)
		calledDefsGet := backend.Mocks.Defs.MockGet_Return(t, &sourcegraph.Def{
			Def: graph.Def{
				Name: "aaa",
				DefKey: graph.DefKey{
					Repo:     "r",
					CommitID: "c",
					UnitType: "t",
					Unit:     "u",
					Path:     "p",
				},
				Exported: true,
				Kind:     "func",
				File:     "f",
			},
			FmtStrings: &graph.DefFormatStrings{
				DefKeyword: "func",
				Name:       graph.QualFormatStrings{ScopeQualified: "NewRouter"},
			},
			DocHTML: &htmlutil.HTML{HTML: "<p><b>hello</b> world!</p>"},
		})
		var calledDefsListRefLocations bool
		backend.Mocks.Defs.ListRefLocations = func(ctx context.Context, op *sourcegraph.DefsListRefLocationsOp) (*sourcegraph.RefLocationsList, error) {
			calledDefsListRefLocations = true
			return &sourcegraph.RefLocationsList{}, nil
		}
		var calledDefsListRefs bool
		backend.Mocks.Defs.ListRefs = func(ctx context.Context, op *sourcegraph.DefsListRefsOp) (*sourcegraph.RefList, error) {
			calledDefsListRefs = true
			return &sourcegraph.RefList{}, nil
		}
		calledRepoTreeGet := backend.Mocks.RepoTree.MockGet_Return_NoCheck(t, &sourcegraph.TreeEntry{
			FileRange: &sourcegraph.FileRange{},
		})
		calledAnnotationsList := backend.Mocks.Annotations.MockList(t)

		wantMeta := meta{
			Title:        test.wantTitlePrefix + " · r · Sourcegraph",
			ShortTitle:   test.wantTitlePrefix,
			Description:  "lang usage examples and docs for imp.scope.name_imp.scope.typeName — hello world!",
			CanonicalURL: "http://example.com" + test.wantCanonURL,
			Index:        test.wantIndex,
			Follow:       test.wantFollow,
		}

		if m, err := getForTest(c, fmt.Sprintf("/r%s/-/info/t/u/-/p", test.rev), http.StatusOK); err != nil {
			t.Errorf("%#v: %s", test, err)
			continue
		} else if !reflect.DeepEqual(m, wantMeta) {
			t.Errorf("%#v: meta mismatch:\n%s", test, metaDiff(m, wantMeta))
		}
		if !*calledReposResolve {
			t.Errorf("%#v: !calledReposResolve", test)
		}
		if !calledGet {
			t.Errorf("%#v: !calledGet", test)
		}
		if !*calledReposResolveRev {
			t.Errorf("%#v: !calledReposResolveRev", test)
		}
		if !*calledReposGetSrclibDataVersionForPath {
			t.Errorf("%#v: !calledReposGetSrclibDataVersionForPath", test)
		}
		if !*calledDefsGet {
			t.Errorf("%#v: !calledDefsGet", test)
		}
		if !calledDefsListRefLocations {
			t.Errorf("%#v: !calledDefsListRefLocations", test)
		}
		if !*calledRepoTreeGet {
			t.Errorf("%#v: !calledRepoTreeGet", test)
		}
		if !*calledAnnotationsList {
			t.Errorf("%#v: !calledAnnotationsList", test)
		}
		if !calledDefsListRefs {
			t.Errorf("%#v: !calledDefsListRefs", test)
		}
	}
}

func TestDefLanding_Error(t *testing.T) {
	c := newTest()

	for url, req := range urls {
		if req.repo == "" || req.rev == "" || req.defUnitType == "" || req.defUnit == "" || req.defPath == "" {
			continue
		}

		calledReposResolve := backend.Mocks.Repos.MockResolve_Local(t, req.repo, 1)
		calledGet := backend.Mocks.Repos.MockGet(t, 1)
		calledReposResolveRev := backend.Mocks.Repos.MockResolveRev_NoCheck(t, "v")
		calledReposGetSrclibDataVersionForPath := backend.Mocks.Repos.MockGetSrclibDataVersionForPath_Current(t)
		var calledDefsGet bool
		backend.Mocks.Defs.Get = func(ctx context.Context, op *sourcegraph.DefsGetOp) (*sourcegraph.Def, error) {
			calledDefsGet = true
			return nil, legacyerr.Errorf(legacyerr.NotFound, "")
		}

		if _, err := getForTest(c, url, http.StatusNotFound); err != nil {
			t.Errorf("%s: %s", url, err)
			continue
		}
		if !*calledReposResolve {
			t.Errorf("%s: !calledReposResolve", url)
		}
		if !*calledGet {
			t.Errorf("%s: !calledGet", url)
		}
		if !*calledReposResolveRev {
			t.Errorf("%s: !calledReposResolveRev", url)
		}
		if !*calledReposGetSrclibDataVersionForPath {
			t.Errorf("%s: !calledReposGetSrclibDataVersionForPath", url)
		}
		if !calledDefsGet {
			t.Errorf("%s: !calledDefsGet", url)
		}
	}
}
