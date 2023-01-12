package scip

import (
	"bytes"
	"context"
	"io"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/conversion"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/upload"
)

const unknownIndexer = "lsif-void"

// ConvertLSIF converts the given raw LSIF reader into a SCIP index.
func ConvertLSIF(ctx context.Context, uploadID int, r io.Reader, root string) (*scip.Index, error) {
	var buf bytes.Buffer
	indexerName, err := upload.ReadIndexerName(io.TeeReader(r, &buf))
	if err != nil {
		indexerName = unknownIndexer
	}

	groupedBundleData, err := conversion.Correlate(ctx, io.MultiReader(bytes.NewReader(buf.Bytes()), r), root, nil)
	if err != nil {
		return nil, err
	}

	resultChunks := map[int]precise.ResultChunkData{}
	for resultChunk := range groupedBundleData.ResultChunks {
		resultChunks[resultChunk.Index] = resultChunk.ResultChunk
	}

	targetRangeFetcher := func(resultID precise.ID) (rangeIDs []precise.ID) {
		if resultID == "" {
			return nil
		}

		resultChunk, ok := resultChunks[precise.HashKey(resultID, groupedBundleData.Meta.NumResultChunks)]
		if !ok {
			return nil
		}

		for _, pair := range resultChunk.DocumentIDRangeIDs[resultID] {
			rangeIDs = append(rangeIDs, pair.RangeID)
		}

		return rangeIDs
	}

	var documents []*scip.Document
	for document := range groupedBundleData.Documents {
		documents = append(documents, ConvertLSIFDocument(
			uploadID,
			targetRangeFetcher,
			indexerName,
			document.Path,
			document.Document,
		))
	}

	metadata := &scip.Metadata{
		Version:              0,
		ToolInfo:             &scip.ToolInfo{Name: indexerName},
		ProjectRoot:          groupedBundleData.ProjectRoot,
		TextDocumentEncoding: scip.TextEncoding_UnspecifiedTextEncoding,
	}

	return &scip.Index{
		Metadata:  metadata,
		Documents: documents,
	}, nil
}
