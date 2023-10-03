package codenav

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
)

func mockedGitTreeTranslator() GitTreeTranslator {
	mockPositionAdjuster := NewMockGitTreeTranslator()
	mockPositionAdjuster.GetTargetCommitPathFromSourcePathFunc.SetDefaultHook(func(ctx context.Context, commit string, path string, _ bool) (string, bool, error) {
		return commit, true, nil
	})
	mockPositionAdjuster.GetTargetCommitPositionFromSourcePositionFunc.SetDefaultHook(func(ctx context.Context, commit string, pos shared.Position, _ bool) (string, shared.Position, bool, error) {
		return commit, pos, true, nil
	})
	mockPositionAdjuster.GetTargetCommitRangeFromSourceRangeFunc.SetDefaultHook(func(ctx context.Context, commit string, path string, rx shared.Range, _ bool) (string, shared.Range, bool, error) {
		return commit, rx, true, nil
	})

	return mockPositionAdjuster
}
