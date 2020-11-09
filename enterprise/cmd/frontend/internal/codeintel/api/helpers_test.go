package api

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
)

var (
	testCommit             = makeCommit(0)
	testDump1              = store.Dump{ID: 42, Root: "sub1/"}
	testDump2              = store.Dump{ID: 50, Root: "sub2/"}
	testDump3              = store.Dump{ID: 51, Root: "sub3/"}
	testDump4              = store.Dump{ID: 52, Root: "sub4/"}
	testMoniker1           = lsifstore.MonikerData{Kind: "import", Scheme: "gomod", Identifier: "pad", PackageInformationID: "1234"}
	testMoniker2           = lsifstore.MonikerData{Kind: "export", Scheme: "gomod", Identifier: "pad", PackageInformationID: "1234"}
	testMoniker3           = lsifstore.MonikerData{Kind: "export", Scheme: "gomod", Identifier: "pad"}
	testPackageInformation = lsifstore.PackageInformationData{Name: "leftpad", Version: "0.1.0"}

	testRange1 = lsifstore.Range{
		Start: lsifstore.Position{Line: 10, Character: 50},
		End:   lsifstore.Position{Line: 10, Character: 55},
	}
	testRange2 = lsifstore.Range{
		Start: lsifstore.Position{Line: 11, Character: 50},
		End:   lsifstore.Position{Line: 11, Character: 55},
	}
	testRange3 = lsifstore.Range{
		Start: lsifstore.Position{Line: 12, Character: 50},
		End:   lsifstore.Position{Line: 12, Character: 55},
	}
	testRange4 = lsifstore.Range{
		Start: lsifstore.Position{Line: 13, Character: 50},
		End:   lsifstore.Position{Line: 13, Character: 55},
	}
	testRange5 = lsifstore.Range{
		Start: lsifstore.Position{Line: 14, Character: 50},
		End:   lsifstore.Position{Line: 14, Character: 55},
	}
)

// makeCommit formats an integer as a 40-character git commit hash.
func makeCommit(i int) string {
	return fmt.Sprintf("%040d", i)
}

func setMockDBStoreGetDumpByID(t *testing.T, mockDBStore *MockDBStore, dumps map[int]store.Dump) {
	mockDBStore.GetDumpByIDFunc.SetDefaultHook(func(ctx context.Context, id int) (store.Dump, bool, error) {
		dump, ok := dumps[id]
		return dump, ok, nil
	})
}

