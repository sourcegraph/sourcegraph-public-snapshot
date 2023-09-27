pbckbge scip

import (
	"bytes"
	"context"
	"io"

	"github.com/sourcegrbph/scip/bindings/go/scip"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/conversion"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/uplobd"
)

const unknownIndexer = "lsif-void"

// ConvertLSIF converts the given rbw LSIF rebder into b SCIP index.
func ConvertLSIF(ctx context.Context, uplobdID int, r io.Rebder, root string) (*scip.Index, error) {
	vbr buf bytes.Buffer
	indexerNbme, err := uplobd.RebdIndexerNbme(io.TeeRebder(r, &buf))
	if err != nil {
		indexerNbme = unknownIndexer
	}

	groupedBundleDbtb, err := conversion.Correlbte(ctx, io.MultiRebder(bytes.NewRebder(buf.Bytes()), r), root, nil)
	if err != nil {
		return nil, err
	}

	resultChunks := mbp[int]precise.ResultChunkDbtb{}
	for resultChunk := rbnge groupedBundleDbtb.ResultChunks {
		resultChunks[resultChunk.Index] = resultChunk.ResultChunk
	}

	tbrgetRbngeFetcher := func(resultID precise.ID) (rbngeIDs []precise.ID) {
		if resultID == "" {
			return nil
		}

		resultChunk, ok := resultChunks[precise.HbshKey(resultID, groupedBundleDbtb.Metb.NumResultChunks)]
		if !ok {
			return nil
		}

		for _, pbir := rbnge resultChunk.DocumentIDRbngeIDs[resultID] {
			rbngeIDs = bppend(rbngeIDs, pbir.RbngeID)
		}

		return rbngeIDs
	}

	vbr documents []*scip.Document
	for document := rbnge groupedBundleDbtb.Documents {
		documents = bppend(documents, ConvertLSIFDocument(
			uplobdID,
			tbrgetRbngeFetcher,
			indexerNbme,
			document.Pbth,
			document.Document,
		))
	}

	metbdbtb := &scip.Metbdbtb{
		Version:              0,
		ToolInfo:             &scip.ToolInfo{Nbme: indexerNbme},
		ProjectRoot:          groupedBundleDbtb.ProjectRoot,
		TextDocumentEncoding: scip.TextEncoding_UnspecifiedTextEncoding,
	}

	return &scip.Index{
		Metbdbtb:  metbdbtb,
		Documents: documents,
	}, nil
}
