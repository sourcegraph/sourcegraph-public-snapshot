package sgconf

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const (
	DefaultFile          = "sg.config.yaml"
	DefaultOverwriteFile = "sg.config.overwrite.yaml"
)

var (
	globalConfOnce sync.Once
	globalConf     *Config
	globalConfErr  output.FancyLine
)

// Get retrieves the global config files. If Config is nil, a line will be provided that
// can be printed for an explanation as to what went wrong.
//
// It must not be called before flag initalization, i.e. when confFile or overwriteFile is
// not set.
func Get(confFile, overwriteFile string) (*Config, output.FancyLine) {
	// If unset, Get was called in an illegal context, since sg.Before validates that the
	// flags are non-empty.
	if confFile == "" || overwriteFile == "" {
		panic("sgconf.Get called before flag initialization")
	}

	globalConfOnce.Do(func() {
		globalConf, globalConfErr = parseConf(confFile, overwriteFile)
	})
	return globalConf, globalConfErr
}

func parseConf(confFile, overwriteFile string) (*Config, output.FancyLine) {
	// Try to determine root of repository, so we can look for config there
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return nil, output.Linef("", output.StyleWarning, "Failed to determine repository root location: %s", err)
	}

	// If the configFlag/overwriteConfigFlag flags have their default value, we
	// take the value as relative to the root of the repository.
	if confFile == DefaultFile {
		confFile = filepath.Join(repoRoot, confFile)
	}
	if overwriteFile == DefaultOverwriteFile {
		overwriteFile = filepath.Join(repoRoot, overwriteFile)
	}

	conf, err := parseConfigFile(confFile)
	if err != nil {
		return nil, output.Linef("", output.StyleWarning, "Failed to parse %s%s%s%s as configuration file:%s\n%s", output.StyleBold, confFile, output.StyleReset, output.StyleWarning, output.StyleReset, err)
	}

	if ok, _ := fileExists(overwriteFile); ok {
		overwriteConf, err := parseConfigFile(overwriteFile)
		if err != nil {
			return nil, output.Linef("", output.StyleWarning, "Failed to parse %s%s%s%s as overwrites configuration file:%s\n%s", output.StyleBold, overwriteFile, output.StyleReset, output.StyleWarning, output.StyleReset, err)
		}
		conf.Merge(overwriteConf)
	}

	return conf, output.FancyLine{}
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
