package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"sourcegraph.com/sourcegraph/srclib"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// Filename is the name of the file that configures a directory tree or
// repository. It is intended to be used by repository authors.
var Filename = "Srcfile"

// Repository represents the config for an entire repository.
type Repository struct {
	// Tree is the configuration for the top-level directory tree in the
	// repository.
	Tree
}

// Tree represents the config for a directory and its subdirectories.
type Tree struct {
	// SourceUnits is a list of source units in the repository, either specified
	// manually in the Srcfile or discovered automatically by the scanner.
	SourceUnits []*unit.SourceUnit `json:",omitempty"`

	// Scanners to use to scan for source units in this tree.
	Scanners []*srclib.ToolRef `json:",omitempty"`

	// SkipDirs is a list of directory trees that are skipped. That is, any
	// source units (produced by scanners) whose Dir is in a skipped dir tree is
	// not processed further.
	SkipDirs []string `json:",omitempty"`

	// SkipUnits is a list of source units that are skipped. That is,
	// any scanned source units whose name and type exactly matches a
	// name and type pair in SkipUnits is skipped.
	SkipUnits []struct{ Name, Type string } `json:",omitempty"`

	// TODO(sqs): Add some type of field that lets the Srcfile and the scanners
	// have input into which tools get used during the execution phase. Right
	// now, we're going to try just using the system defaults (srclib-*) and
	// then add more flexibility when we are more familiar with the system.

	// Config is an arbitrary key-value property map. Properties are copied
	// verbatim to each source unit that is scanned in this tree.
	Config map[string]interface{} `json:",omitempty"`
}

// ReadRepository parses and validates the configuration for a repository. If no
// Srcfile exists, it returns the default configuration for the repository. If
// an overridden configuration is specified for the repository (hard-coded in
// the Go code), then it is used instead of the Srcfile or the default
// configuration.
func ReadRepository(dir string) (*Repository, error) {
	var c *Repository
	if f, err := os.Open(filepath.Join(dir, Filename)); err == nil {
		defer f.Close()
		err = json.NewDecoder(f).Decode(&c)
		if err != nil {
			return nil, err
		}
	} else if os.IsNotExist(err) {
		err = nil
		c = new(Repository)
	} else {
		return nil, err
	}

	return c.finish()
}

func (c *Repository) finish() (*Repository, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}
	return c, nil
}
