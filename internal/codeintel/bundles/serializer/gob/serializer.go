package gob

import (
	"bytes"
	"encoding/gob"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/serializer"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

func init() {
	gob.Register(&types.DocumentData{})
	gob.Register(&types.ResultChunkData{})
}

type gobSerializer struct{}

var _ serializer.Serializer = &gobSerializer{}

func New() serializer.Serializer {
	return &gobSerializer{}
}

// MarshalDocumentData transforms document data into a string of bytes writable to disk.
func (*gobSerializer) MarshalDocumentData(d types.DocumentData) ([]byte, error) {
	buf := bytes.Buffer{}
	err := gob.NewEncoder(&buf).Encode(&d)
	return buf.Bytes(), err
}

// MarshalResultChunkData transforms result chunk data into a string of bytes writable to disk.
func (*gobSerializer) MarshalResultChunkData(rc types.ResultChunkData) ([]byte, error) {
	buf := bytes.Buffer{}
	err := gob.NewEncoder(&buf).Encode(&rc)
	return buf.Bytes(), err
}

// UnmarshalDocumentData is the inverse of MarshalDocumentData.
func (*gobSerializer) UnmarshalDocumentData(data []byte) (types.DocumentData, error) {
	var d types.DocumentData
	err := gob.NewDecoder(bytes.NewReader(data)).Decode(&d)
	return d, err
}

// UnmarshalResultChunkData is the inverse of MarshalResultChunkData.
func (*gobSerializer) UnmarshalResultChunkData(data []byte) (types.ResultChunkData, error) {
	var rc types.ResultChunkData
	err := gob.NewDecoder(bytes.NewReader(data)).Decode(&rc)
	return rc, err
}
