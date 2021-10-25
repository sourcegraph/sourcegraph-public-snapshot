package compute

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type Replace struct {
	MatchPattern   MatchPattern
	ReplacePattern string
}

func (c *Replace) String() string {
	return fmt.Sprintf("Replace in place: (%s) -> (%s)", c.MatchPattern.String(), c.ReplacePattern)
}

func doReplaceInPlace(content []byte, command *Replace) (*Text, error) {
	var newContent []byte
	switch p := command.MatchPattern.(type) {
	case *Regexp:
		newContent = p.Value.ReplaceAll(content, []byte(command.ReplacePattern))
	default:
		return nil, errors.Errorf("unsupported replacement operation for %T", p)
	}
	return &Text{Value: string(newContent), Kind: "replace-in-place"}, nil
}

func ReplaceInPlaceFromFileMatch(ctx context.Context, fm *result.FileMatch, command *Replace) (*Text, error) {
	content, err := git.ReadFile(ctx, fm.Repo.Name, fm.CommitID, fm.Path, 0)
	if err != nil {
		return nil, err
	}
	return doReplaceInPlace(content, command)
}

func (c *Replace) Run(ctx context.Context, fm *result.FileMatch) (Result, error) {
	return ReplaceInPlaceFromFileMatch(ctx, fm, c)
}
