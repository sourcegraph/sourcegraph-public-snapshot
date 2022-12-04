package scip

import (
	"context"
	"io"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/conversion"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// TODO - document
func ConvertLSIF(ctx context.Context, uploadID int, r io.Reader, root, indexerName string) (*scip.Index, error) {
	chans, err := conversion.Correlate(ctx, r, root, nil)
	if err != nil {
		return nil, err
	}

	resultChunks := map[int]precise.ResultChunkData{}
	for resultChunk := range chans.ResultChunks {
		resultChunks[resultChunk.Index] = resultChunk.ResultChunk
	}

	definitionMatcher := func(
		targetPath string,
		targetRangeID precise.ID,
		definitionResultID precise.ID,
	) bool {
		definitionResultChunk, ok := resultChunks[precise.HashKey(precise.ID(definitionResultID), chans.Meta.NumResultChunks)]
		if !ok {
			return false
		}

		for _, pair := range definitionResultChunk.DocumentIDRangeIDs[precise.ID(definitionResultID)] {
			if targetPath == definitionResultChunk.DocumentPaths[pair.DocumentID] && pair.RangeID == precise.ID(targetRangeID) {
				return true
			}
		}

		return false
	}

	var documents []*scip.Document
	for document := range chans.Documents {
		documents = append(documents, ConvertLSIFDocument(
			uploadID,
			definitionMatcher,
			indexerName,
			document.Path,
			document.Document,
		))
	}

	metadata := &scip.Metadata{
		Version:              0,   // TODO
		ToolInfo:             nil, // TODO
		ProjectRoot:          root,
		TextDocumentEncoding: scip.TextEncoding_UnspecifiedTextEncoding,
	}

	return &scip.Index{
		Metadata:  metadata,
		Documents: documents,
	}, nil
}
