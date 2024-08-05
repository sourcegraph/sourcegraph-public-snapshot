package graphql

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/v2/testutil/require"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/scip/bindings/go/scip"
	"github.com/sourcegraph/scip/cmd/scip/tests/reprolang/bindings/go/repro"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers/gitresolvers"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	sgtypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Only exposed for tests, production code should use Unchecked function
// directly for clarity.
var repoRelPath = core.NewRepoRelPathUnchecked

func TestRanges(t *testing.T) {
	mockCodeNavService := NewMockCodeNavService()

	mockRequestState := codenav.RequestState{
		RepositoryID: 1,
		Commit:       "deadbeef1",
		Path:         repoRelPath("/src/main"),
	}
	mockOperations := newOperations(observation.TestContextTB(t))

	resolver := newGitBlobLSIFDataResolver(
		mockCodeNavService,
		nil,
		mockRequestState,
		nil,
		nil,
		nil,
		mockOperations,
	)

	args := &resolverstubs.LSIFRangesArgs{StartLine: 10, EndLine: 20}
	if _, err := resolver.Ranges(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockCodeNavService.GetRangesFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockCodeNavService.GetRangesFunc.History()))
	}
	if val := mockCodeNavService.GetRangesFunc.History()[0].Arg3; val != 10 {
		t.Fatalf("unexpected start line. want=%d have=%d", 10, val)
	}
	if val := mockCodeNavService.GetRangesFunc.History()[0].Arg4; val != 20 {
		t.Fatalf("unexpected end line. want=%d have=%d", 20, val)
	}
}

func TestDefinitions(t *testing.T) {
	mockCodeNavService := NewMockCodeNavService()
	mockRequestState := codenav.RequestState{
		RepositoryID: 1,
		Commit:       "deadbeef1",
		Path:         repoRelPath("/src/main"),
	}
	mockOperations := newOperations(observation.TestContextTB(t))

	resolver := newGitBlobLSIFDataResolver(
		mockCodeNavService,
		nil,
		mockRequestState,
		nil,
		nil,
		nil,
		mockOperations,
	)

	args := &resolverstubs.LSIFQueryPositionArgs{Line: 10, Character: 15}
	if _, err := resolver.Definitions(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockCodeNavService.GetDefinitionsFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockCodeNavService.GetDefinitionsFunc.History()))
	}
	pos, ok := mockCodeNavService.GetDefinitionsFunc.History()[0].Arg1.Matcher.PositionBased()
	require.True(t, ok)
	require.True(t, pos.Compare(scip.Position{Line: 10, Character: 15}) == 0)
}

func TestReferences(t *testing.T) {
	mockCodeNavService := NewMockCodeNavService()
	mockRequestState := codenav.RequestState{
		RepositoryID: 1,
		Commit:       "deadbeef1",
		Path:         repoRelPath("/src/main"),
	}
	mockOperations := newOperations(observation.TestContextTB(t))

	resolver := newGitBlobLSIFDataResolver(
		mockCodeNavService,
		nil,
		mockRequestState,
		nil,
		nil,
		nil,
		mockOperations,
	)

	offset := int32(25)
	mockRefCursor := codenav.PreciseCursor{Phase: "local"}
	encodedCursor := mockRefCursor.Encode()
	mockCursor := base64.StdEncoding.EncodeToString([]byte(encodedCursor))

	args := &resolverstubs.LSIFPagedQueryPositionArgs{
		LSIFQueryPositionArgs: resolverstubs.LSIFQueryPositionArgs{
			Line:      10,
			Character: 15,
		},
		PagedConnectionArgs: resolverstubs.PagedConnectionArgs{ConnectionArgs: resolverstubs.ConnectionArgs{First: &offset}, After: &mockCursor},
	}

	if _, err := resolver.References(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockCodeNavService.GetReferencesFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockCodeNavService.GetReferencesFunc.History()))
	}
	pos, ok := mockCodeNavService.GetReferencesFunc.History()[0].Arg1.Matcher.PositionBased()
	require.True(t, ok)
	require.True(t, pos.Compare(scip.Position{Line: 10, Character: 15}) == 0)
	if val := mockCodeNavService.GetReferencesFunc.History()[0].Arg1; val.Limit != 25 {
		t.Fatalf("unexpected character. want=%v have=%v", 25, val)
	}
	if val := mockCodeNavService.GetReferencesFunc.History()[0].Arg1; val.RawCursor != encodedCursor {
		t.Fatalf("unexpected character. want=%v have=%v", "test-cursor", val)
	}
}

