package sourcegraph

import (
	"fmt"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

func (s *DefSpec) RouteVars() map[string]string {
	m := map[string]string{"Repo": s.Repo, "UnitType": s.UnitType, "Unit": s.Unit, "Path": s.Path}
	if s.CommitID != "" {
		m["Rev"] = s.CommitID
	}
	return m
}

// DefKey returns the def key specified by s, using the Repo, UnitType,
// Unit, and Path fields of s.
func (s *DefSpec) DefKey() graph.DefKey {
	if s.Repo == "" {
		panic("Repo is empty")
	}
	if s.UnitType == "" {
		panic("UnitType is empty")
	}
	if s.Unit == "" {
		panic("Unit is empty")
	}
	return graph.DefKey{
		Repo:     s.Repo,
		CommitID: s.CommitID,
		UnitType: s.UnitType,
		Unit:     s.Unit,
		Path:     s.Path,
	}
}

// NewDefSpecFromDefKey returns a DefSpec that specifies the same
// def as the given key.
func NewDefSpecFromDefKey(key graph.DefKey) DefSpec {
	return DefSpec{
		Repo:     key.Repo,
		CommitID: key.CommitID,
		UnitType: key.UnitType,
		Unit:     key.Unit,
		Path:     key.Path,
	}
}

// DefSpec returns the DefSpec that specifies s.
func (s *Def) DefSpec() DefSpec {
	spec := NewDefSpecFromDefKey(s.Def.DefKey)
	return spec
}

type Refs []*Ref

func (r *Ref) sortKey() string     { return fmt.Sprintf("%+v", r) }
func (vs Refs) Len() int           { return len(vs) }
func (vs Refs) Swap(i, j int)      { vs[i], vs[j] = vs[j], vs[i] }
func (vs Refs) Less(i, j int) bool { return vs[i].sortKey() < vs[j].sortKey() }

type Examples []*Example

func (r *Example) sortKey() string     { return fmt.Sprintf("%+v", r) }
func (vs Examples) Len() int           { return len(vs) }
func (vs Examples) Swap(i, j int)      { vs[i], vs[j] = vs[j], vs[i] }
func (vs Examples) Less(i, j int) bool { return vs[i].sortKey() < vs[j].sortKey() }

type DefAuthorsByBytes []*DefAuthor

func (v DefAuthorsByBytes) Len() int           { return len(v) }
func (v DefAuthorsByBytes) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v DefAuthorsByBytes) Less(i, j int) bool { return v[i].Bytes < v[j].Bytes }
