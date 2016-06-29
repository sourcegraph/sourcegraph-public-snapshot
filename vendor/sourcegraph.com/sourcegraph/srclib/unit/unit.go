package unit

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"strings"

	"sourcegraph.com/sourcegraph/srclib"
	"sourcegraph.com/sourcegraph/srclib/buildstore"
)

func init() {
	buildstore.RegisterDataType("unit", SourceUnit{})
}

// UnitRepoUnresolved is a sentinel value that indicates this unit's
// repository is unresolved
const UnitRepoUnresolved = "?"

func (u Key) IsResolved() bool {
	return u.Repo == UnitRepoUnresolved
}

// ContainsAny returns true if u contains any files in filesnames. Currently
// doesn't process globs.
func (u SourceUnit) ContainsAny(filenames []string) bool {
	if len(filenames) == 0 {
		return false
	}
	files := make(map[string]bool)
	for _, f := range filenames {
		files[f] = true
	}
	for _, uf := range u.Files {
		if files[uf] {
			return true
		}
	}
	return false
}

// idSeparator joins a source unit's name and type in its ID string.
const idSeparator = "@"

// ID returns an opaque identifier for this source unit that is guaranteed to be
// unique among all other source units in the same repository.
func (u SourceUnit) ID() ID {
	return ID(fmt.Sprintf("%s%s%s", url.QueryEscape(u.Name), idSeparator, u.Type))
}

func (u *SourceUnit) ID2() ID2 {
	return u.Key.ID2()
}

func (u Key) ID2() ID2 {
	return ID2{Type: u.Type, Name: u.Name}
}

// ParseID parses the name and type from a source unit ID (from
// (*SourceUnit).ID()).
func ParseID(unitID string) (name, typ string, err error) {
	at := strings.Index(unitID, idSeparator)
	if at == -1 {
		return "", "", fmt.Errorf("no %q in source unit ID", idSeparator)
	}

	name, err = url.QueryUnescape(unitID[:at])
	if err != nil {
		return "", "", err
	}
	typ = unitID[at+len(idSeparator):]
	return name, typ, nil
}

// ID is a source unit ID. It is only unique within a repository.
type ID string

// ID2 is a source unit ID. It is only unique within a repository.
type ID2 struct {
	Type string
	Name string
}

func (v ID2) String() string { return fmt.Sprintf("{%s %s}", v.Type, v.Name) }

// ExpandPaths interprets paths, which contains paths (optionally with
// filepath.Glob-compatible globs) that are relative to base. A list of actual
// files that are referenced is returned.
func ExpandPaths(base string, paths []string) ([]string, error) {
	var expanded []string
	for _, path := range paths {
		hits, err := filepath.Glob(filepath.Join(base, path))
		if err != nil {
			return nil, err
		}
		for _, hit := range hits {
			expanded = append(expanded, filepath.ToSlash(hit))
		}
	}
	return expanded, nil
}

type SourceUnits []*SourceUnit

func (v SourceUnits) Len() int           { return len(v) }
func (v SourceUnits) Less(i, j int) bool { return v[i].String() < v[j].String() }
func (v SourceUnits) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }

// sourceUnit is the struct for JSON serialization used to preserve
// backcompat with the old SourceUnit JSON serialization format
//
// DEPRECATED: this should be removed after all srclib toolchain
// Docker images have been updated to the new source unit schema.
type sourceUnit struct {
	Name         string
	Type         string
	Repo         string   `json:",omitempty"`
	CommitID     string   `json:",omitempty"`
	Globs        []string `json:",omitempty"`
	Files        []string
	Dir          string                      `json:",omitempty"`
	Dependencies []json.RawMessage           `json:",omitempty"`
	Info         *Info                       `json:",omitempty"`
	Data         *json.RawMessage            `json:",omitempty"`
	Config       map[string]*json.RawMessage `json:",omitempty"`
	Ops          map[string]*srclib.ToolRef  `json:",omitempty"`
}

var _ json.Marshaler = (*SourceUnit)(nil)
var _ json.Unmarshaler = (*SourceUnit)(nil)

func (u *SourceUnit) MarshalJSON() ([]byte, error) {
	deps := make([]json.RawMessage, len(u.Dependencies))
	for i := range u.Dependencies {
		var err error
		deps[i], err = json.Marshal(u.Dependencies[i])
		if err != nil {
			return nil, err
		}
	}
	cfg := make(map[string]*json.RawMessage)
	for k, v := range u.Config {
		if b, err := json.Marshal(v); err == nil {
			b_ := json.RawMessage(b)
			cfg[k] = &b_
		} else {
			return nil, err
		}
	}

	ops := make(map[string]*srclib.ToolRef)
	for k := range u.Ops {
		ops[k] = nil
	}
	var data *json.RawMessage
	if u.Data != nil {
		d := json.RawMessage(u.Data)
		data = &d
	}
	return json.Marshal(sourceUnit{
		Name:         u.Name,
		Type:         u.Type,
		Repo:         u.Repo,
		CommitID:     u.CommitID,
		Files:        u.Files,
		Dir:          u.Dir,
		Dependencies: deps,
		Data:         data,
		Config:       cfg,
		Ops:          ops,
	})
}

func (u *SourceUnit) UnmarshalJSON(b []byte) error {
	var su sourceUnit
	if err := json.Unmarshal(b, &su); err != nil {
		return fmt.Errorf("could not unmarshal source unit: %s; JSON was: %s", err, string(b))
	}
	deps := make([]*Key, len(su.Dependencies))
	for i, depJSON := range su.Dependencies {
		var dep Key
		if err := json.Unmarshal(depJSON, &dep); err == nil {
			deps[i] = &dep
		} else if err != nil {
			var s string
			if err := json.Unmarshal(depJSON, &s); err != nil {
				return fmt.Errorf("could not unmarshal dependency: %s; JSON was %v", err, depJSON)
			}
			deps[i] = &Key{
				Repo: UnitRepoUnresolved,
				Name: s,
				Type: u.Type,
			}
		}
	}
	cfg := make(map[string]string)
	for k, vJSON := range su.Config {
		var v string
		if err := json.Unmarshal(*vJSON, &v); err != nil {
			log.Printf("warning: could not unmarshal config string: %s, JSON was %v", err, vJSON)
			continue
		}
		cfg[k] = v
	}
	ops := make(map[string][]byte)
	for k := range su.Ops {
		ops[k] = nil
	}
	u.Name = su.Name
	u.Type = su.Type
	u.Repo = su.Repo
	u.CommitID = su.CommitID
	u.Files = su.Files
	u.Dir = su.Dir
	u.Dependencies = deps
	if su.Data != nil {
		u.Data = *su.Data
	}
	u.Config = cfg
	u.Ops = ops
	return nil
}
