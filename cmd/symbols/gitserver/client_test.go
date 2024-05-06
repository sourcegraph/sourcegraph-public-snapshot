package gitserver

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestParseGitDiffOutput(t *testing.T) {
	testCases := []struct {
		output          []byte
		expectedChanges Changes
		shouldError     bool
	}{
		{
			output: combineBytes(
				[]byte("A"), NUL, []byte("added1.json"), NUL,
				[]byte("M"), NUL, []byte("modified1.json"), NUL,
				[]byte("D"), NUL, []byte("deleted1.json"), NUL,
				[]byte("A"), NUL, []byte("added2.json"), NUL,
				[]byte("M"), NUL, []byte("modified2.json"), NUL,
				[]byte("D"), NUL, []byte("deleted2.json"), NUL,
				[]byte("A"), NUL, []byte("added3.json"), NUL,
				[]byte("M"), NUL, []byte("modified3.json"), NUL,
				[]byte("D"), NUL, []byte("deleted3.json"), NUL,
			),
			expectedChanges: Changes{
				Added:    []string{"added1.json", "added2.json", "added3.json"},
				Modified: []string{"modified1.json", "modified2.json", "modified3.json"},
				Deleted:  []string{"deleted1.json", "deleted2.json", "deleted3.json"},
			},
		},
		{
			output: combineBytes(
				[]byte("A"), NUL, []byte("added1.json"), NUL,
				[]byte("M"), NUL, []byte("modified1.json"), NUL,
				[]byte("D"), NUL,
			),
			shouldError: true,
		},
		{
			output: []byte{},
		},
	}

	for _, testCase := range testCases {
		changes, err := parseGitDiffOutput(testCase.output)
		if err != nil {
			if !testCase.shouldError {
				t.Fatalf("unexpected error parsing git diff output: %s", err)
			}
		} else if testCase.shouldError {
			t.Fatalf("expected error, got none")
		}

		if diff := cmp.Diff(testCase.expectedChanges, changes); diff != "" {
			t.Errorf("unexpected changes (-want +got):\n%s", diff)
		}
	}
}

func combineBytes(bss ...[]byte) (combined []byte) {
	for _, bs := range bss {
		combined = append(combined, bs...)
	}

	return combined
}

func TestGitserverClient_PaginatedRevList(t *testing.T) {
	allCommits := []*gitdomain.Commit{
		{ID: "4ac04f2761285633cd35188c696a6e08de03c00c"},
		{ID: "e7d0b23cb4e2e975ad657b163793bc83926c21b2"},
		{ID: "a04652fa1998a0a7d2f2f77ecb7021de943d3aab"},
	}

	allCommitIDs := []api.CommitID{}
	for _, c := range allCommits {
		allCommitIDs = append(allCommitIDs, c.ID)
	}

	t.Run("returns commits in reverse chronological order", func(t *testing.T) {
		inner := gitserver.NewMockClient()
		inner.CommitsFunc.SetDefaultReturn(allCommits, nil)
		client := &gitserverClient{
			innerClient: inner,
			operations:  newOperations(&observation.TestContext),
		}
		commits, _, err := client.paginatedRevList(context.Background(), "repo", "HEAD", 999)
		require.NoError(t, err)
		require.Equal(t, allCommitIDs, commits)
	})

	t.Run("returns next cursor when more commits exist", func(t *testing.T) {
		inner := gitserver.NewMockClient()
		for i := range allCommits {
			if len(allCommits) > i+1 {
				inner.CommitsFunc.PushReturn(allCommits[i:i+2], nil)
			} else {
				inner.CommitsFunc.PushReturn(allCommits[i:i+1], nil)
			}
		}
		client := &gitserverClient{
			innerClient: inner,
			operations:  newOperations(&observation.TestContext),
		}

		nextCursor := "HEAD"
		haveCommits := []api.CommitID{}
		for {
			commits, next, err := client.paginatedRevList(context.Background(), "repo", nextCursor, 1)
			require.NoError(t, err)
			nextCursor = next
			haveCommits = append(haveCommits, commits...)
			if nextCursor == "" {
				break
			}
		}
		require.Equal(t, allCommitIDs, haveCommits)
	})
}
