package api

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	bundlemocks "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/mocks"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	dbmocks "github.com/sourcegraph/sourcegraph/internal/codeintel/db/mocks"
)

var (
	testCommit             = makeCommit(0)
	testDump1              = db.Dump{ID: 42, Root: "sub1/"}
	testDump2              = db.Dump{ID: 50, Root: "sub2/"}
	testDump3              = db.Dump{ID: 51, Root: "sub3/"}
	testDump4              = db.Dump{ID: 52, Root: "sub4/"}
	testMoniker1           = bundles.MonikerData{Kind: "import", Scheme: "gomod", Identifier: "pad", PackageInformationID: "1234"}
	testMoniker2           = bundles.MonikerData{Kind: "export", Scheme: "gomod", Identifier: "pad", PackageInformationID: "1234"}
	testMoniker3           = bundles.MonikerData{Kind: "export", Scheme: "gomod", Identifier: "pad"}
	testPackageInformation = bundles.PackageInformationData{Name: "leftpad", Version: "0.1.0"}

	testRange1 = bundles.Range{
		Start: bundles.Position{Line: 10, Character: 50},
		End:   bundles.Position{Line: 10, Character: 55},
	}
	testRange2 = bundles.Range{
		Start: bundles.Position{Line: 11, Character: 50},
		End:   bundles.Position{Line: 11, Character: 55},
	}
	testRange3 = bundles.Range{
		Start: bundles.Position{Line: 12, Character: 50},
		End:   bundles.Position{Line: 12, Character: 55},
	}
	testRange4 = bundles.Range{
		Start: bundles.Position{Line: 13, Character: 50},
		End:   bundles.Position{Line: 13, Character: 55},
	}
	testRange5 = bundles.Range{
		Start: bundles.Position{Line: 14, Character: 50},
		End:   bundles.Position{Line: 14, Character: 55},
	}
)

// makeCommit formats an integer as a 40-character git commit hash.
func makeCommit(i int) string {
	return fmt.Sprintf("%040d", i)
}

func setMockDBGetDumpByID(t *testing.T, mockDB *dbmocks.MockDB, dumps map[int]db.Dump) {
	mockDB.GetDumpByIDFunc.SetDefaultHook(func(ctx context.Context, id int) (db.Dump, bool, error) {
		dump, ok := dumps[id]
		return dump, ok, nil
	})
}

func setMockDBGetPackage(t *testing.T, mockDB *dbmocks.MockDB, expectedScheme, expectedName, expectedVersion string, dump db.Dump, exists bool) {
	mockDB.GetPackageFunc.SetDefaultHook(func(ctx context.Context, scheme, name, version string) (db.Dump, bool, error) {
		if scheme != expectedScheme {
			t.Errorf("unexpected scheme for GetPackage. want=%s have=%s", expectedScheme, scheme)
		}
		if name != expectedName {
			t.Errorf("unexpected name for GetPackage. want=%s have=%s", expectedName, name)
		}
		if version != expectedVersion {
			t.Errorf("unexpected version for GetPackage. want=%s have=%s", expectedVersion, version)
		}
		return dump, exists, nil
	})
}

func setMockDBFindClosestDumps(t *testing.T, mockDB *dbmocks.MockDB, expectedRepositoryID int, expectedCommit, expectedFile string, dumps []db.Dump) {
	mockDB.FindClosestDumpsFunc.SetDefaultHook(func(ctx context.Context, repositoryID int, commit, file string) ([]db.Dump, error) {
		if repositoryID != expectedRepositoryID {
			t.Errorf("unexpected repository id for FindClosestDumps. want=%d have=%d", expectedRepositoryID, repositoryID)
		}
		if commit != expectedCommit {
			t.Errorf("unexpected commit for FindClosestDumps. want=%s have=%s", expectedCommit, commit)
		}
		if file != expectedFile {
			t.Errorf("unexpected file for FindClosestDumps. want=%s have=%s", expectedFile, file)
		}
		return dumps, nil
	})
}

func setMockDBSameRepoPager(t *testing.T, mockDB *dbmocks.MockDB, expectedRepositoryID int, expectedCommit, expectedScheme, expectedName, expectedVersion string, expectedLimit, totalCount int, pager db.ReferencePager) {
	mockDB.SameRepoPagerFunc.SetDefaultHook(func(ctx context.Context, repositoryID int, commit, scheme, name, version string, limit int) (int, db.ReferencePager, error) {
		if repositoryID != expectedRepositoryID {
			t.Errorf("unexpected repository id for SameRepoPager. want=%v have=%v", expectedRepositoryID, repositoryID)
		}
		if commit != expectedCommit {
			t.Errorf("unexpected commit for SameRepoPager. want=%s have=%s", expectedCommit, commit)
		}
		if scheme != expectedScheme {
			t.Errorf("unexpected scheme for SameRepoPager. want=%s have=%s", expectedScheme, scheme)
		}
		if name != expectedName {
			t.Errorf("unexpected name for SameRepoPager. want=%s have=%s", expectedName, name)
		}
		if version != expectedVersion {
			t.Errorf("unexpected version for SameRepoPager. want=%s have=%s", expectedVersion, version)
		}
		if limit != expectedLimit {
			t.Errorf("unexpected limit for SameRepoPager. want=%d have=%d", expectedLimit, limit)
		}
		return totalCount, pager, nil
	})
}

