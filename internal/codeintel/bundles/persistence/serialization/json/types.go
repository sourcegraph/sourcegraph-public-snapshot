package json

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

//
// The following types are used during marshalling

type SerializingTaggedValue struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type SerializingRange struct {
	StartLine          int                    `json:"startLine"`
	StartCharacter     int                    `json:"startCharacter"`
	EndLine            int                    `json:"endLine"`
	EndCharacter       int                    `json:"endCharacter"`
	DefinitionResultID types.ID               `json:"definitionResultId"`
	ReferenceResultID  types.ID               `json:"referenceResultId"`
	HoverResultID      types.ID               `json:"hoverResultId"`
	MonikerIDs         SerializingTaggedValue `json:"monikerIds"`
}

type SerializingDocument struct {
	Ranges             SerializingTaggedValue `json:"ranges"`
	HoverResults       SerializingTaggedValue `json:"hoverResults"`
	Monikers           SerializingTaggedValue `json:"monikers"`
	PackageInformation SerializingTaggedValue `json:"packageInformation"`
}

type SerializingResultChunk struct {
	DocumentPaths      SerializingTaggedValue `json:"documentPaths"`
	DocumentIDRangeIDs SerializingTaggedValue `json:"documentIdRangeIds"`
}

type SerializingLocation struct {
	URI            string `json:"uri"`
	StartLine      int    `json:"startLine"`
	StartCharacter int    `json:"startCharacter"`
	EndLine        int    `json:"endLine"`
	EndCharacter   int    `json:"endCharacter"`
}

//
// The following types are used during unmarshalling

type SerializedTaggedValue struct {
	Type  string            `json:"type"`
	Value []json.RawMessage `json:"value"`
}

type SerializedRange struct {
	StartLine          int                   `json:"startLine"`
	StartCharacter     int                   `json:"startCharacter"`
	EndLine            int                   `json:"endLine"`
	EndCharacter       int                   `json:"endCharacter"`
	DefinitionResultID ID                    `json:"definitionResultId"`
	ReferenceResultID  ID                    `json:"referenceResultId"`
	HoverResultID      ID                    `json:"hoverResultId"`
	MonikerIDs         SerializedTaggedValue `json:"monikerIds"`
}

type SerializedDocument struct {
	Ranges             SerializedTaggedValue `json:"ranges"`
	HoverResults       SerializedTaggedValue `json:"hoverResults"`
	Monikers           SerializedTaggedValue `json:"monikers"`
	PackageInformation SerializedTaggedValue `json:"packageInformation"`
}

type SerializedResultChunk struct {
	DocumentPaths      SerializedTaggedValue `json:"documentPaths"`
	DocumentIDRangeIDs SerializedTaggedValue `json:"documentIdRangeIds"`
}

type SerializedLocation = SerializingLocation

type SerializedMoniker struct {
	Kind                 string `json:"kind"`
	Scheme               string `json:"scheme"`
	Identifier           string `json:"identifier"`
	PackageInformationID ID     `json:"packageInformationId"`
}

type SerializedPackageInformation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type SerializedDocumentIDRangeID struct {
	DocumentID ID `json:"documentId"`
	RangeID    ID `json:"rangeId"`
}
