package context

import (
	"context"

	"github.com/sourcegraph/scip/bindings/go/scip"

	codenavtypes "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav"
	codenavshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
)

type CodeNavService interface {
	GetClosestDumpsForBlob(ctx context.Context, repositoryID int, commit, path string, exactPath bool, indexer string) (_ []uploadsshared.Dump, err error)
	// GetSCIPDocumentsBySymbolNames(
	// 	ctx context.Context,
	// 	uploads []uploadsshared.Dump,
	// 	symbolNames []string,
	// 	rangeMap map[string][]int32,
	// 	repoName api.RepoName,
	// 	repoID api.RepoID,
	// 	commitID api.CommitID,
	// 	path string,
	// ) (content []string, err error)
	GetSCIPDocumentsBySymbolNames(ctx context.Context, uploads int, symbolNames []string) (documents []*scip.Document, err error)
	GetDefinitions(ctx context.Context, args codenavtypes.RequestArgs, requestState codenavtypes.RequestState) (_ []codenavshared.UploadLocation, err error)
}
