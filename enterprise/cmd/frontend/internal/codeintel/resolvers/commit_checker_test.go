package resolvers

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
)

func TestCachedCommitChecker(t *testing.T) {
	t.Run("uncached", func(t *testing.T) {
		commits := []gitserver.RepositoryCommit{
			{RepositoryID: 150, Commit: "deadbeef1"},
			{RepositoryID: 151, Commit: "deadbeef2"},
			{RepositoryID: 152, Commit: "deadbeef3"},
			{RepositoryID: 153, Commit: "deadbeef4"},
		}
		expectedExists := []bool{
			true,
			false,
			true,
			false,
		}

		gitserverClient := NewMockGitserverClient()
		gitserverClient.CommitsExistFunc.SetDefaultReturn(expectedExists, nil)

		ctx := context.Background()
		commitChecker := newCachedCommitChecker(gitserverClient)

		exists, err := commitChecker.existsBatch(ctx, commits)
		if err != nil {
			t.Fatalf("unexpected error checking commit batch: %s", err)
		}
		if diff := cmp.Diff(expectedExists, exists); diff != "" {
			t.Errorf("unexpected exists slice (-want +got):\n%s", diff)
		}
	})

	t.Run("fully cached", func(t *testing.T) {
		commits := []gitserver.RepositoryCommit{
			{RepositoryID: 150, Commit: "deadbeef1"},
			{RepositoryID: 151, Commit: "deadbeef2"},
			{RepositoryID: 152, Commit: "deadbeef3"},
			{RepositoryID: 153, Commit: "deadbeef4"},
		}
		expectedExists := []bool{
			true,
			false,
			true,
			false,
		}

		gitserverClient := NewMockGitserverClient()
		gitserverClient.CommitsExistFunc.SetDefaultReturn(expectedExists, nil)

		ctx := context.Background()
		commitChecker := newCachedCommitChecker(gitserverClient)

		exists1, err := commitChecker.existsBatch(ctx, commits)
		if err != nil {
			t.Fatalf("unexpected error checking commit batch: %s", err)
		}
		if diff := cmp.Diff(expectedExists, exists1); diff != "" {
			t.Errorf("unexpected exists slice (-want +got):\n%s", diff)
		}

		// Should be fully cached
		exists2, err := commitChecker.existsBatch(ctx, commits)
		if err != nil {
			t.Fatalf("unexpected error checking commit batch: %s", err)
		}
		if diff := cmp.Diff(expectedExists, exists2); diff != "" {
			t.Errorf("unexpected exists slice (-want +got):\n%s", diff)
		}

		// Should not have called underlying gitserver method twice
		if callCount := len(gitserverClient.CommitsExistFunc.History()); callCount != 1 {
			t.Errorf("unexpected call count. want=%d have=%d", 1, callCount)
		}
	})

	t.Run("partially cached", func(t *testing.T) {
		gitserverClient := NewMockGitserverClient()
		gitserverClient.CommitsExistFunc.SetDefaultHook(func(ctx context.Context, rcs []gitserver.RepositoryCommit) (exists []bool, _ error) {
			for _, rc := range rcs {
				exists = append(exists, len(rc.Commit)%2 == 0)
			}
			return
		})

		ctx := context.Background()
		commitChecker := newCachedCommitChecker(gitserverClient)

		commits1 := []gitserver.RepositoryCommit{
			{RepositoryID: 151, Commit: "even"},
			{RepositoryID: 151, Commit: "odd"},
			{RepositoryID: 152, Commit: "even"},
			{RepositoryID: 152, Commit: "odd"},
		}
		exists1, err := commitChecker.existsBatch(ctx, commits1)
		if err != nil {
			t.Fatalf("unexpected error checking commit batch: %s", err)
		}
		expected1 := []bool{
			true,
			false,
			true,
			false,
		}
		if diff := cmp.Diff(expected1, exists1); diff != "" {
			t.Errorf("unexpected exists slice (-want +got):\n%s", diff)
		}

		commits2 := []gitserver.RepositoryCommit{
			{RepositoryID: 152, Commit: "odd"}, // cached
			{RepositoryID: 153, Commit: "odd"},
			{RepositoryID: 153, Commit: "even"},
			{RepositoryID: 152, Commit: "even"}, // cached
		}
		exists2, err := commitChecker.existsBatch(ctx, commits2)
		if err != nil {
			t.Fatalf("unexpected error checking commit batch: %s", err)
		}
		expected2 := []bool{
			false, // cached
			false,
			true,
			true, // cached
		}
		if diff := cmp.Diff(expected2, exists2); diff != "" {
			t.Errorf("unexpected exists slice (-want +got):\n%s", diff)
		}

		// Should not have called underlying gitserver method twice
		if callCount := len(gitserverClient.CommitsExistFunc.History()); callCount != 2 {
			t.Errorf("unexpected call count. want=%d have=%d", 2, callCount)
		} else {
			calls := gitserverClient.CommitsExistFunc.History()

			if diff := cmp.Diff(commits1, calls[0].Arg1); diff != "" {
				t.Errorf("unexpected commits argument (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(commits2[1:3], calls[1].Arg1); diff != "" {
				t.Errorf("unexpected commits argument (-want +got):\n%s", diff)
			}
		}
	})
}
