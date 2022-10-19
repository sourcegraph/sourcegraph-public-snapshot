package graphql

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestRanges(t *testing.T) {
	mockCodeNavService := NewMockCodeNavService()
	// mockGitBlobResolver := NewMockGitBlobResolver()
	mockAutoIndexingSvc := NewMockAutoIndexingService()
	mockUploadsService := NewMockUploadsService()
	mockPolicyService := NewMockPolicyService()

	mockRequestState := codenav.RequestState{
		RepositoryID: 1,
		Commit:       "deadbeef1",
		Path:         "/src/main",
	}
	mockOperations := newOperations(&observation.TestContext)

	resolver := NewGitBlobLSIFDataResolverQueryResolver(
		mockCodeNavService,
		mockAutoIndexingSvc,
		mockUploadsService,
		mockPolicyService,
		mockRequestState,
		observation.NewErrorCollector(),
		mockOperations,
	)

	args := &LSIFRangesArgs{StartLine: 10, EndLine: 20}
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
	// mockGitBlobResolver := NewMockGitBlobResolver()
	mockAutoIndexingSvc := NewMockAutoIndexingService()
	mockUploadsService := NewMockUploadsService()
	mockPolicyService := NewMockPolicyService()

	// resolver := NewGitBlobLSIFDataResolverQueryResolver(
	// 	mockAutoIndexingSvc,
	// 	mockUploadsService,
	// 	mockPolicyService,
	// 	mockGitBlobResolver,
	// 	observation.NewErrorCollector(),
	// )
	mockCodeNavService := NewMockCodeNavService()
	mockRequestState := codenav.RequestState{
		RepositoryID: 1,
		Commit:       "deadbeef1",
		Path:         "/src/main",
	}
	mockOperations := newOperations(&observation.TestContext)

	resolver := NewGitBlobLSIFDataResolverQueryResolver(
		mockCodeNavService,
		mockAutoIndexingSvc,
		mockUploadsService,
		mockPolicyService,
		mockRequestState,
		observation.NewErrorCollector(),
		mockOperations,
	)

	args := &LSIFQueryPositionArgs{Line: 10, Character: 15}
	if _, err := resolver.Definitions(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockCodeNavService.GetDefinitionsFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockCodeNavService.GetDefinitionsFunc.History()))
	}
	if val := mockCodeNavService.GetDefinitionsFunc.History()[0].Arg1; val.Line != 10 {
		t.Fatalf("unexpected line. want=%v have=%v", 10, val)
	}
	if val := mockCodeNavService.GetDefinitionsFunc.History()[0].Arg1; val.Character != 15 {
		t.Fatalf("unexpected character. want=%d have=%v", 15, val)
	}
}

func TestReferences(t *testing.T) {
	// mockGitBlobResolver := NewMockGitBlobResolver()
	mockAutoIndexingSvc := NewMockAutoIndexingService()
	mockUploadsService := NewMockUploadsService()
	mockPolicyService := NewMockPolicyService()

	// resolver := NewGitBlobLSIFDataResolverQueryResolver(
	// 	mockAutoIndexingSvc,
	// 	mockUploadsService,
	// 	mockPolicyService,
	// 	mockGitBlobResolver,
	// 	observation.NewErrorCollector(),
	// )
	mockCodeNavService := NewMockCodeNavService()
	mockRequestState := codenav.RequestState{
		RepositoryID: 1,
		Commit:       "deadbeef1",
		Path:         "/src/main",
	}
	mockOperations := newOperations(&observation.TestContext)

	resolver := NewGitBlobLSIFDataResolverQueryResolver(
		mockCodeNavService,
		mockAutoIndexingSvc,
		mockUploadsService,
		mockPolicyService,
		mockRequestState,
		observation.NewErrorCollector(),
		mockOperations,
	)

	offset := int32(25)
	cursor := base64.StdEncoding.EncodeToString([]byte("test-cursor"))
	// cursor := "test-cursor"

	args := &LSIFPagedQueryPositionArgs{
		LSIFQueryPositionArgs: LSIFQueryPositionArgs{
			Line:      10,
			Character: 15,
		},
		ConnectionArgs: ConnectionArgs{First: &offset},
		After:          &cursor,
	}

	if _, err := resolver.References(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockCodeNavService.GetReferencesFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockCodeNavService.GetReferencesFunc.History()))
	}
	if val := mockCodeNavService.GetReferencesFunc.History()[0].Arg1; val.Line != 10 {
		t.Fatalf("unexpected line. want=%v have=%v", 10, val)
	}
	if val := mockCodeNavService.GetReferencesFunc.History()[0].Arg1; val.Character != 15 {
		t.Fatalf("unexpected character. want=%v have=%v", 15, val)
	}
	if val := mockCodeNavService.GetReferencesFunc.History()[0].Arg1; val.Limit != 25 {
		t.Fatalf("unexpected character. want=%v have=%v", 25, val)
	}
	if val := mockCodeNavService.GetReferencesFunc.History()[0].Arg1; val.RawCursor != "test-cursor" {
		t.Fatalf("unexpected character. want=%v have=%v", "test-cursor", val)
	}
}

