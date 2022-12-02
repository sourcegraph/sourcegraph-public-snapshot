package background

import (
	"context"

	otlog "github.com/opentracing/opentracing-go/log"

	codeinteltypes "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// writeSCIPData transactionally writes the given correlated SCIP data into the given store targeting
// the codeintel-db.
func writeSCIPData(
	ctx context.Context,
	lsifStore lsifstore.LsifStore,
	upload codeinteltypes.Upload,
	repo *types.Repo,
	isDefaultBranch bool,
	correlatedSCIPData lsifstore.ProcessedSCIPData,
	trace observation.TraceLogger,
) (err error) {
	tx, err := lsifStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.InsertMetadata(ctx, upload.ID, correlatedSCIPData.Metadata); err != nil {
		return err
	}

	symbolWriter, err := tx.NewSymbolWriter(ctx, upload.ID)
	if err != nil {
		return err
	}

	var numDocuments uint32
	for document := range correlatedSCIPData.Documents {
		documentLookupID, err := tx.InsertSCIPDocument(
			ctx,
			upload.ID,
			document.DocumentPath,
			document.Hash,
			document.RawSCIPPayload,
		)
		if err != nil {
			return err
		}

		if err := symbolWriter.WriteSCIPSymbols(ctx, documentLookupID, document.Symbols); err != nil {
			return err
		}
		numDocuments += 1
	}
	trace.Log(otlog.Uint32("numDocuments", numDocuments))

	count, err := symbolWriter.Flush(ctx)
	if err != nil {
		return err
	}
	trace.Log(otlog.Uint32("numSymbols", count))

	return nil
}
