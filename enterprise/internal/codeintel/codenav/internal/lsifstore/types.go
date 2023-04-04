package lsifstore

import (
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
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

func translateRange(r *scip.Range) shared.Range {
	return shared.Range{
		Start: shared.Position{
			Line:      int(r.Start.Line),
			Character: int(r.Start.Character),
		},
		End: shared.Position{
			Line:      int(r.End.Line),
			Character: int(r.End.Character),
		},
	}
}
