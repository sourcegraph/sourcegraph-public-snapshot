package worker

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bloomfilter"
	bundlemocks "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client/mocks"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	dbmocks "github.com/sourcegraph/sourcegraph/internal/codeintel/db/mocks"
	gitservermocks "github.com/sourcegraph/sourcegraph/internal/codeintel/gitserver/mocks"
	"github.com/sourcegraph/sourcegraph/internal/sqliteutil"
)

func init() {
	sqliteutil.SetLocalLibpath()
	sqliteutil.MustRegisterSqlite3WithPcre()
}

func TestProcess(t *testing.T) {
	upload := db.Upload{
		ID:           42,
		Root:         "root/",
		Commit:       makeCommit(1),
		RepositoryID: 50,
		Indexer:      "lsif-go",
	}

	mockDB := dbmocks.NewMockDB()
	bundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	gitserverClient := gitservermocks.NewMockClient()

	// Give correlation package a valid input dump
	bundleManagerClient.GetUploadFunc.SetDefaultHook(copyTestDump)

	// Whitelist all files in dump
	gitserverClient.DirectoryChildrenFunc.SetDefaultReturn(map[string][]string{
		"": {"foo.go", "bar.go"},
	}, nil)

	// Set a different tip commit
	gitserverClient.HeadFunc.SetDefaultReturn(makeCommit(30), nil)

	// Return some ancestors for each commit args
	gitserverClient.CommitsNearFunc.SetDefaultHook(func(ctx context.Context, db db.DB, repositoryID int, commit string) (map[string][]string, error) {
		offset, err := strconv.ParseInt(commit, 10, 64)
		if err != nil {
			return nil, err
		}

		commits := map[string][]string{}
		for i := 0; i < 10; i++ {
			commits[makeCommit(int(offset)+i)] = []string{makeCommit(int(offset) + i + 1)}
		}

		return commits, nil
	})

	processor := &processor{
		bundleManagerClient: bundleManagerClient,
		gitserverClient:     gitserverClient,
	}

	err := processor.Process(context.Background(), mockDB, upload)
	if err != nil {
		t.Fatalf("unexpected error processing upload: %s", err)
	}

	expectedPackages := []types.Package{
		{DumpID: 42,
			Scheme:  "scheme B",
			Name:    "pkg B",
			Version: "v1.2.3",
		},
	}
	if len(mockDB.UpdatePackagesFunc.History()) != 1 {
		t.Errorf("unexpected number of UpdatePackages calls. want=%d have=%d", 1, len(mockDB.UpdatePackagesFunc.History()))
	} else if diff := cmp.Diff(expectedPackages, mockDB.UpdatePackagesFunc.History()[0].Arg1); diff != "" {
		t.Errorf("unexpected UpdatePackagesFuncargs (-want +got):\n%s", diff)
	}

	filter, err := bloomfilter.CreateFilter([]string{"ident A"})
	if err != nil {
		t.Fatalf("unexpected error creating filter: %s", err)
	}
	expectedPackageReferences := []types.PackageReference{
		{DumpID: 42,
			Scheme:  "scheme A",
			Name:    "pkg A",
			Version: "v0.1.0",
			Filter:  filter,
		},
	}
	if len(mockDB.UpdatePackageReferencesFunc.History()) != 1 {
		t.Errorf("unexpected number of UpdatePackageReferences calls. want=%d have=%d", 1, len(mockDB.UpdatePackageReferencesFunc.History()))
	} else if diff := cmp.Diff(expectedPackageReferences, mockDB.UpdatePackageReferencesFunc.History()[0].Arg1); diff != "" {
		t.Errorf("unexpected UpdatePackageReferencesFunc args (-want +got):\n%s", diff)
	}

	if len(mockDB.DeleteOverlappingDumpsFunc.History()) != 1 {
		t.Errorf("unexpected number of DeleteOverlappingDumps calls. want=%d have=%d", 1, len(mockDB.DeleteOverlappingDumpsFunc.History()))
	} else if mockDB.DeleteOverlappingDumpsFunc.History()[0].Arg1 != 50 {
		t.Errorf("unexpected value for repository id. want=%d have=%d", 50, mockDB.DeleteOverlappingDumpsFunc.History()[0].Arg1)
	} else if mockDB.DeleteOverlappingDumpsFunc.History()[0].Arg2 != makeCommit(1) {
		t.Errorf("unexpected value for commit. want=%s have=%s", makeCommit(1), mockDB.DeleteOverlappingDumpsFunc.History()[0].Arg2)
	} else if mockDB.DeleteOverlappingDumpsFunc.History()[0].Arg3 != "root/" {
		t.Errorf("unexpected value for root. want=%s have=%s", "root/", mockDB.DeleteOverlappingDumpsFunc.History()[0].Arg3)
	} else if mockDB.DeleteOverlappingDumpsFunc.History()[0].Arg4 != "lsif-go" {
		t.Errorf("unexpected value for indexer. want=%s have=%s", "lsif-go", mockDB.DeleteOverlappingDumpsFunc.History()[0].Arg4)
	}

	offsets := []int{1, 30}
	expectedCommits := map[string][]string{}
	for i := 0; i < 10; i++ {
		for _, offset := range offsets {
			expectedCommits[makeCommit(offset+i)] = []string{makeCommit(offset + i + 1)}
		}
	}
	if len(mockDB.UpdateCommitsFunc.History()) != 1 {
		t.Errorf("unexpected number of update UpdateCommits calls. want=%d have=%d", 1, len(mockDB.UpdateCommitsFunc.History()))
	} else if diff := cmp.Diff(expectedCommits, mockDB.UpdateCommitsFunc.History()[0].Arg2); diff != "" {
		t.Errorf("unexpected update UpdateCommitsFunc args (-want +got):\n%s", diff)
	}

	if len(mockDB.UpdateDumpsVisibleFromTipFunc.History()) != 1 {
		t.Errorf("unexpected number of UpdateDumpsVisibleFromTip calls. want=%d have=%d", 1, len(mockDB.UpdateDumpsVisibleFromTipFunc.History()))
	} else if mockDB.UpdateDumpsVisibleFromTipFunc.History()[0].Arg1 != 50 {
		t.Errorf("unexpected value for repository id. want=%d have=%d", 50, mockDB.UpdateDumpsVisibleFromTipFunc.History()[0].Arg1)
	} else if mockDB.UpdateDumpsVisibleFromTipFunc.History()[0].Arg2 != makeCommit(30) {
		t.Errorf("unexpected value for tip commit. want=%s have=%s", makeCommit(30), mockDB.UpdateDumpsVisibleFromTipFunc.History()[0].Arg2)
	}

	if len(bundleManagerClient.SendDBFunc.History()) != 1 {
		t.Errorf("unexpected number of SendDB calls. want=%d have=%d", 1, len(bundleManagerClient.SendDBFunc.History()))
	} else if bundleManagerClient.SendDBFunc.History()[0].Arg1 != 42 {
		t.Errorf("unexpected SendDBFunc args. want=%d have=%d", 42, bundleManagerClient.SendDBFunc.History()[0].Arg1)
	}
}