func setMockDBStoreGetPackage(t *testing.T, mockDBStore *MockDBStore, expectedScheme, expectedName, expectedVersion string, dump store.Dump, exists bool) {
	mockDBStore.GetPackageFunc.SetDefaultHook(func(ctx context.Context, scheme, name, version string) (store.Dump, bool, error) {
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

func setMockDBStoreFindClosestDumps(t *testing.T, mockDBStore *MockDBStore, expectedRepositoryID int, expectedCommit, expectedFile string, expectedrootMustEnclosePath bool, expectedIndexer string, dumps []store.Dump) {
	mockDBStore.FindClosestDumpsFunc.SetDefaultHook(func(ctx context.Context, repositoryID int, commit, file string, rootMustEnclosePath bool, indexer string) ([]store.Dump, error) {
		if repositoryID != expectedRepositoryID {
			t.Errorf("unexpected repository id for FindClosestDumps. want=%d have=%d", expectedRepositoryID, repositoryID)
		}
		if commit != expectedCommit {
			t.Errorf("unexpected commit for FindClosestDumps. want=%s have=%s", expectedCommit, commit)
		}
		if file != expectedFile {
			t.Errorf("unexpected file for FindClosestDumps. want=%s have=%s", expectedFile, file)
		}
		if rootMustEnclosePath != expectedrootMustEnclosePath {
			t.Errorf("unexpected rootMustEnclosePath for FindClosestDumps. want=%v have=%v", expectedrootMustEnclosePath, rootMustEnclosePath)
		}
		if indexer != expectedIndexer {
			t.Errorf("unexpected indexer for FindClosestDumps. want=%s have=%s", expectedIndexer, indexer)
		}
		return dumps, nil
	})
}

func setMockDBStoreFindClosestDumpsFromGraphFragment(t *testing.T, mockDBStore *MockDBStore, expectedRepositoryID int, expectedCommit, expectedFile string, expectedrootMustEnclosePath bool, expectedIndexer string, expectedGraph map[string][]string, dumps []store.Dump) {
	mockDBStore.FindClosestDumpsFromGraphFragmentFunc.SetDefaultHook(func(ctx context.Context, repositoryID int, commit, file string, rootMustEnclosePath bool, indexer string, graph map[string][]string) ([]store.Dump, error) {
		if repositoryID != expectedRepositoryID {
			t.Errorf("unexpected repository id for FindClosestDumps. want=%d have=%d", expectedRepositoryID, repositoryID)
		}
		if commit != expectedCommit {
			t.Errorf("unexpected commit for FindClosestDumps. want=%s have=%s", expectedCommit, commit)
		}
		if file != expectedFile {
			t.Errorf("unexpected file for FindClosestDumps. want=%s have=%s", expectedFile, file)
		}
		if rootMustEnclosePath != expectedrootMustEnclosePath {
			t.Errorf("unexpected rootMustEnclosePath for FindClosestDumps. want=%v have=%v", expectedrootMustEnclosePath, rootMustEnclosePath)
		}
		if indexer != expectedIndexer {
			t.Errorf("unexpected indexer for FindClosestDumps. want=%s have=%s", expectedIndexer, indexer)
		}
		if diff := cmp.Diff(expectedGraph, graph); diff != "" {
			t.Errorf("unexpected graph (-want +got):\n%s", diff)
		}
		return dumps, nil
	})
}

func setMockDBStoreSameRepoPager(t *testing.T, mockDBStore *MockDBStore, expectedRepositoryID int, expectedCommit, expectedScheme, expectedName, expectedVersion string, expectedLimit, totalCount int, pager ReferencePager) {
	mockDBStore.SameRepoPagerFunc.SetDefaultHook(func(ctx context.Context, repositoryID int, commit, scheme, name, version string, limit int) (int, ReferencePager, error) {
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

func setMockDBStorePackageReferencePager(t *testing.T, mockDBStore *MockDBStore, expectedScheme, expectedName, expectedVersion string, expectedRepositoryID, expectedLimit int, totalCount int, pager ReferencePager) {
	mockDBStore.PackageReferencePagerFunc.SetDefaultHook(func(ctx context.Context, scheme, name, version string, repositoryID, limit int) (int, ReferencePager, error) {
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

func setMockDBStoreHasRepository(t *testing.T, mockDBStore *MockDBStore, expectedRepositoryID int, exists bool) {
	mockDBStore.HasRepositoryFunc.SetDefaultHook(func(ctx context.Context, repositoryID int) (bool, error) {
		if repositoryID != expectedRepositoryID {
			t.Errorf("unexpected repository id for HasRepository. want=%d have=%d", expectedRepositoryID, repositoryID)
		}
		return exists, nil
	})
}

func setMockDBStoreHasCommit(t *testing.T, mockDBStore *MockDBStore, expectedRepositoryID int, expectedCommit string, exists bool) {
	mockDBStore.HasCommitFunc.SetDefaultHook(func(ctx context.Context, repositoryID int, commit string) (bool, error) {
		if repositoryID != expectedRepositoryID {
			t.Errorf("unexpected repository id for HasCommit. want=%d have=%d", expectedRepositoryID, repositoryID)
		}
		if commit != expectedCommit {
			t.Errorf("unexpected commit for HasCommit. want=%s have=%s", expectedCommit, commit)
		}
		return exists, nil
	})
}

func setMockReferencePagerPageFromOffset(t *testing.T, mockReferencePager *MockReferencePager, expectedOffset int, references []lsifstore.PackageReference) {
	mockReferencePager.PageFromOffsetFunc.SetDefaultHook(func(ctx context.Context, offset int) ([]lsifstore.PackageReference, error) {
		if offset != expectedOffset {
			t.Errorf("unexpected offset for PageFromOffset. want=%d have=%d", expectedOffset, offset)
		}
		return references, nil
	})
}

func setmockLSIFStoreExists(t *testing.T, mockLSIFStore *MockLSIFStore, expectedDumpID int, expectedPath string, exists bool) {
	mockLSIFStore.ExistsFunc.SetDefaultHook(func(ctx context.Context, dumpID int, path string) (bool, error) {
		if dumpID != expectedDumpID {
			t.Errorf("unexpected id for Exists. want=%d have=%d", expectedDumpID, dumpID)
		}
		if path != expectedPath {
			t.Errorf("unexpected path for Exists. want=%s have=%s", expectedPath, path)
		}
		return exists, nil
	})
}

type existsSpec struct {
	dumpID int
	path   string
	exists bool
}

func setMultimockLSIFStoreExists(t *testing.T, mockLSIFStore *MockLSIFStore, specs ...existsSpec) {
	mockLSIFStore.ExistsFunc.SetDefaultHook(func(ctx context.Context, dumpID int, path string) (bool, error) {
		for _, spec := range specs {
			if dumpID == spec.dumpID && path == spec.path {
				return spec.exists, nil
			}
		}

		t.Errorf("unexpected args for Exists. dumpID=%d path=%s", dumpID, path)
		return false, nil
	})
}

func setmockLSIFStoreRanges(t *testing.T, mockLSIFStore *MockLSIFStore, expectedDumpID int, expectedPath string, expectedStartLine, expectedEndLine int, ranges []lsifstore.CodeIntelligenceRange) {
	mockLSIFStore.RangesFunc.SetDefaultHook(func(ctx context.Context, dumpID int, path string, startLine, endLine int) ([]lsifstore.CodeIntelligenceRange, error) {
		if dumpID != expectedDumpID {
			t.Errorf("unexpected id for Ranges. want=%d have=%d", expectedDumpID, dumpID)
		}
		if path != expectedPath {
			t.Errorf("unexpected path for Ranges. want=%s have=%s", expectedPath, path)
		}
		if startLine != expectedStartLine {
			t.Errorf("unexpected start line for Ranges. want=%d have=%d", expectedStartLine, startLine)
		}
		if endLine != expectedEndLine {
			t.Errorf("unexpected end line for Ranges. want=%d have=%d", expectedEndLine, endLine)
		}
		return ranges, nil
	})
}

func setmockLSIFStoreDefinitions(t *testing.T, mockLSIFStore *MockLSIFStore, expectedDumpID int, expectedPath string, expectedLine, expectedCharacter int, locations []lsifstore.Location) {
	mockLSIFStore.DefinitionsFunc.SetDefaultHook(func(ctx context.Context, dumpID int, path string, line, character int) ([]lsifstore.Location, error) {
		if dumpID != expectedDumpID {
			t.Errorf("unexpected id for Definitions. want=%d have=%d", expectedDumpID, dumpID)
		}
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

func setmockLSIFStoreReferences(t *testing.T, mockLSIFStore *MockLSIFStore, expectedDumpID int, expectedPath string, expectedLine, expectedCharacter int, locations []lsifstore.Location) {
	mockLSIFStore.ReferencesFunc.SetDefaultHook(func(ctx context.Context, dumpID int, path string, line, character int) ([]lsifstore.Location, error) {
		if dumpID != expectedDumpID {
			t.Errorf("unexpected id for References. want=%d have=%d", expectedDumpID, dumpID)
		}
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

func setmockLSIFStoreHover(t *testing.T, mockLSIFStore *MockLSIFStore, expectedDumpID int, expectedPath string, expectedLine, expectedCharacter int, text string, r lsifstore.Range, exists bool) {
	mockLSIFStore.HoverFunc.SetDefaultHook(func(ctx context.Context, dumpID int, path string, line, character int) (string, lsifstore.Range, bool, error) {
		if dumpID != expectedDumpID {
			t.Errorf("unexpected id for Hover. want=%d have=%d", expectedDumpID, dumpID)
		}
		if path != expectedPath {
			t.Errorf("unexpected path for Hover. want=%s have=%s", expectedPath, path)
		}
		if line != expectedLine {
			t.Errorf("unexpected line for Hover. want=%d have=%d", expectedLine, line)
		}
		if character != expectedCharacter {
			t.Errorf("unexpected character for Hover. want=%d have=%d", expectedCharacter, character)
		}
		return text, r, exists, nil
	})
}

type hoverSpec struct {
	dumpID    int
	path      string
	line      int
	character int
	text      string
	r         lsifstore.Range
	exists    bool
}

func setMultimockLSIFStoreHover(t *testing.T, mockLSIFStore *MockLSIFStore, specs ...hoverSpec) {
	mockLSIFStore.HoverFunc.SetDefaultHook(func(ctx context.Context, dumpID int, path string, line, character int) (string, lsifstore.Range, bool, error) {
		for _, spec := range specs {
			if dumpID == spec.dumpID && path == spec.path && line == spec.line && character == spec.character {
				return spec.text, spec.r, spec.exists, nil
			}
		}

		t.Errorf("unexpected args for Hover. dumpID=%d path=%s line=%d character=%d", dumpID, path, line, character)
		return "", lsifstore.Range{}, false, nil
	})
}

func setmockLSIFStoreDiagnostics(t *testing.T, mockLSIFStore *MockLSIFStore, expectedDumpID int, expectedPrefix string, expectedSkip, expectedTake int, diagnostics []lsifstore.Diagnostic, totalCount int) {
	mockLSIFStore.DiagnosticsFunc.SetDefaultHook(func(ctx context.Context, dumpID int, prefix string, skip, take int) ([]lsifstore.Diagnostic, int, error) {
		if dumpID != expectedDumpID {
			t.Errorf("unexpected id for Diagnostics. want=%d have=%d", expectedDumpID, dumpID)
		}
		if prefix != expectedPrefix {
			t.Errorf("unexpected prefix for Diagnostics. want=%s have=%s", expectedPrefix, prefix)
		}
		if skip != expectedSkip {
			t.Errorf("unexpected skip for Diagnostics. want=%d have=%d", expectedSkip, skip)
		}
		if take != expectedTake {
			t.Errorf("unexpected take for Diagnostics. want=%d have=%d", expectedTake, take)
		}
		return diagnostics, totalCount, nil
	})
}

func setmockLSIFStoreMonikersByPosition(t *testing.T, mockLSIFStore *MockLSIFStore, expectedDumpID int, expectedPath string, expectedLine, expectedCharacter int, monikers [][]lsifstore.MonikerData) {
	mockLSIFStore.MonikersByPositionFunc.SetDefaultHook(func(ctx context.Context, dumpID int, path string, line, character int) ([][]lsifstore.MonikerData, error) {
		if dumpID != expectedDumpID {
			t.Errorf("unexpected id for MonikersByPosition. want=%d have=%d", expectedDumpID, dumpID)
		}
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

func setmockLSIFStoreMonikerResults(t *testing.T, mockLSIFStore *MockLSIFStore, expectedDumpID int, expectedModelType, expectedScheme, expectedIdentifier string, expectedSkip, expectedTake int, locations []lsifstore.Location, totalCount int) {
	mockLSIFStore.MonikerResultsFunc.SetDefaultHook(func(ctx context.Context, dumpID int, modelType, scheme, identifier string, skip, take int) ([]lsifstore.Location, int, error) {
		if dumpID != expectedDumpID {
			t.Errorf("unexpected id for MonikerResults. want=%d have=%d", expectedDumpID, dumpID)
		}
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

type monikerResultsSpec struct {
	dumpID     int
	modelType  string
	scheme     string
	identifier string
	skip       int
	take       int
	locations  []lsifstore.Location
	totalCount int
}

func setMultimockLSIFStoreMonikerResults(t *testing.T, mockLSIFStore *MockLSIFStore, specs ...monikerResultsSpec) {
	mockLSIFStore.MonikerResultsFunc.SetDefaultHook(func(ctx context.Context, dumpID int, modelType, scheme, identifier string, skip, take int) ([]lsifstore.Location, int, error) {
		for _, spec := range specs {
			if dumpID == spec.dumpID && modelType == spec.modelType && scheme == spec.scheme && identifier == spec.identifier && skip == spec.skip && take == spec.take {
				return spec.locations, spec.totalCount, nil
			}
		}

		t.Errorf("unexpected args for MonikerResults. dumpID=%d modelType=%s scheme=%s identifier=%s skip=%d take=%d", dumpID, modelType, scheme, identifier, skip, take)
		return nil, 0, nil
	})
}

func setmockLSIFStorePackageInformation(t *testing.T, mockLSIFStore *MockLSIFStore, expectedDumpID int, expectedPath, expectedPackageInformationID string, packageInformation lsifstore.PackageInformationData) {
	mockLSIFStore.PackageInformationFunc.SetDefaultHook(func(ctx context.Context, dumpID int, path, packageInformationID string) (lsifstore.PackageInformationData, bool, error) {
		if dumpID != expectedDumpID {
			t.Errorf("unexpected id for PackageInformation. want=%d have=%d", expectedDumpID, dumpID)
		}
		if path != expectedPath {
			t.Errorf("unexpected path for PackageInformation. want=%s have=%s", expectedPath, path)
		}
		if packageInformationID != expectedPackageInformationID {
			t.Errorf("unexpected package information id for PackageInformation. want=%s have=%s", expectedPackageInformationID, packageInformationID)
		}
		return packageInformation, true, nil
	})
}

func readTestFilter(t *testing.T, dirname, filename string) []byte {
	content, err := ioutil.ReadFile(fmt.Sprintf("./testdata/filters/%s/%s", dirname, filename))
	if err != nil {
		t.Fatalf("unexpected error reading: %s", err)
	}

	raw, err := hex.DecodeString(strings.TrimSpace(string(content)))
	if err != nil {
		t.Fatalf("unexpected error decoding: %s", err)
	}

	return raw
}

func setMockGitserverCommitGraph(t *testing.T, mockGitserverClient *MockGitserverClient, expectedRepositoryID int, graph map[string][]string) {
	mockGitserverClient.CommitGraphFunc.SetDefaultHook(func(ctx context.Context, s DBStore, repositoryID int, options gitserver.CommitGraphOptions) (map[string][]string, error) {
		if repositoryID != expectedRepositoryID {
			t.Errorf("unexpected repository identifier for CommitGraph. want=%d have=%d", expectedRepositoryID, repositoryID)
		}

		return graph, nil
	})
}
