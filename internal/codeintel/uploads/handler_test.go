package uploads

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	codeinteltypes "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	uploadstoremocks "github.com/sourcegraph/sourcegraph/internal/uploadstore/mocks"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestHandle(t *testing.T) {
	setupRepoMocks(t)

	upload := codeinteltypes.Upload{
		ID:           42,
		Root:         "root/",
		Commit:       "deadbeef",
		RepositoryID: 50,
		Indexer:      "lsif-go",
	}

	mockWorkerStore := NewMockWorkerStore()
	mockDBStore := NewMockStore()
	mockRepoStore := NewMockRepoStore()
	mockLSIFStore := NewMockLsifStore()
	mockUploadStore := uploadstoremocks.NewMockStore()
	gitserverClient := NewMockGitserverClient()

	// Set default transaction behavior
	mockDBStore.TransactFunc.SetDefaultReturn(mockDBStore, nil)
	mockDBStore.DoneFunc.SetDefaultHook(func(err error) error { return err })

	// Set default transaction behavior
	mockLSIFStore.TransactFunc.SetDefaultReturn(mockLSIFStore, nil)
	mockDBStore.DoneFunc.SetDefaultHook(func(err error) error { return err })

	// Give correlation package a valid input dump
	mockUploadStore.GetFunc.SetDefaultHook(copyTestDump)

	// Allowlist all files in dump
	gitserverClient.DirectoryChildrenFunc.SetDefaultReturn(map[string][]string{
		"": {"foo.go", "bar.go"},
	}, nil)

	expectedCommitDate := time.Unix(1587396557, 0).UTC()
	expectedCommitDateStr := expectedCommitDate.Format(time.RFC3339)
	gitserverClient.CommitDateFunc.SetDefaultReturn("deadbeef", expectedCommitDate, true, nil)

	handler := &handler{
		dbStore:         mockDBStore,
		repoStore:       mockRepoStore,
		workerStore:     mockWorkerStore,
		lsifStore:       mockLSIFStore,
		uploadStore:     mockUploadStore,
		gitserverClient: gitserverClient,
	}

	requeued, err := handler.handle(context.Background(), logtest.Scoped(t), upload, observation.TestTraceLogger(logtest.Scoped(t)))
	if err != nil {
		t.Fatalf("unexpected error handling upload: %s", err)
	} else if requeued {
		t.Errorf("unexpected requeue")
	}

	if calls := mockDBStore.UpdateCommittedAtFunc.History(); len(calls) != 1 {
		t.Errorf("unexpected number of UpdateCommitedAt calls. want=%d have=%d", 1, len(mockDBStore.UpdatePackagesFunc.History()))
	} else if calls[0].Arg1 != 50 {
		t.Errorf("unexpected UpdateCommitedAt repository id. want=%d have=%d", 50, calls[0].Arg1)
	} else if calls[0].Arg2 != "deadbeef" {
		t.Errorf("unexpected UpdateCommitedAt commit. want=%s have=%s", "deadbeef", calls[0].Arg2)
	} else if calls[0].Arg3 != expectedCommitDateStr {
		t.Errorf("unexpected UpdateCommitedAt commit date. want=%s have=%s", expectedCommitDate, calls[0].Arg3)
	}

	expectedPackagesDumpID := 42
	expectedPackages := []precise.Package{
		{
			Scheme:  "scheme B",
			Name:    "pkg B",
			Version: "v1.2.3",
		},
	}
	if len(mockDBStore.UpdatePackagesFunc.History()) != 1 {
		t.Errorf("unexpected number of UpdatePackages calls. want=%d have=%d", 1, len(mockDBStore.UpdatePackagesFunc.History()))
	} else if diff := cmp.Diff(expectedPackagesDumpID, mockDBStore.UpdatePackagesFunc.History()[0].Arg1); diff != "" {
		t.Errorf("unexpected UpdatePackagesFunc args (-want +got):\n%s", diff)
	} else if diff := cmp.Diff(expectedPackages, mockDBStore.UpdatePackagesFunc.History()[0].Arg2); diff != "" {
		t.Errorf("unexpected UpdatePackagesFunc args (-want +got):\n%s", diff)
	}

	expectedPackageReferencesDumpID := 42
	expectedPackageReferences := []precise.PackageReference{
		{
			Package: precise.Package{
				Scheme:  "scheme A",
				Name:    "pkg A",
				Version: "v0.1.0",
			},
		},
	}
	if len(mockDBStore.UpdatePackageReferencesFunc.History()) != 1 {
		t.Errorf("unexpected number of UpdatePackageReferences calls. want=%d have=%d", 1, len(mockDBStore.UpdatePackageReferencesFunc.History()))
	} else if diff := cmp.Diff(expectedPackageReferencesDumpID, mockDBStore.UpdatePackageReferencesFunc.History()[0].Arg1); diff != "" {
		t.Errorf("unexpected UpdatePackageReferencesFunc args (-want +got):\n%s", diff)
	} else if diff := cmp.Diff(expectedPackageReferences, mockDBStore.UpdatePackageReferencesFunc.History()[0].Arg2); diff != "" {
		t.Errorf("unexpected UpdatePackageReferencesFunc args (-want +got):\n%s", diff)
	}

	expectedIDsForRefcountUpdate := []int{42}
	if len(mockDBStore.UpdateUploadsReferenceCountsFunc.History()) != 1 {
		t.Errorf("unexpected number of UpdateReferenceCounts calls. want=%d have=%d", 1, len(mockDBStore.UpdateUploadsReferenceCountsFunc.History()))
	} else if diff := cmp.Diff(expectedIDsForRefcountUpdate, mockDBStore.UpdateUploadsReferenceCountsFunc.History()[0].Arg1); diff != "" {
		t.Errorf("unexpected UpdateReferenceCounts args (-want +got):\n%s", diff)
	} else if diff := cmp.Diff(shared.DependencyReferenceCountUpdateTypeAdd, mockDBStore.UpdateUploadsReferenceCountsFunc.History()[0].Arg2); diff != "" {
		t.Errorf("unexpected UpdateReferenceCounts args (-want +got):\n%s", diff)
	}

	if len(mockDBStore.InsertDependencySyncingJobFunc.History()) != 1 {
		t.Errorf("unexpected number of InsertDependencyIndexingJob calls. want=%d have=%d", 1, len(mockDBStore.InsertDependencySyncingJobFunc.History()))
	} else if mockDBStore.InsertDependencySyncingJobFunc.History()[0].Arg1 != 42 {
		t.Errorf("unexpected value for upload id. want=%d have=%d", 42, mockDBStore.InsertDependencySyncingJobFunc.History()[0].Arg1)
	}

	if len(mockDBStore.DeleteOverlappingDumpsFunc.History()) != 1 {
		t.Errorf("unexpected number of DeleteOverlappingDumps calls. want=%d have=%d", 1, len(mockDBStore.DeleteOverlappingDumpsFunc.History()))
	} else if mockDBStore.DeleteOverlappingDumpsFunc.History()[0].Arg1 != 50 {
		t.Errorf("unexpected value for repository id. want=%d have=%d", 50, mockDBStore.DeleteOverlappingDumpsFunc.History()[0].Arg1)
	} else if mockDBStore.DeleteOverlappingDumpsFunc.History()[0].Arg2 != "deadbeef" {
		t.Errorf("unexpected value for commit. want=%s have=%s", "deadbeef", mockDBStore.DeleteOverlappingDumpsFunc.History()[0].Arg2)
	} else if mockDBStore.DeleteOverlappingDumpsFunc.History()[0].Arg3 != "root/" {
		t.Errorf("unexpected value for root. want=%s have=%s", "root/", mockDBStore.DeleteOverlappingDumpsFunc.History()[0].Arg3)
	} else if mockDBStore.DeleteOverlappingDumpsFunc.History()[0].Arg4 != "lsif-go" {
		t.Errorf("unexpected value for indexer. want=%s have=%s", "lsif-go", mockDBStore.DeleteOverlappingDumpsFunc.History()[0].Arg4)
	}

	if len(mockDBStore.SetRepositoryAsDirtyFunc.History()) != 1 {
		t.Errorf("unexpected number of MarkRepositoryAsDirty calls. want=%d have=%d", 1, len(mockDBStore.SetRepositoryAsDirtyFunc.History()))
	} else if mockDBStore.SetRepositoryAsDirtyFunc.History()[0].Arg1 != 50 {
		t.Errorf("unexpected value for repository id. want=%d have=%d", 50, mockDBStore.SetRepositoryAsDirtyFunc.History()[0].Arg1)
	}

	if len(mockUploadStore.DeleteFunc.History()) != 1 {
		t.Errorf("unexpected number of Delete calls. want=%d have=%d", 1, len(mockUploadStore.DeleteFunc.History()))
	}
}

