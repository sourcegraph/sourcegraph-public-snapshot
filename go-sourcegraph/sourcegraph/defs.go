package sourcegraph

import (
	"fmt"
	"log"
	"path"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/store"
	"sourcegraph.com/sourcegraph/srclib/unit"
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

func (o *DefListOptions) DefFilters() []store.DefFilter {
	var fs []store.DefFilter
	if o.DefKeys != nil {
		fs = append(fs, store.DefFilterFunc(func(def *graph.Def) bool {
			for _, dk := range o.DefKeys {
				if (def.Repo == "" || def.Repo == dk.Repo) && (def.CommitID == "" || def.CommitID == dk.CommitID) &&
					(def.UnitType == "" || def.UnitType == dk.UnitType) && (def.Unit == "" || def.Unit == dk.Unit) &&
					def.Path == dk.Path {
					return true
				}
			}
			return false
		}))
	}
	if o.Name != "" {
		fs = append(fs, store.DefFilterFunc(func(def *graph.Def) bool {
			return def.Name == o.Name
		}))
	}
	if o.ByteEnd != 0 {
		fs = append(fs, store.DefFilterFunc(func(d *graph.Def) bool {
			return d.DefStart == o.ByteStart && d.DefEnd == o.ByteEnd
		}))
	}
	if o.Query != "" {
		fs = append(fs, store.ByDefQuery(o.Query))
	}
	if len(o.RepoRevs) > 0 {
		vs := make([]store.Version, len(o.RepoRevs))
		for i, repoRev := range o.RepoRevs {
			repo, commitID := ParseRepoAndCommitID(repoRev)
			if len(commitID) != 40 {
				log.Printf("WARNING: In DefListOptions.DefFilters, o.RepoRevs[%d]==%q has no commit ID or a non-absolute commit ID. No defs will match it.", i, repoRev)
			}
			vs[i] = store.Version{Repo: repo, CommitID: commitID}
		}
		fs = append(fs, store.ByRepoCommitIDs(vs...))
	}
	if o.Unit != "" && o.UnitType != "" {
		fs = append(fs, store.ByUnits(unit.ID2{Type: o.UnitType, Name: o.Unit}))
	}
	if (o.UnitType != "" && o.Unit == "") || (o.UnitType == "" && o.Unit != "") {
		log.Println("WARNING: DefListOptions.DefFilter: must specify either both or neither of --type and --name (to filter by source unit)")
	}
	if o.File != "" {
		fs = append(fs, store.ByFiles(path.Clean(o.File)))
	}
	if o.FilePathPrefix != "" {
		fs = append(fs, store.ByFiles(path.Clean(o.FilePathPrefix)))
	}
	if len(o.Kinds) > 0 {
		fs = append(fs, store.DefFilterFunc(func(def *graph.Def) bool {
			for _, kind := range o.Kinds {
				if def.Kind == kind {
					return true
				}
			}
			return false
		}))
	}
	if o.Exported {
		fs = append(fs, store.DefFilterFunc(func(def *graph.Def) bool {
			return def.Exported
		}))
	}
	if o.Nonlocal {
		fs = append(fs, store.DefFilterFunc(func(def *graph.Def) bool {
			return !def.Local
		}))
	}
	if !o.IncludeTest {
		fs = append(fs, store.DefFilterFunc(func(def *graph.Def) bool {
			return !def.Test
		}))
	}
	switch o.Sort {
	case "key":
		fs = append(fs, store.DefsSortByKey{})
	case "name":
		fs = append(fs, store.DefsSortByName{})
	}
	return fs
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
