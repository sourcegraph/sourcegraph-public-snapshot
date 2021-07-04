package resolvers

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/semantic"
)

const slowSymbolRequestThreshold = 5 * time.Second

func (r *queryResolver) Symbol(ctx context.Context, scheme, identifier string) (_ *AdjustedSymbol, _ []int, err error) {
	ctx, traceLog, endObservation := observeResolver(ctx, &err, "Symbol", r.operations.symbol, slowSymbolRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", r.repositoryID),
			log.String("commit", r.commit),
			log.String("uploadIDs", strings.Join(r.uploadIDs(), ", ")),
			log.String("scheme", scheme),
			log.String("identifier", identifier),
		},
	})
	defer endObservation()

	if r.path != "" {
		// return nil, nil, errors.New("unable to get symbol for non-root")
		r.path = "" // TODO(sqs): hacky
	}

	adjustedUploads, err := r.adjustUploadPaths(ctx)
	if err != nil {
		return nil, nil, err
	}

	for i := range adjustedUploads {
		traceLog(log.Int("uploadID", adjustedUploads[i].Upload.ID))

		symbol, treePath, err := r.lsifStore.Symbol(
			ctx,
			adjustedUploads[i].Upload.ID,
			scheme,
			identifier,
		)
		if err != nil {
			return nil, nil, err
		}

		// TODO(sqs): handle case when multiple symbols have same moniker

		if symbol != nil {
			adjustedSymbol, err := r.adjustSymbol(ctx, adjustedUploads[i], *symbol)
			if err != nil {
				return nil, nil, err
			}
			return adjustedSymbol, treePath, nil
		}
	}

	return nil, nil, nil
}

// adjustSymbol translates a symbol (relative to the indexed commit) into an equivalent symbol in
// the requested commit.
func (r *queryResolver) adjustSymbol(ctx context.Context, adjustedUpload adjustedUpload, symbol semantic.SymbolData) (*AdjustedSymbol, error) {
	rn := lsifstore.Range{
		Start: lsifstore.Position{
			Line:      symbol.Location.StartLine,
			Character: symbol.Location.StartCharacter,
		},
		End: lsifstore.Position{
			Line:      symbol.Location.EndLine,
			Character: symbol.Location.EndCharacter,
		},
	}

	// Adjust path in symbol before reading it. This value is used in the adjustRange
	// call below, and is also reflected in the embedded symbol value in the return.
	symbol.Location.URI = adjustedUpload.Upload.Root + symbol.Location.URI

	adjustedCommit, adjustedRange, _, err := r.adjustRange(
		ctx,
		adjustedUpload.Upload.RepositoryID,
		adjustedUpload.Upload.Commit,
		symbol.Location.URI,
		rn,
	)
	if err != nil {
		return nil, err
	}

	return &AdjustedSymbol{
		SymbolData: symbol,
		AdjustedLocation: AdjustedLocation{
			Dump:           adjustedUpload.Upload,
			Path:           symbol.Location.URI,
			AdjustedCommit: adjustedCommit,
			AdjustedRange:  adjustedRange,
		},
		// TODO(sqs): children
	}, nil
}

// TODO(sqs): hack
func (r *queryResolver) TmpWithPath(path string) QueryResolver {
	tmp := *r
	tmp.path = path
	return &tmp
}

// uploadIDs returns a slice of this query's matched upload identifiers.
func (r *queryResolver) uploadIDs() []string {
	uploadIDs := make([]string, 0, len(r.uploads))
	for i := range r.uploads {
		uploadIDs = append(uploadIDs, strconv.Itoa(r.uploads[i].ID))
	}

	return uploadIDs
}
