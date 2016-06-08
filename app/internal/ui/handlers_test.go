package ui

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/kr/pretty"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/apptest"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httptestutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/testutil/srclibtest"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sqs/pbtypes"
)

func newTest() (*httptestutil.Client, *httptestutil.MockClients) {
	return apptest.New()
}

func getForTest(c interface {
	Get(url string) (*http.Response, error)
}, url string, wantStatusCode int) (meta, error) {
	resp, err := c.Get(url)
	if err != nil {
		return meta{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != wantStatusCode {
		return meta{}, fmt.Errorf("got HTTP %d, want %d", resp.StatusCode, wantStatusCode)
	}

	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return meta{}, err
	}
	m, err := parseMeta(html)
	if err != nil {
		return meta{}, err
	}

	if wantStatusCode != http.StatusOK {
		// Check that title contains error.
		if want := http.StatusText(wantStatusCode); !strings.Contains(m.Title, want) {
			return meta{}, fmt.Errorf("got title %q, want it to contain %q", m.Title, want)
		}
	}

	return *m, nil
}

func TestCatchAll(t *testing.T) {
	c, _ := newTest()
	m, err := getForTest(c, "/tools", http.StatusOK)
	if err != nil {
		t.Fatal(err)
	}
	if want := "Sourcegraph"; m.Title != want {
		t.Errorf("got title %q, want %q", m.Title, want)
	}
}

var urls = map[string]struct {
	repo string // repo is necessary (but not sufficient) for this route
	rev  string // rev is necessary (but not sufficient) for this route
	tree string // tree is necessary (but not sufficient) for this route
	blob string // blob is necessary (but not sufficient) for this route

	defUnitType, defUnit, defPath string // def is necessary (but not sufficient) for this route
}{
	"/r":                  {repo: "r"},
	"/r@v":                {repo: "r", rev: "v"},
	"/r@v/-/tree/d":       {repo: "r", rev: "v", tree: "d"},
	"/r@v/-/blob/f":       {repo: "r", rev: "v", blob: "f"},
	"/r@v/-/def/t/u/-/p":  {repo: "r", rev: "v", defUnitType: "t", defUnit: "u", defPath: "p"},
	"/r@v/-/info/t/u/-/p": {repo: "r", rev: "v", defUnitType: "t", defUnit: "u", defPath: "p"},
	"/r/-/builds":         {repo: "r"},
	"/r/-/builds/2":       {repo: "r"},
}

func metaDiff(a, b meta) string { return strings.Join(pretty.Diff(a, b), "\n") }

func init() {
	graph.RegisterMakeDefFormatter("t", func(*graph.Def) graph.DefFormatter {
		return srclibtest.Formatter{}
	})
}

func TestRepo_OK(t *testing.T) {
	c, mock := newTest()

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	var calledGet bool
	mock.Repos.Get_ = func(ctx context.Context, op *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		calledGet = true
		return &sourcegraph.Repo{
			ID:          1,
			URI:         "r",
			Description: "d",
		}, nil
	}
	// (Should not try to resolve the revision; see serveRepo for why.)

	wantMeta := meta{
		Title:        "r: d · Sourcegraph",
		ShortTitle:   "r",
		Description:  "d",
		CanonicalURL: "http://example.com/r",
		Index:        false,
		Follow:       false,
	}

	if m, err := getForTest(c, "/r", http.StatusOK); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(m, wantMeta) {
		t.Fatalf("meta mismatch:\n%s", metaDiff(m, wantMeta))
	}
	if !*calledReposResolve {
		t.Error("!calledReposResolve")
	}
	if !calledGet {
		t.Error("!calledGet")
	}
}

func TestRepo_Error_Resolve(t *testing.T) {
	c, mock := newTest()

	for url, req := range urls {
		if req.repo == "" {
			continue
		}

		calledReposResolve := mock.Repos.MockResolve_NotFound(t, req.repo)

		if _, err := getForTest(c, url, http.StatusNotFound); err != nil {
			t.Errorf("%s: %s", url, err)
			continue
		}
		if !*calledReposResolve {
			t.Errorf("%s: !calledReposResolve", url)
		}
	}
}

func TestRepo_Error_Get(t *testing.T) {
	c, mock := newTest()

	for url, req := range urls {
		if req.repo == "" {
			continue
		}

		calledReposResolve := mock.Repos.MockResolve_Local(t, req.repo, 1)
		var calledGet bool
		mock.Repos.Get_ = func(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
			calledGet = true
			return nil, grpc.Errorf(codes.NotFound, "")
		}

		if _, err := getForTest(c, url, http.StatusNotFound); err != nil {
			t.Errorf("%s: %s", url, err)
			continue
		}
		if !*calledReposResolve {
			t.Errorf("%s: !calledReposResolve", url)
		}
		if !calledGet {
			t.Errorf("%s: !calledGet", url)
		}
	}
}

func TestRepoRev_OK(t *testing.T) {
	c, mock := newTest()

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	var calledGet bool
	mock.Repos.Get_ = func(ctx context.Context, op *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		calledGet = true
		return &sourcegraph.Repo{
			ID:          1,
			URI:         "r",
			Description: "d",
		}, nil
	}
	calledReposResolveRev := mock.Repos.MockResolveRev_NoCheck(t, "c")

	wantMeta := meta{
		Title:        "r: d · Sourcegraph",
		ShortTitle:   "r",
		Description:  "d",
		CanonicalURL: "http://example.com/r@c",
		Index:        false,
		Follow:       false,
	}

	if m, err := getForTest(c, "/r@v", http.StatusOK); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(m, wantMeta) {
		t.Fatalf("meta mismatch:\n%s", metaDiff(m, wantMeta))
	}
	if !*calledReposResolve {
		t.Error("!calledReposResolve")
	}
	if !calledGet {
		t.Error("!calledGet")
	}
	if !*calledReposResolveRev {
		t.Error("!calledReposResolveRev")
	}
}

func TestRepoRev_Error(t *testing.T) {
	c, mock := newTest()

	for url, req := range urls {
		if req.repo == "" || req.rev == "" {
			continue
		}

		calledReposResolve := mock.Repos.MockResolve_Local(t, req.repo, 1)
		calledGet := mock.Repos.MockGet(t, 1)
		var calledReposResolveRev bool
		mock.Repos.ResolveRev_ = func(ctx context.Context, op *sourcegraph.ReposResolveRevOp) (*sourcegraph.ResolvedRev, error) {
			calledReposResolveRev = true
			return nil, grpc.Errorf(codes.NotFound, "")
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
		if !calledReposResolveRev {
			t.Errorf("%s: !calledReposResolveRev", url)
		}
	}
}

func TestBlob_OK(t *testing.T) {
	c, mock := newTest()

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	var calledGet bool
	mock.Repos.Get_ = func(ctx context.Context, op *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		calledGet = true
		return &sourcegraph.Repo{
			ID:          1,
			URI:         "r",
			Description: "desc",
		}, nil
	}
	calledReposResolveRev := mock.Repos.MockResolveRev_NoCheck(t, "c")
	calledRepoTreeGet := mock.RepoTree.MockGet_Return_NoCheck(t, &sourcegraph.TreeEntry{
		BasicTreeEntry: &sourcegraph.BasicTreeEntry{
			Name: "f",
			Type: sourcegraph.FileEntry,
		},
	})

	wantMeta := meta{
		Title:        "f · r · Sourcegraph",
		ShortTitle:   "f",
		Description:  "r — desc",
		CanonicalURL: "http://example.com/r@c/-/blob/f",
		Index:        false,
		Follow:       false,
	}

	if m, err := getForTest(c, "/r@v/-/blob/f", http.StatusOK); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(m, wantMeta) {
		t.Fatalf("meta mismatch:\n%s", metaDiff(m, wantMeta))
	}
	if !*calledReposResolve {
		t.Error("!calledReposResolve")
	}
	if !calledGet {
		t.Error("!calledGet")
	}
	if !*calledReposResolveRev {
		t.Error("!calledReposResolveRev")
	}
	if !*calledRepoTreeGet {
		t.Error("!calledRepoTreeGet")
	}
}

func TestBlob_NotFound_NonFile(t *testing.T) {
	c, mock := newTest()

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	calledGet := mock.Repos.MockGet(t, 1)
	calledReposResolveRev := mock.Repos.MockResolveRev_NoCheck(t, "v")
	calledRepoTreeGet := mock.RepoTree.MockGet_Return_NoCheck(t, &sourcegraph.TreeEntry{
		BasicTreeEntry: &sourcegraph.BasicTreeEntry{
			Name: "d",
			Type: sourcegraph.DirEntry,
		},
	})

	if _, err := getForTest(c, "/r@v/-/blob/d", http.StatusNotFound); err != nil {
		t.Fatal(err)
	}
	if !*calledReposResolve {
		t.Error("!calledReposResolve")
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if !*calledReposResolveRev {
		t.Error("!calledReposResolveRev")
	}
	if !*calledRepoTreeGet {
		t.Error("!calledRepoTreeGet")
	}
}

func TestBlob_Error(t *testing.T) {
	c, mock := newTest()

	for url, req := range urls {
		if req.repo == "" || req.rev == "" || req.blob == "" {
			continue
		}

		calledReposResolve := mock.Repos.MockResolve_Local(t, req.repo, 1)
		calledGet := mock.Repos.MockGet(t, 1)
		calledReposResolveRev := mock.Repos.MockResolveRev_NoCheck(t, "v")
		var calledRepoTreeGet bool
		mock.RepoTree.Get_ = func(ctx context.Context, op *sourcegraph.RepoTreeGetOp) (*sourcegraph.TreeEntry, error) {
			calledRepoTreeGet = true
			return nil, grpc.Errorf(codes.NotFound, "")
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
		if !calledRepoTreeGet {
			t.Errorf("%s: !calledRepoTreeGet", url)
		}
	}
}

func TestDef_OK(t *testing.T) {
	c, mock := newTest()

	tests := []struct {
		rev       string
		defOrInfo string // "def" (for Def route) or "info" (for DefInfo route)

		wantCanonURL    string
		wantTitlePrefix string
		wantIndex       bool
		wantFollow      bool
	}{
		{"@v", "def", "/r@c/-/blob/f", "imp.scope.name · f", false, false},
		{"@v", "info", "/r@c/-/info/t/u/-/p", "imp.scope.name", false, false},
		{"@b", "def", "/r/-/blob/f", "imp.scope.name · f", false, true},
		{"@b", "info", "/r/-/info/t/u/-/p", "imp.scope.name", true, true},
		{"", "def", "/r/-/blob/f", "imp.scope.name · f", false, true},
		{"", "info", "/r/-/info/t/u/-/p", "imp.scope.name", true, true},
	}

	for _, test := range tests {
		calledReposResolve := mock.Repos.MockResolve_Local(t, "r", 1)
		var calledGet bool
		mock.Repos.Get_ = func(ctx context.Context, op *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
			calledGet = true
			return &sourcegraph.Repo{
				ID:            1,
				URI:           "r",
				Description:   "desc",
				DefaultBranch: "b",
			}, nil
		}
		calledReposResolveRev := mock.Repos.MockResolveRev_NoCheck(t, "c")
		calledReposGetSrclibDataVersionForPath := mock.Repos.MockGetSrclibDataVersionForPath_Current(t)
		calledDefsGet := mock.Defs.MockGet_Return(t, &sourcegraph.Def{
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
			DocHTML: &pbtypes.HTML{HTML: "<p><b>hello</b> world!</p>"},
		})

		wantMeta := meta{
			Title:        test.wantTitlePrefix + " · r · Sourcegraph",
			ShortTitle:   test.wantTitlePrefix,
			Description:  "imp.scope.name_imp.scope.typeName — hello world!",
			CanonicalURL: "http://example.com" + test.wantCanonURL,
			Index:        test.wantIndex,
			Follow:       test.wantFollow,
		}

		if m, err := getForTest(c, fmt.Sprintf("/r%s/-/%s/t/u/-/p", test.rev, test.defOrInfo), http.StatusOK); err != nil {
			t.Errorf("%#v: %s", test, err)
			continue
		} else if !reflect.DeepEqual(m, wantMeta) {
			t.Fatalf("%#v: meta mismatch:\n%s", test, metaDiff(m, wantMeta))
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
	}
}

func TestDef_Error(t *testing.T) {
	c, mock := newTest()

	for url, req := range urls {
		if req.repo == "" || req.rev == "" || req.defUnitType == "" || req.defUnit == "" || req.defPath == "" {
			continue
		}

		calledReposResolve := mock.Repos.MockResolve_Local(t, req.repo, 1)
		calledGet := mock.Repos.MockGet(t, 1)
		calledReposResolveRev := mock.Repos.MockResolveRev_NoCheck(t, "v")
		calledReposGetSrclibDataVersionForPath := mock.Repos.MockGetSrclibDataVersionForPath_Current(t)
		var calledDefsGet bool
		mock.Defs.Get_ = func(ctx context.Context, op *sourcegraph.DefsGetOp) (*sourcegraph.Def, error) {
			calledDefsGet = true
			return nil, grpc.Errorf(codes.NotFound, "")
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

func TestTree_OK(t *testing.T) {
	c, mock := newTest()

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	var calledGet bool
	mock.Repos.Get_ = func(ctx context.Context, op *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		calledGet = true
		return &sourcegraph.Repo{
			ID:          1,
			URI:         "r",
			Description: "desc",
		}, nil
	}
	calledReposResolveRev := mock.Repos.MockResolveRev_NoCheck(t, "c")
	calledRepoTreeGet := mock.RepoTree.MockGet_Return_NoCheck(t, &sourcegraph.TreeEntry{
		BasicTreeEntry: &sourcegraph.BasicTreeEntry{
			Name: "d",
			Type: sourcegraph.DirEntry,
		},
	})

	wantMeta := meta{
		Title:        "d · r · Sourcegraph",
		ShortTitle:   "d",
		Description:  "r — desc",
		CanonicalURL: "http://example.com/r@c/-/tree/d",
		Index:        false,
		Follow:       false,
	}

	if m, err := getForTest(c, "/r@v/-/tree/d", http.StatusOK); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(m, wantMeta) {
		t.Fatalf("meta mismatch:\n%s", metaDiff(m, wantMeta))
	}
	if !*calledReposResolve {
		t.Error("!calledReposResolve")
	}
	if !calledGet {
		t.Error("!calledGet")
	}
	if !*calledReposResolveRev {
		t.Error("!calledReposResolveRev")
	}
	if !*calledRepoTreeGet {
		t.Error("!calledRepoTreeGet")
	}
}

func TestTree_NotFound_NonDir(t *testing.T) {
	c, mock := newTest()

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	calledGet := mock.Repos.MockGet(t, 1)
	calledReposResolveRev := mock.Repos.MockResolveRev_NoCheck(t, "v")
	calledRepoTreeGet := mock.RepoTree.MockGet_Return_NoCheck(t, &sourcegraph.TreeEntry{
		BasicTreeEntry: &sourcegraph.BasicTreeEntry{
			Name: "f",
			Type: sourcegraph.FileEntry,
		},
	})

	if _, err := getForTest(c, "/r@v/-/tree/f", http.StatusNotFound); err != nil {
		t.Fatal(err)
	}
	if !*calledReposResolve {
		t.Error("!calledReposResolve")
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if !*calledReposResolveRev {
		t.Error("!calledReposResolveRev")
	}
	if !*calledRepoTreeGet {
		t.Error("!calledRepoTreeGet")
	}
}

func TestTree_Error(t *testing.T) {
	c, mock := newTest()

	for url, req := range urls {
		if req.repo == "" || req.rev == "" || req.tree == "" {
			continue
		}

		calledReposResolve := mock.Repos.MockResolve_Local(t, req.repo, 1)
		calledGet := mock.Repos.MockGet(t, 1)
		calledReposResolveRev := mock.Repos.MockResolveRev_NoCheck(t, "v")
		var calledRepoTreeGet bool
		mock.RepoTree.Get_ = func(ctx context.Context, op *sourcegraph.RepoTreeGetOp) (*sourcegraph.TreeEntry, error) {
			calledRepoTreeGet = true
			return nil, grpc.Errorf(codes.NotFound, "")
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
		if !calledRepoTreeGet {
			t.Errorf("%s: !calledRepoTreeGet", url)
		}
	}
}
