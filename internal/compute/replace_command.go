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

func replace(content []byte, matchPattern MatchPattern, replacePattern string) (*Text, error) {
	var newContent []byte
	switch p := matchPattern.(type) {
	case *Regexp:
		newContent = p.Value.ReplaceAll(content, []byte(replacePattern))
	default:
		return nil, errors.Errorf("unsupported replacement operation for %T", p)
	}
	return &Text{Value: string(newContent), Kind: "replace-in-place"}, nil
}

func (c *Replace) Run(ctx context.Context, fm *result.FileMatch) (Result, error) {
	content, err := git.ReadFile(ctx, fm.Repo.Name, fm.CommitID, fm.Path, 0)
	if err != nil {
		return nil, err
	}
	return replace(content, c.MatchPattern, c.ReplacePattern)
}
