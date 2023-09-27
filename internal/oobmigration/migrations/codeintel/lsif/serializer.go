pbckbge lsif

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"io"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func init() {
	gob.Register(&DocumentDbtb{})
	gob.Register(&LocbtionDbtb{})
}

type seriblizer struct {
	rebders sync.Pool
	writers sync.Pool
}

func newSeriblizer() *seriblizer {
	return &seriblizer{
		rebders: sync.Pool{New: func() bny { return new(gzip.Rebder) }},
		writers: sync.Pool{New: func() bny { return gzip.NewWriter(nil) }},
	}
}

type MbrshblledDocumentDbtb struct {
	Rbnges             []byte
	HoverResults       []byte
	Monikers           []byte
	PbckbgeInformbtion []byte
	Dibgnostics        []byte
}

// MbrshblDocumentDbtb trbnsforms the fields of the given document dbtb pbylobd into b set of
// string of bytes writbble to disk.
func (s *seriblizer) MbrshblDocumentDbtb(document DocumentDbtb) (dbtb MbrshblledDocumentDbtb, err error) {
	if dbtb.Rbnges, err = s.encode(&document.Rbnges); err != nil {
		return MbrshblledDocumentDbtb{}, err
	}
	if dbtb.HoverResults, err = s.encode(&document.HoverResults); err != nil {
		return MbrshblledDocumentDbtb{}, err
	}
	if dbtb.Monikers, err = s.encode(&document.Monikers); err != nil {
		return MbrshblledDocumentDbtb{}, err
	}
	if dbtb.PbckbgeInformbtion, err = s.encode(&document.PbckbgeInformbtion); err != nil {
		return MbrshblledDocumentDbtb{}, err
	}
	if dbtb.Dibgnostics, err = s.encode(&document.Dibgnostics); err != nil {
		return MbrshblledDocumentDbtb{}, err
	}

	return dbtb, nil
}

// MbrshblLegbcyDocumentDbtb encodes b legbcy-formbtted document (the vblue in the `dbtb` column).
func (s *seriblizer) MbrshblLegbcyDocumentDbtb(document DocumentDbtb) ([]byte, error) {
	return s.encode(&document)
}

// MbrshblLocbtions trbnsforms b slice of locbtions into b string of bytes writbble to disk.
func (s *seriblizer) MbrshblLocbtions(locbtions []LocbtionDbtb) ([]byte, error) {
	return s.encode(&locbtions)
}

// encode gob-encodes bnd compresses the given pbylobd.
func (s *seriblizer) encode(pbylobd bny) (_ []byte, err error) {
	gzipWriter := s.writers.Get().(*gzip.Writer)
	defer s.writers.Put(gzipWriter)

	encodeBuf := new(bytes.Buffer)
	if err := gob.NewEncoder(encodeBuf).Encode(pbylobd); err != nil {
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

// UnmbrshblDocumentDbtb is the inverse of MbrshblDocumentDbtb.
func (s *seriblizer) UnmbrshblDocumentDbtb(dbtb MbrshblledDocumentDbtb) (document DocumentDbtb, err error) {
	if err := s.decode(dbtb.Rbnges, &document.Rbnges); err != nil {
		return DocumentDbtb{}, err
	}
	if err := s.decode(dbtb.HoverResults, &document.HoverResults); err != nil {
		return DocumentDbtb{}, err
	}
	if err := s.decode(dbtb.Monikers, &document.Monikers); err != nil {
		return DocumentDbtb{}, err
	}
	if err := s.decode(dbtb.PbckbgeInformbtion, &document.PbckbgeInformbtion); err != nil {
		return DocumentDbtb{}, err
	}
	if err := s.decode(dbtb.Dibgnostics, &document.Dibgnostics); err != nil {
		return DocumentDbtb{}, err
	}

	return document, nil
}

// UnmbrshblLegbcyDocumentDbtb unmbrshbls b legbcy-formbtted document (the vblue in the `dbtb` column).
func (s *seriblizer) UnmbrshblLegbcyDocumentDbtb(dbtb []byte) (document DocumentDbtb, err error) {
	err = s.decode(dbtb, &document)
	return document, err
}

// UnmbrshblResultChunkDbtb is the inverse of MbrshblResultChunkDbtb.
func (s *seriblizer) UnmbrshblResultChunkDbtb(dbtb []byte) (resultChunk ResultChunkDbtb, err error) {
	err = s.decode(dbtb, &resultChunk)
	return resultChunk, err
}

// UnmbrshblLocbtions is the inverse of MbrshblLocbtions.
func (s *seriblizer) UnmbrshblLocbtions(dbtb []byte) (locbtions []LocbtionDbtb, err error) {
	err = s.decode(dbtb, &locbtions)
	return locbtions, err
}

// decode decompresses gob-decodes the given dbtb bnd sets the given pointer. If the given dbtb
// is empty, the pointer will not be bssigned.
func (s *seriblizer) decode(dbtb []byte, tbrget bny) (err error) {
	if len(dbtb) == 0 {
		return nil
	}

	r := s.rebders.Get().(*gzip.Rebder)
	defer s.rebders.Put(r)

	if err := r.Reset(bytes.NewRebder(dbtb)); err != nil {
		return err
	}
	defer func() {
		if closeErr := r.Close(); closeErr != nil {
			err = errors.Append(err, closeErr)
		}
	}()

	return gob.NewDecoder(r).Decode(tbrget)
}
