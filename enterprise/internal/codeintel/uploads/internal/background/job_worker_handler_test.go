package background

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	codeinteltypes "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/api"
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
		ContentType:  "application/x-ndjson+lsif",
	}

	mockWorkerStore := NewMockWorkerStore[codeinteltypes.Upload]()
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

	svc := &handler{
		store:           mockDBStore,
		lsifStore:       mockLSIFStore,
		gitserverClient: gitserverClient,
		repoStore:       mockRepoStore,
		workerStore:     mockWorkerStore,
	}

	requeued, err := svc.HandleRawUpload(context.Background(), logtest.Scoped(t), upload, mockUploadStore, observation.TestTraceLogger(logtest.Scoped(t)))
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

func TestHandleSCIP(t *testing.T) {
	setupRepoMocks(t)

	upload := codeinteltypes.Upload{
		ID:           42,
		Root:         "",
		Commit:       "deadbeef",
		RepositoryID: 50,
		Indexer:      "lsif-go",
		ContentType:  "application/x-protobuf+scip",
	}

	mockWorkerStore := NewMockWorkerStore[codeinteltypes.Upload]()
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

	// Track writes to symbols table
	scipWriter := NewMockSCIPWriter()
	mockLSIFStore.NewSCIPWriterFunc.SetDefaultReturn(scipWriter, nil)

	scipWriter.InsertDocumentFunc.SetDefaultHook(func(_ context.Context, _ string, _ *scip.Document) error {
		return nil
	})

	// Give correlation package a valid input dump
	mockUploadStore.GetFunc.SetDefaultHook(copyTestDumpScip)

	// Allowlist all files in dump
	gitserverClient.DirectoryChildrenFunc.SetDefaultReturn(scipDirectoryChildren, nil)

	expectedCommitDate := time.Unix(1587396557, 0).UTC()
	expectedCommitDateStr := expectedCommitDate.Format(time.RFC3339)
	gitserverClient.CommitDateFunc.SetDefaultReturn("deadbeef", expectedCommitDate, true, nil)

	svc := &handler{
		store:           mockDBStore,
		lsifStore:       mockLSIFStore,
		gitserverClient: gitserverClient,
		repoStore:       mockRepoStore,
		workerStore:     mockWorkerStore,
	}

	requeued, err := svc.HandleRawUpload(context.Background(), logtest.Scoped(t), upload, mockUploadStore, observation.TestTraceLogger(logtest.Scoped(t)))
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
			Scheme:  "scip-typescript",
			Manager: "npm",
			Name:    "template",
			Version: "0.0.0-DEVELOPMENT",
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
		{Package: precise.Package{
			Scheme:  "scip-typescript",
			Manager: "npm",
			Name:    "typescript",
			Version: "4.9.3",
		}},
		{Package: precise.Package{
			Scheme:  "scip-typescript",
			Manager: "npm",
			Name:    "sourcegraph",
			Version: "25.5.0",
		}},
		{Package: precise.Package{
			Scheme:  "scip-typescript",
			Manager: "npm",
			Name:    "js-base64",
			Version: "3.7.1",
		}},
		{Package: precise.Package{
			Scheme:  "scip-typescript",
			Manager: "npm",
			Name:    "tagged-template-noop",
			Version: "2.1.01",
		}},
		{Package: precise.Package{
			Scheme:  "scip-typescript",
			Manager: "npm",
			Name:    "@types/mocha",
			Version: "9.0.0",
		}},
		{Package: precise.Package{
			Scheme:  "scip-typescript",
			Manager: "npm",
			Name:    "@types/node",
			Version: "14.17.15",
		}},
		{Package: precise.Package{
			Scheme:  "scip-typescript",
			Manager: "npm",
			Name:    "@types/lodash",
			Version: "4.14.178",
		}},
		{Package: precise.Package{
			Scheme:  "scip-typescript",
			Manager: "npm",
			Name:    "rxjs",
			Version: "6.6.7",
		}},
	}
	if len(mockDBStore.UpdatePackageReferencesFunc.History()) != 1 {
		t.Errorf("unexpected number of UpdatePackageReferences calls. want=%d have=%d", 1, len(mockDBStore.UpdatePackageReferencesFunc.History()))
	} else if diff := cmp.Diff(expectedPackageReferencesDumpID, mockDBStore.UpdatePackageReferencesFunc.History()[0].Arg1); diff != "" {
		t.Errorf("unexpected UpdatePackageReferencesFunc args (-want +got):\n%s", diff)
	} else {
		sort.Slice(expectedPackageReferences, func(i, j int) bool {
			return expectedPackageReferences[i].Name < expectedPackageReferences[j].Name
		})

		if diff := cmp.Diff(expectedPackageReferences, mockDBStore.UpdatePackageReferencesFunc.History()[0].Arg2); diff != "" {
			t.Errorf("unexpected UpdatePackageReferencesFunc args (-want +got):\n%s", diff)
		}
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
	} else if mockDBStore.DeleteOverlappingDumpsFunc.History()[0].Arg3 != "" {
		t.Errorf("unexpected value for root. want=%s have=%s", "", mockDBStore.DeleteOverlappingDumpsFunc.History()[0].Arg3)
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

	if len(mockLSIFStore.InsertMetadataFunc.History()) != 1 {
		t.Errorf("unexpected number of of InsertMetadataFunc.History() calls. want=%d have=%d", 1, len(mockLSIFStore.InsertMetadataFunc.History()))
	} else {
		call := mockLSIFStore.InsertMetadataFunc.History()[0]
		if call.Arg1 != 42 {
			t.Fatalf("unexpected value for upload id. want=%d have=%d", 42, call.Arg1)
		}

		expectedMetadata := lsifstore.ProcessedMetadata{
			TextDocumentEncoding: "UTF8",
			ToolName:             "scip-typescript",
			ToolVersion:          "0.3.3",
			ToolArguments:        nil,
			ProtocolVersion:      0,
		}
		if diff := cmp.Diff(expectedMetadata, call.Arg2); diff != "" {
			t.Errorf("unexpected processed metadata args (-want +got):\n%s", diff)
		}
	}
	if len(scipWriter.InsertDocumentFunc.History()) != 11 {
		t.Errorf("unexpected number of of InsertDocumentFunc.History() calls. want=%d have=%d", 11, len(scipWriter.InsertDocumentFunc.History()))
	} else {
		foundDocument1 := false
		foundDocument2 := false

		for _, call := range scipWriter.InsertDocumentFunc.History() {
			switch call.Arg1 {
			case "template/src/util/promise.ts":
				payload, _ := proto.Marshal(call.Arg2)
				hash := sha256.New()
				_, _ = hash.Write(payload)

				foundDocument1 = true
				expectedHash := "TTQ+xW2zU2O1b+MEGtkYLhjB3dbHRpHM3CXoS6pqqvI="
				if diff := cmp.Diff(expectedHash, base64.StdEncoding.EncodeToString(hash.Sum(nil))); diff != "" {
					t.Errorf("unexpected hash (-want +got):\n%s", diff)
				}

			case "template/src/util/graphql.ts":
				foundDocument2 = true
				if diff := cmp.Diff(testedInvertedRangeIndex, codeinteltypes.ExtractSymbolIndexes(call.Arg2)); diff != "" {
					t.Errorf("unexpected inverted range index (-want +got):\n%s", diff)
				}
			}
		}
		if !foundDocument1 {
			t.Fatalf("target path #1 not found")
		}
		if !foundDocument2 {
			t.Fatalf("target path #2 not found")
		}
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
		ContentType:  "application/x-ndjson+lsif",
	}

	mockWorkerStore := NewMockWorkerStore[codeinteltypes.Upload]()
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

	svc := &handler{
		store:     mockDBStore,
		lsifStore: mockLSIFStore,
		// lsifstore:       mockLSIFStore,
		gitserverClient: gitserverClient,
		repoStore:       mockRepoStore,
		workerStore:     mockWorkerStore,
	}

	requeued, err := svc.HandleRawUpload(context.Background(), logtest.Scoped(t), upload, mockUploadStore, observation.TestTraceLogger(logtest.Scoped(t)))
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

func TestHandleErrorScip(t *testing.T) {
	setupRepoMocks(t)

	upload := codeinteltypes.Upload{
		ID:           42,
		Root:         "root/",
		Commit:       "deadbeef",
		RepositoryID: 50,
		Indexer:      "lsif-go",
		ContentType:  "application/x-protobuf+scip",
	}

	mockWorkerStore := NewMockWorkerStore[codeinteltypes.Upload]()
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

	// Track writes to symbols table
	scipWriter := NewMockSCIPWriter()
	mockLSIFStore.NewSCIPWriterFunc.SetDefaultReturn(scipWriter, nil)

	// Give correlation package a valid input dump
	mockUploadStore.GetFunc.SetDefaultHook(copyTestDumpScip)

	// Supply non-nil commit date
	gitserverClient.CommitDateFunc.SetDefaultReturn("deadbeef", time.Now(), true, nil)

	// Set a different tip commit
	mockDBStore.SetRepositoryAsDirtyFunc.SetDefaultReturn(errors.Errorf("uh-oh!"))

	svc := &handler{
		store:     mockDBStore,
		lsifStore: mockLSIFStore,
		// lsifstore:       mockLSIFStore,
		gitserverClient: gitserverClient,
		repoStore:       mockRepoStore,
		workerStore:     mockWorkerStore,
	}

	requeued, err := svc.HandleRawUpload(context.Background(), logtest.Scoped(t), upload, mockUploadStore, observation.TestTraceLogger(logtest.Scoped(t)))
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
		ContentType:  "application/x-ndjson+lsif",
	}

	mockWorkerStore := NewMockWorkerStore[codeinteltypes.Upload]()
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

	svc := &handler{
		store:           mockDBStore,
		gitserverClient: gitserverClient,
		repoStore:       mockRepoStore,
		workerStore:     mockWorkerStore,
	}

	requeued, err := svc.HandleRawUpload(context.Background(), logtest.Scoped(t), upload, mockUploadStore, observation.TestTraceLogger(logtest.Scoped(t)))
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

func copyTestDumpScip(ctx context.Context, key string) (io.ReadCloser, error) {
	return os.Open("./testdata/index1.scip.gz")
}

var scipDirectoryChildren = map[string][]string{
	"": {
		"template",
	},
	"template": {
		"template/src",
	},
	"template/src": {
		"template/src/extension.ts",
		"template/src/indicators.ts",
		"template/src/language.ts",
		"template/src/logging.ts",
		"template/src/util",
	},
	"template/src/util": {
		"template/src/util/api.ts",
		"template/src/util/graphql.ts",
		"template/src/util/ix.test.ts",
		"template/src/util/ix.ts",
		"template/src/util/promise.ts",
		"template/src/util/uri.test.ts",
		"template/src/util/uri.ts",
	},
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