func TestHandleError(t *testing.T) {
	setupRepoMocks(t)

	upload := codeinteltypes.Upload{
		ID:           42,
		Root:         "root/",
		Commit:       "deadbeef",
		RepositoryID: 50,
		Indexer:      "lsif-go",
	}

	mockWorkerStore := NewMockWorkerStore()
	mockDBStore := NewMockStore()
	mockRepoStore := NewMockRepoStore()
	mockLSIFStore := NewMockLsifStore()
	mockUploadStore := uploadstoremocks.NewMockStore()
	gitserverClient := NewMockGitserverClient()

	// Set default transaction behavior
	mockDBStore.TransactFunc.SetDefaultReturn(mockDBStore, nil)
	mockDBStore.DoneFunc.SetDefaultHook(func(err error) error { return err })

	// Set default transaction behavior
	mockLSIFStore.TransactFunc.SetDefaultReturn(mockLSIFStore, nil)
	mockDBStore.DoneFunc.SetDefaultHook(func(err error) error { return err })

	// Give correlation package a valid input dump
	mockUploadStore.GetFunc.SetDefaultHook(copyTestDump)

	// Supply non-nil commit date
	gitserverClient.CommitDateFunc.SetDefaultReturn("deadbeef", time.Now(), true, nil)

	// Set a different tip commit
	mockDBStore.SetRepositoryAsDirtyFunc.SetDefaultReturn(errors.Errorf("uh-oh!"))

	handler := &handler{
		dbStore:         mockDBStore,
		repoStore:       mockRepoStore,
		workerStore:     mockWorkerStore,
		lsifStore:       mockLSIFStore,
		uploadStore:     mockUploadStore,
		gitserverClient: gitserverClient,
	}

	requeued, err := handler.handle(context.Background(), logtest.Scoped(t), upload, observation.TestTraceLogger(logtest.Scoped(t)))
	if err == nil {
		t.Fatalf("unexpected nil error handling upload")
	} else if !strings.Contains(err.Error(), "uh-oh!") {
		t.Fatalf("unexpected error: %s", err)
	} else if requeued {
		t.Errorf("unexpected requeue")
	}

	if len(mockDBStore.DoneFunc.History()) != 1 {
		t.Errorf("unexpected number of Done calls. want=%d have=%d", 1, len(mockDBStore.DoneFunc.History()))
	}

	if len(mockUploadStore.DeleteFunc.History()) != 0 {
		t.Errorf("unexpected number of Delete calls. want=%d have=%d", 0, len(mockUploadStore.DeleteFunc.History()))
	}
}

