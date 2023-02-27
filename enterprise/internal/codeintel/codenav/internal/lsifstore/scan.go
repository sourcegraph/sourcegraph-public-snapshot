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
	var uploadID int
	var path string
	var compressedSCIPPayload []byte

	if err := rows.Scan(&uploadID, &path, &compressedSCIPPayload); err != nil {
		return QualifiedDocumentData{}, err
	}

	scipPayload, err := decompressor.decompress(bytes.NewReader(compressedSCIPPayload))
	if err != nil {
		return QualifiedDocumentData{}, err
	}

	var data scip.Document
	if err := proto.Unmarshal(scipPayload, &data); err != nil {
		return QualifiedDocumentData{}, err
	}

	qualifiedData := QualifiedDocumentData{
		UploadID: uploadID,
		Path:     path,
		SCIPData: &data,
	}
	return qualifiedData, nil
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
	var record QualifiedMonikerLocations

	if err := rows.Scan(&record.DumpID, &record.Scheme, &record.Identifier, &scipPayload, &uri); err != nil {
		return QualifiedMonikerLocations{}, err
	}

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
