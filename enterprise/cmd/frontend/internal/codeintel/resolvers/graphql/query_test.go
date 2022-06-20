package graphql

import (
	"context"
	"encoding/base64"
	"testing"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	resolvermocks "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers/mocks"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestRanges(t *testing.T) {
	db := database.NewDB(nil)

	mockQueryResolver := resolvermocks.NewMockQueryResolver()
	mockResolver := resolvermocks.NewMockResolver()
	resolver := NewQueryResolver(nil, mockQueryResolver, mockResolver, NewCachedLocationResolver(db), nil)

	args := &gql.LSIFRangesArgs{StartLine: 10, EndLine: 20}
	if _, err := resolver.Ranges(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockQueryResolver.RangesFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockQueryResolver.RangesFunc.History()))
	}
	if val := mockQueryResolver.RangesFunc.History()[0].Arg1; val != 10 {
		t.Fatalf("unexpected start line. want=%d have=%d", 10, val)
	}
	if val := mockQueryResolver.RangesFunc.History()[0].Arg2; val != 20 {
		t.Fatalf("unexpected end line. want=%d have=%d", 20, val)
	}
}

func TestDefinitions(t *testing.T) {
	db := database.NewDB(nil)

	mockQueryResolver := resolvermocks.NewMockQueryResolver()
	mockResolver := resolvermocks.NewMockResolver()
	resolver := NewQueryResolver(nil, mockQueryResolver, mockResolver, NewCachedLocationResolver(db), nil)

	args := &gql.LSIFQueryPositionArgs{Line: 10, Character: 15}
	if _, err := resolver.Definitions(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockQueryResolver.DefinitionsFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockQueryResolver.DefinitionsFunc.History()))
	}
	if val := mockQueryResolver.DefinitionsFunc.History()[0].Arg1; val != 10 {
		t.Fatalf("unexpected line. want=%d have=%d", 10, val)
	}
	if val := mockQueryResolver.DefinitionsFunc.History()[0].Arg2; val != 15 {
		t.Fatalf("unexpected character. want=%d have=%d", 15, val)
	}
}

func TestReferences(t *testing.T) {
	db := database.NewDB(nil)

	mockQueryResolver := resolvermocks.NewMockQueryResolver()
	mockResolver := resolvermocks.NewMockResolver()
	resolver := NewQueryResolver(nil, mockQueryResolver, mockResolver, NewCachedLocationResolver(db), nil)

	offset := int32(25)
	cursor := base64.StdEncoding.EncodeToString([]byte("test-cursor"))

	args := &gql.LSIFPagedQueryPositionArgs{
		LSIFQueryPositionArgs: gql.LSIFQueryPositionArgs{
			Line:      10,
			Character: 15,
		},
		ConnectionArgs: graphqlutil.ConnectionArgs{First: &offset},
		After:          &cursor,
	}

	if _, err := resolver.References(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockQueryResolver.ReferencesFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockQueryResolver.ReferencesFunc.History()))
	}
	if val := mockQueryResolver.ReferencesFunc.History()[0].Arg1; val != 10 {
		t.Fatalf("unexpected line. want=%d have=%d", 10, val)
	}
	if val := mockQueryResolver.ReferencesFunc.History()[0].Arg2; val != 15 {
		t.Fatalf("unexpected character. want=%d have=%d", 15, val)
	}
	if val := mockQueryResolver.ReferencesFunc.History()[0].Arg3; val != 25 {
		t.Fatalf("unexpected character. want=%d have=%d", 25, val)
	}
	if val := mockQueryResolver.ReferencesFunc.History()[0].Arg4; val != "test-cursor" {
		t.Fatalf("unexpected character. want=%s have=%s", "test-cursor", val)
	}
}

func TestReferencesDefaultLimit(t *testing.T) {
	db := database.NewDB(nil)

	mockQueryResolver := resolvermocks.NewMockQueryResolver()
	mockResolver := resolvermocks.NewMockResolver()
	resolver := NewQueryResolver(nil, mockQueryResolver, mockResolver, NewCachedLocationResolver(db), nil)

	args := &gql.LSIFPagedQueryPositionArgs{
		LSIFQueryPositionArgs: gql.LSIFQueryPositionArgs{
			Line:      10,
			Character: 15,
		},
		ConnectionArgs: graphqlutil.ConnectionArgs{},
	}

	if _, err := resolver.References(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockQueryResolver.ReferencesFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockQueryResolver.DiagnosticsFunc.History()))
	}
	if val := mockQueryResolver.ReferencesFunc.History()[0].Arg3; val != DefaultReferencesPageSize {
		t.Fatalf("unexpected limit. want=%d have=%d", DefaultReferencesPageSize, val)
	}
}

