package correlation

import (
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/correlation/datastructures"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/correlation/lsif"
)

// State is an in-memory representation of an uploaded LSIF index.
type State struct {
	LSIFVersion            string
	ProjectRoot            string
	DocumentData           map[int]lsif.Document
	RangeData              map[int]lsif.Range
	ResultSetData          map[int]lsif.ResultSet
	DefinitionData         map[int]datastructures.DefaultIDSetMap
	ReferenceData          map[int]datastructures.DefaultIDSetMap
	HoverData              map[int]string
	MonikerData            map[int]lsif.Moniker
	PackageInformationData map[int]lsif.PackageInformation
	Diagnostics            map[int]lsif.DiagnosticResult
	NextData               map[int]int                  // maps vertices related via next edges
	ImportedMonikers       *datastructures.IDSet        // moniker ids that have kind "import"
	ExportedMonikers       *datastructures.IDSet        // moniker ids that have kind "export"
	LinkedMonikers         datastructures.DisjointIDSet // tracks which moniker ids are related via next edges
	LinkedReferenceResults datastructures.DisjointIDSet // tracks which reference result ids are related via next edges
}

// newState create a new State with zero-valued map fields.
func newState() *State {
	return &State{
		DocumentData:           map[int]lsif.Document{},
		RangeData:              map[int]lsif.Range{},
		ResultSetData:          map[int]lsif.ResultSet{},
		DefinitionData:         map[int]datastructures.DefaultIDSetMap{},
		ReferenceData:          map[int]datastructures.DefaultIDSetMap{},
		HoverData:              map[int]string{},
		MonikerData:            map[int]lsif.Moniker{},
		PackageInformationData: map[int]lsif.PackageInformation{},
		Diagnostics:            map[int]lsif.DiagnosticResult{},
		NextData:               map[int]int{},
		ImportedMonikers:       datastructures.NewIDSet(),
		ExportedMonikers:       datastructures.NewIDSet(),
		LinkedMonikers:         datastructures.DisjointIDSet{},
		LinkedReferenceResults: datastructures.DisjointIDSet{},
	}
}
