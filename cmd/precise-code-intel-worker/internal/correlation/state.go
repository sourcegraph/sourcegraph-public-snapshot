package correlation

import (
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/datastructures"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/lsif"
)

// State is an in-memory representation of an uploaded LSIF index.
type State struct {
	LSIFVersion            string
	ProjectRoot            string
	DocumentData           map[string]lsif.Document
	RangeData              map[string]lsif.Range
	ResultSetData          map[string]lsif.ResultSet
	DefinitionData         map[string]datastructures.DefaultIDSetMap
	ReferenceData          map[string]datastructures.DefaultIDSetMap
	HoverData              map[string]string
	MonikerData            map[string]lsif.Moniker
	PackageInformationData map[string]lsif.PackageInformation
	NextData               map[string]string            // maps vertices related via next edges
	ImportedMonikers       datastructures.IDSet         // moniker ids that have kind "import"
	ExportedMonikers       datastructures.IDSet         // moniker ids that have kind "export"
	LinkedMonikers         datastructures.DisjointIDSet // tracks which moniker ids are related via next edges
	LinkedReferenceResults datastructures.DisjointIDSet // tracks which reference result ids are related via next edges
}

// newState create a new State with zero-valued map fields.
func newState() *State {
	return &State{
		DocumentData:           map[string]lsif.Document{},
		RangeData:              map[string]lsif.Range{},
		ResultSetData:          map[string]lsif.ResultSet{},
		DefinitionData:         map[string]datastructures.DefaultIDSetMap{},
		ReferenceData:          map[string]datastructures.DefaultIDSetMap{},
		HoverData:              map[string]string{},
		MonikerData:            map[string]lsif.Moniker{},
		PackageInformationData: map[string]lsif.PackageInformation{},
		NextData:               map[string]string{},
		ImportedMonikers:       datastructures.IDSet{},
		ExportedMonikers:       datastructures.IDSet{},
		LinkedMonikers:         datastructures.DisjointIDSet{},
		LinkedReferenceResults: datastructures.DisjointIDSet{},
	}
}
