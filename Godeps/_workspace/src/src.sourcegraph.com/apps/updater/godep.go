package updater

import (
	"bytes"
	"encoding/json"
)

// Godeps describes what a package needs to be rebuilt reproducibly.
// It's the same information stored in file Godeps.
type Godeps struct {
	ImportPath string
	GoVersion  string
	Packages   []string `json:",omitempty"` // Arguments to save, if any.
	Deps       []Dependency
}

// A Dependency is a specific revision of a package.
type Dependency struct {
	ImportPath string
	Comment    string `json:",omitempty"` // Description of commit, if present.
	Rev        string // VCS-specific commit ID.
}

// parseGodeps parses a Godeps.json file.
func parseGodeps(content []byte) (Godeps, error) {
	r := bytes.NewReader(content)
	var g Godeps
	err := json.NewDecoder(r).Decode(&g)
	return g, err
}