func setMockDBPackageReferencePager(t *testing.T, mockDB *dbmocks.MockDB, expectedScheme, expectedName, expectedVersion string, expectedRepositoryID, expectedLimit int, totalCount int, pager db.ReferencePager) {
	mockDB.PackageReferencePagerFunc.SetDefaultHook(func(ctx context.Context, scheme, name, version string, repositoryID, limit int) (int, db.ReferencePager, error) {
		if scheme != expectedScheme {
			t.Errorf("unexpected scheme for PackageReferencePager. want=%s have=%s", expectedScheme, scheme)
		}
		if name != expectedName {
			t.Errorf("unexpected name for PackageReferencePager. want=%s have=%s", expectedName, name)
		}
		if version != expectedVersion {
			t.Errorf("unexpected version for PackageReferencePager. want=%s have=%s", expectedVersion, version)
		}
		if repositoryID != expectedRepositoryID {
			t.Errorf("unexpected repository id for PackageReferencePager. want=%d have=%d", expectedRepositoryID, repositoryID)
		}
		if limit != expectedLimit {
			t.Errorf("unexpected limit for PackageReferencePager. want=%d have=%d", expectedLimit, limit)
		}
		return totalCount, pager, nil
	})
}

func setMockDBHasCommit(t *testing.T, mockDB *dbmocks.MockDB, expectedRepositoryID int, expectedCommit string, exists bool) {
	mockDB.HasCommitFunc.SetDefaultHook(func(ctx context.Context, repositoryID int, commit string) (bool, error) {
		if repositoryID != expectedRepositoryID {
			t.Errorf("unexpected repository id for HasCommit. want=%d have=%d", expectedRepositoryID, repositoryID)
		}
		if commit != expectedCommit {
			t.Errorf("unexpected commit for HasCommit. want=%s have=%s", expectedCommit, commit)
		}
		return exists, nil
	})
}

func setMockReferencePagerPageFromOffset(t *testing.T, mockReferencePager *dbmocks.MockReferencePager, expectedOffset int, references []types.PackageReference) {
	mockReferencePager.PageFromOffsetFunc.SetDefaultHook(func(ctx context.Context, offset int) ([]types.PackageReference, error) {
		if offset != expectedOffset {
			t.Errorf("unexpected offset for PageFromOffset. want=%d have=%d", expectedOffset, offset)
		}
		return references, nil
	})
}

func setMockBundleManagerClientBundleClient(t *testing.T, mockBundleManagerClient *bundlemocks.MockBundleManagerClient, bundleClients map[int]bundles.BundleClient) {
	mockBundleManagerClient.BundleClientFunc.SetDefaultHook(func(bundleID int) bundles.BundleClient {
		bundleClient, ok := bundleClients[bundleID]
		if !ok {
			t.Errorf("unexpected bundle id for BundleClient: %d", bundleID)
		}

		return bundleClient
	})
}

func setMockBundleClientExists(t *testing.T, mockBundleClient *bundlemocks.MockBundleClient, expectedPath string, exists bool) {
	mockBundleClient.ExistsFunc.SetDefaultHook(func(ctx context.Context, path string) (bool, error) {
		if path != expectedPath {
			t.Errorf("unexpected path for Exists. want=%s have=%s", expectedPath, path)
		}
		return exists, nil
	})
}

func setMockBundleClientDefinitions(t *testing.T, mockBundleClient *bundlemocks.MockBundleClient, expectedPath string, expectedLine, expectedCharacter int, locations []bundles.Location) {
	mockBundleClient.DefinitionsFunc.SetDefaultHook(func(ctx context.Context, path string, line, character int) ([]bundles.Location, error) {
		if path != expectedPath {
			t.Errorf("unexpected path for Definitions. want=%s have=%s", expectedPath, path)
		}
		if line != expectedLine {
			t.Errorf("unexpected line for Definitions. want=%d have=%d", expectedLine, line)
		}
		if character != expectedCharacter {
			t.Errorf("unexpected character for Definitions. want=%d have=%d", expectedCharacter, character)
		}
		return locations, nil
	})
}

