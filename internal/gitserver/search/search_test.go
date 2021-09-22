package search

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

// initGitRepository initializes a new Git repository and runs cmds in a new
// temporary directory (returned as dir).
func initGitRepository(t testing.TB, cmds ...string) string {
	t.Helper()
	dir := t.TempDir()
	cmds = append([]string{"git init"}, cmds...)
	for _, cmd := range cmds {
		out, err := gitCommand(dir, "bash", "-c", cmd).CombinedOutput()
		if err != nil {
			t.Fatalf("Command %q failed. Output was:\n\n%s", cmd, out)
		}
	}
	return dir
}

func gitCommand(dir, name string, args ...string) *exec.Cmd {
	c := exec.Command(name, args...)
	c.Dir = dir
	c.Env = append(os.Environ(), "GIT_CONFIG="+path.Join(dir, ".git", "config"))
	return c
}

func TestSearch(t *testing.T) {
	cmds := []string{
		"echo lorem ipsum dolor sit amet > file1",
		"git add -A",
		"GIT_COMMITTER_NAME=camden1 " +
			"GIT_COMMITTER_EMAIL=camden1@ccheek.com " +
			"GIT_COMMITTER_DATE=2006-01-02T15:04:05Z " +
			"GIT_AUTHOR_NAME=camden1 " +
			"GIT_AUTHOR_EMAIL=camden1@ccheek.com " +
			"GIT_AUTHOR_DATE=2006-01-02T15:04:05Z " +
			"git commit -m commit1 ",
		"echo consectetur adipiscing elit > file2",
		"git add -A",
		"GIT_COMMITTER_NAME=camden2 " +
			"GIT_COMMITTER_EMAIL=camden2@ccheek.com " +
			"GIT_COMMITTER_DATE=2006-01-02T15:04:05Z " +
			"GIT_AUTHOR_NAME=camden2 " +
			"GIT_AUTHOR_EMAIL=camden2@ccheek.com " +
			"GIT_AUTHOR_DATE=2006-01-02T15:04:05Z " +
			"git commit -m commit2",
	}
	dir := initGitRepository(t, cmds...)

	t.Run("match one", func(t *testing.T) {
		query := &protocol.MessageMatches{protocol.Regexp{regexp.MustCompile("commit2")}}
		tree := ToMatchTree(query)
		var commits []*LazyCommit
		var highlights []*protocol.CommitHighlights
		err := Search(context.Background(), dir, nil, tree, func(lc *LazyCommit, hl *protocol.CommitHighlights) bool {
			commits = append(commits, lc)
			highlights = append(highlights, hl)
			return true
		})
		require.NoError(t, err)
		require.Len(t, commits, 1)
		require.Len(t, highlights, 1)
	})

	t.Run("match both", func(t *testing.T) {
		query := &protocol.MessageMatches{protocol.Regexp{regexp.MustCompile("c")}}
		tree := ToMatchTree(query)
		var commits []*LazyCommit
		var highlights []*protocol.CommitHighlights
		err := Search(context.Background(), dir, nil, tree, func(lc *LazyCommit, hl *protocol.CommitHighlights) bool {
			commits = append(commits, lc)
			highlights = append(highlights, hl)
			return true
		})
		require.NoError(t, err)
		require.Len(t, commits, 2)
		require.Len(t, highlights, 2)
	})
}

func TestCommitScanner(t *testing.T) {
	cases := []struct {
		input    []byte
		expected []*RawCommit
	}{{
		input: []byte(
			"2061ba96d63cba38f20a76f039cf29ef68736b8a\x00\x00HEAD\x00Camden Cheek\x00camden@sourcegraph.com\x001632251505\x00Camden Cheek\x00camden@sourcegraph.com\x001632251505\x00fix import\n\x005230097b75dcbb2c214618dd171da4053aff18a6\x00\x00" +
				"5230097b75dcbb2c214618dd171da4053aff18a6\x00\x00HEAD\x00Camden Cheek\x00camden@sourcegraph.com\x001632248499\x00Camden Cheek\x00camden@sourcegraph.com\x001632248499\x00only set matches if they exist\n\x00\x00",
		),
		expected: []*RawCommit{{
			Hash:           []byte("2061ba96d63cba38f20a76f039cf29ef68736b8a"),
			RefNames:       []byte(""),
			SourceRefs:     []byte("HEAD"),
			AuthorName:     []byte("Camden Cheek"),
			AuthorEmail:    []byte("camden@sourcegraph.com"),
			AuthorDate:     []byte("1632251505"),
			CommitterName:  []byte("Camden Cheek"),
			CommitterEmail: []byte("camden@sourcegraph.com"),
			CommitterDate:  []byte("1632251505"),
			Message:        []byte("fix import"),
			ParentHashes:   []byte("5230097b75dcbb2c214618dd171da4053aff18a6"),
		}, {
			Hash:           []byte("5230097b75dcbb2c214618dd171da4053aff18a6"),
			RefNames:       []byte(""),
			SourceRefs:     []byte("HEAD"),
			AuthorName:     []byte("Camden Cheek"),
			AuthorEmail:    []byte("camden@sourcegraph.com"),
			AuthorDate:     []byte("1632248499"),
			CommitterName:  []byte("Camden Cheek"),
			CommitterEmail: []byte("camden@sourcegraph.com"),
			CommitterDate:  []byte("1632248499"),
			Message:        []byte("only set matches if they exist"),
			ParentHashes:   []byte(""),
		}},
	}}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			scanner := NewCommitScanner(bytes.NewReader(tc.input))
			var output []*RawCommit
			for scanner.Scan() {
				output = append(output, scanner.NextRawCommit())
			}
			require.NoError(t, scanner.Err())
			require.Equal(t, tc.expected, output)
		})
	}
}
