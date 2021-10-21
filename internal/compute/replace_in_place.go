package compute

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func doReplaceInPlace(content []byte, op *ReplaceInPlace) (*Text, error) {
	var newContent []byte
	switch p := op.MatchPattern.(type) {
	case *Regexp:
		newContent = p.Value.ReplaceAll(content, []byte(op.ReplacePattern))
	default:
		return nil, errors.Errorf("unsupported replacement operation for %T", p)
	}
	return &Text{Value: string(newContent), Kind: "replace-in-place"}, nil
}

func ReplaceInPlaceFromFileMatch(ctx context.Context, fm *result.FileMatch, op *ReplaceInPlace) (*Text, error) {
	content, err := git.ReadFile(ctx, fm.Repo.Name, fm.CommitID, fm.Path, 0)
	if err != nil {
		return nil, err
	}
	return doReplaceInPlace(content, op)
}
