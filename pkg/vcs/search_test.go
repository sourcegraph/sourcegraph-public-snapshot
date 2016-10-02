package vcs_test

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/util"
)

func TestRepository_Search_LongLine(t *testing.T) {
	t.Parallel()
	// TODO(sqs): implement hg Searcher

	tmp, longline, err := createLongFile()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp)

	searchOpt := vcs.SearchOptions{
		Query:        "ab",
		QueryType:    vcs.FixedQuery,
		ContextLines: 1,
	}
	wantRes := []*vcs.SearchResult{
		{
			File:      "f1",
			StartLine: 1,
			EndLine:   1,
			Match:     longline,
		},
	}

	gitCommands := []string{
		"cp " + filepath.ToSlash(tmp) + " f1",
		"git add f1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit f1 -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}

	testGitRepositorySearch(t, gitCommands, searchOpt, wantRes)
}

func TestRepository_Search(t *testing.T) {
	t.Parallel()
	// TODO(sqs): implement hg Searcher

	searchOpt := vcs.SearchOptions{
		Query:        "xy",
		QueryType:    vcs.FixedQuery,
		ContextLines: 1,
	}
	wantRes := []*vcs.SearchResult{
		{
			File:      "f1",
			StartLine: 2,
			EndLine:   3,
			Match:     []byte("def\nxyz"),
		},
		{
			File:      "f2",
			StartLine: 1,
			EndLine:   1,
			Match:     []byte("xyz"),
		},
	}

	gitCommands := []string{
		"echo abc > f1",
		"echo def >> f1",
		"echo xyz >> f1",
		"echo xyz > f2",
		"git add f1 f2",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit f1 f2 -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}

	testGitRepositorySearch(t, gitCommands, searchOpt, wantRes)
}

// testGitRepositorySearch is a helper that tests repository search
// over a git repository specified by the initializtion in
// repoInitCommands
func testGitRepositorySearch(t *testing.T, repoInitCmds []string, searchOpt vcs.SearchOptions, wantRes []*vcs.SearchResult) {
	tests := map[string]struct {
		repo        vcs.Repository
		spec        vcs.CommitID
		opt         vcs.SearchOptions
		wantResults []*vcs.SearchResult
	}{
		"git cmd": {
			repo:        makeGitRepositoryCmd(t, repoInitCmds...),
			spec:        "master",
			opt:         searchOpt,
			wantResults: wantRes,
		},
	}

	for label, test := range tests {
		res, err := test.repo.Search(ctx, test.spec, test.opt)
		if err != nil {
			t.Errorf("%s: Search: %s", label, err)
			continue
		}

		if !reflect.DeepEqual(res, test.wantResults) {
			t.Errorf("%s: got results == %v, want %v", label, asJSON(res), asJSON(test.wantResults))
		}
	}
}

// Makes long file with abbbb... content. Returns file name, file's content, and error if any
func createLongFile() (file string, line []byte, err error) {

	longline := make([]byte, bufio.MaxScanTokenSize+1)
	for i := 0; i < len(longline); i++ {
		if i == 0 {
			longline[i] = 'a'
		} else {
			longline[i] = 'b'
		}
	}

	tmp, err := ioutil.TempFile("", "test")
	if err != nil {
		return "", longline, err
	}
	err = util.WriteFileWithPermissions(tmp.Name(), longline, 0666)
	return tmp.Name(), longline, err
}
