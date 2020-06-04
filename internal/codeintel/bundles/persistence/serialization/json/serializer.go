package json

import (
	"encoding/json"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

type jsonSerializer struct{}

var _ serialization.Serializer = &jsonSerializer{}

func New() serialization.Serializer {
	return &jsonSerializer{}
}

func (*jsonSerializer) MarshalDocumentData(document types.DocumentData) ([]byte, error) {
	rangePairs := make([]interface{}, 0, len(document.Ranges))
	for k, v := range document.Ranges {
		if v.MonikerIDs == nil {
			v.MonikerIDs = []types.ID{}
		}

		vs := SerializingRange{
			StartLine:          v.StartLine,
			StartCharacter:     v.StartCharacter,
			EndLine:            v.EndLine,
			EndCharacter:       v.EndCharacter,
			DefinitionResultID: v.DefinitionResultID,
			ReferenceResultID:  v.ReferenceResultID,
			HoverResultID:      v.HoverResultID,
			MonikerIDs:         SerializingTaggedValue{Type: "set", Value: v.MonikerIDs},
		}

		rangePairs = append(rangePairs, []interface{}{k, vs})
	}

	hoverResultPairs := make([]interface{}, 0, len(document.HoverResults))
	for k, v := range document.HoverResults {
		hoverResultPairs = append(hoverResultPairs, []interface{}{k, v})
	}

	monikerPairs := make([]interface{}, 0, len(document.Monikers))
	for k, v := range document.Monikers {
		monikerPairs = append(monikerPairs, []interface{}{k, v})
	}

	packageInformationPairs := make([]interface{}, 0, len(document.PackageInformation))
	for k, v := range document.PackageInformation {
		packageInformationPairs = append(packageInformationPairs, []interface{}{k, v})
	}

	encoded, err := json.Marshal(SerializingDocument{
		Ranges:             SerializingTaggedValue{Type: "map", Value: rangePairs},
		HoverResults:       SerializingTaggedValue{Type: "map", Value: hoverResultPairs},
		Monikers:           SerializingTaggedValue{Type: "map", Value: monikerPairs},
		PackageInformation: SerializingTaggedValue{Type: "map", Value: packageInformationPairs},
	})
	if err != nil {
		return nil, err
	}

	return compress(encoded)
}

func (*jsonSerializer) MarshalResultChunkData(resultChunk types.ResultChunkData) ([]byte, error) {
	documentPathPairs := make([]interface{}, 0, len(resultChunk.DocumentPaths))
	for k, v := range resultChunk.DocumentPaths {
		documentPathPairs = append(documentPathPairs, []interface{}{k, v})
	}

	documentIDRangeIDPairs := make([]interface{}, 0, len(resultChunk.DocumentIDRangeIDs))
	for k, v := range resultChunk.DocumentIDRangeIDs {
		documentIDRangeIDPairs = append(documentIDRangeIDPairs, []interface{}{k, v})
	}

	encoded, err := json.Marshal(SerializingResultChunk{
		DocumentPaths:      SerializingTaggedValue{Type: "map", Value: documentPathPairs},
		DocumentIDRangeIDs: SerializingTaggedValue{Type: "map", Value: documentIDRangeIDPairs},
	})
	if err != nil {
		return nil, err
	}

	return compress(encoded)
}

func (*jsonSerializer) MarshalLocations(locations []types.Location) ([]byte, error) {
	serializingLocations := make([]SerializingLocation, 0, len(locations))
	for _, location := range locations {
		serializingLocations = append(serializingLocations, SerializingLocation{
			URI:            location.URI,
			StartLine:      location.StartLine,
			StartCharacter: location.StartCharacter,
			EndLine:        location.EndLine,
			EndCharacter:   location.EndCharacter,
		})
	}

	encoded, err := json.Marshal(serializingLocations)
	if err != nil {
		return nil, err
	}

	return compress(encoded)
}

func (*jsonSerializer) UnmarshalDocumentData(data []byte) (types.DocumentData, error) {
	var payload SerializedDocument
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

func (*jsonSerializer) UnmarshalResultChunkData(data []byte) (types.ResultChunkData, error) {
	var payload SerializedResultChunk
	if err := unmarshalGzippedJSON(data, &payload); err != nil {
		return types.ResultChunkData{}, err
	}

	documentPaths, err := unmarshalWrappedDocumentPaths(payload.DocumentPaths.Value)
	if err != nil {
		return types.ResultChunkData{}, err
	}

	documentIDRangeIDs, err := unmarshalWrappedDocumentIDRangeIDs(payload.DocumentIDRangeIDs.Value)
	if err != nil {
		return types.ResultChunkData{}, err
	}

	return types.ResultChunkData{
		DocumentPaths:      documentPaths,
		DocumentIDRangeIDs: documentIDRangeIDs,
	}, nil
}

func (*jsonSerializer) UnmarshalLocations(data []byte) ([]types.Location, error) {
	var payload []SerializedLocation
	if err := unmarshalGzippedJSON(data, &payload); err != nil {
		return nil, err
	}

	locations := make([]types.Location, 0, len(payload))
	for _, location := range payload {
		locations = append(locations, types.Location{
			URI:            location.URI,
			StartLine:      location.StartLine,
			StartCharacter: location.StartCharacter,
			EndLine:        location.EndLine,
			EndCharacter:   location.EndCharacter,
		})
	}

	return locations, nil
}

func unmarshalWrappedRanges(pairs []json.RawMessage) (map[types.ID]types.RangeData, error) {
	m := map[types.ID]types.RangeData{}
	for _, pair := range pairs {
		var id ID
		var value SerializedRange

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		var monikerIDs []types.ID
		for _, value := range value.MonikerIDs.Value {
			var id ID
			if err := json.Unmarshal([]byte(value), &id); err != nil {
				return nil, err
			}

			monikerIDs = append(monikerIDs, types.ID(id))
		}

		m[types.ID(id)] = types.RangeData{
			StartLine:          value.StartLine,
			StartCharacter:     value.StartCharacter,
			EndLine:            value.EndLine,
			EndCharacter:       value.EndCharacter,
			DefinitionResultID: types.ID(value.DefinitionResultID),
			ReferenceResultID:  types.ID(value.ReferenceResultID),
			HoverResultID:      types.ID(value.HoverResultID),
			MonikerIDs:         monikerIDs,
		}
	}

	return m, nil
}

func unmarshalWrappedHoverResults(pairs []json.RawMessage) (map[types.ID]string, error) {
	m := map[types.ID]string{}
	for _, pair := range pairs {
		var id ID
		var value string

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		m[types.ID(id)] = value
	}

	return m, nil
}

func unmarshalWrappedMonikers(pairs []json.RawMessage) (map[types.ID]types.MonikerData, error) {
	m := map[types.ID]types.MonikerData{}
	for _, pair := range pairs {
		var id ID
		var value SerializedMoniker

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		m[types.ID(id)] = types.MonikerData{
			Kind:                 value.Kind,
			Scheme:               value.Scheme,
			Identifier:           value.Identifier,
			PackageInformationID: types.ID(value.PackageInformationID),
		}
	}

	return m, nil
}

func unmarshalWrappedPackageInformation(pairs []json.RawMessage) (map[types.ID]types.PackageInformationData, error) {
	m := map[types.ID]types.PackageInformationData{}
	for _, pair := range pairs {
		var id ID
		var value SerializedPackageInformation

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		m[types.ID(id)] = types.PackageInformationData{
			Name:    value.Name,
			Version: value.Version,
		}
	}

	return m, nil
}

func unmarshalWrappedDocumentPaths(pairs []json.RawMessage) (map[types.ID]string, error) {
	m := map[types.ID]string{}
	for _, pair := range pairs {
		var id ID
		var value string

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		m[types.ID(id)] = value
	}

	return m, nil
}

func unmarshalWrappedDocumentIDRangeIDs(pairs []json.RawMessage) (map[types.ID][]types.DocumentIDRangeID, error) {
	m := map[types.ID][]types.DocumentIDRangeID{}
	for _, pair := range pairs {
		var id ID
		var value []SerializedDocumentIDRangeID

		target := []interface{}{&id, &value}
		if err := json.Unmarshal([]byte(pair), &target); err != nil {
			return nil, err
		}

		var documentIDRangeIDs []types.DocumentIDRangeID
		for _, v := range value {
			documentIDRangeIDs = append(documentIDRangeIDs, types.DocumentIDRangeID{
				DocumentID: types.ID(v.DocumentID),
				RangeID:    types.ID(v.RangeID),
			})
		}

		m[types.ID(id)] = documentIDRangeIDs
	}

	return m, nil
}

type ID string

func (id *ID) UnmarshalJSON(raw []byte) error {
	if raw[0] == '"' {
		var v string
		if err := json.Unmarshal(raw, &v); err != nil {
			return err
		}

		*id = ID(v)
		return nil
	}

	var v int64
	if err := json.Unmarshal(raw, &v); err != nil {
		return err
	}

	*id = ID(strconv.FormatInt(v, 10))
	return nil
}
