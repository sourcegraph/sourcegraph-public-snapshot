package toolchain

import (
	"encoding/json"
	"os"
	"path/filepath"

	"sourcegraph.com/sourcegraph/srclib"
)

// Dir returns the directory where the named toolchain lives (under
// the SRCLIBPATH). If the toolchain already exists in any of the
// entries of SRCLIBPATH, that directory is returned. Otherwise a
// nonexistent directory in the first SRCLIBPATH entry is returned.
func Dir(toolchainPath string) (string, error) {
	toolchainPath = filepath.Clean(toolchainPath)

	dir, err := lookupToolchain(toolchainPath)
	if os.IsNotExist(err) {
		return filepath.Join(filepath.SplitList(srclib.Path)[0], toolchainPath), nil
	}
	if err != nil {
		err = &os.PathError{Op: "toolchain.Dir", Path: toolchainPath, Err: err}
	}
	return dir, err
}

// Info describes a toolchain.
type Info struct {
	// Path is the toolchain's path (not a directory path) underneath the
	// SRCLIBPATH. It consists of the URI of this repository's toolchain plus
	// its subdirectory path within the repository. E.g., "github.com/foo/bar"
	// for a toolchain defined in the root directory of that repository.
	Path string

	// Dir is the filesystem directory that defines this toolchain.
	Dir string

	// ConfigFile is the path to the Srclibtoolchain file, relative to Dir.
	ConfigFile string

	// Program is the path to the executable program (relative to Dir) to run to
	// invoke this toolchain.
	Program string `json:",omitempty"`
}

// ReadConfig reads and parses the Srclibtoolchain config file for the
// toolchain.
func (t *Info) ReadConfig() (*Config, error) {
	f, err := os.Open(filepath.Join(t.Dir, t.ConfigFile))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var c *Config
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		return nil, err
	}
	return c, nil
}

// Command returns the path to the executable program for the
// toolchain with the given path.
func Command(path string) (string, error) {
	tc, err := Lookup(path)
	if err != nil {
		return "", err
	}
	return filepath.Join(tc.Dir, tc.Program), nil
}
