package lsifstore

import (
	"bytes"
	"database/sql"

	"github.com/lib/pq"
	genslices "github.com/life4/genesis/slices"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codegraph"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/ranges"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
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
	var customEncodedRangesList pq.ByteaArray
	var documentPaths pq.StringArray
	if err := rows.Scan(&uploadID, &symbol, &customEncodedRangesList, &documentPaths); err != nil {
		return codegraph.UploadSymbolLoci{}, err
	}

	loci := []codegraph.Locus{}
	for i, docPath := range documentPaths {
		scipRanges, err := ranges.DecodeRanges(customEncodedRangesList[i])
		if err != nil {
			return codegraph.UploadSymbolLoci{}, err
		}
		loci = append(loci, genslices.Map(scipRanges, func(r scip.Range) codegraph.Locus {
			return codegraph.Locus{Path: core.NewUploadRelPathUnchecked(docPath), Range: r}
		})...)
	}

	return codegraph.UploadSymbolLoci{
		UploadID: uploadID,
		Symbol:   symbol,
		Loci:     loci,
	}, nil
}
