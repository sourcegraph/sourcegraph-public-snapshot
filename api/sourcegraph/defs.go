package sourcegraph

import "sourcegraph.com/sourcegraph/srclib/graph"

// NewDefSpecFromDefKey returns a DefSpec that specifies the same def
// as the given key. The caller must provide the numeric repo ID
// corresponding to key.Repo (using the Repos.Resolve API method).
func NewDefSpecFromDefKey(key graph.DefKey, repo int32) DefSpec {
	return DefSpec{
		Repo:     repo,
		CommitID: key.CommitID,
		UnitType: key.UnitType,
		Unit:     key.Unit,
		Path:     key.Path,
	}
}

// DefSpec returns the DefSpec that specifies s. The caller must
// provide the numeric repo ID corresponding to s.Repo.
func (s *Def) DefSpec(repo int32) DefSpec {
	spec := NewDefSpecFromDefKey(s.Def.DefKey, repo)
	spec.Repo = repo
	return spec
}

// DefKey returns the def key specified by s, using the UnitType,
// Unit, and Path fields of s. The caller must provide the string repo
// path corresponding to s.Repo.
func (s *DefSpec) DefKey(repo string) graph.DefKey {
	if s.Repo == 0 {
		panic("Repo is empty")
	}
	if s.UnitType == "" {
		panic("UnitType is empty")
	}
	if s.Unit == "" {
		panic("Unit is empty")
	}
	return graph.DefKey{
		Repo:     repo,
		CommitID: s.CommitID,
		UnitType: s.UnitType,
		Unit:     s.Unit,
		Path:     s.Path,
	}
}