func TestHandleCloneInProgress(t *testing.T) {
	upload := codeinteltypes.Upload{
		ID:           42,
		Root:         "root/",
		Commit:       "deadbeef",
		RepositoryID: 50,
		Indexer:      "lsif-go",
	}

	mockWorkerStore := NewMockWorkerStore()
	mockDBStore := NewMockStore()
	mockRepoStore := NewMockRepoStore()
	mockUploadStore := uploadstoremocks.NewMockStore()
	gitserverClient := NewMockGitserverClient()

	mockRepoStore.GetFunc.SetDefaultHook(func(ctx context.Context, repoID api.RepoID) (*types.Repo, error) {
		if repoID != api.RepoID(50) {
			t.Errorf("unexpected repository name. want=%d have=%d", 50, repoID)
		}
		return &types.Repo{ID: repoID}, nil
	})
	mockRepoStore.ResolveRevFunc.SetDefaultHook(func(ctx context.Context, repo *types.Repo, _ string) (api.CommitID, error) {
		return "", &gitdomain.RepoNotExistError{Repo: repo.Name, CloneInProgress: true}
	})

	handler := &handler{
		dbStore:         mockDBStore,
		repoStore:       mockRepoStore,
		workerStore:     mockWorkerStore,
		uploadStore:     mockUploadStore,
		gitserverClient: gitserverClient,
	}

	requeued, err := handler.handle(context.Background(), logtest.Scoped(t), upload, observation.TestTraceLogger(logtest.Scoped(t)))
	if err != nil {
		t.Fatalf("unexpected error handling upload: %s", err)
	} else if !requeued {
		t.Errorf("expected upload to be requeued")
	}

	if len(mockWorkerStore.RequeueFunc.History()) != 1 {
		t.Errorf("unexpected number of Requeue calls. want=%d have=%d", 1, len(mockWorkerStore.RequeueFunc.History()))
	}
}

//
//

func copyTestDump(ctx context.Context, key string) (io.ReadCloser, error) {
	return os.Open("./testdata/dump1.lsif.gz")
}

func setupRepoMocks(t *testing.T) {
	t.Cleanup(func() {
		backend.Mocks.Repos.Get = nil
		backend.Mocks.Repos.ResolveRev = nil
	})

	backend.Mocks.Repos.Get = func(ctx context.Context, repoID api.RepoID) (*types.Repo, error) {
		if repoID != api.RepoID(50) {
			t.Errorf("unexpected repository name. want=%d have=%d", 50, repoID)
		}
		return &types.Repo{ID: repoID}, nil
	}

	backend.Mocks.Repos.ResolveRev = func(ctx context.Context, repo *types.Repo, rev string) (api.CommitID, error) {
		if rev != "deadbeef" {
			t.Errorf("unexpected commit. want=%s have=%s", "deadbeef", rev)
		}
		return "", nil
	}
}
