package gitcli

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

func (g *gitCLIBackend) RawDiff(ctx context.Context, base string, head string, typ git.GitDiffComparisonType, paths ...string) (io.ReadCloser, error) {
	baseOID, err := g.revParse(ctx, base)
	if err != nil {
		return nil, err
	}
	headOID, err := g.revParse(ctx, head)
	if err != nil {
		return nil, err
	}

	return g.NewCommand(ctx, WithArguments(buildRawDiffArgs(baseOID, headOID, typ, paths)...))
}

func buildRawDiffArgs(base, head api.CommitID, typ git.GitDiffComparisonType, paths []string) []string {
	var rangeType string
	switch typ {
	case git.GitDiffComparisonTypeIntersection:
		rangeType = "..."
	case git.GitDiffComparisonTypeOnlyInHead:
		rangeType = ".."
	}
	rangeSpec := string(base) + rangeType + string(head)

	return append([]string{
		"diff",
		"--find-renames",
		"--full-index",
		"--inter-hunk-context=3",
		"--no-prefix",
		rangeSpec,
		"--",
	}, paths...)
}
