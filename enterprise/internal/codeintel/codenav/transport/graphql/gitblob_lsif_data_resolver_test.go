package graphql

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestRanges(t *testing.T) {
	mockCodeNavService := NewMockCodeNavService()
	mockAutoIndexingSvc := NewMockAutoIndexingService()
	mockUploadsService := NewMockUploadsService()
	mockPolicyService := NewMockPolicyService()

	mockRequestState := codenav.RequestState{
		RepositoryID: 1,
		Commit:       "deadbeef1",
		Path:         "/src/main",
	}
	mockOperations := newOperations(&observation.TestContext)

	resolver := NewGitBlobLSIFDataResolver(
		mockCodeNavService,
		mockAutoIndexingSvc,
		mockUploadsService,
		mockPolicyService,
		mockRequestState,
		observation.NewErrorCollector(),
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
	mockAutoIndexingSvc := NewMockAutoIndexingService()
	mockUploadsService := NewMockUploadsService()
	mockPolicyService := NewMockPolicyService()
	mockCodeNavService := NewMockCodeNavService()
	mockRequestState := codenav.RequestState{
		RepositoryID: 1,
		Commit:       "deadbeef1",
		Path:         "/src/main",
	}
	mockOperations := newOperations(&observation.TestContext)

	resolver := NewGitBlobLSIFDataResolver(
		mockCodeNavService,
		mockAutoIndexingSvc,
		mockUploadsService,
		mockPolicyService,
		mockRequestState,
		observation.NewErrorCollector(),
		mockOperations,
	)

	args := &resolverstubs.LSIFQueryPositionArgs{Line: 10, Character: 15}
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
	mockAutoIndexingSvc := NewMockAutoIndexingService()
	mockUploadsService := NewMockUploadsService()
	mockPolicyService := NewMockPolicyService()
	mockCodeNavService := NewMockCodeNavService()
	mockRequestState := codenav.RequestState{
		RepositoryID: 1,
		Commit:       "deadbeef1",
		Path:         "/src/main",
	}
	mockOperations := newOperations(&observation.TestContext)

	resolver := NewGitBlobLSIFDataResolver(
		mockCodeNavService,
		mockAutoIndexingSvc,
		mockUploadsService,
		mockPolicyService,
		mockRequestState,
		observation.NewErrorCollector(),
		mockOperations,
	)

	offset := int32(25)
	mockRefCursor := shared.ReferencesCursor{Phase: "local"}
	encodedCursor := encodeReferencesCursor(mockRefCursor)
	mockCursor := base64.StdEncoding.EncodeToString([]byte(encodedCursor))

	args := &resolverstubs.LSIFPagedQueryPositionArgs{
		LSIFQueryPositionArgs: resolverstubs.LSIFQueryPositionArgs{
			Line:      10,
			Character: 15,
		},
		ConnectionArgs: resolverstubs.ConnectionArgs{First: &offset},
		After:          &mockCursor,
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
	if val := mockCodeNavService.GetReferencesFunc.History()[0].Arg1; val.RawCursor != encodedCursor {
		t.Fatalf("unexpected character. want=%v have=%v", "test-cursor", val)
	}
}

func TestReferencesDefaultLimit(t *testing.T) {
	mockAutoIndexingSvc := NewMockAutoIndexingService()
	mockUploadsService := NewMockUploadsService()
	mockPolicyService := NewMockPolicyService()
	mockCodeNavService := NewMockCodeNavService()
	mockRequestState := codenav.RequestState{
		RepositoryID: 1,
		Commit:       "deadbeef1",
		Path:         "/src/main",
	}
	mockOperations := newOperations(&observation.TestContext)

	resolver := NewGitBlobLSIFDataResolver(
		mockCodeNavService,
		mockAutoIndexingSvc,
		mockUploadsService,
		mockPolicyService,
		mockRequestState,
		observation.NewErrorCollector(),
		mockOperations,
	)

	args := &resolverstubs.LSIFPagedQueryPositionArgs{
		LSIFQueryPositionArgs: resolverstubs.LSIFQueryPositionArgs{
			Line:      10,
			Character: 15,
		},
		ConnectionArgs: resolverstubs.ConnectionArgs{},
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
	mockAutoIndexingSvc := NewMockAutoIndexingService()
	mockUploadsService := NewMockUploadsService()
	mockPolicyService := NewMockPolicyService()
	mockCodeNavService := NewMockCodeNavService()
	mockRequestState := codenav.RequestState{
		RepositoryID: 1,
		Commit:       "deadbeef1",
		Path:         "/src/main",
	}
	mockOperations := newOperations(&observation.TestContext)

	resolver := NewGitBlobLSIFDataResolver(
		mockCodeNavService,
		mockAutoIndexingSvc,
		mockUploadsService,
		mockPolicyService,
		mockRequestState,
		observation.NewErrorCollector(),
		mockOperations,
	)

	offset := int32(-1)
	args := &resolverstubs.LSIFPagedQueryPositionArgs{
		LSIFQueryPositionArgs: resolverstubs.LSIFQueryPositionArgs{
			Line:      10,
			Character: 15,
		},
		ConnectionArgs: resolverstubs.ConnectionArgs{First: &offset},
	}

	if _, err := resolver.References(context.Background(), args); err != ErrIllegalLimit {
		t.Fatalf("unexpected error. want=%q have=%q", ErrIllegalLimit, err)
	}
}

func TestHover(t *testing.T) {
	mockAutoIndexingSvc := NewMockAutoIndexingService()
	mockUploadsService := NewMockUploadsService()
	mockPolicyService := NewMockPolicyService()
	mockCodeNavService := NewMockCodeNavService()
	mockRequestState := codenav.RequestState{
		RepositoryID: 1,
		Commit:       "deadbeef1",
		Path:         "/src/main",
	}
	mockOperations := newOperations(&observation.TestContext)

	resolver := NewGitBlobLSIFDataResolver(
		mockCodeNavService,
		mockAutoIndexingSvc,
		mockUploadsService,
		mockPolicyService,
		mockRequestState,
		observation.NewErrorCollector(),
		mockOperations,
	)

	mockCodeNavService.GetHoverFunc.SetDefaultReturn("text", types.Range{}, true, nil)
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
	mockAutoIndexingSvc := NewMockAutoIndexingService()
	mockUploadsService := NewMockUploadsService()
	mockPolicyService := NewMockPolicyService()
	mockCodeNavService := NewMockCodeNavService()
	mockRequestState := codenav.RequestState{
		RepositoryID: 1,
		Commit:       "deadbeef1",
		Path:         "/src/main",
	}
	mockOperations := newOperations(&observation.TestContext)

	resolver := NewGitBlobLSIFDataResolver(
		mockCodeNavService,
		mockAutoIndexingSvc,
		mockUploadsService,
		mockPolicyService,
		mockRequestState,
		observation.NewErrorCollector(),
		mockOperations,
	)

	offset := int32(25)
	args := &resolverstubs.LSIFDiagnosticsArgs{
		ConnectionArgs: resolverstubs.ConnectionArgs{First: &offset},
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
	mockAutoIndexingSvc := NewMockAutoIndexingService()
	mockUploadsService := NewMockUploadsService()
	mockPolicyService := NewMockPolicyService()
	mockCodeNavService := NewMockCodeNavService()
	mockRequestState := codenav.RequestState{
		RepositoryID: 1,
		Commit:       "deadbeef1",
		Path:         "/src/main",
	}
	mockOperations := newOperations(&observation.TestContext)

	resolver := NewGitBlobLSIFDataResolver(
		mockCodeNavService,
		mockAutoIndexingSvc,
		mockUploadsService,
		mockPolicyService,
		mockRequestState,
		observation.NewErrorCollector(),
		mockOperations,
	)

	args := &resolverstubs.LSIFDiagnosticsArgs{
		ConnectionArgs: resolverstubs.ConnectionArgs{},
	}

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
	mockAutoIndexingSvc := NewMockAutoIndexingService()
	mockUploadsService := NewMockUploadsService()
	mockPolicyService := NewMockPolicyService()
	mockCodeNavService := NewMockCodeNavService()
	mockRequestState := codenav.RequestState{
		RepositoryID: 1,
		Commit:       "deadbeef1",
		Path:         "/src/main",
	}
	mockOperations := newOperations(&observation.TestContext)

	resolver := NewGitBlobLSIFDataResolver(
		mockCodeNavService,
		mockAutoIndexingSvc,
		mockUploadsService,
		mockPolicyService,
		mockRequestState,
		observation.NewErrorCollector(),
		mockOperations,
	)

	offset := int32(-1)
	args := &resolverstubs.LSIFDiagnosticsArgs{
		ConnectionArgs: resolverstubs.ConnectionArgs{First: &offset},
	}

	if _, err := resolver.Diagnostics(context.Background(), args); err != ErrIllegalLimit {
		t.Fatalf("unexpected error. want=%q have=%q", ErrIllegalLimit, err)
	}
}
