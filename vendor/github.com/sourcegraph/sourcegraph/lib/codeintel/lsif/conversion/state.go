package conversion

import (
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/conversion/datastructures"
)

// State is an in-memory representation of an uploaded LSIF index.
type State struct {
	LSIFVersion            string                                  // The LSIF version of this dump. This is unused.
	ProjectRoot            string                                  // The root of all files in this dump (e.g. `file:///`). Values of DocumentData are relative to this.
	DocumentData           map[int]string                          // maps document ID -> path relative to the project root
	RangeData              map[int]Range                           // maps range ID -> Range (which has start/end line/character and *Result IDs)
	ResultSetData          map[int]ResultSet                       // maps resultSet ID -> ResultSet (which has *Result IDs)
	DefinitionData         map[int]*datastructures.DefaultIDSetMap // maps definitionResult ID -> document ID -> range ID
	ReferenceData          map[int]*datastructures.DefaultIDSetMap // maps referenceResult ID -> document ID -> range ID
	ImplementationData     map[int]*datastructures.DefaultIDSetMap // maps implementationResult ID -> document ID -> range ID
	HoverData              map[int]string                          // maps hoverResult ID -> hover string
	MonikerData            map[int]Moniker                         // maps moniker ID -> Moniker (which has kind, scheme, identifier, and packageInformation ID)
	PackageInformationData map[int]PackageInformation              // maps packageInformation ID -> PackageInformation (which has name and version)
	DiagnosticResults      map[int][]Diagnostic                    // maps diagnosticResult ID -> []Diagnostic
	NextData               map[int]int                             // maps (range ID | resultSet ID) -> resultSet ID related via next edges
	ImportedMonikers       *datastructures.IDSet                   // set of moniker IDs that have kind "import"
	ExportedMonikers       *datastructures.IDSet                   // set of moniker IDs that have kind "export"
	ImplementedMonikers    *datastructures.IDSet                   // set of moniker IDs that have kind "implementation"
	LinkedMonikers         *datastructures.DisjointIDSet           // tracks which moniker IDs are related via next edges
	LinkedReferenceResults map[int][]int                           // tracks which referenceResult IDs are related via item edges
	Monikers               *datastructures.DefaultIDSetMap         // maps (range ID | resultSet ID) -> moniker IDs
	Contains               *datastructures.DefaultIDSetMap         // maps document ID -> range IDs that are contained in the document
	Diagnostics            *datastructures.DefaultIDSetMap         // maps document ID -> diagnostic IDs
}

// newState create a new State with zero-valued map fields.
func newState() *State {
	return &State{
		DocumentData:           map[int]string{},
		RangeData:              map[int]Range{},
		ResultSetData:          map[int]ResultSet{},
		DefinitionData:         map[int]*datastructures.DefaultIDSetMap{},
		ReferenceData:          map[int]*datastructures.DefaultIDSetMap{},
		ImplementationData:     map[int]*datastructures.DefaultIDSetMap{},
		HoverData:              map[int]string{},
		MonikerData:            map[int]Moniker{},
		PackageInformationData: map[int]PackageInformation{},
		DiagnosticResults:      map[int][]Diagnostic{},
		NextData:               map[int]int{},
		ImportedMonikers:       datastructures.NewIDSet(),
		ExportedMonikers:       datastructures.NewIDSet(),
		ImplementedMonikers:    datastructures.NewIDSet(),
		LinkedMonikers:         datastructures.NewDisjointIDSet(),
		LinkedReferenceResults: map[int][]int{},
		Monikers:               datastructures.NewDefaultIDSetMap(),
		Contains:               datastructures.NewDefaultIDSetMap(),
		Diagnostics:            datastructures.NewDefaultIDSetMap(),
	}
}
