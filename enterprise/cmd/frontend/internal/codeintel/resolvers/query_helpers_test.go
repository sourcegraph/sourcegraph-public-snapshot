package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
)

var (
	testRange1 = lsifstore.Range{Start: lsifstore.Position{Line: 11, Character: 21}, End: lsifstore.Position{Line: 31, Character: 41}}
	testRange2 = lsifstore.Range{Start: lsifstore.Position{Line: 12, Character: 22}, End: lsifstore.Position{Line: 32, Character: 42}}
	testRange3 = lsifstore.Range{Start: lsifstore.Position{Line: 13, Character: 23}, End: lsifstore.Position{Line: 33, Character: 43}}
	testRange4 = lsifstore.Range{Start: lsifstore.Position{Line: 14, Character: 24}, End: lsifstore.Position{Line: 34, Character: 44}}
	testRange5 = lsifstore.Range{Start: lsifstore.Position{Line: 15, Character: 25}, End: lsifstore.Position{Line: 35, Character: 45}}
)

func noopPositionAdjuster() PositionAdjuster {
	mockPositionAdjuster := NewMockPositionAdjuster()
	mockPositionAdjuster.AdjustPathFunc.SetDefaultHook(func(ctx context.Context, commit string, path string, _ bool) (string, bool, error) {
		return commit, true, nil
	})
	mockPositionAdjuster.AdjustPositionFunc.SetDefaultHook(func(ctx context.Context, commit string, path string, pos lsifstore.Position, _ bool) (string, lsifstore.Position, bool, error) {
		return commit, pos, true, nil
	})

	return mockPositionAdjuster
}
