package lsifstore

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func init() {
	gob.Register(&precise.DocumentData{})
	gob.Register(&precise.ResultChunkData{})
	gob.Register(&precise.LocationData{})
}

type Serializer struct {
	readers sync.Pool
	writers sync.Pool
}

func NewSerializer() *Serializer {
	return &Serializer{
		readers: sync.Pool{New: func() any { return new(gzip.Reader) }},
		writers: sync.Pool{New: func() any { return gzip.NewWriter(nil) }},
	}
}

// decode decompresses gob-decodes the given data and sets the given pointer. If the given data
// is empty, the pointer will not be assigned.
func (s *Serializer) decode(data []byte, target any) (err error) {
	if len(data) == 0 {
		return nil
	}

	r := s.readers.Get().(*gzip.Reader)
	defer s.readers.Put(r)

	if err := r.Reset(bytes.NewReader(data)); err != nil {
		return err
	}
	defer func() {
		if closeErr := r.Close(); closeErr != nil {
			err = errors.Append(err, closeErr)
		}
	}()

	return gob.NewDecoder(r).Decode(target)
}

// UnmarshalLegacyDocumentData unmarshals a legacy-formatted document (the value in the `data` column).
func (s *Serializer) UnmarshalLegacyDocumentData(data []byte) (document precise.DocumentData, err error) {
	err = s.decode(data, &document)
	return document, err
}

// UnmarshalDocumentData is the inverse of MarshalDocumentData.
func (s *Serializer) UnmarshalDocumentData(data MarshalledDocumentData) (document precise.DocumentData, err error) {
	if err := s.decode(data.Ranges, &document.Ranges); err != nil {
		return precise.DocumentData{}, err
	}
	if err := s.decode(data.HoverResults, &document.HoverResults); err != nil {
		return precise.DocumentData{}, err
	}
	if err := s.decode(data.Monikers, &document.Monikers); err != nil {
		return precise.DocumentData{}, err
	}
	if err := s.decode(data.PackageInformation, &document.PackageInformation); err != nil {
		return precise.DocumentData{}, err
	}
	if err := s.decode(data.Diagnostics, &document.Diagnostics); err != nil {
		return precise.DocumentData{}, err
	}

	return document, nil
}

// UnmarshalResultChunkData is the inverse of MarshalResultChunkData.
func (s *Serializer) UnmarshalResultChunkData(data []byte) (resultChunk precise.ResultChunkData, err error) {
	err = s.decode(data, &resultChunk)
	return resultChunk, err
}

// UnmarshalLocations is the inverse of MarshalLocations.
func (s *Serializer) UnmarshalLocations(data []byte) (locations []precise.LocationData, err error) {
	err = s.decode(data, &locations)
	return locations, err
}