func setMockBundleClientReferences(t *testing.T, mockBundleClient *bundlemocks.MockBundleClient, expectedPath string, expectedLine, expectedCharacter int, locations []bundles.Location) {
	mockBundleClient.ReferencesFunc.SetDefaultHook(func(ctx context.Context, path string, line, character int) ([]bundles.Location, error) {
		if path != expectedPath {
			t.Errorf("unexpected path for References. want=%s have=%s", expectedPath, path)
		}
		if line != expectedLine {
			t.Errorf("unexpected line for References. want=%d have=%d", expectedLine, line)
		}
		if character != expectedCharacter {
			t.Errorf("unexpected character for References. want=%d have=%d", expectedCharacter, character)
		}
		return locations, nil
	})
}

func setMockBundleClientHover(t *testing.T, mockBundleClient *bundlemocks.MockBundleClient, expectedPath string, expectedLine, expectedCharacter int, text string, r bundles.Range, exists bool) {
	mockBundleClient.HoverFunc.SetDefaultHook(func(ctx context.Context, path string, line, character int) (string, bundles.Range, bool, error) {
		if path != expectedPath {
			t.Errorf("unexpected path Hover. want=%s have=%s", expectedPath, path)
		}
		if line != expectedLine {
			t.Errorf("unexpected line Hover. want=%d have=%d", expectedLine, line)
		}
		if character != expectedCharacter {
			t.Errorf("unexpected character Hover. want=%d have=%d", expectedCharacter, character)
		}
		return text, r, exists, nil
	})
}

func setMockBundleClientMonikersByPosition(t *testing.T, mockBundleClient *bundlemocks.MockBundleClient, expectedPath string, expectedLine, expectedCharacter int, monikers [][]bundles.MonikerData) {
	mockBundleClient.MonikersByPositionFunc.SetDefaultHook(func(ctx context.Context, path string, line, character int) ([][]bundles.MonikerData, error) {
		if path != expectedPath {
			t.Fatalf("unexpected path for MonikersByPosition. want=%s have=%s", expectedPath, path)
		}
		if line != expectedLine {
			t.Fatalf("unexpected line for MonikersByPosition. want=%v have=%v", expectedLine, line)
		}
		if character != expectedCharacter {
			t.Fatalf("unexpected character for MonikersByPosition. want=%v have=%v", expectedCharacter, character)
		}

		return monikers, nil
	})
}

func setMockBundleClientMonikerResults(t *testing.T, mockBundleClient *bundlemocks.MockBundleClient, expectedModelType, expectedScheme, expectedIdentifier string, expectedSkip, expectedTake int, locations []bundles.Location, totalCount int) {
	mockBundleClient.MonikerResultsFunc.SetDefaultHook(func(ctx context.Context, modelType, scheme, identifier string, skip, take int) ([]bundles.Location, int, error) {
		if modelType != expectedModelType {
			t.Errorf("unexpected model type for MonikerResults. want=%s have=%s", expectedModelType, modelType)
		}
		if scheme != expectedScheme {
			t.Errorf("unexpected scheme for MonikerResults. want=%s have=%s", expectedScheme, scheme)
		}
		if identifier != expectedIdentifier {
			t.Errorf("unexpected identifier for MonikerResults. want=%s have=%s", expectedIdentifier, identifier)
		}
		if skip != expectedSkip {
			t.Errorf("unexpected skip for MonikerResults. want=%d have=%d", expectedSkip, skip)
		}
		if take != expectedTake {
			t.Errorf("unexpected take for MonikerResults. want=%d have=%d", expectedTake, take)
		}
		return locations, totalCount, nil
	})
}

func setMockBundleClientPackageInformation(t *testing.T, mockBundleClient *bundlemocks.MockBundleClient, expectedPath, expectedPackageInformationID string, packageInformation bundles.PackageInformationData) {
	mockBundleClient.PackageInformationFunc.SetDefaultHook(func(ctx context.Context, path, packageInformationID string) (bundles.PackageInformationData, error) {
		if path != expectedPath {
			t.Errorf("unexpected path for PackageInformation. want=%s have=%s", expectedPath, path)
		}
		if packageInformationID != expectedPackageInformationID {
			t.Errorf("unexpected package information id for PackageInformation. want=%s have=%s", expectedPackageInformationID, packageInformationID)
		}
		return packageInformation, nil
	})
}

func readTestFilter(t *testing.T, dirname, filename string) []byte {
	content, err := ioutil.ReadFile(fmt.Sprintf("../../testdata/filters/%s/%s", dirname, filename))
	if err != nil {
		t.Fatalf("unexpected error reading: %s", err)
	}

	raw, err := hex.DecodeString(strings.TrimSpace(string(content)))
	if err != nil {
		t.Fatalf("unexpected error decoding: %s", err)
	}

	return raw
}
