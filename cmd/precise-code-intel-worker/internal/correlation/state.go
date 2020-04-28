package correlation

import (
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/datastructures"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/lsif"
)

// State is an in-memory representation of an uploaded LSIF index.
type State struct {
	LSIFVersion            string
	ProjectRoot            string
	DocumentData           map[string]lsif.DocumentData
	RangeData              map[string]lsif.RangeData
	ResultSetData          map[string]lsif.ResultSetData
	DefinitionData         map[string]datastructures.DefaultIDSetMap
	ReferenceData          map[string]datastructures.DefaultIDSetMap
	HoverData              map[string]string
	MonikerData            map[string]lsif.MonikerData
	PackageInformationData map[string]lsif.PackageInformationData
	NextData               map[string]string
	ImportedMonikers       datastructures.IDSet
	ExportedMonikers       datastructures.IDSet
	LinkedMonikers         datastructures.DisjointIDSet
	LinkedReferenceResults datastructures.DisjointIDSet
}

// newState create a new State with zero-valued map fields.
func newState() *State {
	return &State{
		DocumentData:           map[string]lsif.DocumentData{},
		RangeData:              map[string]lsif.RangeData{},
		ResultSetData:          map[string]lsif.ResultSetData{},
		DefinitionData:         map[string]datastructures.DefaultIDSetMap{},
		ReferenceData:          map[string]datastructures.DefaultIDSetMap{},
		HoverData:              map[string]string{},
		MonikerData:            map[string]lsif.MonikerData{},
		PackageInformationData: map[string]lsif.PackageInformationData{},
		NextData:               map[string]string{},
		ImportedMonikers:       datastructures.IDSet{},
		ExportedMonikers:       datastructures.IDSet{},
		LinkedMonikers:         datastructures.DisjointIDSet{},
		LinkedReferenceResults: datastructures.DisjointIDSet{},
	}
}
