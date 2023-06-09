package lsif

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"io"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func init() {
	gob.Register(&DocumentData{})
	gob.Register(&LocationData{})
}

type serializer struct {
	readers sync.Pool
	writers sync.Pool
}

func newSerializer() *serializer {
	return &serializer{
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
func (s *serializer) MarshalDocumentData(document DocumentData) (data MarshalledDocumentData, err error) {
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
func (s *serializer) MarshalLegacyDocumentData(document DocumentData) ([]byte, error) {
	return s.encode(&document)
}

// MarshalLocations transforms a slice of locations into a string of bytes writable to disk.
func (s *serializer) MarshalLocations(locations []LocationData) ([]byte, error) {
	return s.encode(&locations)
}

// encode gob-encodes and compresses the given payload.
func (s *serializer) encode(payload any) (_ []byte, err error) {
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
func (s *serializer) UnmarshalDocumentData(data MarshalledDocumentData) (document DocumentData, err error) {
	if err := s.decode(data.Ranges, &document.Ranges); err != nil {
		return DocumentData{}, err
	}
	if err := s.decode(data.HoverResults, &document.HoverResults); err != nil {
		return DocumentData{}, err
	}
	if err := s.decode(data.Monikers, &document.Monikers); err != nil {
		return DocumentData{}, err
	}
	if err := s.decode(data.PackageInformation, &document.PackageInformation); err != nil {
		return DocumentData{}, err
	}
	if err := s.decode(data.Diagnostics, &document.Diagnostics); err != nil {
		return DocumentData{}, err
	}

	return document, nil
}

// UnmarshalLegacyDocumentData unmarshals a legacy-formatted document (the value in the `data` column).
func (s *serializer) UnmarshalLegacyDocumentData(data []byte) (document DocumentData, err error) {
	err = s.decode(data, &document)
	return document, err
}

// UnmarshalResultChunkData is the inverse of MarshalResultChunkData.
func (s *serializer) UnmarshalResultChunkData(data []byte) (resultChunk ResultChunkData, err error) {
	err = s.decode(data, &resultChunk)
	return resultChunk, err
}

// UnmarshalLocations is the inverse of MarshalLocations.
func (s *serializer) UnmarshalLocations(data []byte) (locations []LocationData, err error) {
	err = s.decode(data, &locations)
	return locations, err
}

// decode decompresses gob-decodes the given data and sets the given pointer. If the given data
// is empty, the pointer will not be assigned.
func (s *serializer) decode(data []byte, target any) (err error) {
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
