package lsifstore

import "github.com/sourcegraph/sourcegraph/lib/codeintel/precise"

type MarshalledDocumentData struct {
	Ranges             []byte
	HoverResults       []byte
	Monikers           []byte
	PackageInformation []byte
	Diagnostics        []byte
}

type QualifiedMonikerLocations struct {
	DumpID int
	precise.MonikerLocations
}

type QualifiedDocumentData struct {
	UploadID int
	precise.KeyedDocumentData
}