func TestReferencesDefaultLimit(t *testing.T) {
	mockCodeNavService := NewMockCodeNavService()
	mockRequestState := codenav.RequestState{
		RepositoryID: 1,
		Commit:       "deadbeef1",
		Path:         repoRelPath("/src/main"),
	}
	mockOperations := newOperations(observation.TestContextTB(t))

	resolver := newGitBlobLSIFDataResolver(
		mockCodeNavService,
		nil,
		mockRequestState,
		nil,
		nil,
		nil,
		mockOperations,
	)

	args := &resolverstubs.LSIFPagedQueryPositionArgs{
		LSIFQueryPositionArgs: resolverstubs.LSIFQueryPositionArgs{
			Line:      10,
			Character: 15,
		},
		PagedConnectionArgs: resolverstubs.PagedConnectionArgs{},
	}

	if _, err := resolver.References(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockCodeNavService.GetReferencesFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockCodeNavService.GetDiagnosticsFunc.History()))
	}
	if val := mockCodeNavService.GetReferencesFunc.History()[0].Arg1; val.Limit != DefaultReferencesPageSize {
		t.Fatalf("unexpected limit. want=%v have=%v", DefaultReferencesPageSize, val)
	}
}

func TestReferencesDefaultIllegalLimit(t *testing.T) {
	mockCodeNavService := NewMockCodeNavService()
	mockRequestState := codenav.RequestState{
		RepositoryID: 1,
		Commit:       "deadbeef1",
		Path:         repoRelPath("/src/main"),
	}
	mockOperations := newOperations(observation.TestContextTB(t))

	resolver := newGitBlobLSIFDataResolver(
		mockCodeNavService,
		nil,
		mockRequestState,
		nil,
		nil,
		nil,
		mockOperations,
	)

	offset := int32(-1)
	args := &resolverstubs.LSIFPagedQueryPositionArgs{
		LSIFQueryPositionArgs: resolverstubs.LSIFQueryPositionArgs{
			Line:      10,
			Character: 15,
		},
		PagedConnectionArgs: resolverstubs.PagedConnectionArgs{ConnectionArgs: resolverstubs.ConnectionArgs{First: &offset}},
	}

	if _, err := resolver.References(context.Background(), args); err != ErrIllegalLimit {
		t.Fatalf("unexpected error. want=%q have=%q", ErrIllegalLimit, err)
	}
}

func TestHover(t *testing.T) {
	mockCodeNavService := NewMockCodeNavService()
	mockRequestState := codenav.RequestState{
		RepositoryID: 1,
		Commit:       "deadbeef1",
		Path:         repoRelPath("/src/main"),
	}
	mockOperations := newOperations(observation.TestContextTB(t))

	resolver := newGitBlobLSIFDataResolver(
		mockCodeNavService,
		nil,
		mockRequestState,
		nil,
		nil,
		nil,
		mockOperations,
	)

	mockCodeNavService.GetHoverFunc.SetDefaultReturn("text", shared.Range{}, true, nil)
	args := &resolverstubs.LSIFQueryPositionArgs{Line: 10, Character: 15}
	if _, err := resolver.Hover(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockCodeNavService.GetHoverFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockCodeNavService.GetHoverFunc.History()))
	}
	if val := mockCodeNavService.GetHoverFunc.History()[0].Arg1; val.Line != 10 {
		t.Fatalf("unexpected line. want=%v have=%v", 10, val)
	}
	if val := mockCodeNavService.GetHoverFunc.History()[0].Arg1; val.Character != 15 {
		t.Fatalf("unexpected character. want=%v have=%v", 15, val)
	}
}

func TestDiagnostics(t *testing.T) {
	mockCodeNavService := NewMockCodeNavService()
	mockRequestState := codenav.RequestState{
		RepositoryID: 1,
		Commit:       "deadbeef1",
		Path:         repoRelPath("/src/main"),
	}
	mockOperations := newOperations(observation.TestContextTB(t))

	resolver := newGitBlobLSIFDataResolver(
		mockCodeNavService,
		nil,
		mockRequestState,
		nil,
		nil,
		nil,
		mockOperations,
	)

	offset := int32(25)
	args := &resolverstubs.LSIFDiagnosticsArgs{
		First: &offset,
	}

	if _, err := resolver.Diagnostics(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockCodeNavService.GetDiagnosticsFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockCodeNavService.GetDiagnosticsFunc.History()))
	}
	if val := mockCodeNavService.GetDiagnosticsFunc.History()[0].Arg1; val.Limit != 25 {
		t.Fatalf("unexpected limit. want=%v have=%v", 25, val)
	}
}

