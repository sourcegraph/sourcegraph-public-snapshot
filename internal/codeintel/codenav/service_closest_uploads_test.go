package codenav

import (
	"context"
	"hash/fnv"
	"slices"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	lsifstoremocks "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/internal/lsifstore/mocks"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/collections"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type idPathPair struct {
	uploadID int
	path     core.UploadRelPath
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
	repoRootPath := core.NewRepoRelPathUnchecked
	uploadRootPath := core.NewUploadRelPathUnchecked

	testCases := []TestCase{
		{
			closestUploads: []shared.CompletedUpload{
				{ID: 22, Commit: missingCommitSHA, Root: ""},
				{ID: 23, Commit: presentCommitSHA, Root: "subdir/"},
			},
			lsifStoreAllowedPaths: []idPathPair{{22, uploadRootPath("a.c")}},
			matchingOptions:       shared.UploadMatchingOptions{Commit: "C2", Path: repoRootPath("a.c")},
			// bug fix: doesn't have upload for which path check fails
			expectUploadIDs: []int{},
		},
		{
			closestUploads: []shared.CompletedUpload{
				{ID: 22, Commit: missingCommitSHA, Root: "subdir/"},
				{ID: 23, Commit: presentCommitSHA, Root: ""},
			},
			lsifStoreAllowedPaths: []idPathPair{{23, uploadRootPath("a.c")}},
			matchingOptions:       shared.UploadMatchingOptions{Commit: "C2", Path: repoRootPath("a.c")},
			// bug fix: has upload for which path check succeeds
			expectUploadIDs: []int{23},
		},
	}
	for _, testCase := range testCases {
		// Set up mocks
		const testRepoName = "yummy.com/cake"
		fakeRepoStore := FakeMinimalRepoStore{data: map[api.RepoID]*types.Repo{repoID: {ID: repoID, Name: testRepoName}}}
		mockLsifStore := lsifstoremocks.NewMockLsifStore()
		mockUploadSvc := NewMockUploadService()
		mockGitserverClient := gitserver.NewMockClient()
		mockSearchClient := client.NewMockSearchClient()

		// Init service
		svc := newService(observation.TestContextTB(t), fakeRepoStore, mockLsifStore, mockUploadSvc, mockGitserverClient, mockSearchClient, log.NoOp())

		closestUploads := slices.Clone(testCase.closestUploads)
		for i := range closestUploads {
			closestUploads[i].RepositoryID = repoID
		}

		mockUploadSvc.InferClosestUploadsFunc.SetDefaultReturn(closestUploads, nil)

		mockGitserverClient.GetCommitFunc.SetDefaultHook(func(_ context.Context, repoName api.RepoName, commitID api.CommitID) (*gitdomain.Commit, error) {
			// C1 is deliberately missing from gitserver
			if string(repoName) == testRepoName && commitID == presentCommitSHA {
				return &gitdomain.Commit{ID: presentCommitSHA}, nil
			}
			return nil, &gitdomain.RevisionNotFoundError{}
		})

		mockLsifStore.FindDocumentIDsFunc.SetDefaultHook(findDocumentIDsFuncLimited(testCase.lsifStoreAllowedPaths))

		testCase.matchingOptions.RepositoryID = repoID
		testCase.matchingOptions.RootToPathMatching = shared.RootMustEnclosePath
		filtered, err := svc.GetClosestCompletedUploadsForBlob(context.Background(), testCase.matchingOptions)
		require.NoError(t, err)

		gotIDs := []int{}
		for _, upload := range filtered {
			gotIDs = append(gotIDs, upload.ID)
		}
		if diff := cmp.Diff(testCase.expectUploadIDs, gotIDs, cmp.Transformer("Sort", func(in []int) []int {
			out := append([]int(nil), in...)
			slices.Sort(out)
			return out
		})); diff != "" {
			t.Errorf("unexpected filtered uploads (-want +got):\n%s", diff)
		}
	}
}

type findDocumentIDsFuncType = func(_ context.Context, uploadIDToPathMap map[int]core.UploadRelPath) (map[int]int, error)

func findDocumentIDsFuncAllowAny() findDocumentIDsFuncType {
	return findDocumentIDsFuncImpl(true, nil)
}

func findDocumentIDsFuncLimited(allowedPaths []idPathPair) findDocumentIDsFuncType {
	return findDocumentIDsFuncImpl(false, allowedPaths)
}

func findDocumentIDsFuncImpl(allowAny bool, allowedPaths []idPathPair) findDocumentIDsFuncType {
	return func(_ context.Context, uploadIDToPathMap map[int]core.UploadRelPath) (map[int]int, error) {
		uploadPathSet := collections.NewSet(allowedPaths...)
		uploadIDToDocumentID := map[int]int{}
		for uploadID, path := range uploadIDToPathMap {
			if allowAny || uploadPathSet.Has(idPathPair{uploadID: uploadID, path: path}) {
				hasher := fnv.New64()
				hasher.Write([]byte(strconv.Itoa(uploadID)))
				hasher.Write([]byte(path.RawValue()))
				uploadIDToDocumentID[uploadID] = int(hasher.Sum64())
			}
		}
		return uploadIDToDocumentID, nil
	}
}
