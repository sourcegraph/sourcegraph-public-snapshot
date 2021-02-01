package lsifstore

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type QualifiedDocumentData struct {
	UploadID int
	KeyedDocumentData
}

// scanDocumentData reads qualified document data from the given row object.
func (s *Store) scanDocumentData(rows *sql.Rows, queryErr error) (_ []QualifiedDocumentData, err error) {
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

// scanFirstDocumentData reads qualified document data values from the given row
// object and returns the first one. If no rows match the query, a false-valued
// flag is returned.
func (s *Store) scanFirstDocumentData(rows *sql.Rows, queryErr error) (_ QualifiedDocumentData, _ bool, err error) {
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

// scanSingleDocumentDataObject populates a qualified document data value from the
// given cursor.
func (s *Store) scanSingleDocumentDataObject(rows *sql.Rows) (QualifiedDocumentData, error) {
	var rawData []byte
	var record QualifiedDocumentData
	if err := rows.Scan(&record.UploadID, &record.Path, &rawData); err != nil {
		return QualifiedDocumentData{}, err
	}

	data, err := s.serializer.UnmarshalDocumentData(rawData)
	if err != nil {
		return QualifiedDocumentData{}, err
	}
	record.Document = data

	return record, nil
}

type QualifiedResultChunkData struct {
	UploadID int
	IndexedResultChunkData
}

// scanQualifiedResultChunkData reads qualified result chunk data from the given
// row object.
func (s *Store) scanQualifiedResultChunkData(rows *sql.Rows, queryErr error) (_ []QualifiedResultChunkData, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var values []QualifiedResultChunkData
	for rows.Next() {
		record, err := s.scanSingleResultChunkDataObject(rows)
		if err != nil {
			return nil, err
		}

		values = append(values, record)
	}

	return values, nil
}

// scanFirstQualifiedResultChunkData reads qualified result chunk data values from
// the given row object and returns the first one. If no rows match the query, a
// false-valued flag is returned.
func (s *Store) scanFirstQualifiedResultChunkData(rows *sql.Rows, queryErr error) (_ QualifiedResultChunkData, _ bool, err error) {
	if queryErr != nil {
		return QualifiedResultChunkData{}, false, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if rows.Next() {
		record, err := s.scanSingleResultChunkDataObject(rows)
		if err != nil {
			return QualifiedResultChunkData{}, false, err
		}

		return record, true, nil
	}

	return QualifiedResultChunkData{}, false, nil
}

// scanSingleResultChunkDataObject populates a qualified result chunk data value from
// the given cursor.
func (s *Store) scanSingleResultChunkDataObject(rows *sql.Rows) (QualifiedResultChunkData, error) {
	var rawData []byte
	var record QualifiedResultChunkData
	if err := rows.Scan(&record.UploadID, &record.Index, &rawData); err != nil {
		return QualifiedResultChunkData{}, err
	}

	data, err := s.serializer.UnmarshalResultChunkData(rawData)
	if err != nil {
		return QualifiedResultChunkData{}, err
	}
	record.ResultChunk = data

	return record, nil
}

// scanQualifiedResultChunkData reads moniker locations values from the given row object.
func (s *Store) scanLocations(rows *sql.Rows, queryErr error) (_ []MonikerLocations, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var values []MonikerLocations
	for rows.Next() {
		record, err := s.scanSingleMonikerLocationsObject(rows)
		if err != nil {
			return nil, err
		}

		values = append(values, record)
	}

	return values, nil
}

// scanFirstLocations reads a moniker locations value from the given row object and
// returns the first one. If no rows match the query, a false-valued flag is returned.
func (s *Store) scanFirstLocations(rows *sql.Rows, queryErr error) (_ MonikerLocations, _ bool, err error) {
	if queryErr != nil {
		return MonikerLocations{}, false, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if rows.Next() {
		record, err := s.scanSingleMonikerLocationsObject(rows)
		if err != nil {
			return MonikerLocations{}, false, err
		}

		return record, true, nil
	}

	return MonikerLocations{}, false, nil
}

// scanSingleMonikerLocationsObject populates a moniker locations value from the
// given cursor.
func (s *Store) scanSingleMonikerLocationsObject(rows *sql.Rows) (MonikerLocations, error) {
	var rawData []byte
	var record MonikerLocations
	if err := rows.Scan(&record.Scheme, &record.Identifier, &rawData); err != nil {
		return MonikerLocations{}, err
	}

	data, err := s.serializer.UnmarshalLocations(rawData)
	if err != nil {
		return MonikerLocations{}, err
	}
	record.Locations = data

	return record, nil
}
