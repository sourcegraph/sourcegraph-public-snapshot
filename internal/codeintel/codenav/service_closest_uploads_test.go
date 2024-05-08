package codenav

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/collections"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type idPathPair struct {
	uploadID int
	path     string
}

type TestCase struct {
	closestUploads        []shared.CompletedUpload
	lsifStoreAllowedPaths []idPathPair
	matchingOptions       shared.UploadMatchingOptions
	expectUploadIDs       []int
}

func TestGetClosestCompletedUploadsForBlob(t *testing.T) {
	const repoID = 37
	const missingCommitSHA = "C1"
	const presentCommitSHA = "C2"
	testCases := []TestCase{
		{
			closestUploads: []shared.CompletedUpload{
				{ID: 22, Commit: missingCommitSHA, Root: ""},
				{ID: 23, Commit: presentCommitSHA, Root: "subdir/"},
			},
			lsifStoreAllowedPaths: []idPathPair{{22, "a.c"}},
			matchingOptions:       shared.UploadMatchingOptions{Commit: "C2", Path: "a.c"},
			// bug fix: doesn't have upload for which path check fails
			expectUploadIDs: []int{},
		},
		{
			closestUploads: []shared.CompletedUpload{
				{ID: 22, Commit: missingCommitSHA, Root: "subdir/"},
				{ID: 23, Commit: presentCommitSHA, Root: ""},
			},
			lsifStoreAllowedPaths: []idPathPair{{23, "a.c"}},
			matchingOptions:       shared.UploadMatchingOptions{Commit: "C2", Path: "a.c"},
			// bug fix: has upload for which path check succeeds
			expectUploadIDs: []int{23},
		},
	}
	for _, testCase := range testCases {
		// Set up mocks
		mockRepoStore := defaultMockRepoStore()
		mockLsifStore := NewMockLsifStore()
		mockUploadSvc := NewMockUploadService()
		mockGitserverClient := gitserver.NewMockClient()

		// Init service
		svc := newService(observation.TestContextTB(t), mockRepoStore, mockLsifStore, mockUploadSvc, mockGitserverClient)

		closestUploads := slices.Clone(testCase.closestUploads)
		for i := range closestUploads {
			closestUploads[i].RepositoryID = repoID
		}

		mockUploadSvc.InferClosestUploadsFunc.SetDefaultReturn(closestUploads, nil)
		const testRepoName = "yummy.com/cake"
		mockRepoStore.GetReposSetByIDsFunc.PushReturn(map[api.RepoID]*types.Repo{
			repoID: {ID: repoID, Name: testRepoName},
		}, nil)

		mockGitserverClient.GetCommitFunc.SetDefaultHook(func(_ context.Context, repoName api.RepoName, commitID api.CommitID) (*gitdomain.Commit, error) {
			// C1 is deliberately missing from gitserver
			if string(repoName) == testRepoName && commitID == presentCommitSHA {
				return &gitdomain.Commit{ID: presentCommitSHA}, nil
			}
			return nil, &gitdomain.RevisionNotFoundError{}
		})

		mockLsifStore.GetPathExistsFunc.SetDefaultHook(func(_ context.Context, uploadID int, path string) (bool, error) {
			return collections.NewSet(testCase.lsifStoreAllowedPaths...).Has(
				idPathPair{uploadID: uploadID, path: path}), nil
		})

		testCase.matchingOptions.RepositoryID = repoID
		testCase.matchingOptions.RootToPathMatching = shared.RootMustEnclosePath
		filtered, err := svc.GetClosestCompletedUploadsForBlob(context.Background(), testCase.matchingOptions)
		require.NoError(t, err)

		gotIDs := []int{}
		for _, upload := range filtered {
			gotIDs = append(gotIDs, upload.ID)
		}
		if diff := cmp.Diff(gotIDs, testCase.expectUploadIDs, cmp.Transformer("Sort", func(in []int) []int {
			out := append([]int(nil), in...)
			slices.Sort(out)
			return out
		})); diff != "" {
			t.Errorf("unexpected filtered uploads (-want +got):\n%s", diff)
		}
	}
}
