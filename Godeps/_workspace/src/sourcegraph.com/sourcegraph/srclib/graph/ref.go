package graph

import "strconv"

type RefKey struct {
	DefRepo     string `json:",omitempty"`
	DefUnitType string `json:",omitempty"`
	DefUnit     string `json:",omitempty"`
	DefPath     string `json:",omitempty"`
	Def         bool   `json:",omitempty"`
	Repo        string `json:",omitempty"`
	UnitType    string `json:",omitempty"`
	Unit        string `json:",omitempty"`
	File        string `json:",omitempty"`
	CommitID    string `json:",omitempty"`
	Start       uint32 `json:",omitempty"`
	End         uint32 `json:",omitempty"`
}

func (r *RefKey) RefDefKey() RefDefKey {
	return RefDefKey{
		DefRepo:     r.DefRepo,
		DefUnitType: r.DefUnitType,
		DefUnit:     r.DefUnit,
		DefPath:     r.DefPath,
	}
}

func (r *Ref) RefKey() RefKey {
	return RefKey{
		DefRepo:     r.DefRepo,
		DefUnitType: r.DefUnitType,
		DefUnit:     r.DefUnit,
		DefPath:     r.DefPath,
		Def:         r.Def,
		Repo:        r.Repo,
		UnitType:    r.UnitType,
		Unit:        r.Unit,
		File:        r.File,
		Start:       r.Start,
		End:         r.End,
	}
}

func (r *Ref) RefDefKey() RefDefKey {
	return RefDefKey{
		DefRepo:     r.DefRepo,
		DefUnitType: r.DefUnitType,
		DefUnit:     r.DefUnit,
		DefPath:     r.DefPath,
	}
}

func (r *Ref) DefKey() DefKey {
	return DefKey{
		Repo:     r.DefRepo,
		UnitType: r.DefUnitType,
		Unit:     r.DefUnit,
		Path:     r.DefPath,
	}
}

func (r *Ref) SetFromDefKey(k DefKey) {
	r.DefPath = k.Path
	r.DefUnitType = k.UnitType
	r.DefUnit = k.Unit
	r.DefRepo = k.Repo
}

// Sorting

type Refs []*Ref

func (r *Ref) sortKey() string {
	return r.DefPath + r.DefRepo + r.DefUnitType + r.DefUnit + r.Repo + r.UnitType + r.Unit + r.File + strconv.Itoa(int(r.Start)) + strconv.Itoa(int(r.End))
}
func (vs Refs) Len() int           { return len(vs) }
func (vs Refs) Swap(i, j int)      { vs[i], vs[j] = vs[j], vs[i] }
func (vs Refs) Less(i, j int) bool { return vs[i].sortKey() < vs[j].sortKey() }

// RefSet is a set of Refs. It can used to determine whether a grapher emits
// duplicate refs.
type RefSet struct {
	refs map[Ref]struct{}
}

func NewRefSet() *RefSet {
	return &RefSet{make(map[Ref]struct{})}
}

// AddAndCheckUnique adds ref to the set of seen refs, and returns whether the
// ref already existed in the set.
func (c *RefSet) AddAndCheckUnique(ref Ref) (duplicate bool) {
	_, present := c.refs[ref]
	if present {
		return true
	}
	c.refs[ref] = struct{}{}
	return false
}
