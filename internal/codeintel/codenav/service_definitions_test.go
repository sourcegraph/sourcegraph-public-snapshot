package codenav

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
)

func mockedGitTreeTranslator() GitTreeTranslator {
	mockPositionAdjuster := NewMockGitTreeTranslator()
	mockPositionAdjuster.GetTargetCommitPositionFromSourcePositionFunc.SetDefaultHook(func(ctx context.Context, commit string, path string, pos shared.Position, _ bool) (shared.Position, bool, error) {
		return pos, true, nil
	})
	mockPositionAdjuster.GetTargetCommitRangeFromSourceRangeFunc.SetDefaultHook(func(ctx context.Context, commit string, path string, rx shared.Range, _ bool) (shared.Range, bool, error) {
		return rx, true, nil
	})

	return mockPositionAdjuster
}
