package ui

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/kr/pretty"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/tmpl"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httptestutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	vcstest "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/testing"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/util"
)

func newTest() *httptestutil.Client {
	backend.TestSkipZap = true
	tmpl.LoadOnce()
	backend.Mocks = backend.MockServices{}
	return httptestutil.NewTest(router)
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
	c := newTest()

	m, err := getForTest(c, "/tools", http.StatusOK)
	if err != nil {
		t.Fatal(err)
	}
	if want := "Sourcegraph"; m.Title != want {
		t.Errorf("got title %q, want %q", m.Title, want)
	}
}

var urls = map[string]struct {
	repo       string // repo is necessary (but not sufficient) for this route
	rev        string // rev is necessary (but not sufficient) for this route
	tree       string // tree is necessary (but not sufficient) for this route
	blob       string // blob is necessary (but not sufficient) for this route
	mode       string // blob is necessary (but not sufficient) for this route
	importPath string // blob is necessary (but not sufficient) for this route
	symbolPath string // blob is necessary (but not sufficient) for this route

	defUnitType, defUnit, defPath string // def is necessary (but not sufficient) for this route
}{
	"/r":            {repo: "r"},
	"/r@v":          {repo: "r", rev: "v"},
	"/r@v/-/tree/d": {repo: "r", rev: "v", tree: "d"},
	"/r@v/-/blob/f": {repo: "r", rev: "v", blob: "f"},
	"/m/i/-/s":      {mode: "m", importPath: "i", symbolPath: "s"},
}

func metaDiff(a, b meta) string { return strings.Join(pretty.Diff(a, b), "\n") }

func TestRepo_OK(t *testing.T) {
	c := newTest()

	var calledGet bool
	backend.Mocks.Repos.GetByURI = func(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
		calledGet = true
		return &sourcegraph.Repo{
			ID:          1,
			URI:         "r",
			Description: "d",
		}, nil
	}

	// (Should not try to resolve the revision; see serveRepo for why.)

	wantMeta := meta{
		Title:        "r: d - Sourcegraph",
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
	if !calledGet {
		t.Error("!calledGet")
	}
}

func TestRepo_Error_Resolve(t *testing.T) {
	c := newTest()

	for url, req := range urls {
		if req.repo == "" {
			continue
		}

		var calledGet bool
		backend.Mocks.Repos.GetByURI = func(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
			calledGet = true
			return nil, legacyerr.Errorf(legacyerr.NotFound, "not found")
		}

		if _, err := getForTest(c, url, http.StatusNotFound); err != nil {
			t.Errorf("%s: %s", url, err)
			continue
		}
		if !calledGet {
			t.Errorf("%s: !calledGet", url)
		}
	}
}

func TestRepo_Error_Get(t *testing.T) {
	c := newTest()

	for url, req := range urls {
		if req.repo == "" {
			continue
		}

		var calledGet bool
		backend.Mocks.Repos.GetByURI = func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
			calledGet = true
			return nil, legacyerr.Errorf(legacyerr.NotFound, "")
		}

		if _, err := getForTest(c, url, http.StatusNotFound); err != nil {
			t.Errorf("%s: %s", url, err)
			continue
		}
		if !calledGet {
			t.Errorf("%s: !calledGet", url)
		}
	}
}

