// Package config exists to avoid dependency cycles when keeping most of gojs
// code internal.
package config

import (
	"os"
	"path/filepath"

	"github.com/tetratelabs/wazero/internal/platform"
)

type Config struct {
	OsWorkdir bool

	// Workdir is the actual working directory value.
	Workdir string
	Umask   uint32
}

func NewConfig() *Config {
	return &Config{
		OsWorkdir: false,
		Workdir:   "/",
		Umask:     uint32(0o0022),
	}
}

func (c *Config) Clone() *Config {
	ret := *c // copy except maps which share a ref
	return &ret
}

func (c *Config) Init() error {
	if c.OsWorkdir {
		workdir, err := os.Getwd()
		if err != nil {
			return err
		}
		// Ensure if used on windows, the input path is translated to a POSIX one.
		workdir = platform.ToPosixPath(workdir)
		// Strip the volume of the path, for example C:\
		c.Workdir = workdir[len(filepath.VolumeName(workdir)):]
	}
	return nil
}
