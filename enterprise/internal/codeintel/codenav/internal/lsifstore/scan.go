package lsifstore

import (
	"bytes"
	"database/sql"

	"github.com/sourcegraph/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// scanFirstDocumentData reads qualified document data from its given row object and returns
// the first one. If no rows match the query, a false-valued flag is returned.
func (s *store) scanFirstDocumentData(rows *sql.Rows, queryErr error) (_ QualifiedDocumentData, _ bool, err error) {
	if queryErr != nil {
		return QualifiedDocumentData{}, false, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if rows.Next() {
		record, err := s.scanSingleDocumentDataObject(rows)
		if err != nil {
			return QualifiedDocumentData{}, false, err
		}

		return record, true, nil
	}

	return QualifiedDocumentData{}, false, nil
}

// scanSingleDocumentDataObject populates a qualified document data value from the given cursor.
func (s *store) scanSingleDocumentDataObject(rows *sql.Rows) (QualifiedDocumentData, error) {
	var rawData []byte
	var uploadID int
	var path string
	var encoded MarshalledDocumentData
	var compressedSCIPPayload []byte

	if err := rows.Scan(
		&uploadID,
		&path,
		&rawData,
		&encoded.Ranges,
		&encoded.HoverResults,
		&encoded.Monikers,
		&encoded.PackageInformation,
		&encoded.Diagnostics,
		&compressedSCIPPayload,
	); err != nil {
		return QualifiedDocumentData{}, err
	}

	qualifiedData := QualifiedDocumentData{
		UploadID: uploadID,
		Path:     path,
	}

	if len(compressedSCIPPayload) != 0 {
		scipPayload, err := decompressor.decompress(bytes.NewReader(compressedSCIPPayload))
		if err != nil {
			return QualifiedDocumentData{}, err
		}

		var data scip.Document
		if err := proto.Unmarshal(scipPayload, &data); err != nil {
			return QualifiedDocumentData{}, err
		}

		qualifiedData.SCIPData = &data
	} else if len(rawData) != 0 {
		data, err := s.serializer.UnmarshalLegacyDocumentData(rawData)
		if err != nil {
			return QualifiedDocumentData{}, err
		}
		qualifiedData.LSIFData = &data
	} else {
		data, err := s.serializer.UnmarshalDocumentData(encoded)
		if err != nil {
			return QualifiedDocumentData{}, err
		}
		qualifiedData.LSIFData = &data
	}

	return qualifiedData, nil
}

// makeResultChunkVisitor returns a function that accepts a mapping function, reads
// result chunk values from the given row object and calls the mapping function on
// each decoded result set.
func (s *store) makeResultChunkVisitor(rows *sql.Rows, queryErr error) func(func(int, precise.ResultChunkData)) error {
	return func(f func(int, precise.ResultChunkData)) (err error) {
		if queryErr != nil {
			return queryErr
		}
		defer func() { err = basestore.CloseRows(rows, err) }()

		var rawData []byte
		for rows.Next() {
			var index int
			if err := rows.Scan(&index, &rawData); err != nil {
				return err
			}

			data, err := s.serializer.UnmarshalResultChunkData(rawData)
			if err != nil {
				return err
			}

			f(index, data)
		}

		return nil
	}
}

// makeDocumentVisitor returns a function that calls the given visitor function over each
// matching decoded document value.
func (s *store) makeDocumentVisitor(f func(string, precise.DocumentData)) func(rows *sql.Rows, queryErr error) error {
	return func(rows *sql.Rows, queryErr error) (err error) {
		if queryErr != nil {
			return queryErr
		}
		defer func() { err = basestore.CloseRows(rows, err) }()

		for rows.Next() {
			record, err := s.scanSingleDocumentDataObject(rows)
			if err != nil {
				return err
			}

			f(record.Path, *record.LSIFData)
		}

		return nil
	}
}

// scanQualifiedMonikerLocations reads moniker locations values from the given row object.
func (s *store) scanQualifiedMonikerLocations(rows *sql.Rows, queryErr error) (_ []QualifiedMonikerLocations, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var values []QualifiedMonikerLocations
	for rows.Next() {
		record, err := s.scanSingleQualifiedMonikerLocationsObject(rows)
		if err != nil {
			return nil, err
		}

		values = append(values, record)
	}

	return values, nil
}

// scanSingleQualifiedMonikerLocationsObject populates a qualified moniker locations value
// from the given cursor.
func (s *store) scanSingleQualifiedMonikerLocationsObject(rows *sql.Rows) (QualifiedMonikerLocations, error) {
	var uri string
	var scipPayload []byte
	var rawData []byte
	var record QualifiedMonikerLocations

	if err := rows.Scan(&record.DumpID, &record.Scheme, &record.Identifier, &rawData, &scipPayload, &uri); err != nil {
		return QualifiedMonikerLocations{}, err
	}

	if len(scipPayload) != 0 {
		ranges, err := types.DecodeRanges(scipPayload)
		if err != nil {
			return QualifiedMonikerLocations{}, err
		}

		locations := make([]precise.LocationData, 0, len(ranges))
		for _, r := range ranges {
			locations = append(locations, precise.LocationData{
				URI:            uri,
				StartLine:      int(r.Start.Line),
				StartCharacter: int(r.Start.Character),
				EndLine:        int(r.End.Line),
				EndCharacter:   int(r.End.Character),
			})
		}

		record.Locations = locations
	} else {
		data, err := s.serializer.UnmarshalLocations(rawData)
		if err != nil {
			return QualifiedMonikerLocations{}, err
		}
		record.Locations = data
	}

	return record, nil
}

// scanDocumentData reads qualified document data from the given row object.
func (s *store) scanDocumentData(rows *sql.Rows, queryErr error) (_ []QualifiedDocumentData, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var values []QualifiedDocumentData
	for rows.Next() {
		record, err := s.scanSingleDocumentDataObject(rows)
		if err != nil {
			return nil, err
		}

		values = append(values, record)
	}

	return values, nil
}
