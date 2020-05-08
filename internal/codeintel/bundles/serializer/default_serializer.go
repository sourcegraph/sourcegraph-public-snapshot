package serializer

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	_ "strconv"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

type defaultSerializer struct{}

var _ Serializer = &defaultSerializer{}

func NewDefaultSerializer() Serializer {
	return &defaultSerializer{}
}

func (*defaultSerializer) MarshalDocumentData(d types.DocumentData) ([]byte, error) {
	rangePairs := []interface{}{}
	for k, v := range d.Ranges {
		if v.MonikerIDs == nil {
			v.MonikerIDs = []types.ID{}
		}

		vs := map[string]interface{}{
			"startLine":          v.StartLine,
			"startCharacter":     v.StartCharacter,
			"endLine":            v.EndLine,
			"endCharacter":       v.EndCharacter,
			"definitionResultId": v.DefinitionResultID,
			"referenceResultId":  v.ReferenceResultID,
			"hoverResultId":      v.HoverResultID,
			"monikerIds":         map[string]interface{}{"type": "set", "value": v.MonikerIDs},
		}

		rangePairs = append(rangePairs, []interface{}{k, vs})
	}

	hoverResultPairs := []interface{}{}
	for k, v := range d.HoverResults {
		hoverResultPairs = append(hoverResultPairs, []interface{}{k, v})
	}

	monikerPairs := []interface{}{}
	for k, v := range d.Monikers {
		monikerPairs = append(monikerPairs, []interface{}{k, v})
	}

	packageInformationPairs := []interface{}{}
	for k, v := range d.PackageInformation {
		packageInformationPairs = append(packageInformationPairs, []interface{}{k, v})
	}

	encoded, err := json.Marshal(map[string]interface{}{
		"ranges":             map[string]interface{}{"type": "map", "value": rangePairs},
		"hoverResults":       map[string]interface{}{"type": "map", "value": hoverResultPairs},
		"monikers":           map[string]interface{}{"type": "map", "value": monikerPairs},
		"packageInformation": map[string]interface{}{"type": "map", "value": packageInformationPairs},
	})
	if err != nil {
		return nil, err
	}

	return compress(encoded)
}

func (defaultSerializer) MarshalResultChunkData(rc types.ResultChunkData) ([]byte, error) {
	documentPathPairs := []interface{}{}
	for k, v := range rc.DocumentPaths {
		documentPathPairs = append(documentPathPairs, []interface{}{k, v})
	}

	documentIDRangeIDPairs := []interface{}{}
	for k, v := range rc.DocumentIDRangeIDs {
		documentIDRangeIDPairs = append(documentIDRangeIDPairs, []interface{}{k, v})
	}

	encoded, err := json.Marshal(map[string]interface{}{
		"documentPaths":      map[string]interface{}{"type": "map", "value": documentPathPairs},
		"documentIdRangeIds": map[string]interface{}{"type": "map", "value": documentIDRangeIDPairs},
	})
	if err != nil {
		return nil, err
	}

	return compress(encoded)
}

func (defaultSerializer) UnmarshalDocumentData(data []byte) (types.DocumentData, error) {
	payload := struct {
		Ranges             wrappedMapValue `json:"ranges"`
		HoverResults       wrappedMapValue `json:"hoverResults"`
		Monikers           wrappedMapValue `json:"monikers"`
		PackageInformation wrappedMapValue `json:"packageInformation"`
	}{}

	if err := unmarshalGzippedJSON(data, &payload); err != nil {
		return types.DocumentData{}, err
	}

	ranges, err := unmarshalWrappedRanges(payload.Ranges.Value)
	if err != nil {
		return types.DocumentData{}, err
	}

	hoverResults, err := unmarshalWrappedHoverResults(payload.HoverResults.Value)
	if err != nil {
		return types.DocumentData{}, err
	}

	monikers, err := unmarshalWrappedMonikers(payload.Monikers.Value)
	if err != nil {
		return types.DocumentData{}, err
	}

	packageInformation, err := unmarshalWrappedPackageInformation(payload.PackageInformation.Value)
	if err != nil {
		return types.DocumentData{}, err
	}

	return types.DocumentData{
		Ranges:             ranges,
		HoverResults:       hoverResults,
		Monikers:           monikers,
		PackageInformation: packageInformation,
	}, nil
}

