package background

import (
	"crypto/sha256"

	"github.com/sourcegraph/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/lsifstore"
)

// processDocument canonicalizes and serializes the given document for persistence.
func processDocument(document *scip.Document, externalSymbolsByName map[string]*scip.SymbolInformation) lsifstore.ProcessedSCIPDocument {
	// Stash path here as canonicalization removes it
	path := document.RelativePath
	canonicalizeDocument(document, externalSymbolsByName)

	payload, err := proto.Marshal(document)
	if err != nil {
		return lsifstore.ProcessedSCIPDocument{
			DocumentPath: path,
			Err:          err,
		}
	}

	return lsifstore.ProcessedSCIPDocument{
		DocumentPath:   path,
		Hash:           hashPayload(payload),
		RawSCIPPayload: payload,
		Symbols:        types.ExtractSymbolIndexes(document),
	}
}

// hashPayload returns a sha256 checksum of the given payload.
func hashPayload(payload []byte) []byte {
	hash := sha256.New()
	_, _ = hash.Write(payload)
	return hash.Sum(nil)
}
