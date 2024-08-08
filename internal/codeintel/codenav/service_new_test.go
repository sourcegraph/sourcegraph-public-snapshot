package codenav

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	genslices "github.com/life4/genesis/slices"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/internal/lsifstore"
	lsifstoremocks "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/internal/lsifstore/mocks"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	sgtypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func posMatcher(line int, char int) shared.Matcher {
	return shared.NewStartPositionMatcher(scip.Position{Line: int32(line), Character: int32(char)})
}

func TestGetDefinitions(t *testing.T) {
	t.Run("local", func(t *testing.T) {
		// Set up mocks
		fakeRepoStore := AllPresentFakeRepoStore{}
		mockLsifStore := lsifstoremocks.NewMockLsifStore()
		mockUploadSvc := NewMockUploadService()
		mockGitserverClient := gitserver.NewMockClient()
		mockSearchClient := client.NewMockSearchClient()

		// Init service
		svc := newService(observation.TestContextTB(t), fakeRepoStore, mockLsifStore, mockUploadSvc, mockGitserverClient, mockSearchClient, log.NoOp())

		// Set up request state
		lookupPath := core.NewRepoRelPathUnchecked("sub2/a.go")
		mockRequestState := RequestState{Path: lookupPath}
		mockRequestState.SetLocalCommitCache(fakeRepoStore, mockGitserverClient)

		mockRequestState.GitTreeTranslator = noopTranslator()
		mockRequest := OccurrenceRequestArgs{
			RepositoryID: 51,
			Commit:       mockCommit,
			Limit:        50,
			Path:         lookupPath,
			Matcher:      posMatcher(10, 20),
		}
		mockCommit := string(mockCommit)

		uploads := []uploadsshared.CompletedUpload{
			{ID: 50, Commit: mockCommit, Root: "sub2/"},
			{ID: 51, Commit: mockCommit, Root: "sub2/"},
			{ID: 52, Commit: mockCommit, Root: "sub2/"},
		}
		mockRequestState.SetUploadsDataLoader(uploads)

		locations := genslices.Map([]shared.Range{testRange1, testRange2, testRange3, testRange4, testRange5},
			func(range_ shared.Range) shared.UsageBuilder {
				occ := scip.Occurrence{Range: range_.ToSCIPRange().SCIPRange(), SymbolRoles: int32(scip.SymbolRole_Definition)}
				return shared.NewUsageBuilder(&occ)
			})
		mockLsifStore.ExtractDefinitionLocationsFromPositionFunc.SetDefaultHook(func(ctx context.Context, key lsifstore.FindUsagesKey) ([]shared.UsageBuilder, []string, error) {
			if key.UploadID == 51 {
				return locations, nil, nil
			}
			return nil, nil, nil
		})

		adjustedLocations, _, err := svc.GetDefinitions(context.Background(), mockRequest, mockRequestState, PreciseCursor{})
		if err != nil {
			t.Fatalf("unexpected error querying definitions: %s", err)
		}
		expectedLocations := []shared.UploadUsage{
			{Upload: uploads[1], Path: repoRelPath("sub2/a.go"), TargetCommit: mockCommit, TargetRange: testRange1, Kind: shared.UsageKindDefinition},
			{Upload: uploads[1], Path: repoRelPath("sub2/a.go"), TargetCommit: mockCommit, TargetRange: testRange2, Kind: shared.UsageKindDefinition},
			{Upload: uploads[1], Path: repoRelPath("sub2/a.go"), TargetCommit: mockCommit, TargetRange: testRange3, Kind: shared.UsageKindDefinition},
			{Upload: uploads[1], Path: repoRelPath("sub2/a.go"), TargetCommit: mockCommit, TargetRange: testRange4, Kind: shared.UsageKindDefinition},
			{Upload: uploads[1], Path: repoRelPath("sub2/a.go"), TargetCommit: mockCommit, TargetRange: testRange5, Kind: shared.UsageKindDefinition},
		}

		if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
			t.Errorf("unexpected locations (-want +got):\n%s", diff)
		}
	})

	t.Run("remote", func(t *testing.T) {
		// Set up mocks
		fakeRepoStore := AllPresentFakeRepoStore{}
		mockLsifStore := lsifstoremocks.NewMockLsifStore()
		mockUploadSvc := NewMockUploadService()
		mockGitserverClient := gitserver.NewMockClient()
		mockSearchClient := client.NewMockSearchClient()

		// Init service
		svc := newService(observation.TestContextTB(t), fakeRepoStore, mockLsifStore, mockUploadSvc, mockGitserverClient, mockSearchClient, log.NoOp())

		// Set up request state
		lookupPath := core.NewRepoRelPathUnchecked("sub2/a.go")
		mockRequestState := RequestState{Path: lookupPath}
		mockRequestState.SetLocalCommitCache(fakeRepoStore, mockGitserverClient)
		mockRequestState.SetLocalGitTreeTranslator(mockGitserverClient, &sgtypes.Repo{ID: 42})
		mockRequestState.GitTreeTranslator = noopTranslator()
		uploads1 := []uploadsshared.CompletedUpload{
			{ID: 50, Commit: "deadbeef", Root: "sub1/"},
			{ID: 51, Commit: "deadbeef", Root: "sub2/"},
			{ID: 52, Commit: "deadbeef", Root: "sub3/"},
			{ID: 53, Commit: "deadbeef", Root: "sub4/"},
		}
		mockRequestState.SetUploadsDataLoader(uploads1)

		uploads2 := []uploadsshared.CompletedUpload{
			{ID: 150, Commit: "deadbeef1", Root: "sub1/"},
			{ID: 151, Commit: "deadbeef2", Root: "sub2/"},
			{ID: 152, Commit: "deadbeef3", Root: "sub3/"},
			{ID: 153, Commit: "deadbeef4", Root: "sub4/"},
		}
		mockUploadSvc.GetCompletedUploadsWithDefinitionsForMonikersFunc.PushReturn(uploads2, nil)

		// upload #150's commit no longer exists; all others do
		mockGitserverClient.GetCommitFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName, ci api.CommitID) (*gitdomain.Commit, error) {
			if ci == "deadbeef1" {
				return nil, &gitdomain.RevisionNotFoundError{Repo: rn, Spec: string(ci)}
			}
			return &gitdomain.Commit{ID: ci}, nil
		})

		symbolNames := []string{
			"tsc npm leftpad 0.1.0 padLeft.",
			"local pad_left",
			"tsc npm leftpad 0.2.0 pad-left.",
			"local left_pad",
		}
		mockLsifStore.ExtractDefinitionLocationsFromPositionFunc.PushReturn(nil, symbolNames, nil)

		usages := []shared.Usage{
			{UploadID: 151, Path: uploadRelPath("a.go"), Range: testRange1},
			{UploadID: 151, Path: uploadRelPath("b.go"), Range: testRange2},
			{UploadID: 151, Path: uploadRelPath("a.go"), Range: testRange3},
			{UploadID: 151, Path: uploadRelPath("b.go"), Range: testRange4},
			{UploadID: 151, Path: uploadRelPath("c.go"), Range: testRange5},
		}
		mockLsifStore.GetSymbolUsagesFunc.PushReturn(usages, len(usages), nil)

		mockRequest := OccurrenceRequestArgs{
			RepositoryID: 42,
			Commit:       mockCommit,
			Limit:        50,
			Path:         mockPath,
			Matcher:      posMatcher(10, 20),
		}
		remoteUploads := uploads2
		adjustedLocations, _, err := svc.GetDefinitions(context.Background(), mockRequest, mockRequestState, PreciseCursor{})
		if err != nil {
			t.Fatalf("unexpected error querying definitions: %s", err)
		}

		xLocations := []shared.UploadUsage{
			{Upload: remoteUploads[1], Path: repoRelPath("sub2/a.go"), TargetCommit: "deadbeef2", TargetRange: testRange1},
			{Upload: remoteUploads[1], Path: repoRelPath("sub2/b.go"), TargetCommit: "deadbeef2", TargetRange: testRange2},
			{Upload: remoteUploads[1], Path: repoRelPath("sub2/a.go"), TargetCommit: "deadbeef2", TargetRange: testRange3},
			{Upload: remoteUploads[1], Path: repoRelPath("sub2/b.go"), TargetCommit: "deadbeef2", TargetRange: testRange4},
			{Upload: remoteUploads[1], Path: repoRelPath("sub2/c.go"), TargetCommit: "deadbeef2", TargetRange: testRange5},
		}

		if diff := cmp.Diff(xLocations, adjustedLocations); diff != "" {
			t.Errorf("unexpected locations (-want +got):\n%s", diff)
		}

		if history := mockUploadSvc.GetCompletedUploadsWithDefinitionsForMonikersFunc.History(); len(history) != 1 {
			t.Fatalf("unexpected call count for dbstore.DefinitionDump. want=%d have=%d", 1, len(history))
		} else {
			expectedMonikers := []precise.QualifiedMonikerData{
				{
					MonikerData:            precise.MonikerData{Kind: "", Scheme: "tsc", Identifier: "tsc npm leftpad 0.1.0 padLeft."},
					PackageInformationData: precise.PackageInformationData{Manager: "npm", Name: "leftpad", Version: "0.1.0"},
				},
				{
					MonikerData:            precise.MonikerData{Kind: "", Scheme: "tsc", Identifier: "tsc npm leftpad 0.2.0 pad-left."},
					PackageInformationData: precise.PackageInformationData{Manager: "npm", Name: "leftpad", Version: "0.2.0"},
				},
			}
			if diff := cmp.Diff(expectedMonikers, history[0].Arg1); diff != "" {
				t.Errorf("unexpected monikers (-want +got):\n%s", diff)
			}
		}

		if history := mockLsifStore.GetSymbolUsagesFunc.History(); len(history) != 1 {
			t.Fatalf("unexpected call count for lsifstore.BulkMonikerResults. want=%d have=%d", 1, len(history))
		} else {
			options := history[0].Arg1
			if diff := cmp.Diff([]int{50, 51, 52, 53, 151, 152, 153}, options.UploadIDs); diff != "" {
				t.Errorf("unexpected ids (-want +got):\n%s", diff)
			}
			expectedSymbolNames := []string{"tsc npm leftpad 0.1.0 padLeft.", "tsc npm leftpad 0.2.0 pad-left."}
			if diff := cmp.Diff(expectedSymbolNames, options.LookupSymbols); diff != "" {
				t.Errorf("unexpected ids (-want +got):\n%s", diff)
			}
		}
	})
}

