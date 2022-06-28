package lsifstore

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"io"
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

type MarshalledDocumentData struct {
	Ranges             []byte
	HoverResults       []byte
	Monikers           []byte
	PackageInformation []byte
	Diagnostics        []byte
}

// MarshalDocumentData transforms the fields of the given document data payload into a set of
// string of bytes writable to disk.
func (s *Serializer) MarshalDocumentData(document precise.DocumentData) (data MarshalledDocumentData, err error) {
	if data.Ranges, err = s.encode(&document.Ranges); err != nil {
		return MarshalledDocumentData{}, err
	}
	if data.HoverResults, err = s.encode(&document.HoverResults); err != nil {
		return MarshalledDocumentData{}, err
	}
	if data.Monikers, err = s.encode(&document.Monikers); err != nil {
		return MarshalledDocumentData{}, err
	}
	if data.PackageInformation, err = s.encode(&document.PackageInformation); err != nil {
		return MarshalledDocumentData{}, err
	}
	if data.Diagnostics, err = s.encode(&document.Diagnostics); err != nil {
		return MarshalledDocumentData{}, err
	}

	return data, nil
}

// MarshalLegacyDocumentData encodes a legacy-formatted document (the value in the `data` column).
func (s *Serializer) MarshalLegacyDocumentData(document precise.DocumentData) ([]byte, error) {
	return s.encode(&document)
}

// MarshalResultChunkData transforms result chunk data into a string of bytes writable to disk.
func (s *Serializer) MarshalResultChunkData(resultChunks precise.ResultChunkData) ([]byte, error) {
	return s.encode(&resultChunks)
}

// MarshalLocations transforms a slice of locations into a string of bytes writable to disk.
func (s *Serializer) MarshalLocations(locations []precise.LocationData) ([]byte, error) {
	return s.encode(&locations)
}

// encode gob-encodes and compresses the given payload.
func (s *Serializer) encode(payload any) (_ []byte, err error) {
	gzipWriter := s.writers.Get().(*gzip.Writer)
	defer s.writers.Put(gzipWriter)

	encodeBuf := new(bytes.Buffer)
	if err := gob.NewEncoder(encodeBuf).Encode(payload); err != nil {
		return nil, err
	}

	compressBuf := new(bytes.Buffer)
	gzipWriter.Reset(compressBuf)

	if _, err := io.Copy(gzipWriter, encodeBuf); err != nil {
		return nil, err
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}

	return compressBuf.Bytes(), nil
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

// UnmarshalLegacyDocumentData unmarshals a legacy-formatted document (the value in the `data` column).
func (s *Serializer) UnmarshalLegacyDocumentData(data []byte) (document precise.DocumentData, err error) {
	err = s.decode(data, &document)
	return document, err
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

// encode decompresses gob-decodes the given data and sets the given pointer. If the given data
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
