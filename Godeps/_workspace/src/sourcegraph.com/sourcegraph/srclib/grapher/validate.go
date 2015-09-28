package grapher

import (
	"fmt"

	"strings"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

func ValidateRefs(refs []*graph.Ref) (errs MultiError) {
	refKeys := make(map[graph.RefKey]struct{})
	for _, ref := range refs {
		key := ref.RefKey()
		if _, in := refKeys[key]; in {
			errs = append(errs, fmt.Errorf("duplicate ref key: %+v", key))
		} else {
			refKeys[key] = struct{}{}
		}
	}
	return
}

func ValidateDefs(defs []*graph.Def) (errs MultiError) {
	defKeys := make(map[graph.DefKey]struct{})
	for _, def := range defs {
		key := def.DefKey
		if _, in := defKeys[key]; in {
			errs = append(errs, fmt.Errorf("duplicate def key: %+v", key))
		} else {
			defKeys[key] = struct{}{}
		}
	}
	return
}

func ValidateDocs(docs []*graph.Doc) (errs MultiError) {
	docKeys := make(map[graph.DocKey]struct{})
	for _, doc := range docs {
		key := doc.Key()
		if _, in := docKeys[key]; in {
			errs = append(errs, fmt.Errorf("duplicate doc key: %+v", key))
		} else {
			docKeys[key] = struct{}{}
		}
		// TODO(samer): check that Start and End do not equal each
		// other if DefKey is empty in linter.
	}
	return
}

type MultiError []error

func (e MultiError) Error() string {
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "\n")
}

// UnresolvedInternalRefs returns a map of unresolved internal refs,
// keyed on the (nonexistent) defs they point to. CurrentRepoURI must
// be the repo URI of the repo the refs and defs were built from.
// CurrentRepoURI may be empty. It is used to determine whether a ref
// is an internal ref or not. Only internal refs can be checked in
// this way because checking resolution to external defs would require
// loading external data, which is outside the scope of this function.
func UnresolvedInternalRefs(currentRepoURI string, refs []*graph.Ref, defs []*graph.Def) map[graph.DefKey][]*graph.Ref {
	defKeys := map[graph.DefKey]*graph.Def{}
	for _, def := range defs {
		// Remove CommitID because the internal refs don't have a
		// DefCommitID (it's implied).
		defKeyInMapKey := def.DefKey
		defKeyInMapKey.CommitID = ""

		defKeys[defKeyInMapKey] = def
	}

	unresolvedInternalRefsByDefKey := map[graph.DefKey][]*graph.Ref{}
	for _, ref := range refs {
		// We can only check internal refs easily here, since we've
		// pulled only the data we need to do so already. Checking
		// xrefs would also be useful but it would take a lot longer
		// and require fetching external data.
		if graph.URIEqual(ref.DefRepo, currentRepoURI) {
			defKey := ref.DefKey()
			if _, resolved := defKeys[defKey]; !resolved {
				unresolvedInternalRefsByDefKey[defKey] = append(unresolvedInternalRefsByDefKey[defKey], ref)
			}
		}
	}

	return unresolvedInternalRefsByDefKey
}

// PopulateImpliedFields fills in fields on graph data objects that
// individual toolchains leave blank but that are implied by the
// source unit the graph data objects were built from.
func PopulateImpliedFields(repo, commitID, unitType, unit string, o *graph.Output) {
	for _, def := range o.Defs {
		def.UnitType = unitType
		def.Unit = unit
		def.Repo = repo
		def.CommitID = commitID
		if len(def.Data) == 0 {
			def.Data = []byte(`{}`)
		}
	}
	for _, ref := range o.Refs {
		ref.Repo = repo
		ref.UnitType = unitType
		ref.Unit = unit
		ref.CommitID = commitID

		// Treat an empty repository URI as referring to the current
		// repository.
		if ref.DefRepo == "" {
			ref.DefRepo = repo
			if ref.DefUnit == "" {
				ref.DefUnitType = unitType
				ref.DefUnit = unit
			}
		}
		if ref.DefUnitType == "" {
			// default DefUnitType to same unit type as the ref itself
			ref.DefUnitType = unitType
		}
	}
	for _, doc := range o.Docs {
		doc.UnitType = unitType
		doc.Unit = unit
		doc.Repo = repo
		doc.CommitID = commitID
	}

	for _, ann := range o.Anns {
		ann.UnitType = unitType
		ann.Unit = unit
		ann.Repo = repo
		ann.CommitID = commitID
	}
}
