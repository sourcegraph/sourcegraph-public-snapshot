package gob

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

func init() {
	gob.Register(&types.DocumentData{})
	gob.Register(&types.ResultChunkData{})
	gob.Register(&types.Location{})
}

type gobSerializer struct{}

var _ serialization.Serializer = &gobSerializer{}

func New() serialization.Serializer {
	return &gobSerializer{}
}

// MarshalDocumentData transforms document data into a string of bytes writable to disk.
func (*gobSerializer) MarshalDocumentData(document types.DocumentData) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(&document); err != nil {
		return nil, err
	}

	return compress(&buf)
}

// MarshalResultChunkData transforms result chunk data into a string of bytes writable to disk.
func (*gobSerializer) MarshalResultChunkData(resultChunks types.ResultChunkData) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(&resultChunks); err != nil {
		return nil, err
	}

	return compress(&buf)
}

// MarshalLocations transforms a slice of locations into a string of bytes writable to disk.
func (*gobSerializer) MarshalLocations(locations []types.Location) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(&locations); err != nil {
		return nil, err
	}

	return compress(&buf)
}

// UnmarshalDocumentData is the inverse of MarshalDocumentData.
func (*gobSerializer) UnmarshalDocumentData(data []byte) (document types.DocumentData, err error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return types.DocumentData{}, err
	}

	err = gob.NewDecoder(r).Decode(&document)
	return document, err
}

// UnmarshalResultChunkData is the inverse of MarshalResultChunkData.
func (*gobSerializer) UnmarshalResultChunkData(data []byte) (resultChunk types.ResultChunkData, err error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return types.ResultChunkData{}, err
	}

	err = gob.NewDecoder(r).Decode(&resultChunk)
	return resultChunk, err
}

// UnmarshalLocations is the inverse of MarshalLocations.
func (*gobSerializer) UnmarshalLocations(data []byte) (locations []types.Location, err error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	err = gob.NewDecoder(r).Decode(&locations)
	return locations, err
}

// compress gzips the bytes in the given reader.
func compress(r io.Reader) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(gzipWriter, r); err != nil {
		return nil, err
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
