package gob

import (
	"bytes"
	"encoding/gob"
	"io"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/klauspost/compress/gzip"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/serialization"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/types"
)

func init() {
	gob.Register(&types.DocumentData{})
	gob.Register(&types.ResultChunkData{})
	gob.Register(&types.Location{})
}

type gobSerializer struct {
	readers *sync.Pool
	writers *sync.Pool
}

var _ serialization.Serializer = &gobSerializer{}

func New() serialization.Serializer {
	return &gobSerializer{
		readers: &sync.Pool{New: func() interface{} { return new(gzip.Reader) }},
		writers: &sync.Pool{New: func() interface{} { return gzip.NewWriter(nil) }},
	}
}

// MarshalDocumentData transforms document data into a string of bytes writable to disk.
func (s *gobSerializer) MarshalDocumentData(document types.DocumentData) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(&document); err != nil {
		return nil, err
	}

	return s.compress(&buf)
}

// MarshalResultChunkData transforms result chunk data into a string of bytes writable to disk.
func (s *gobSerializer) MarshalResultChunkData(resultChunks types.ResultChunkData) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(&resultChunks); err != nil {
		return nil, err
	}

	return s.compress(&buf)
}

// MarshalLocations transforms a slice of locations into a string of bytes writable to disk.
func (s *gobSerializer) MarshalLocations(locations []types.Location) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(&locations); err != nil {
		return nil, err
	}

	return s.compress(&buf)
}

// UnmarshalDocumentData is the inverse of MarshalDocumentData.
func (s *gobSerializer) UnmarshalDocumentData(data []byte) (document types.DocumentData, err error) {
	err = s.withDecoder(data, func(decoder *gob.Decoder) error { return decoder.Decode(&document) })
	return document, err
}

// UnmarshalResultChunkData is the inverse of MarshalResultChunkData.
func (s *gobSerializer) UnmarshalResultChunkData(data []byte) (resultChunk types.ResultChunkData, err error) {
	err = s.withDecoder(data, func(decoder *gob.Decoder) error { return decoder.Decode(&resultChunk) })
	return resultChunk, err
}

// UnmarshalLocations is the inverse of MarshalLocations.
func (s *gobSerializer) UnmarshalLocations(data []byte) (locations []types.Location, err error) {
	err = s.withDecoder(data, func(decoder *gob.Decoder) error { return decoder.Decode(&locations) })
	return locations, err
}

//
//
//

// TODO
func (s *gobSerializer) compress(r io.Reader) ([]byte, error) {
	w := s.writers.Get().(*gzip.Writer)
	defer s.writers.Put(w)

	var buf bytes.Buffer
	w.Reset(&buf)

	if _, err := io.Copy(w, r); err != nil {
		w.Close()
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// TODO
func (s *gobSerializer) withDecoder(data []byte, f func(decoder *gob.Decoder) error) (err error) {
	r := s.readers.Get().(*gzip.Reader)
	defer s.readers.Put(r)

	if err := r.Reset(bytes.NewReader(data)); err != nil {
		return err
	}
	defer func() {
		if closeErr := r.Close(); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}
	}()

	return f(gob.NewDecoder(r))
}