// func TestReferencesDefaultLimit(t *testing.T) {
// 	mockGitBlobResolver := NewMockGitBlobResolver()
// 	mockAutoIndexingSvc := NewMockAutoIndexingService()
// 	mockUploadsService := NewMockUploadsService()
// 	mockPolicyService := NewMockPolicyService()

// 	resolver := NewGitBlobLSIFDataResolverQueryResolver(
// 		mockAutoIndexingSvc,
// 		mockUploadsService,
// 		mockPolicyService,
// 		mockGitBlobResolver,
// 		observation.NewErrorCollector(),
// 	)

// 	args := &LSIFPagedQueryPositionArgs{
// 		LSIFQueryPositionArgs: LSIFQueryPositionArgs{
// 			Line:      10,
// 			Character: 15,
// 		},
// 		ConnectionArgs: ConnectionArgs{},
// 	}

// 	if _, err := resolver.References(context.Background(), args); err != nil {
// 		t.Fatalf("unexpected error: %s", err)
// 	}

// 	if len(mockGitBlobResolver.ReferencesFunc.History()) != 1 {
// 		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockGitBlobResolver.DiagnosticsFunc.History()))
// 	}
// 	if val := mockGitBlobResolver.ReferencesFunc.History()[0].Arg3; val != DefaultReferencesPageSize {
// 		t.Fatalf("unexpected limit. want=%d have=%d", DefaultReferencesPageSize, val)
// 	}
// }

// func TestReferencesDefaultIllegalLimit(t *testing.T) {
// 	mockGitBlobResolver := NewMockGitBlobResolver()
// 	mockAutoIndexingSvc := NewMockAutoIndexingService()
// 	mockUploadsService := NewMockUploadsService()
// 	mockPolicyService := NewMockPolicyService()

// 	resolver := NewGitBlobLSIFDataResolverQueryResolver(
// 		mockAutoIndexingSvc,
// 		mockUploadsService,
// 		mockPolicyService,
// 		mockGitBlobResolver,
// 		observation.NewErrorCollector(),
// 	)

// 	offset := int32(-1)
// 	args := &LSIFPagedQueryPositionArgs{
// 		LSIFQueryPositionArgs: LSIFQueryPositionArgs{
// 			Line:      10,
// 			Character: 15,
// 		},
// 		ConnectionArgs: ConnectionArgs{First: &offset},
// 	}

// 	if _, err := resolver.References(context.Background(), args); err != ErrIllegalLimit {
// 		t.Fatalf("unexpected error. want=%q have=%q", ErrIllegalLimit, err)
// 	}
// }

// func TestHover(t *testing.T) {
// 	mockGitBlobResolver := NewMockGitBlobResolver()
// 	mockAutoIndexingSvc := NewMockAutoIndexingService()
// 	mockUploadsService := NewMockUploadsService()
// 	mockPolicyService := NewMockPolicyService()

// 	resolver := NewGitBlobLSIFDataResolverQueryResolver(
// 		mockAutoIndexingSvc,
// 		mockUploadsService,
// 		mockPolicyService,
// 		mockGitBlobResolver,
// 		observation.NewErrorCollector(),
// 	)
// 	mockGitBlobResolver.HoverFunc.SetDefaultReturn("text", types.Range{}, true, nil)
// 	args := &LSIFQueryPositionArgs{Line: 10, Character: 15}
// 	if _, err := resolver.Hover(context.Background(), args); err != nil {
// 		t.Fatalf("unexpected error: %s", err)
// 	}

