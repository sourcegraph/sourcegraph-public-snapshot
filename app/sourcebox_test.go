package app_test

import (
	"io/ioutil"
	"net/http"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"regexp"
	"strings"

	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/apptest"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func init() {
	handlerutil.AllowBrowsingUnbuiltRepo = func(*sourcegraph.Repo) bool { return false }
}

func treeEntryFixture(sourceContents []string) *sourcegraph.TreeEntry {
	entry := &sourcegraph.TreeEntry{
		TreeEntry: &vcsclient.TreeEntry{
			Contents: []byte(strings.Join(sourceContents, "\n")),
			Type:     vcsclient.FileEntry,
		},
		SourceCode: &sourcegraph.SourceCode{
			Lines: []*sourcegraph.SourceCodeLine{},
		},
		FileRange: &vcsclient.FileRange{
			StartLine: 1,
			EndLine:   int64(len(sourceContents) + 1),
		},
	}

	for _, lineContents := range sourceContents {
		entry.SourceCode.Lines = append(entry.SourceCode.Lines, &sourcegraph.SourceCodeLine{
			Tokens: []*sourcegraph.SourceCodeToken{
				{Label: lineContents},
			},
		})
	}
	return entry
}

func TestSourceboxDef(t *testing.T) {
	c, mock := apptest.New()

	def := &sourcegraph.Def{Def: graph.Def{DefKey: graph.DefKey{Repo: "my/repo", UnitType: "GoPackage", Unit: "u", Path: "p"}}}

	entry := treeEntryFixture([]string{"foo1234"})

	calledReposGet := mockRepoGet(mock, "my/repo")
	calledReposGetCommit := mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	calledDefsGet := mock.Defs.MockGet_Return(t, def)
	calledRepoTreeGet := mockTreeEntryGet(mock, entry)
	mockCurrentRepoBuild(mock)
	mockEnabledRepoConfig(mock)

	resp, err := c.GetOK(router.Rel.URLToSourceboxDef(def.DefKey, "js").String())
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), string(entry.Contents)) {
		t.Errorf("got body that does not contain %q (body was: %q)", entry.Contents, b)
	}

	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
	if !*calledReposGetCommit {
		t.Error("!calledReposGetCommit")
	}
	if !*calledDefsGet {
		t.Error("!calledDefsGet")
	}
	if !*calledRepoTreeGet {
		t.Error("!calledRepoTreeGet")
	}
}

func TestSourceboxDef_unbuiltDisplayEmpty(t *testing.T) {
	c, mock := apptest.New()

	calledReposGet := mockRepoGet(mock, "my/repo")
	calledReposGetCommit := mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	mock.Defs.Get_ = func(ctx context.Context, op *sourcegraph.DefsGetOp) (*sourcegraph.Def, error) {
		return nil, grpc.Errorf(codes.NotFound, "")
	}
	mockNoRepoBuild(mock)
	mockEnabledRepoConfig(mock)

	resp, err := c.Get(router.Rel.URLToSourceboxDef(graph.DefKey{Repo: "my/repo", UnitType: "GoPackage", Unit: "u", Path: "p"}, "js").String())
	if err != nil {
		t.Fatal(err)
	}
	if want := http.StatusNotFound; resp.StatusCode != want {
		t.Errorf("got HTTP %d, want %d", resp.StatusCode, want)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "Not built.") {
		t.Errorf("got non-empty body %q", b)
	}

	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
	if !*calledReposGetCommit {
		t.Error("!calledReposGetCommit")
	}
}