func TestProcessError(t *testing.T) {
	upload := db.Upload{
		ID:           42,
		Root:         "root/",
		Commit:       makeCommit(1),
		RepositoryID: 50,
		Indexer:      "lsif-go",
	}

	mockDB := dbmocks.NewMockDB()
	bundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	gitserverClient := gitservermocks.NewMockClient()

	// Give correlation package a valid input dump
	bundleManagerClient.GetUploadFunc.SetDefaultHook(copyTestDump)

	// Set a different tip commit
	gitserverClient.HeadFunc.SetDefaultReturn("", fmt.Errorf("uh-oh!"))

	processor := &processor{
		bundleManagerClient: bundleManagerClient,
		gitserverClient:     gitserverClient,
	}

	err := processor.Process(context.Background(), mockDB, upload)
	if err == nil {
		t.Fatalf("unexpected nil error processing upload")
	} else if !strings.Contains(err.Error(), "uh-oh!") {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockDB.RollbackToSavepointFunc.History()) != 1 {
		t.Errorf("unexpected number of RollbackToLastSavepoint calls. want=%d have=%d", 1, len(mockDB.RollbackToSavepointFunc.History()))
	}

	if len(bundleManagerClient.DeleteUploadFunc.History()) != 1 {
		t.Errorf("unexpected number of DeleteUpload calls. want=%d have=%d", 1, len(mockDB.RollbackToSavepointFunc.History()))
	}

}

//
//

func makeCommit(i int) string {
	return fmt.Sprintf("%040d", i)
}

func copyTestDump(ctx context.Context, uploadID int, dir string) (string, error) {
	src, err := os.Open("../../testdata/dump1.lsif")
	if err != nil {
		return "", err
	}
	defer src.Close()

	filename := filepath.Join(dir, "dump.lsif")
	dst, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	gzipWriter := gzip.NewWriter(dst)
	defer gzipWriter.Close()

	if _, err := io.Copy(gzipWriter, src); err != nil {
		return "", err
	}

	return filename, nil
}
