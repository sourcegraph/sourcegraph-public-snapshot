package serialization

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

type Serializer interface {
	// MarshalDocumentData transforms document data into a string of bytes writable to disk.
	MarshalDocumentData(document types.DocumentData) ([]byte, error)

	// MarshalResultChunkData transforms result chunk data into a string of bytes writable to disk.
	MarshalResultChunkData(resultChunk types.ResultChunkData) ([]byte, error)

	// MarshalLocations transforms a slice of locations into a string of bytes writable to disk.
	MarshalLocations(locations []types.Location) ([]byte, error)

	// UnmarshalDocumentData is the inverse of MarshalDocumentData.
	UnmarshalDocumentData(data []byte) (types.DocumentData, error)

	// UnmarshalResultChunkData is the inverse of MarshalResultChunkData.
	UnmarshalResultChunkData(data []byte) (types.ResultChunkData, error)

	// UnmarshalLocations is the inverse of MarshalLocations.
	UnmarshalLocations(data []byte) ([]types.Location, error)
}
