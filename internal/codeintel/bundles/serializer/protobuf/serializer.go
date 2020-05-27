package protobuf

import (
	"github.com/gogo/protobuf/proto"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/serializer"
	bundleproto "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/serializer/protobuf/proto"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

type protobufSerializer struct{}

var _ serializer.Serializer = &protobufSerializer{}

func New() serializer.Serializer {
	return &protobufSerializer{}
}

// MarshalDocumentData transforms document data into a string of bytes writable to disk.
func (*protobufSerializer) MarshalDocumentData(d types.DocumentData) ([]byte, error) {
	ranges := make(map[string]*bundleproto.RangeData, len(d.Ranges))
	for key, value := range d.Ranges {
		monikerIDs := make([]string, 0, len(value.MonikerIDs))
		for _, monikerID := range value.MonikerIDs {
			monikerIDs = append(monikerIDs, string(monikerID))
		}

		ranges[string(key)] = &bundleproto.RangeData{
			StartLine:          int32(value.StartLine),
			StartCharacter:     int32(value.StartCharacter),
			EndLine:            int32(value.EndLine),
			EndCharacter:       int32(value.EndCharacter),
			DefinitionResultID: string(value.DefinitionResultID),
			ReferenceResultID:  string(value.ReferenceResultID),
			HoverResultID:      string(value.HoverResultID),
			MonikerIDs:         monikerIDs,
		}
	}

	hoverResults := make(map[string]string, len(d.HoverResults))
	for key, value := range d.HoverResults {
		hoverResults[string(key)] = value
	}

	monikers := make(map[string]*bundleproto.MonikerData, len(d.Monikers))
	for key, value := range d.Monikers {
		monikers[string(key)] = &bundleproto.MonikerData{
			Kind:                 value.Kind,
			Scheme:               value.Scheme,
			Identifier:           value.Identifier,
			PackageInformationID: string(value.PackageInformationID),
		}
	}

	packageInformation := make(map[string]*bundleproto.PackageInformationData, len(d.PackageInformation))
	for key, value := range d.PackageInformation {
		packageInformation[string(key)] = &bundleproto.PackageInformationData{
			Name:    value.Name,
			Version: value.Version,
		}
	}

	return proto.Marshal(&bundleproto.DocumentData{
		Ranges:             ranges,
		HoverResults:       hoverResults,
		Monikers:           monikers,
		PackageInformation: packageInformation,
	})
}

// MarshalResultChunkData transforms result chunk data into a string of bytes writable to disk.
func (*protobufSerializer) MarshalResultChunkData(rc types.ResultChunkData) ([]byte, error) {
	documentPaths := make(map[string]string, len(rc.DocumentPaths))
	for key, value := range rc.DocumentPaths {
		documentPaths[string(key)] = value
	}

	documentIDRangeIDs := make(map[string]*bundleproto.DocumentIDRangeIDs, len(rc.DocumentIDRangeIDs))
	for key, values := range rc.DocumentIDRangeIDs {
		// TODO - rename/restructure
		q := make([]*bundleproto.DocumentIDRangeID, 0, len(values))

		for _, value := range values {
			q = append(q, &bundleproto.DocumentIDRangeID{
				DocumentID: string(value.DocumentID),
				RangeID:    string(value.RangeID),
			})
		}

		documentIDRangeIDs[string(key)] = &bundleproto.DocumentIDRangeIDs{
			DocumentIDRangeID: q,
		}
	}

	return proto.Marshal(&bundleproto.ResultChunkData{
		DocumentPaths:      documentPaths,
		DocumentIDRangeIDs: documentIDRangeIDs,
	})
}

// UnmarshalDocumentData is the inverse of MarshalDocumentData.
func (*protobufSerializer) UnmarshalDocumentData(data []byte) (types.DocumentData, error) {
	var element bundleproto.DocumentData
	if err := proto.Unmarshal(data, &element); err != nil {
		return types.DocumentData{}, err
	}

	ranges := make(map[types.ID]types.RangeData, len(element.Ranges))
	for key, value := range element.Ranges {
		monikerIDs := make([]types.ID, 0, len(value.MonikerIDs))
		for _, monikerID := range value.MonikerIDs {
			monikerIDs = append(monikerIDs, types.ID(monikerID))
		}

		ranges[types.ID(key)] = types.RangeData{
			StartLine:          int(value.StartLine),
			StartCharacter:     int(value.StartCharacter),
			EndLine:            int(value.EndLine),
			EndCharacter:       int(value.EndCharacter),
			DefinitionResultID: types.ID(value.DefinitionResultID),
			ReferenceResultID:  types.ID(value.ReferenceResultID),
			HoverResultID:      types.ID(value.HoverResultID),
			MonikerIDs:         monikerIDs,
		}
	}

	hoverResults := make(map[types.ID]string, len(element.HoverResults))
	for key, value := range element.HoverResults {
		hoverResults[types.ID(key)] = value
	}

	monikers := make(map[types.ID]types.MonikerData, len(element.Monikers))
	for key, value := range element.Monikers {
		monikers[types.ID(key)] = types.MonikerData{
			Kind:                 value.Kind,
			Scheme:               value.Scheme,
			Identifier:           value.Identifier,
			PackageInformationID: types.ID(value.PackageInformationID),
		}
	}

	packageInformation := make(map[types.ID]types.PackageInformationData, len(element.PackageInformation))
	for key, value := range element.PackageInformation {
		packageInformation[types.ID(key)] = types.PackageInformationData{
			Name:    value.Name,
			Version: value.Version,
		}
	}

	return types.DocumentData{
		Ranges:             ranges,
		HoverResults:       hoverResults,
		Monikers:           monikers,
		PackageInformation: packageInformation,
	}, nil
}

// UnmarshalResultChunkData is the inverse of MarshalResultChunkData.
func (*protobufSerializer) UnmarshalResultChunkData(data []byte) (types.ResultChunkData, error) {
	var element bundleproto.ResultChunkData
	if err := proto.Unmarshal(data, &element); err != nil {
		return types.ResultChunkData{}, err
	}

	documentPaths := make(map[types.ID]string, len(element.DocumentPaths))
	for key, value := range element.DocumentPaths {
		documentPaths[types.ID(key)] = value
	}

	documentIDRangeIDs := make(map[types.ID][]types.DocumentIDRangeID, len(element.DocumentIDRangeIDs))
	for key, value := range element.DocumentIDRangeIDs {
		qs := make([]types.DocumentIDRangeID, 0)

		for _, value := range value.DocumentIDRangeID {
			// TODO - rename/restructure
			qs = append(qs, types.DocumentIDRangeID{
				DocumentID: types.ID(value.DocumentID),
				RangeID:    types.ID(value.RangeID),
			})
		}

		documentIDRangeIDs[types.ID(key)] = qs
	}

	return types.ResultChunkData{
		DocumentPaths:      documentPaths,
		DocumentIDRangeIDs: documentIDRangeIDs,
	}, nil
}
