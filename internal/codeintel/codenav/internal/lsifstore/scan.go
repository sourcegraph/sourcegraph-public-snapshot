package lsifstore

import (
	"bytes"
	"database/sql"

	"github.com/sourcegraph/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codegraph"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/ranges"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/collections"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type qualifiedDocumentData struct {
	UploadID int
	Path     string
	SCIPData *scip.Document
}

func (s *store) scanDocumentData(rows *sql.Rows, queryErr error) (_ []qualifiedDocumentData, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var values []qualifiedDocumentData
	for rows.Next() {
		record, err := s.scanSingleDocumentDataObject(rows)
		if err != nil {
			return nil, err
		}

		values = append(values, record)
	}

	return values, nil
}

func (s *store) scanFirstDocumentData(rows *sql.Rows, queryErr error) (_ qualifiedDocumentData, _ bool, err error) {
	if queryErr != nil {
		return qualifiedDocumentData{}, false, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if rows.Next() {
		record, err := s.scanSingleDocumentDataObject(rows)
		if err != nil {
			return qualifiedDocumentData{}, false, err
		}

		return record, true, nil
	}

	return qualifiedDocumentData{}, false, nil
}

func (s *store) scanSingleDocumentDataObject(rows *sql.Rows) (qualifiedDocumentData, error) {
	var uploadID int
	var path string
	var compressedSCIPPayload []byte

	if err := rows.Scan(&uploadID, &path, &compressedSCIPPayload); err != nil {
		return qualifiedDocumentData{}, err
	}

	scipPayload, err := shared.Decompressor.Decompress(bytes.NewReader(compressedSCIPPayload))
	if err != nil {
		return qualifiedDocumentData{}, err
	}

	var data scip.Document
	if err := proto.Unmarshal(scipPayload, &data); err != nil {
		return qualifiedDocumentData{}, err
	}

	qualifiedData := qualifiedDocumentData{
		UploadID: uploadID,
		Path:     path,
		SCIPData: &data,
	}
	return qualifiedData, nil
}

type qualifiedMonikerLocations struct {
	UploadID int
	precise.MonikerLocations
}

// Post-condition: number of entries == number of unique (upload, symbol, document) triples.
func (s *store) scanUploadSymbolLoci(rows *sql.Rows, queryErr error) (_ []codegraph.UploadSymbolLoci, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var values []codegraph.UploadSymbolLoci
	for rows.Next() {
		record, err := s.scanSingleUploadSymbolLoci(rows)
		if err != nil {
			return nil, err
		}

		values = append(values, record)
	}

	return values, nil
}

func (s *store) scanSingleUploadSymbolLoci(rows *sql.Rows) (codegraph.UploadSymbolLoci, error) {
	var uploadID int
	var symbol string
	var customEncodedRanges []byte
	var documentPath string
	if err := rows.Scan(&uploadID, &symbol, &customEncodedRanges, &documentPath); err != nil {
		return codegraph.UploadSymbolLoci{}, err
	}

	ranges, err := ranges.DecodeRanges(customEncodedRanges)
	if err != nil {
		return codegraph.UploadSymbolLoci{}, err
	}

	locations := make([]codegraph.Locus, 0, len(ranges))
	for _, r := range ranges {
		locations = append(locations, codegraph.Locus{
			Path:  core.NewUploadRelPathUnchecked(documentPath),
			Range: scip.NewRangeUnchecked([]int32{r.Start.Line, r.Start.Character, r.End.Line, r.End.Character}),
		})
	}

	return codegraph.UploadSymbolLoci{
		UploadID: uploadID,
		Symbol:   symbol,
		Loci:     locations,
	}, nil
}

//
//

// Post-condition: Returns one entry per upload.
func (s *store) scanDeduplicatedUploadLoci(rows *sql.Rows, queryErr error) (_ []codegraph.UploadLoci, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var values []codegraph.UploadLoci
	for rows.Next() {
		record, err := s.scanSingleUploadLoci(rows)
		if err != nil {
			return nil, err
		}

		// TODO: Also use the ordering guarantees for document paths + range sorting
		// on insertion to replace the Deduplicate with some simple checks here.
		if n := len(values) - 1; n >= 0 && values[n].UploadID == record.UploadID {
			values[n].Loci = append(values[n].Loci, record.Loci...)
		} else {
			values = append(values, record)
		}
	}
	for i := range values {
		values[i].Loci = collections.Deduplicate(values[i].Loci)
	}

	return values, nil
}

func (s *store) scanSingleUploadLoci(rows *sql.Rows) (codegraph.UploadLoci, error) {
	var uploadID int
	var customEncodedRanges []byte
	var documentPath string
	if err := rows.Scan(&uploadID, &customEncodedRanges, &documentPath); err != nil {
		return codegraph.UploadLoci{}, err
	}

	ranges, err := ranges.DecodeRanges(customEncodedRanges)
	if err != nil {
		return codegraph.UploadLoci{}, err
	}

	locations := make([]codegraph.Locus, 0, len(ranges))
	for _, r := range ranges {
		locations = append(locations, codegraph.Locus{
			Path:  core.NewUploadRelPathUnchecked(documentPath),
			Range: scip.NewRangeUnchecked([]int32{r.Start.Line, r.Start.Character, r.End.Line, r.End.Character}),
		})
	}

	return codegraph.UploadLoci{
		UploadID: uploadID,
		Loci:     locations,
	}, nil
}