func TestGetReferences(t *testing.T) {
	t.Run("local", func(t *testing.T) {
		// Set up mocks
		fakeRepoStore := AllPresentFakeRepoStore{}
		mockLsifStore := lsifstoremocks.NewMockLsifStore()
		mockUploadSvc := NewMockUploadService()
		mockGitserverClient := gitserver.NewMockClient()
		mockSearchClient := client.NewMockSearchClient()

		// Init service
		svc := newService(observation.TestContextTB(t), fakeRepoStore, mockLsifStore, mockUploadSvc, mockGitserverClient, mockSearchClient, log.NoOp())

		// Set up request state
		lookupPath := core.NewRepoRelPathUnchecked("sub2/a.go")
		mockRequestState := RequestState{Path: lookupPath}
		mockRequestState.SetLocalCommitCache(fakeRepoStore, mockGitserverClient)
		mockRequestState.GitTreeTranslator = noopTranslator()
		uploads := []uploadsshared.CompletedUpload{
			{ID: 50, Commit: "deadbeef", Root: "sub2/"},
			{ID: 51, Commit: "deadbeef", Root: "sub2/"},
			{ID: 52, Commit: "deadbeef", Root: "sub2/"},
			{ID: 53, Commit: "deadbeef", Root: "sub2/"},
		}
		mockRequestState.SetUploadsDataLoader(uploads)

		// Empty result set (prevents nil pointer as scanner is always non-nil)
		mockUploadSvc.GetUploadIDsWithReferencesFunc.PushReturn([]int{}, 0, 0, nil)

		locations := genslices.Map([]shared.Range{testRange1, testRange2, testRange3, testRange4, testRange5},
			func(range_ shared.Range) shared.UsageBuilder {
				occ := scip.Occurrence{Range: range_.ToSCIPRange().SCIPRange(), SymbolRoles: 0}
				return shared.NewUsageBuilder(&occ)
			})
		callCount := -1
		mockLsifStore.ExtractReferenceLocationsFromPositionFunc.SetDefaultHook(func(ctx context.Context, key lsifstore.FindUsagesKey) ([]shared.UsageBuilder, []string, error) {
			callCount++
			switch callCount {
			case 0: // uploadID = 50
				return locations[:1], nil, nil
			case 1: // uploadID = 51
				return locations[1:4], nil, nil
			case 2: // uploadID = 52
				return locations[4:], nil, nil
			}
			return nil, nil, nil
		})

		mockCursor := PreciseCursor{}
		mockRequest := OccurrenceRequestArgs{
			RepositoryID: 42,
			Commit:       mockCommit,
			Limit:        50,
			Path:         lookupPath,
			Matcher:      posMatcher(10, 20),
		}
		adjustedLocations, _, err := svc.GetReferences(context.Background(), mockRequest, mockRequestState, mockCursor)
		if err != nil {
			t.Fatalf("unexpected error querying references: %s", err)
		}

		expectedLocations := genslices.Map([]shared.UploadUsage{
			{Upload: uploads[0], TargetRange: testRange1},
			{Upload: uploads[1], TargetRange: testRange2},
			{Upload: uploads[1], TargetRange: testRange3},
			{Upload: uploads[1], TargetRange: testRange4},
			{Upload: uploads[2], TargetRange: testRange5},
		}, func(uu shared.UploadUsage) shared.UploadUsage {
			uu.Path = repoRelPath("sub2/a.go")
			uu.TargetCommit = "deadbeef"
			uu.Kind = shared.UsageKindReference
			return uu
		})
		if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
			t.Errorf("unexpected locations (-want +got):\n%s", diff)
		}
	})

	t.Run("remote", func(t *testing.T) {
		// Set up mocks
		fakeRepoStore := AllPresentFakeRepoStore{}
		mockLsifStore := lsifstoremocks.NewMockLsifStore()
		mockUploadSvc := NewMockUploadService()
		mockGitserverClient := gitserver.NewMockClient()
		mockSearchClient := client.NewMockSearchClient()

		// Init service
		svc := newService(observation.TestContextTB(t), fakeRepoStore, mockLsifStore, mockUploadSvc, mockGitserverClient, mockSearchClient, log.NoOp())

		// Set up request state
		lookupPath := core.NewRepoRelPathUnchecked("sub2/a.go")
		mockRequestState := RequestState{Path: lookupPath}
		mockRequestState.SetLocalCommitCache(fakeRepoStore, mockGitserverClient)
		mockRequestState.GitTreeTranslator = noopTranslator()
		uploads := []uploadsshared.CompletedUpload{
			{ID: 50, Commit: "deadbeef", Root: "sub1/"},
			{ID: 51, Commit: "deadbeef", Root: "sub2/"},
			{ID: 52, Commit: "deadbeef", Root: "sub3/"},
			{ID: 53, Commit: "deadbeef", Root: "sub4/"},
		}
		mockRequestState.SetUploadsDataLoader(uploads)

		definitionUploads := []uploadsshared.CompletedUpload{
			{ID: 150, Commit: "deadbeef1", Root: "sub1/"},
			{ID: 151, Commit: "deadbeef2", Root: "sub2/"},
			{ID: 152, Commit: "deadbeef3", Root: "sub3/"},
			{ID: 153, Commit: "deadbeef4", Root: "sub4/"},
		}
		mockUploadSvc.GetCompletedUploadsWithDefinitionsForMonikersFunc.PushReturn(definitionUploads, nil)

		referenceUploads := []uploadsshared.CompletedUpload{
			{ID: 250, Commit: "deadbeef1", Root: "sub1/"},
			{ID: 251, Commit: "deadbeef2", Root: "sub2/"},
			{ID: 252, Commit: "deadbeef3", Root: "sub3/"},
			{ID: 253, Commit: "deadbeef4", Root: "sub4/"},
		}
		mockUploadSvc.GetCompletedUploadsByIDsFunc.PushReturn(nil, nil) // empty
		mockUploadSvc.GetCompletedUploadsByIDsFunc.PushReturn(referenceUploads[:2], nil)
		mockUploadSvc.GetCompletedUploadsByIDsFunc.PushReturn(referenceUploads[2:], nil)

		mockUploadSvc.GetUploadIDsWithReferencesFunc.PushReturn([]int{250, 251}, 0, 4, nil)
		mockUploadSvc.GetUploadIDsWithReferencesFunc.PushReturn([]int{252, 253}, 0, 2, nil)

		// upload #150/#250's commits no longer exists; all others do
		mockGitserverClient.GetCommitFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName, ci api.CommitID) (*gitdomain.Commit, error) {
			if ci == "deadbeef1" {
				return nil, &gitdomain.RevisionNotFoundError{Repo: rn, Spec: string(ci)}
			}
			return &gitdomain.Commit{ID: ci}, nil
		})

		symbols := []precise.MonikerData{
			{Scheme: "tsc", Identifier: "tsc npm leftpad 0.1.0 padLeft."},
			{Scheme: "tsc", Identifier: "tsc npm leftpad 0.2.0 pad_left."},
			{Scheme: "tsc", Identifier: "tsc npm leftpad 0.3.0 pad-left."},
		}

		packageInformation1 := precise.PackageInformationData{Manager: "npm", Name: "leftpad", Version: "0.1.0"}
		packageInformation2 := precise.PackageInformationData{Manager: "npm", Name: "leftpad", Version: "0.2.0"}
		packageInformation3 := precise.PackageInformationData{Manager: "npm", Name: "leftpad", Version: "0.3.0"}

		ranges := []shared.Range{testRange1, testRange2, testRange3, testRange4, testRange5}
		locations := genslices.Map(ranges, func(range_ shared.Range) shared.UsageBuilder {
			occ := scip.Occurrence{Range: range_.ToSCIPRange().SCIPRange(), SymbolRoles: 0} // reference
			return shared.NewUsageBuilder(&occ)
		})
		symbolNames := []string{
			"tsc npm leftpad 0.1.0 padLeft.",
			"tsc npm leftpad 0.2.0 pad_left.",
			"tsc npm leftpad 0.3.0 pad-left.",
		}
		mockLsifStore.ExtractReferenceLocationsFromPositionFunc.SetDefaultHook(func(ctx context.Context, key lsifstore.FindUsagesKey) ([]shared.UsageBuilder, []string, error) {
			if key.UploadID == 51 {
				return locations, symbolNames, nil
			}
			return nil, nil, nil
		})

		returnedUsages := []shared.Usage{
			{UploadID: 53, Path: uploadRelPath("a.go"), Range: testRange1, Kind: shared.UsageKindDefinition},
			{UploadID: 53, Path: uploadRelPath("b.go"), Range: testRange2, Kind: shared.UsageKindReference},
			{UploadID: 53, Path: uploadRelPath("a.go"), Range: testRange3, Kind: shared.UsageKindReference},
			{UploadID: 53, Path: uploadRelPath("b.go"), Range: testRange4, Kind: shared.UsageKindReference},
			{UploadID: 53, Path: uploadRelPath("c.go"), Range: testRange5, Kind: shared.UsageKindReference},
		}
		mockLsifStore.GetSymbolUsagesFunc.PushReturn(returnedUsages[0:1], 1, nil) // defs
		mockLsifStore.GetSymbolUsagesFunc.PushReturn(returnedUsages[1:2], 1, nil) // refs batch 1
		mockLsifStore.GetSymbolUsagesFunc.PushReturn(returnedUsages[2:], 3, nil)  // refs batch 2

		mockCursor := PreciseCursor{}
		mockRequest := OccurrenceRequestArgs{
			RepositoryID: 42,
			Commit:       mockCommit,
			Limit:        50,
			Path:         mockPath,
			Matcher:      posMatcher(10, 20),
		}
		adjustedLocations, _, err := svc.GetReferences(context.Background(), mockRequest, mockRequestState, mockCursor)
		if err != nil {
			t.Fatalf("unexpected error querying references: %s", err)
		}

		expectedLocations := genslices.Map([]shared.UploadUsage{
			{Upload: uploads[1], Path: repoRelPath("sub2/a.go"), TargetRange: testRange1},
			{Upload: uploads[1], Path: repoRelPath("sub2/a.go"), TargetRange: testRange2},
			{Upload: uploads[1], Path: repoRelPath("sub2/a.go"), TargetRange: testRange3},
			{Upload: uploads[1], Path: repoRelPath("sub2/a.go"), TargetRange: testRange4},
			{Upload: uploads[1], Path: repoRelPath("sub2/a.go"), TargetRange: testRange5},
			{Upload: uploads[3], Path: repoRelPath("sub4/a.go"), TargetRange: testRange1, Kind: shared.UsageKindDefinition},
			{Upload: uploads[3], Path: repoRelPath("sub4/b.go"), TargetRange: testRange2},
			{Upload: uploads[3], Path: repoRelPath("sub4/a.go"), TargetRange: testRange3},
			{Upload: uploads[3], Path: repoRelPath("sub4/b.go"), TargetRange: testRange4},
			{Upload: uploads[3], Path: repoRelPath("sub4/c.go"), TargetRange: testRange5},
		}, func(usage shared.UploadUsage) shared.UploadUsage {
			usage.TargetCommit = "deadbeef"
			if usage.Kind != shared.UsageKindDefinition {
				usage.Kind = shared.UsageKindReference
			}
			return usage
		})
		if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
			t.Errorf("unexpected locations (-want +got):\n%s", diff)
		}

		if history := mockUploadSvc.GetCompletedUploadsWithDefinitionsForMonikersFunc.History(); len(history) != 1 {
			t.Fatalf("unexpected call count for dbstore.DefinitionDump. want=%d have=%d", 1, len(history))
		} else {
			expectedMonikers := []precise.QualifiedMonikerData{
				{MonikerData: symbols[0], PackageInformationData: packageInformation1},
				{MonikerData: symbols[1], PackageInformationData: packageInformation2},
				{MonikerData: symbols[2], PackageInformationData: packageInformation3},
			}
			if diff := cmp.Diff(expectedMonikers, history[0].Arg1); diff != "" {
				t.Errorf("unexpected monikers (-want +got):\n%s", diff)
			}
		}

		if history := mockLsifStore.GetSymbolUsagesFunc.History(); len(history) != 3 {
			t.Fatalf("unexpected call count for lsifstore.BulkMonikerResults. want=%d have=%d", 3, len(history))
		} else {
			if diff := cmp.Diff([]int{50, 51, 52, 53, 151, 152, 153}, history[0].Arg1.UploadIDs); diff != "" {
				t.Errorf("unexpected ids (-want +got):\n%s", diff)
			}

			expectedSymbolNames := []string{
				symbols[0].Identifier,
				symbols[1].Identifier,
				symbols[2].Identifier,
			}
			if diff := cmp.Diff(expectedSymbolNames, history[0].Arg1.LookupSymbols); diff != "" {
				t.Errorf("unexpected symbols (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff([]int{250, 251}, history[1].Arg1.UploadIDs); diff != "" {
				t.Errorf("unexpected ids (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(expectedSymbolNames, history[1].Arg1.LookupSymbols); diff != "" {
				t.Errorf("unexpected symbols (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff([]int{252, 253}, history[2].Arg1.UploadIDs); diff != "" {
				t.Errorf("unexpected ids (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(expectedSymbolNames, history[2].Arg1.LookupSymbols); diff != "" {
				t.Errorf("unexpected symbols (-want +got):\n%s", diff)
			}
		}
	})
}

func TestGetImplementations(t *testing.T) {
	t.Run("local", func(t *testing.T) {
		// Set up mocks
		fakeRepoStore := AllPresentFakeRepoStore{}
		mockLsifStore := lsifstoremocks.NewMockLsifStore()
		mockUploadSvc := NewMockUploadService()
		mockGitserverClient := gitserver.NewMockClient()
		mockSearchClient := client.NewMockSearchClient()

		// Init service
		svc := newService(observation.TestContextTB(t), fakeRepoStore, mockLsifStore, mockUploadSvc, mockGitserverClient, mockSearchClient, log.NoOp())

		// Set up request state
		lookupPath := core.NewRepoRelPathUnchecked("sub2/a.go")
		mockRequestState := RequestState{Path: lookupPath}
		mockRequestState.SetLocalCommitCache(fakeRepoStore, mockGitserverClient)
		mockRequestState.GitTreeTranslator = noopTranslator()

		// Empty result set (prevents nil pointer as scanner is always non-nil)
		mockUploadSvc.GetUploadIDsWithReferencesFunc.PushReturn([]int{}, 0, 0, nil)

		ranges := []shared.Range{testRange1, testRange2, testRange3, testRange4, testRange5}
		locations := genslices.Map(ranges,
			func(range_ shared.Range) shared.UsageBuilder {
				occ := scip.Occurrence{Range: range_.ToSCIPRange().SCIPRange(), SymbolRoles: int32(scip.SymbolRole_Definition)}
				return shared.NewUsageBuilder(&occ)
			})
		mockLsifStore.ExtractImplementationLocationsFromPositionFunc.SetDefaultHook(func(ctx context.Context, key lsifstore.FindUsagesKey) ([]shared.UsageBuilder, []string, error) {
			if key.UploadID == 51 {
				return locations, nil, nil
			}
			return nil, nil, nil
		})

		uploads := []uploadsshared.CompletedUpload{
			{ID: 50, Commit: "deadbeef", Root: "sub2/"},
			{ID: 51, Commit: "deadbeef", Root: "sub2/"},
			{ID: 52, Commit: "deadbeef", Root: "sub2/"},
		}
		mockRequestState.SetUploadsDataLoader(uploads)
		mockCursor := PreciseCursor{}
		mockRequest := OccurrenceRequestArgs{
			RepositoryID: 99,
			Commit:       "deadbeef",
			Limit:        50,
			Path:         lookupPath,
			Matcher:      posMatcher(10, 20),
		}
		adjustedLocations, _, err := svc.GetImplementations(context.Background(), mockRequest, mockRequestState, mockCursor)
		if err != nil {
			t.Fatalf("unexpected error querying implementations: %s", err)
		}

		expectedLocations := genslices.Map(ranges, func(range_ shared.Range) shared.UploadUsage {
			return shared.UploadUsage{Upload: uploads[1], Path: lookupPath, TargetCommit: "deadbeef", TargetRange: range_, Kind: shared.UsageKindImplementation}
		})
		if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
			t.Errorf("unexpected locations (-want +got):\n%s", diff)
		}
	})
}