func TestDiagnosticsDefaultLimit(t *testing.T) {
	mockCodeNavService := NewMockCodeNavService()
	mockRequestState := codenav.RequestState{
		RepositoryID: 1,
		Commit:       "deadbeef1",
		Path:         repoRelPath("/src/main"),
	}
	mockOperations := newOperations(observation.TestContextTB(t))

	resolver := newGitBlobLSIFDataResolver(
		mockCodeNavService,
		nil,
		mockRequestState,
		nil,
		nil,
		nil,
		mockOperations,
	)

	args := &resolverstubs.LSIFDiagnosticsArgs{}

	if _, err := resolver.Diagnostics(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockCodeNavService.GetDiagnosticsFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockCodeNavService.GetDiagnosticsFunc.History()))
	}
	if val := mockCodeNavService.GetDiagnosticsFunc.History()[0].Arg1; val.Limit != DefaultDiagnosticsPageSize {
		t.Fatalf("unexpected limit. want=%v have=%v", DefaultDiagnosticsPageSize, val)
	}
}

func TestDiagnosticsDefaultIllegalLimit(t *testing.T) {
	mockCodeNavService := NewMockCodeNavService()
	mockRequestState := codenav.RequestState{
		RepositoryID: 1,
		Commit:       "deadbeef1",
		Path:         repoRelPath("/src/main"),
	}
	mockOperations := newOperations(observation.TestContextTB(t))

	resolver := newGitBlobLSIFDataResolver(
		mockCodeNavService,
		nil,
		mockRequestState,
		nil,
		nil,
		nil,
		mockOperations,
	)

	offset := int32(-1)
	args := &resolverstubs.LSIFDiagnosticsArgs{
		First: &offset,
	}

	if _, err := resolver.Diagnostics(context.Background(), args); err != ErrIllegalLimit {
		t.Fatalf("unexpected error. want=%q have=%q", ErrIllegalLimit, err)
	}
}

func TestResolveLocations(t *testing.T) {
	repos := dbmocks.NewStrictMockRepoStore()
	repos.GetFunc.SetDefaultHook(func(_ context.Context, id api.RepoID) (*sgtypes.Repo, error) {
		return &sgtypes.Repo{ID: id, Name: api.RepoName(fmt.Sprintf("repo%d", id))}, nil
	})

	gsClient := gitserver.NewMockClient()
	gsClient.ResolveRevisionFunc.SetDefaultHook(func(_ context.Context, _ api.RepoName, spec string, _ gitserver.ResolveRevisionOptions) (api.CommitID, error) {
		if spec == "deadbeef3" {
			return "", &gitdomain.RevisionNotFoundError{}
		}
		return api.CommitID(spec), nil
	})

	factory := gitresolvers.NewCachedLocationResolverFactory(repos, gsClient)
	locationResolver := factory.Create()

	r1 := shared.Range{Start: shared.Position{Line: 11, Character: 12}, End: shared.Position{Line: 13, Character: 14}}
	r2 := shared.Range{Start: shared.Position{Line: 21, Character: 22}, End: shared.Position{Line: 23, Character: 24}}
	r3 := shared.Range{Start: shared.Position{Line: 31, Character: 32}, End: shared.Position{Line: 33, Character: 34}}
	r4 := shared.Range{Start: shared.Position{Line: 41, Character: 42}, End: shared.Position{Line: 43, Character: 44}}

	locations, err := resolveLocations(context.Background(), locationResolver, []shared.UploadLocation{
		{Upload: uploadsshared.CompletedUpload{RepositoryID: 50}, TargetCommit: "deadbeef1", TargetRange: r1, Path: repoRelPath("p1")},
		{Upload: uploadsshared.CompletedUpload{RepositoryID: 51}, TargetCommit: "deadbeef2", TargetRange: r2, Path: repoRelPath("p2")},
		{Upload: uploadsshared.CompletedUpload{RepositoryID: 52}, TargetCommit: "deadbeef3", TargetRange: r3, Path: repoRelPath("p3")},
		{Upload: uploadsshared.CompletedUpload{RepositoryID: 53}, TargetCommit: "deadbeef4", TargetRange: r4, Path: repoRelPath("p4")},
	})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	mockrequire.Called(t, repos.GetFunc)

	if len(locations) != 3 {
		t.Fatalf("unexpected length. want=%d have=%d", 3, len(locations))
	}
	if url := locations[0].CanonicalURL(); url != "/repo50@deadbeef1/-/blob/p1?L12:13-14:15" {
		t.Errorf("unexpected canonical url. want=%s have=%s", "/repo50@deadbeef1/-/blob/p1?L12:13-14:15", url)
	}
	if url := locations[1].CanonicalURL(); url != "/repo51@deadbeef2/-/blob/p2?L22:23-24:25" {
		t.Errorf("unexpected canonical url. want=%s have=%s", "/repo51@deadbeef2/-/blob/p2?L22:23-24:25", url)
	}
	if url := locations[2].CanonicalURL(); url != "/repo53@deadbeef4/-/blob/p4?L42:43-44:45" {
		t.Errorf("unexpected canonical url. want=%s have=%s", "/repo53@deadbeef4/-/blob/p4?L42:43-44:45", url)
	}
}