func TestRepoRev_OK(t *testing.T) {
	c := newTest()

	var calledGet bool
	backend.Mocks.Repos.GetByURI = func(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
		calledGet = true
		return &sourcegraph.Repo{
			ID:          1,
			URI:         "r",
			Description: "d",
		}, nil
	}
	calledReposResolveRev := backend.Mocks.Repos.MockResolveRev_NoCheck(t, "c")

	wantMeta := meta{
		Title:        "r: d - Sourcegraph",
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
	if !calledGet {
		t.Error("!calledGet")
	}
	if !*calledReposResolveRev {
		t.Error("!calledReposResolveRev")
	}
}

func TestRepoRev_Error(t *testing.T) {
	c := newTest()

	for url, req := range urls {
		if req.repo == "" || req.rev == "" {
			continue
		}

		calledGet := backend.Mocks.Repos.MockGetByURI(t, req.repo, 1)
		var calledReposResolveRev bool
		backend.Mocks.Repos.ResolveRev = func(ctx context.Context, op *sourcegraph.ReposResolveRevOp) (*sourcegraph.ResolvedRev, error) {
			calledReposResolveRev = true
			return nil, legacyerr.Errorf(legacyerr.NotFound, "")
		}

		if _, err := getForTest(c, url, http.StatusNotFound); err != nil {
			t.Errorf("%s: %s", url, err)
			continue
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
	c := newTest()

	var calledGet bool
	backend.Mocks.Repos.GetByURI = func(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
		calledGet = true
		return &sourcegraph.Repo{
			ID:          1,
			URI:         "r",
			Description: "desc",
		}, nil
	}
	calledReposResolveRev := backend.Mocks.Repos.MockResolveRev_NoCheck(t, "c")

	mockRepo := vcstest.MockRepository{}
	mockRepo.Lstat_ = func(ctx context.Context, commit vcs.CommitID, name string) (os.FileInfo, error) {
		return &util.FileInfo{Name_: "f"}, nil
	}
	localstore.Mocks.RepoVCS.Open = func(ctx context.Context, repo int32) (vcs.Repository, error) {
		return mockRepo, nil
	}

	wantMeta := meta{
		Title:        "f · r - Sourcegraph",
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
	if !calledGet {
		t.Error("!calledGet")
	}
	if !*calledReposResolveRev {
		t.Error("!calledReposResolveRev")
	}
}

func TestBlob_NotFound_NonFile(t *testing.T) {
	c := newTest()

	calledGet := backend.Mocks.Repos.MockGetByURI(t, "r", 1)
	calledReposResolveRev := backend.Mocks.Repos.MockResolveRev_NoCheck(t, "v")

	mockRepo := vcstest.MockRepository{}
	mockRepo.Lstat_ = func(ctx context.Context, commit vcs.CommitID, name string) (os.FileInfo, error) {
		return &util.FileInfo{Name_: "d", Mode_: os.ModeDir}, nil
	}
	localstore.Mocks.RepoVCS.Open = func(ctx context.Context, repo int32) (vcs.Repository, error) {
		return mockRepo, nil
	}

	if _, err := getForTest(c, "/r@v/-/blob/d", http.StatusNotFound); err != nil {
		t.Fatal(err)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if !*calledReposResolveRev {
		t.Error("!calledReposResolveRev")
	}
}

func TestBlob_Error(t *testing.T) {
	c := newTest()

	for url, req := range urls {
		if req.repo == "" || req.rev == "" || req.blob == "" {
			continue
		}

		calledGet := backend.Mocks.Repos.MockGetByURI(t, req.repo, 1)
		calledReposResolveRev := backend.Mocks.Repos.MockResolveRev_NoCheck(t, "v")

		mockRepo := vcstest.MockRepository{}
		mockRepo.Lstat_ = func(ctx context.Context, commit vcs.CommitID, name string) (os.FileInfo, error) {
			return nil, legacyerr.Errorf(legacyerr.NotFound, "")
		}
		localstore.Mocks.RepoVCS.Open = func(ctx context.Context, repo int32) (vcs.Repository, error) {
			return mockRepo, nil
		}

		if _, err := getForTest(c, url, http.StatusNotFound); err != nil {
			t.Errorf("%s: %s", url, err)
			continue
		}
		if !*calledGet {
			t.Errorf("%s: !calledGet", url)
		}
		if !*calledReposResolveRev {
			t.Errorf("%s: !calledReposResolveRev", url)
		}
	}
}

func TestDefRedirect_OK(t *testing.T) {
	c := newTest()

	tests := map[string]string{
		"/r/-/refs/t/u/-/p": "/r/-/info/t/u/-/p",
		"/r/-/def/t/u/-/p":  "/r/-/info/t/u/-/p",
	}
	for origURL, wantURL := range tests {
		resp, err := c.GetNoFollowRedirects(origURL)
		if err != nil {
			t.Errorf("%s: Get: %s", origURL, err)
			continue
		}
		if want := http.StatusMovedPermanently; resp.StatusCode != want {
			t.Errorf("%s: got HTTP status code %d, want %d", origURL, resp.StatusCode, want)
		}
		if got := resp.Header.Get("location"); got != wantURL {
			t.Errorf("%s: got redirected to %q, want %q", origURL, got, wantURL)
		}
	}
}

func TestTree_OK(t *testing.T) {
	c := newTest()

	var calledGet bool
	backend.Mocks.Repos.GetByURI = func(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
		calledGet = true
		return &sourcegraph.Repo{
			ID:          1,
			URI:         "r",
			Description: "desc",
		}, nil
	}
	calledReposResolveRev := backend.Mocks.Repos.MockResolveRev_NoCheck(t, "c")

	mockRepo := vcstest.MockRepository{}
	mockRepo.Lstat_ = func(ctx context.Context, commit vcs.CommitID, name string) (os.FileInfo, error) {
		return &util.FileInfo{Name_: "d", Mode_: os.ModeDir}, nil
	}
	localstore.Mocks.RepoVCS.Open = func(ctx context.Context, repo int32) (vcs.Repository, error) {
		return mockRepo, nil
	}

	wantMeta := meta{
		Title:        "d · r - Sourcegraph",
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
	if !calledGet {
		t.Error("!calledGet")
	}
	if !*calledReposResolveRev {
		t.Error("!calledReposResolveRev")
	}
}

func TestTree_NotFound_NonDir(t *testing.T) {
	c := newTest()

	calledGet := backend.Mocks.Repos.MockGetByURI(t, "r", 1)
	calledReposResolveRev := backend.Mocks.Repos.MockResolveRev_NoCheck(t, "v")

	mockRepo := vcstest.MockRepository{}
	mockRepo.Lstat_ = func(ctx context.Context, commit vcs.CommitID, name string) (os.FileInfo, error) {
		return &util.FileInfo{Name_: "f"}, nil
	}
	localstore.Mocks.RepoVCS.Open = func(ctx context.Context, repo int32) (vcs.Repository, error) {
		return mockRepo, nil
	}

	if _, err := getForTest(c, "/r@v/-/tree/f", http.StatusNotFound); err != nil {
		t.Fatal(err)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if !*calledReposResolveRev {
		t.Error("!calledReposResolveRev")
	}
}

func TestTree_Error(t *testing.T) {
	c := newTest()

	for url, req := range urls {
		if req.repo == "" || req.rev == "" || req.tree == "" {
			continue
		}

		calledGet := backend.Mocks.Repos.MockGetByURI(t, req.repo, 1)
		calledReposResolveRev := backend.Mocks.Repos.MockResolveRev_NoCheck(t, "v")

		mockRepo := vcstest.MockRepository{}
		mockRepo.Lstat_ = func(ctx context.Context, commit vcs.CommitID, name string) (os.FileInfo, error) {
			return nil, legacyerr.Errorf(legacyerr.NotFound, "")
		}
		localstore.Mocks.RepoVCS.Open = func(ctx context.Context, repo int32) (vcs.Repository, error) {
			return mockRepo, nil
		}

		if _, err := getForTest(c, url, http.StatusNotFound); err != nil {
			t.Errorf("%s: %s", url, err)
			continue
		}
		if !*calledGet {
			t.Errorf("%s: !calledGet", url)
		}
		if !*calledReposResolveRev {
			t.Errorf("%s: !calledReposResolveRev", url)
		}
	}
}
