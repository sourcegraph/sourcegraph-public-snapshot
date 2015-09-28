package makex

import (
	"flag"
	"os"
	"runtime"

	"sourcegraph.com/sourcegraph/rwvfs"
)

type Config struct {
	FS           FileSystem
	ParallelJobs int
	Verbose      bool
	DryRun       bool
}

var Default = Config{
	ParallelJobs: 1,
}

func (c *Config) fs() FileSystem {
	if c.FS != nil {
		return c.FS
	}
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	return NewFileSystem(rwvfs.OS(dir))
}

func (c *Config) pathExists(path string) (bool, error) {
	_, err := c.fs().Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

// Flags adds makex command-line flags to an existing flag.FlagSet (or the
// global FlagSet if fs is nil).
func Flags(fs *flag.FlagSet, conf *Config, prefix string) {
	if fs == nil {
		fs = flag.CommandLine
	}
	fs.BoolVar(&conf.DryRun, prefix+"n", false, "dry run (don't actually run any commands)")
	fs.IntVar(&conf.ParallelJobs, prefix+"j", runtime.GOMAXPROCS(0), "number of jobs to run in parallel")
	fs.BoolVar(&conf.Verbose, prefix+"v", false, "verbose")
}
