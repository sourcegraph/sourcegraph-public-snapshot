package context

import (
	"context"

	"github.com/sourcegraph/scip/bindings/go/scip"

	codenavtypes "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	codenavshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
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

	GetDefinitionBySymbolName(
		ctx context.Context,
		orderedMonikers []precise.QualifiedMonikerData,
		requestState codenavtypes.RequestState,
		args codenavtypes.RequestArgs,
	) (_ []shared.UploadLocation, err error)

	GetFullSCIPNameByDescriptor(ctx context.Context, uploadID []int, symbolNames []string) (names []*types.SCIPNames, err error)

	GetScipDefinitionsLocation(ctx context.Context, document *scip.Document, occ *scip.Occurrence, uploadID int, path string, limit, offset int) (_ []shared.Location, _ int, err error)

	GetUploadLocations(ctx context.Context, args codenavtypes.RequestArgs, requestState codenavtypes.RequestState, locations []shared.Location, includeFallbackLocations bool) ([]shared.UploadLocation, error)

	GetLocationByExplodedSymbol(ctx context.Context, symbolName string, uploadID int, scipFieldName string) (locations []shared.Location, err error)
}
