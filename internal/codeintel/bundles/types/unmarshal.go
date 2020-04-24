package types

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"strconv"
)

// wrappedMapValue represents a JSON-encoded map with the following form.
// This maintains the same functionality that exists on the TypeScript side.
//
//     {
//       "value": [
//         ["key-1", "value-1"],
//         ["key-2", "value-2"],
//         ...
//       ]
//     }
type wrappedMapValue struct {
	Value []json.RawMessage `json:"value"`
}

// wrappedSetValue represents a JSON-encoded set with the following form.
// This maintains the same functionality that exists on the TypeScript side.
//
//     {
//       "value": [
//         "value-1",
//         "value-2",
//         ...
//       ]
//     }
type wrappedSetValue struct {
	Value []json.RawMessage `json:"value"`
}

// UnmarshalJSON converts a JSON number or string into an identifier. This
// maintains the same functionality that exists on the TypeScript side by
// simply running JSON.parse() on document and result chunk data blobs.
func (id *ID) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		return json.Unmarshal(b, (*string)(id))
	}

	var value int64
	if err := json.Unmarshal(b, &value); err != nil {
		return err
	}

	*id = ID(strconv.FormatInt(value, 10))
	return nil
}

// UnmarshalDocumentData unmarshals document data from a gzipped json-encoded blob.
func UnmarshalDocumentData(data []byte) (DocumentData, error) {
	payload := struct {
		Ranges             wrappedMapValue `json:"ranges"`
		HoverResults       wrappedMapValue `json:"hoverResults"`
		Monikers           wrappedMapValue `json:"monikers"`
		PackageInformation wrappedMapValue `json:"packageInformation"`
	}{}

	if err := unmarshalGzippedJSON(data, &payload); err != nil {
		return DocumentData{}, err
	}

	ranges, err := unmarshalWrappedRanges(payload.Ranges.Value)
	if err != nil {
		return DocumentData{}, err
	}

	hoverResults, err := unmarshalWrappedHoverResults(payload.HoverResults.Value)
	if err != nil {
		return DocumentData{}, err
	}

	monikers, err := unmarshalWrappedMonikers(payload.Monikers.Value)
	if err != nil {
		return DocumentData{}, err
	}

	packageInformation, err := unmarshalWrappedPackageInformation(payload.PackageInformation.Value)
	if err != nil {
		return DocumentData{}, err
	}

	return DocumentData{
		Ranges:             ranges,
		HoverResults:       hoverResults,
		Monikers:           monikers,
		PackageInformation: packageInformation,
	}, nil
}

// UnmarshalDocumentData unmarshals result chunk data from a gzipped json-encoded blob.
func UnmarshalResultChunkData(data []byte) (ResultChunkData, error) {
	payload := struct {
		DocumentPaths      wrappedMapValue `json:"documentPaths"`
		DocumentIDRangeIDs wrappedMapValue `json:"documentIdRangeIds"`
	}{}

	if err := unmarshalGzippedJSON(data, &payload); err != nil {
		return ResultChunkData{}, err
	}

	documentPaths, err := unmarshalWrappedDocumentPaths(payload.DocumentPaths.Value)
	if err != nil {
		return ResultChunkData{}, err
	}

	documentIDRangeIDs, err := unmarshalWrappedDocumentIdRangeIDs(payload.DocumentIDRangeIDs.Value)
	if err != nil {
		return ResultChunkData{}, err
	}

	return ResultChunkData{
		DocumentPaths:      documentPaths,
		DocumentIDRangeIDs: documentIDRangeIDs,
	}, nil
}

func unmarshalWrappedRanges(pairs []json.RawMessage) (map[ID]RangeData, error) {
	m := map[ID]RangeData{}
	for _, pair := range pairs {
		var id ID
		var value struct {
			StartLine          int             `json:"startLine"`
			StartCharacter     int             `json:"startCharacter"`
			EndLine            int             `json:"endLine"`
			EndCharacter       int             `json:"endCharacter"`
			DefinitionResultID ID              `json:"definitionResultID"`
			ReferenceResultID  ID              `json:"referenceResultID"`
			HoverResultID      ID              `json:"hoverResultID"`
			MonikerIDs         wrappedSetValue `json:"monikerIDs"`
		}

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		var monikerIDs []ID
		for _, value := range value.MonikerIDs.Value {
			var id ID
			if err := json.Unmarshal([]byte(value), &id); err != nil {
				return nil, err
			}

			monikerIDs = append(monikerIDs, id)
		}

		m[id] = RangeData{
			StartLine:          value.StartLine,
			StartCharacter:     value.StartCharacter,
			EndLine:            value.EndLine,
			EndCharacter:       value.EndCharacter,
			DefinitionResultID: value.DefinitionResultID,
			ReferenceResultID:  value.ReferenceResultID,
			HoverResultID:      value.HoverResultID,
			MonikerIDs:         monikerIDs,
		}
	}

	return m, nil
}

func unmarshalWrappedHoverResults(pairs []json.RawMessage) (map[ID]string, error) {
	m := map[ID]string{}
	for _, pair := range pairs {
		var id ID
		var value string

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		m[id] = value
	}

	return m, nil
}

func unmarshalWrappedMonikers(pairs []json.RawMessage) (map[ID]MonikerData, error) {
	m := map[ID]MonikerData{}
	for _, pair := range pairs {
		var id ID
		var value struct {
			Kind                 string `json:"kind"`
			Scheme               string `json:"scheme"`
			Identifier           string `json:"identifier"`
			PackageInformationID ID     `json:"packageInformationID"`
		}

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		m[id] = MonikerData{
			Kind:                 value.Kind,
			Scheme:               value.Scheme,
			Identifier:           value.Identifier,
			PackageInformationID: value.PackageInformationID,
		}
	}

	return m, nil
}

func unmarshalWrappedPackageInformation(pairs []json.RawMessage) (map[ID]PackageInformationData, error) {
	m := map[ID]PackageInformationData{}
	for _, pair := range pairs {
		var id ID
		var value struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		m[id] = PackageInformationData{
			Name:    value.Name,
			Version: value.Version,
		}
	}

	return m, nil
}

func unmarshalWrappedDocumentPaths(pairs []json.RawMessage) (map[ID]string, error) {
	m := map[ID]string{}
	for _, pair := range pairs {
		var id ID
		var value string

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		m[id] = value
	}

	return m, nil
}

func unmarshalWrappedDocumentIdRangeIDs(pairs []json.RawMessage) (map[ID][]DocumentIDRangeID, error) {
	m := map[ID][]DocumentIDRangeID{}
	for _, pair := range pairs {
		var id ID
		var value []struct {
			DocumentID ID `json:"documentId"`
			RangeID    ID `json:"rangeId"`
		}

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		var documentIDRangeIDs []DocumentIDRangeID
		for _, v := range value {
			documentIDRangeIDs = append(documentIDRangeIDs, DocumentIDRangeID{
				DocumentID: v.DocumentID,
				RangeID:    v.RangeID,
			})
		}

		m[id] = documentIDRangeIDs
	}

	return m, nil
}

// unmarshalGzippedJSON unmarshals the gzip+json encoded data.
func unmarshalGzippedJSON(data []byte, payload interface{}) error {
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}

	return json.NewDecoder(gzipReader).Decode(&payload)
}
