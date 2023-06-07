package context

import (
	"context"

	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type CodeNavService interface {
	GetClosestDumpsForBlob(ctx context.Context, repositoryID int, commit, path string, exactPath bool, indexer string) (_ []uploadsshared.Dump, err error)
	GetSCIPDocumentsBySymbolNames(
		ctx context.Context,
		uploads []uploadsshared.Dump,
		symbolNames []string,
		rangeMap map[string][]int32,
		repoName api.RepoName,
		repoID api.RepoID,
		commitID api.CommitID,
		path string,
	) (content []string, err error)
}
