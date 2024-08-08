package codenav

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	lsifstoremocks "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/internal/lsifstore/mocks"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	sgtypes "github.com/sourcegraph/sourcegraph/internal/types"
)

const rangesDiff = `
diff --git sub1/changed.go sub1/changed.go
index deadbeef1..deadbeef2 100644
--- sub1/changed.go
+++ sub1/changed.go
@@ -16,2 +16,2 @@ const imageProcWorkers = 1
-       img, err := i.getSpec().imageCache.getOrCreate(i, conf, func() (*imageResource, image.Image, error) {
-                imageProcSem <- true
+       return i.getSpec().imageCache.getOrCreate(i, conf, func() (*imageResource, image.Image, error) {
+                defer func() {
`

func TestRanges(t *testing.T) {
	// Set up mocks
	fakeRepoStore := AllPresentFakeRepoStore{}
	mockLsifStore := lsifstoremocks.NewMockLsifStore()
	mockUploadSvc := NewMockUploadService()
	mockGitserverClient := gitserver.NewMockClient()
	mockGitserverClient.DiffFunc.SetDefaultHook(func(ctx context.Context, repo api.RepoName, opt gitserver.DiffOptions) (*gitserver.DiffFileIterator, error) {
		if len(opt.Paths) > 0 && opt.Paths[0] == "sub1/changed.go" {
			return gitserver.NewDiffFileIterator(io.NopCloser(strings.NewReader(rangesDiff))), nil
		}
		return gitserver.NewDiffFileIterator(io.NopCloser(bytes.NewReader([]byte{}))), nil
	})
	mockSearchClient := client.NewMockSearchClient()

	mockLsifStore.FindDocumentIDsFunc.SetDefaultHook(findDocumentIDsFuncAllowAny())

	// Init service
	svc := newService(observation.TestContextTB(t), fakeRepoStore, mockLsifStore, mockUploadSvc, mockGitserverClient, mockSearchClient, log.NoOp())

	// Set up request state
	mockRequestState := RequestState{}
	mockRequestState.SetLocalCommitCache(fakeRepoStore, mockGitserverClient)
	mockRequestState.SetLocalGitTreeTranslator(mockGitserverClient, &sgtypes.Repo{})
	uploads := []uploadsshared.CompletedUpload{
		{ID: 50, Commit: "deadbeef1", Root: "sub1/", RepositoryID: 42},
		{ID: 51, Commit: "deadbeef1", Root: "sub1/", RepositoryID: 42},
		{ID: 52, Commit: "deadbeef2", Root: "sub1/", RepositoryID: 42},
		{ID: 53, Commit: "deadbeef1", Root: "sub1/", RepositoryID: 42},
	}
	mockRequestState.SetUploadsDataLoader(uploads)

	lookupPath := core.NewRepoRelPathUnchecked("sub1/a.go")
	testLocation1 := shared.Usage{UploadID: 50, Path: uploadRelPath("a.go"), Range: testRange1}
	testLocation2 := shared.Usage{UploadID: 51, Path: uploadRelPath("a.go"), Range: testRange2}
	testLocation3 := shared.Usage{UploadID: 51, Path: uploadRelPath("a.go"), Range: testRange1}
	testLocation4 := shared.Usage{UploadID: 51, Path: uploadRelPath("a.go"), Range: testRange2}
	testLocation5 := shared.Usage{UploadID: 51, Path: uploadRelPath("a.go"), Range: testRange1}
	testLocation6 := shared.Usage{UploadID: 51, Path: uploadRelPath("a.go"), Range: testRange2}
	testLocation7 := shared.Usage{UploadID: 51, Path: uploadRelPath("a.go"), Range: testRange3}
	testLocation8 := shared.Usage{UploadID: 52, Path: uploadRelPath("a.go"), Range: testRange4}
	testLocation9 := shared.Usage{UploadID: 52, Path: uploadRelPath("changed.go"), Range: testRange6}

	ranges := []shared.CodeIntelligenceRange{
		{Range: testRange1, HoverText: "text1", Definitions: nil, References: []shared.Usage{testLocation1}, Implementations: []shared.Usage{}},
		{Range: testRange2, HoverText: "text2", Definitions: []shared.Usage{testLocation2}, References: []shared.Usage{testLocation3}, Implementations: []shared.Usage{}},
		{Range: testRange3, HoverText: "text3", Definitions: []shared.Usage{testLocation4}, References: []shared.Usage{testLocation5}, Implementations: []shared.Usage{}},
		{Range: testRange4, HoverText: "text4", Definitions: []shared.Usage{testLocation6}, References: []shared.Usage{testLocation7}, Implementations: []shared.Usage{}},
		{Range: testRange5, HoverText: "text5", Definitions: []shared.Usage{testLocation8}, References: nil, Implementations: []shared.Usage{}},
		{Range: testRange6, HoverText: "text6", Definitions: []shared.Usage{testLocation9}, References: nil, Implementations: []shared.Usage{}},
	}

	mockLsifStore.GetRangesFunc.PushReturn(ranges[0:1], nil)
	mockLsifStore.GetRangesFunc.PushReturn(ranges[1:4], nil)
	mockLsifStore.GetRangesFunc.PushReturn(ranges[4:], nil)

	mockRequest := PositionalRequestArgs{
		RequestArgs: RequestArgs{
			RepositoryID: 42,
			Commit:       mockCommit,
			Limit:        50,
		},
		Path:      lookupPath,
		Line:      10,
		Character: 20,
	}
	adjustedRanges, err := svc.GetRanges(context.Background(), mockRequest, mockRequestState, 10, 20)
	if err != nil {
		t.Fatalf("unexpected error querying ranges: %s", err)
	}

	adjustedLocation1 := shared.UploadLocation{Upload: uploads[0], Path: repoRelPath("sub1/a.go"), TargetCommit: "deadbeef", TargetRange: testRange1}
	adjustedLocation2 := shared.UploadLocation{Upload: uploads[1], Path: repoRelPath("sub1/a.go"), TargetCommit: "deadbeef", TargetRange: testRange2}
	adjustedLocation3 := shared.UploadLocation{Upload: uploads[1], Path: repoRelPath("sub1/a.go"), TargetCommit: "deadbeef", TargetRange: testRange1}
	adjustedLocation4 := shared.UploadLocation{Upload: uploads[1], Path: repoRelPath("sub1/a.go"), TargetCommit: "deadbeef", TargetRange: testRange2}
	adjustedLocation5 := shared.UploadLocation{Upload: uploads[1], Path: repoRelPath("sub1/a.go"), TargetCommit: "deadbeef", TargetRange: testRange1}
	adjustedLocation6 := shared.UploadLocation{Upload: uploads[1], Path: repoRelPath("sub1/a.go"), TargetCommit: "deadbeef", TargetRange: testRange2}
	adjustedLocation7 := shared.UploadLocation{Upload: uploads[1], Path: repoRelPath("sub1/a.go"), TargetCommit: "deadbeef", TargetRange: testRange3}
	adjustedLocation8 := shared.UploadLocation{Upload: uploads[2], Path: repoRelPath("sub1/a.go"), TargetCommit: "deadbeef", TargetRange: testRange4}

	expectedRanges := []AdjustedCodeIntelligenceRange{
		{Range: testRange1, HoverText: "text1", Definitions: []shared.UploadLocation{}, References: []shared.UploadLocation{adjustedLocation1}, Implementations: []shared.UploadLocation{}},
		{Range: testRange2, HoverText: "text2", Definitions: []shared.UploadLocation{adjustedLocation2}, References: []shared.UploadLocation{adjustedLocation3}, Implementations: []shared.UploadLocation{}},
		{Range: testRange3, HoverText: "text3", Definitions: []shared.UploadLocation{adjustedLocation4}, References: []shared.UploadLocation{adjustedLocation5}, Implementations: []shared.UploadLocation{}},
		{Range: testRange4, HoverText: "text4", Definitions: []shared.UploadLocation{adjustedLocation6}, References: []shared.UploadLocation{adjustedLocation7}, Implementations: []shared.UploadLocation{}},
		{Range: testRange5, HoverText: "text5", Definitions: []shared.UploadLocation{adjustedLocation8}, References: []shared.UploadLocation{}, Implementations: []shared.UploadLocation{}},
		// no definition expected, as the line has been changed and we filter those out from range requests
		{Range: testRange6, HoverText: "text6", Definitions: []shared.UploadLocation{}, References: []shared.UploadLocation{}, Implementations: []shared.UploadLocation{}},
	}
	if diff := cmp.Diff(expectedRanges, adjustedRanges, cmp.Comparer(core.RepoRelPath.Equal), cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("unexpected ranges (-want +got):\n%s", diff)
	}
}
