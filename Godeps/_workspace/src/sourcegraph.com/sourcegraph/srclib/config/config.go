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
	// URI is the repository's clone URI.
	URI string `json:",omitempty"`

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

	// PreConfigCommands is a list of commands (passed to `sh -c`) that should
	// be run on the tree before configuration occurs (after the initial config
	// is read from the Srcfile but before scanners are run). The commands are
	// run at the top-level directory of the tree.
	//
	// If `srclib config` is run with the program execution method ("-m
	// program"), the commands are executed normally. If it is run
	// with the Docker execution method ("-m docker"), the commands
	// are run in a Docker container with the tree mounted read-write
	// at /src. Currently this container runs Ubuntu 14.04 with the
	// git, mercurial, curl, and build-essential packages installed
	// (plus others that are temporarily installed until we develop a
	// better way of handling things that need pre-config commands).
	//
	// This is the only accepted way to modify the tree during processing. Tools
	// shouldn't modify the tree because they run concurrently. If they run in
	// Docker, they're unable to modify the tree because it is mounted
	// read-only.
	PreConfigCommands []string `json:",omitempty"`

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
func ReadRepository(dir string, repoURI string) (*Repository, error) {
	var c *Repository
	if oc, overridden := Overrides[repoURI]; overridden {
		c = oc
	} else if f, err := os.Open(filepath.Join(dir, Filename)); err == nil {
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

	return c.finish(repoURI)
}

func (c *Repository) finish(repoURI string) (*Repository, error) {
	err := c.validate()
	if err != nil {
		return nil, err
	}
	c.URI = repoURI
	return c, nil
}

type Options struct {
	Repo   string `long:"repo" description:"repository URI" value-name:"URI"`
	Subdir string `long:"subdir" description:"subdirectory in repository" value-name:"DIR"`
}