func TestReferencesDefaultIllegalLimit(t *testing.T) {
	db := database.NewDB(nil)

	mockQueryResolver := resolvermocks.NewMockQueryResolver()
	mockResolver := resolvermocks.NewMockResolver()
	resolver := NewQueryResolver(nil, mockQueryResolver, mockResolver, NewCachedLocationResolver(db), observation.NewErrorCollector())

	offset := int32(-1)
	args := &gql.LSIFPagedQueryPositionArgs{
		LSIFQueryPositionArgs: gql.LSIFQueryPositionArgs{
			Line:      10,
			Character: 15,
		},
		ConnectionArgs: graphqlutil.ConnectionArgs{First: &offset},
	}

	if _, err := resolver.References(context.Background(), args); err != ErrIllegalLimit {
		t.Fatalf("unexpected error. want=%q have=%q", ErrIllegalLimit, err)
	}
}

func TestHover(t *testing.T) {
	db := database.NewDB(nil)

	mockQueryResolver := resolvermocks.NewMockQueryResolver()
	mockQueryResolver.HoverFunc.SetDefaultReturn("text", lsifstore.Range{}, true, nil)
	mockResolver := resolvermocks.NewMockResolver()
	resolver := NewQueryResolver(nil, mockQueryResolver, mockResolver, NewCachedLocationResolver(db), nil)

	args := &gql.LSIFQueryPositionArgs{Line: 10, Character: 15}
	if _, err := resolver.Hover(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockQueryResolver.HoverFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockQueryResolver.HoverFunc.History()))
	}
	if val := mockQueryResolver.HoverFunc.History()[0].Arg1; val != 10 {
		t.Fatalf("unexpected line. want=%d have=%d", 10, val)
	}
	if val := mockQueryResolver.HoverFunc.History()[0].Arg2; val != 15 {
		t.Fatalf("unexpected character. want=%d have=%d", 15, val)
	}
}

func TestDiagnostics(t *testing.T) {
	db := database.NewDB(nil)

	mockQueryResolver := resolvermocks.NewMockQueryResolver()
	mockResolver := resolvermocks.NewMockResolver()
	resolver := NewQueryResolver(nil, mockQueryResolver, mockResolver, NewCachedLocationResolver(db), nil)

	offset := int32(25)
	args := &gql.LSIFDiagnosticsArgs{
		ConnectionArgs: graphqlutil.ConnectionArgs{First: &offset},
	}

	if _, err := resolver.Diagnostics(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockQueryResolver.DiagnosticsFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockQueryResolver.DiagnosticsFunc.History()))
	}
	if val := mockQueryResolver.DiagnosticsFunc.History()[0].Arg1; val != 25 {
		t.Fatalf("unexpected limit. want=%d have=%d", 25, val)
	}
}

func TestDiagnosticsDefaultLimit(t *testing.T) {
	db := database.NewDB(nil)

	mockQueryResolver := resolvermocks.NewMockQueryResolver()
	mockResolver := resolvermocks.NewMockResolver()
	resolver := NewQueryResolver(nil, mockQueryResolver, mockResolver, NewCachedLocationResolver(db), nil)

	args := &gql.LSIFDiagnosticsArgs{
		ConnectionArgs: graphqlutil.ConnectionArgs{},
	}

	if _, err := resolver.Diagnostics(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockQueryResolver.DiagnosticsFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockQueryResolver.DiagnosticsFunc.History()))
	}
	if val := mockQueryResolver.DiagnosticsFunc.History()[0].Arg1; val != DefaultDiagnosticsPageSize {
		t.Fatalf("unexpected limit. want=%d have=%d", DefaultDiagnosticsPageSize, val)
	}
}

func TestDiagnosticsDefaultIllegalLimit(t *testing.T) {
	db := database.NewDB(nil)

	mockQueryResolver := resolvermocks.NewMockQueryResolver()
	mockResolver := resolvermocks.NewMockResolver()
	resolver := NewQueryResolver(nil, mockQueryResolver, mockResolver, NewCachedLocationResolver(db), observation.NewErrorCollector())

	offset := int32(-1)
	args := &gql.LSIFDiagnosticsArgs{
		ConnectionArgs: graphqlutil.ConnectionArgs{First: &offset},
	}

	if _, err := resolver.Diagnostics(context.Background(), args); err != ErrIllegalLimit {
		t.Fatalf("unexpected error. want=%q have=%q", ErrIllegalLimit, err)
	}
}