func sampleSourceFiles() []*scip.SourceFile {
	testFiles := []struct {
		path    string
		content string
	}{
		{
			path: "locals.repro",
			content: `definition local_a
reference local_a
`,
		},
	}
	out := []*scip.SourceFile{}
	for _, testFile := range testFiles {
		out = append(out, &scip.SourceFile{
			AbsolutePath: "/var/myproject/" + testFile.path,
			RelativePath: testFile.path,
			Text:         testFile.content,
			Lines:        strings.Split(testFile.content, "\n"),
		})
	}
	return out
}

func unwrap[T any](v T, err error) func(*testing.T) T {
	return func(t *testing.T) T {
		require.NoError(t, err)
		return v
	}
}

func makeTestResolver(t *testing.T) resolverstubs.CodeGraphDataResolver {
	codeNavSvc := NewStrictMockCodeNavService()
	gitTreeTranslator := codenav.NewMockGitTreeTranslator()
	index := unwrap(repro.Index("", "testpkg", sampleSourceFiles(), nil))(t)
	errUploadNotFound := errors.New("upload not found")
	errDocumentNotFound := errors.New("document not found")
	testUpload := uploadsshared.CompletedUpload{ID: 82}
	codeNavSvc.SCIPDocumentFunc.SetDefaultHook(func(_ context.Context, _ codenav.GitTreeTranslator, upload core.UploadLike, _ api.CommitID, path core.RepoRelPath) (*scip.Document, error) {
		if upload.GetID() != testUpload.ID {
			return nil, errUploadNotFound
		}
		for _, d := range index.Documents {
			if path.RawValue() == d.RelativePath {
				return d, nil
			}
		}
		return nil, errDocumentNotFound
	})

	return newCodeGraphDataResolver(
		codeNavSvc, gitTreeTranslator, testUpload,
		&resolverstubs.CodeGraphDataOpts{Repo: &sgtypes.Repo{}, Path: repoRelPath("locals.repro")},
		codenav.ProvenancePrecise, newOperations(&observation.TestContext))
}

func TestOccurrences_BadArgs(t *testing.T) {
	resolver := makeTestResolver(t)
	bgCtx := context.Background()

	t.Run("fetching with undeserializable 'after'", func(t *testing.T) {
		badArgs := resolverstubs.OccurrencesArgs{After: pointers.Ptr("not-a-cursor")}
		normalizedArgs := badArgs.Normalize(10)
		occs := unwrap(resolver.Occurrences(bgCtx, normalizedArgs))(t)
		_, err := occs.Nodes(bgCtx)
		require.Error(t, err)
	})

	t.Run("fetching with out-of-bounds 'after'", func(t *testing.T) {
		oobCursor := unwrap(marshalCursor(cursor{100}))(t)
		badArgs := resolverstubs.OccurrencesArgs{After: oobCursor}
		normalizedArgs := badArgs.Normalize(10)
		occs := unwrap(resolver.Occurrences(bgCtx, normalizedArgs))(t)
		nodes, err := occs.Nodes(bgCtx)
		// TODO: I think this should be an out-of-bounds error, Slack discussion:
		// https://sourcegraph.slack.com/archives/C02UC4WUX1Q/p1716378462737019
		require.NoError(t, err)
		require.Equal(t, 0, len(nodes))
	})
}

