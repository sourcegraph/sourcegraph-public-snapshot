package serialization

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

type Serializer interface {
	// MarshalDocumentData transforms document data into a string of bytes writable to disk.
	MarshalDocumentData(d types.DocumentData) ([]byte, error)

	// MarshalResultChunkData transforms result chunk data into a string of bytes writable to disk.
	MarshalResultChunkData(rc types.ResultChunkData) ([]byte, error)

	// UnmarshalDocumentData is the inverse of MarshalDocumentData.
	UnmarshalDocumentData(data []byte) (types.DocumentData, error)

	// UnmarshalResultChunkData is the inverse of MarshalResultChunkData.
	UnmarshalResultChunkData(data []byte) (types.ResultChunkData, error)
}