func TestSourceboxFile(t *testing.T) {
	c, mock := apptest.New()

	commitID := strings.Repeat("c", 40)

	entry := treeEntryFixture([]string{"foo1234"})

	calledReposGet := mockRepoGet(mock, "my/repo")
	calledReposGetCommit := mock.Repos.MockGetCommit_ByID_NoCheck(t, vcs.CommitID(commitID))
	calledRepoTreeGet := mockTreeEntryGet(mock, entry)
	mockSpecificRepoBuild(mock, commitID)
	mockEnabledRepoConfig(mock)

	entrySpec := sourcegraph.TreeEntrySpec{
		RepoRev: sourcegraph.RepoRevSpec{
			RepoSpec: sourcegraph.RepoSpec{URI: "my/repo"},
			Rev:      "c",
			CommitID: commitID,
		},
		Path: "p",
	}
	resp, err := c.GetOK(router.Rel.URLToSourceboxFile(entrySpec, "js").String())
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), string(entry.Contents)) {
		t.Errorf("got body that does not contain %q (body was: %q)", entry.Contents, b)
	}

	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
	if !*calledReposGetCommit {
		t.Error("!calledReposGetCommit")
	}
	if !*calledRepoTreeGet {
		t.Error("!calledRepoTreeGet")
	}
}

func TestSourceboxFile_unbuiltButStillDisplaysRawFile(t *testing.T) {
	c, mock := apptest.New()

	commitID := strings.Repeat("c", 40)

	entry := treeEntryFixture([]string{"foo1234"})

	calledReposGet := mockRepoGet(mock, "my/repo")
	calledReposGetCommit := mock.Repos.MockGetCommit_ByID_NoCheck(t, vcs.CommitID(commitID))
	calledRepoTreeGet := mockTreeEntryGet(mock, entry)
	mockNoRepoBuild(mock)
	mockEnabledRepoConfig(mock)

	entrySpec := sourcegraph.TreeEntrySpec{
		RepoRev: sourcegraph.RepoRevSpec{
			RepoSpec: sourcegraph.RepoSpec{URI: "my/repo"},
			Rev:      "c",
			CommitID: commitID,
		},
		Path: "p",
	}
	resp, err := c.GetOK(router.Rel.URLToSourceboxFile(entrySpec, "js").String())
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), string(entry.Contents)) {
		t.Errorf("got body that does not contain %q (body was: %q)", entry.Contents, b)
	}

	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
	if !*calledReposGetCommit {
		t.Error("!calledReposGetCommit")
	}
	if !*calledRepoTreeGet {
		t.Error("!calledRepoTreeGet")
	}
}

func TestSourceboxFile_lineNumbersEnabled(t *testing.T) {
	c, mock := apptest.New()

	commitID := strings.Repeat("c", 40)

	entry := treeEntryFixture([]string{"foo line1", "bar line2", "baz line3"})

	calledReposGet := mockRepoGet(mock, "my/repo")
	calledReposGetCommit := mock.Repos.MockGetCommit_ByID_NoCheck(t, vcs.CommitID(commitID))
	calledRepoTreeGet := mockTreeEntryGet(mock, entry)
	mockNoRepoBuild(mock)
	mockEnabledRepoConfig(mock)

	entrySpec := sourcegraph.TreeEntrySpec{
		RepoRev: sourcegraph.RepoRevSpec{
			RepoSpec: sourcegraph.RepoSpec{URI: "my/repo"},
			Rev:      "c",
			CommitID: commitID,
		},
		Path: "p",
	}

	resp, err := c.GetOK(router.Rel.URLToSourceboxFile(entrySpec, "js").String() + "?LineNumbers")
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	expectedLineCount := len(entry.SourceCode.Lines)
	rp := regexp.MustCompile("(?U)td.*class.*line-number.*[1-3].*/td")

	if lc := len(rp.FindAllString(string(b), -1)); lc != expectedLineCount {
		t.Errorf("got body that does not contain %d correct line numbers (actual count was: %d, body was: %s)", expectedLineCount, lc, b)
	}

	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
	if !*calledReposGetCommit {
		t.Error("!calledReposGetCommit")
	}
	if !*calledRepoTreeGet {
		t.Error("!calledRepoTreeGet")
	}
}