// 	if len(mockGitBlobResolver.HoverFunc.History()) != 1 {
// 		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockGitBlobResolver.HoverFunc.History()))
// 	}
// 	if val := mockGitBlobResolver.HoverFunc.History()[0].Arg1; val != 10 {
// 		t.Fatalf("unexpected line. want=%d have=%d", 10, val)
// 	}
// 	if val := mockGitBlobResolver.HoverFunc.History()[0].Arg2; val != 15 {
// 		t.Fatalf("unexpected character. want=%d have=%d", 15, val)
// 	}
// }

// func TestDiagnostics(t *testing.T) {
// 	mockGitBlobResolver := NewMockGitBlobResolver()
// 	mockAutoIndexingSvc := NewMockAutoIndexingService()
// 	mockUploadsService := NewMockUploadsService()
// 	mockPolicyService := NewMockPolicyService()

// 	resolver := NewGitBlobLSIFDataResolverQueryResolver(
// 		mockAutoIndexingSvc,
// 		mockUploadsService,
// 		mockPolicyService,
// 		mockGitBlobResolver,
// 		observation.NewErrorCollector(),
// 	)

// 	offset := int32(25)
// 	args := &LSIFDiagnosticsArgs{
// 		ConnectionArgs: ConnectionArgs{First: &offset},
// 	}

// 	if _, err := resolver.Diagnostics(context.Background(), args); err != nil {
// 		t.Fatalf("unexpected error: %s", err)
// 	}

// 	if len(mockGitBlobResolver.DiagnosticsFunc.History()) != 1 {
// 		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockGitBlobResolver.DiagnosticsFunc.History()))
// 	}
// 	if val := mockGitBlobResolver.DiagnosticsFunc.History()[0].Arg1; val != 25 {
// 		t.Fatalf("unexpected limit. want=%d have=%d", 25, val)
// 	}
// }

// func TestDiagnosticsDefaultLimit(t *testing.T) {
// 	mockGitBlobResolver := NewMockGitBlobResolver()
// 	mockAutoIndexingSvc := NewMockAutoIndexingService()
// 	mockUploadsService := NewMockUploadsService()
// 	mockPolicyService := NewMockPolicyService()

// 	resolver := NewGitBlobLSIFDataResolverQueryResolver(
// 		mockAutoIndexingSvc,
// 		mockUploadsService,
// 		mockPolicyService,
// 		mockGitBlobResolver,
// 		observation.NewErrorCollector(),
// 	)

// 	args := &LSIFDiagnosticsArgs{
// 		ConnectionArgs: ConnectionArgs{},
// 	}

// 	if _, err := resolver.Diagnostics(context.Background(), args); err != nil {
// 		t.Fatalf("unexpected error: %s", err)
// 	}

// 	if len(mockGitBlobResolver.DiagnosticsFunc.History()) != 1 {
// 		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockGitBlobResolver.DiagnosticsFunc.History()))
// 	}
// 	if val := mockGitBlobResolver.DiagnosticsFunc.History()[0].Arg1; val != DefaultDiagnosticsPageSize {
// 		t.Fatalf("unexpected limit. want=%d have=%d", DefaultDiagnosticsPageSize, val)
// 	}
// }

// func TestDiagnosticsDefaultIllegalLimit(t *testing.T) {
// 	mockGitBlobResolver := NewMockGitBlobResolver()
// 	mockAutoIndexingSvc := NewMockAutoIndexingService()
// 	mockUploadsService := NewMockUploadsService()
// 	mockPolicyService := NewMockPolicyService()

// 	resolver := NewGitBlobLSIFDataResolverQueryResolver(
// 		mockAutoIndexingSvc,
// 		mockUploadsService,
// 		mockPolicyService,
// 		mockGitBlobResolver,
// 		observation.NewErrorCollector(),
// 	)

// 	offset := int32(-1)
// 	args := &LSIFDiagnosticsArgs{
// 		ConnectionArgs: ConnectionArgs{First: &offset},
// 	}

// 	if _, err := resolver.Diagnostics(context.Background(), args); err != ErrIllegalLimit {
// 		t.Fatalf("unexpected error. want=%q have=%q", ErrIllegalLimit, err)
// 	}
// }
