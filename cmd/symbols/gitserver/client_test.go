package gitserver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

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
