package unit

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"sourcegraph.com/sourcegraph/srclib"
	"sourcegraph.com/sourcegraph/srclib/buildstore"
)

// Key is the unique key for a source unit.
type Key struct {
	// Repo is the URI of the repository containing this source unit.
	Repo string

	// CommitID is the VCS commit that the source unit exists at.
	CommitID string

	// UnitType is the source unit's Type.
	UnitType string

	// Unit is the source unit's Name.
	Unit string
}

// START SourceUnit OMIT
type SourceUnit struct {
	// Name is an opaque identifier for this source unit that MUST be unique
	// among all other source units of the same type in the same repository.
	//
	// Two source units of different types in a repository may have the same name.
	// To obtain an identifier for a source unit that is guaranteed to be unique
	// repository-wide, use the ID method.
	Name string

	// Type is the type of source unit this represents, such as "GoPackage".
	Type string

	// Repo is the URI of the repository containing this source unit, if any.
	// The scanner tool does not need to set this field - it can be left blank,
	// to be filled in by the `srclib` tool
	Repo string `json:",omitempty"`

	// CommitID is the commit ID of the repository containing this
	// source unit, if any. The scanner tool need not fill this in; it
	// should be left blank, to be filled in by the `srclib` tool.
	CommitID string `json:",omitempty"`

	// Globs is a list of patterns that match files that make up this source
	// unit. It is used to detect when the source unit definition is out of date
	// (e.g., when a file matches the glob but is not in the Files list).
	//
	// TODO(sqs): implement this in the Makefiles
	Globs []string `json:",omitempty"`

	// Files is all of the files that make up this source unit. Filepaths should
	// be relative to the repository root.
	Files []string

	// Dir is the root directory of this source unit. It is optional and maybe
	// empty.
	Dir string `json:",omitempty"`

	// Dependencies is a list of dependencies that this source unit has. The
	// schema for these dependencies is internal to the scanner that produced
	// this source unit. The dependency resolver is expected to know how to
	// interpret this schema.
	//
	// The dependency information stored in this field should be able to be very
	// quickly determined by the scanner. The scanner should not perform any
	// dependency resolution on these entries. This is because the scanner is
	// run frequently and should execute very quickly, and dependency resolution
	// is often slow (requiring network access, etc.).
	Dependencies []interface{} `json:",omitempty"`

	// Info is an optional field that contains additional information used to
	// display the source unit
	Info *Info `json:",omitempty"`

	// Data is additional data dumped by the scanner about this source unit. It
	// typically holds information that the scanner wants to make available to
	// other components in the toolchain (grapher, dep resolver, etc.).
	Data interface{} `json:",omitempty"`

	// Config is an arbitrary key-value property map. The Config map from the
	// tree config is copied verbatim to each source unit. It can be used to
	// pass options from the Srcfile to tools.
	Config map[string]interface{} `json:",omitempty"`

	// Ops enumerates the operations that should be performed on this source
	// unit. Each key is the name of an operation, and the value is the tool to
	// use to perform that operation. If the value is nil, the tool is chosen
	// automatically according to the user's configuration.
	Ops map[string]*srclib.ToolRef `json:",omitempty"`

	// TODO(sqs): add a way to specify the toolchains and tools to use for
	// various tasks on this source unit
}

func (u *SourceUnit) Key() Key {
	return Key{
		Repo:     u.Repo,
		CommitID: u.CommitID,
		UnitType: u.Type,
		Unit:     u.Name,
	}
}

//END SourceUnit OMIT

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

// OpsSorted returns the keys of the Ops map in sorted order.
func (u *SourceUnit) OpsSorted() []string {
	ops := make([]string, len(u.Ops))
	i := 0
	for op := range u.Ops {
		ops[i] = op
		i++
	}
	sort.Strings(ops)
	return ops
}

func init() {
	buildstore.RegisterDataType("unit", SourceUnit{})
}

// idSeparator joins a source unit's name and type in its ID string.
var idSeparator = "@"

// ID returns an opaque identifier for this source unit that is guaranteed to be
// unique among all other source units in the same repository.
func (u SourceUnit) ID() ID {
	return ID(fmt.Sprintf("%s%s%s", url.QueryEscape(u.Name), idSeparator, u.Type))
}

func (u *SourceUnit) ID2() ID2 {
	return ID2{Type: u.Type, Name: u.Name}
}

func (u Key) ID2() ID2 {
	return ID2{Type: u.UnitType, Name: u.Unit}
}

func (u SourceUnit) String() string {
	b, _ := json.Marshal(u)
	return string(b)
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
		expanded = append(expanded, hits...)
	}
	return expanded, nil
}

type SourceUnits []*SourceUnit

func (v SourceUnits) Len() int           { return len(v) }
func (v SourceUnits) Less(i, j int) bool { return v[i].String() < v[j].String() }
func (v SourceUnits) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
