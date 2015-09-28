package graph

// StatType is the name of a def statistic (see below for a listing).
type StatType string

// Stats holds statistics for a def.
type Stats map[StatType]int

const (
	// StatXRefs is the number of external references to a def (i.e.,
	// references from other repositories). It is only computed for abstract
	// defs (see the docs for DefKey) because it is not easy to determine
	// which specific commit a ref references (for external refs).
	StatXRefs = "xrefs"

	// StatRRefs is the number of references to a def from the same
	// repository in which the def is defined. It is inclusive of the
	// StatURefs count. It is only computed for concrete defs (see the docs
	// for DefKey) because otherwise it would count 1 rref for each unique
	// revision of the repository that we have processed. (It is easy to
	// determine which specific commit an internal ref references; we just
	// assume it references a def in the same commit.)
	StatRRefs = "rrefs"

	// StatURefs is the number of references to a def from the same source
	// unit in which the def is defined. It is included in the StatRRefs
	// count. It is only computed for concrete defs (see the docs for
	// DefKey) because otherwise it would count 1 uref for each revision of
	// the repository that we have processed.
	StatURefs = "urefs"

	// StatAuthors is the number of distinct resolved people who contributed
	// code to a def's definition (according to a VCS "blame" of the
	// version). It is only computed for concrete defs (see the docs for
	// DefKey).
	StatAuthors = "authors"

	// StatClients is the number of distinct resolved people who have committed
	// refs that reference a def. It is only computed for abstract defs
	// (see the docs for DefKey) because it is not easy to determine which
	// specific commit a ref references.
	StatClients = "clients"

	// StatDependents is the number of distinct repositories that contain refs
	// that reference a def. It is only computed for abstract defs (see
	// the docs for DefKey) because it is not easy to determine which
	// specific commit a ref references.
	StatDependents = "dependents"

	// StatExportedElements is the number of exported defs whose path is a
	// descendant of this def's path (and that is in the same repository and
	// source unit). It is only computed for concrete defs (see the docs for
	// DefKey) because otherwise it would count 1 exported element for each
	// revision of the repository that we have processed.
	StatExportedElements = "exported_elements"
)

var AllStatTypes = []StatType{StatXRefs, StatRRefs, StatURefs, StatAuthors, StatClients, StatDependents, StatExportedElements}

func (x StatType) IsAbstract() bool {
	switch x {
	case StatXRefs:
		fallthrough
	case StatClients:
		fallthrough
	case StatDependents:
		return true
	default:
		return false
	}
}

// UniqueRefDefs groups refs by the RefDefKey field and returns a map of
// how often each RefDefKey appears. If m is non-nil, counts are incremented
// and a new map is not created.
func UniqueRefDefs(refs []*Ref, m map[RefDefKey]int) map[RefDefKey]int {
	if m == nil {
		m = make(map[RefDefKey]int)
	}
	for _, ref := range refs {
		m[ref.RefDefKey()]++
	}
	return m
}
