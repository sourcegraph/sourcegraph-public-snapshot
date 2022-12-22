package lsifstore

import (
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

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
	Path     string
	LSIFData *precise.DocumentData
	SCIPData *scip.Document
}

func translateRange(r *scip.Range) types.Range {
	return types.Range{
		Start: types.Position{
			Line:      int(r.Start.Line),
			Character: int(r.Start.Character),
		},
		End: types.Position{
			Line:      int(r.End.Line),
			Character: int(r.End.Character),
		},
	}
}