func (defaultSerializer) UnmarshalResultChunkData(data []byte) (types.ResultChunkData, error) {
	payload := struct {
		DocumentPaths      wrappedMapValue `json:"documentPaths"`
		DocumentIDRangeIDs wrappedMapValue `json:"documentIdRangeIds"`
	}{}

	if err := unmarshalGzippedJSON(data, &payload); err != nil {
		return types.ResultChunkData{}, err
	}

	documentPaths, err := unmarshalWrappedDocumentPaths(payload.DocumentPaths.Value)
	if err != nil {
		return types.ResultChunkData{}, err
	}

	documentIDRangeIDs, err := unmarshalWrappedDocumentIdRangeIDs(payload.DocumentIDRangeIDs.Value)
	if err != nil {
		return types.ResultChunkData{}, err
	}

	return types.ResultChunkData{
		DocumentPaths:      documentPaths,
		DocumentIDRangeIDs: documentIDRangeIDs,
	}, nil
}

//
//
//

func unmarshalWrappedRanges(pairs []json.RawMessage) (map[types.ID]types.RangeData, error) {
	m := map[types.ID]types.RangeData{}
	for _, pair := range pairs {
		var id types.ID
		var value struct {
			StartLine          int             `json:"startLine"`
			StartCharacter     int             `json:"startCharacter"`
			EndLine            int             `json:"endLine"`
			EndCharacter       int             `json:"endCharacter"`
			DefinitionResultID types.ID        `json:"definitionResultId"`
			ReferenceResultID  types.ID        `json:"referenceResultId"`
			HoverResultID      types.ID        `json:"hoverResultId"`
			MonikerIDs         wrappedSetValue `json:"monikerIds"`
		}

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		var monikerIDs []types.ID
		for _, value := range value.MonikerIDs.Value {
			var id types.ID
			if err := json.Unmarshal([]byte(value), &id); err != nil {
				return nil, err
			}

			monikerIDs = append(monikerIDs, id)
		}

		m[id] = types.RangeData{
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

func unmarshalWrappedHoverResults(pairs []json.RawMessage) (map[types.ID]string, error) {
	m := map[types.ID]string{}
	for _, pair := range pairs {
		var id types.ID
		var value string

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		m[id] = value
	}

	return m, nil
}

func unmarshalWrappedMonikers(pairs []json.RawMessage) (map[types.ID]types.MonikerData, error) {
	m := map[types.ID]types.MonikerData{}
	for _, pair := range pairs {
		var id types.ID
		var value struct {
			Kind                 string   `json:"kind"`
			Scheme               string   `json:"scheme"`
			Identifier           string   `json:"identifier"`
			PackageInformationID types.ID `json:"packageInformationId"`
		}

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		m[id] = types.MonikerData{
			Kind:                 value.Kind,
			Scheme:               value.Scheme,
			Identifier:           value.Identifier,
			PackageInformationID: value.PackageInformationID,
		}
	}

	return m, nil
}

func unmarshalWrappedPackageInformation(pairs []json.RawMessage) (map[types.ID]types.PackageInformationData, error) {
	m := map[types.ID]types.PackageInformationData{}
	for _, pair := range pairs {
		var id types.ID
		var value struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		m[id] = types.PackageInformationData{
			Name:    value.Name,
			Version: value.Version,
		}
	}

	return m, nil
}

func unmarshalWrappedDocumentPaths(pairs []json.RawMessage) (map[types.ID]string, error) {
	m := map[types.ID]string{}
	for _, pair := range pairs {
		var id types.ID
		var value string

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		m[id] = value
	}

	return m, nil
}

func unmarshalWrappedDocumentIdRangeIDs(pairs []json.RawMessage) (map[types.ID][]types.DocumentIDRangeID, error) {
	m := map[types.ID][]types.DocumentIDRangeID{}
	for _, pair := range pairs {
		var id types.ID
		var value []struct {
			DocumentID types.ID `json:"documentId"`
			RangeID    types.ID `json:"rangeId"`
		}

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		var documentIDRangeIDs []types.DocumentIDRangeID
		for _, v := range value {
			documentIDRangeIDs = append(documentIDRangeIDs, types.DocumentIDRangeID{
				DocumentID: v.DocumentID,
				RangeID:    v.RangeID,
			})
		}

		m[id] = documentIDRangeIDs
	}

	return m, nil
}

// TODO(efritz) - document
func compress(uncompressed []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	if _, err := io.Copy(gzipWriter, bytes.NewReader(uncompressed)); err != nil {
		return nil, err
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// unmarshalGzippedJSON unmarshals the gzip+json encoded data.
func unmarshalGzippedJSON(data []byte, payload interface{}) error {
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}

	return json.NewDecoder(gzipReader).Decode(&payload)
}

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
