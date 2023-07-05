package context

import (
	"context"

	codenavtypes "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/symbols"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
)

type CodeNavService interface {
	GetClosestDumpsForBlob(ctx context.Context, repositoryID int, commit, path string, exactPath bool, indexer string) (_ []uploadsshared.Dump, err error)
	GetFullSCIPNameByDescriptor(ctx context.Context, uploadID []int, symbolNames []string) (names []*symbols.ExplodedSymbol, err error)
	NewGetDefinitionsBySymbolNames(ctx context.Context, args codenavtypes.RequestArgs, requestState codenavtypes.RequestState, symbolNames []string) (_ []shared.UploadLocation, err error)
}