func TestOccurrences_Pages(t *testing.T) {
	resolver := makeTestResolver(t)
	bgCtx := context.Background()

	type TestCase struct {
		name        string
		initialArgs *resolverstubs.OccurrencesArgs
		// Run with go test <path> -update to update the wantPages values
		wantPages autogold.Value
	}

	type occurrenceNode struct {
		Symbol string
		Range  []int32
		Roles  []string
	}

	testCases := []TestCase{
		{
			name:        "Single page",
			initialArgs: (&resolverstubs.OccurrencesArgs{}).Normalize(10),
			wantPages: autogold.Expect([][]occurrenceNode{{
				{
					Symbol: "local _a",
					Range: []int32{
						0,
						11,
						0,
						18,
					},
					Roles: []string{"DEFINITION"},
				},
				{
					Symbol: "local _a",
					Range: []int32{
						1,
						10,
						1,
						17,
					},
					Roles: []string{"REFERENCE"},
				},
			}}),
		},
		{
			name:        "Multiple pages",
			initialArgs: (&resolverstubs.OccurrencesArgs{}).Normalize(1),
			wantPages: autogold.Expect([][]occurrenceNode{
				{
					{
						Symbol: "local _a",
						Range: []int32{
							0,
							11,
							0,
							18,
						},
						Roles: []string{"DEFINITION"},
					},
				},
				{{
					Symbol: "local _a",
					Range: []int32{
						1,
						10,
						1,
						17,
					},
					Roles: []string{"REFERENCE"},
				}},
			}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			allOccurrences := [][]occurrenceNode{}
			args := testCase.initialArgs
			const maxIters = 10
			i := 0
			for ; i < maxIters; i++ {
				connx := unwrap(resolver.Occurrences(bgCtx, args))(t)
				occs := unwrap(connx.Nodes(bgCtx))(t)
				var nodes []occurrenceNode
				for _, occ := range occs {
					s := unwrap(occ.Symbol())(t)
					r := unwrap(occ.Range())(t)
					roles := unwrap(occ.Roles())(t)
					var rolesStrs []string
					for _, role := range *roles {
						rolesStrs = append(rolesStrs, string(role))
					}
					nodes = append(nodes, occurrenceNode{
						Symbol: *s,
						Range:  []int32{r.Start().Line(), r.Start().Character(), r.End().Line(), r.End().Character()},
						Roles:  rolesStrs,
					})
				}
				allOccurrences = append(allOccurrences, nodes)
				pages := unwrap(connx.PageInfo(bgCtx))(t)
				if pages.HasNextPage() {
					endCursor := unwrap(pages.EndCursor())(t)
					args.After = endCursor
				} else {
					break
				}
			}
			require.Less(t, i, maxIters)
			testCase.wantPages.Equal(t, allOccurrences)
		})
	}
}

func TestSubRepoPerms(t *testing.T) {
	checker := authz.NewMockSubRepoPermissionChecker()
	oldChecker := authz.DefaultSubRepoPermsChecker
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{
		ExperimentalFeatures: &schema.ExperimentalFeatures{ScipBasedAPIs: pointers.Ptr(true)},
	}})
	authz.DefaultSubRepoPermsChecker = checker
	t.Cleanup(func() {
		conf.Mock(nil)
		authz.DefaultSubRepoPermsChecker = oldChecker
	})

	a := &actor.Actor{UID: 1}
	ctx := actor.WithActor(context.Background(), a)
	repoName := api.RepoName("foo")

	checker.EnabledFunc.SetDefaultReturn(true)
	permsFunc := func(path string) (authz.Perms, error) {
		switch path {
		case "can-access.txt":
			return authz.Read, nil
		default:
			return authz.None, nil
		}
	}
	checker.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
		return permsFunc(content.Path)
	})
	checker.EnabledForRepoFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName) (bool, error) {
		if rn == repoName {
			return true, nil
		}
		return false, nil
	})

	mockCodeNavService := NewMockCodeNavService()
	observationCtx := observation.NewContext(nil)

	resolver := NewRootResolver(
		observationCtx, mockCodeNavService, nil, nil, nil, nil,
		nil, nil, nil, nil, 10)
	repo := sgtypes.Repo{ID: 0, Name: repoName}
	opts := resolverstubs.CodeGraphDataOpts{Args: nil, Repo: &repo, Commit: ""}
	{
		opts := opts
		opts.Path = core.NewRepoRelPathUnchecked("can-access.txt")
		data, err := resolver.CodeGraphData(ctx, &opts)
		require.NoError(t, err)
		require.Empty(t, data)
	}
	{
		opts := opts
		opts.Path = core.NewRepoRelPathUnchecked("cannot-access.txt")
		data, err := resolver.CodeGraphData(ctx, &opts)
		require.ErrorIs(t, err, os.ErrNotExist)
		require.Empty(t, data)
	}
}
