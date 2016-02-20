package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"

	"sourcegraph.com/sourcegraph/go-flags"

	"sourcegraph.com/sourcegraph/rwvfs"
	"sourcegraph.com/sourcegraph/srclib/buildstore"
	"sourcegraph.com/sourcegraph/srclib/config"
	"sourcegraph.com/sourcegraph/srclib/plan"

	"sourcegraph.com/sourcegraph/srclib/unit"
)

func init() {
	cliInit = append(cliInit, func(cli *flags.Command) {
		c, err := cli.AddCommand("config",
			"reads & scans for project configuration",
			`Produces a configuration file suitable for building the repository or directory tree rooted at DIR (or the current directory if not specified).

The steps are:

1. Read user srclib config (SRCLIBPATH/.srclibconfig), if present.

2. Read configuration from the current directory's Srcfile (if present).

3. Scan for source units in the directory tree rooted at the current directory (or the root of the repository containing the current directory), using the scanners specified in either the user srclib config or the Srcfile (or otherwise the defaults).
`,
			&configCmd,
		)
		if err != nil {
			log.Fatal(err)
		}
		c.Aliases = []string{"c"}
	})
}

// getInitialConfig gets the initial config (i.e., the config that comes solely
// from the Srcfile, if any, and the external user config, before running the
// scanners).
func getInitialConfig(dir string) (*config.Repository, error) {
	r, err := OpenRepo(dir)
	if err != nil {
		return nil, err
	}

	cfg, err := config.ReadRepository(r.RootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read repository at %s: %s", r.RootDir, err)
	}

	if cfg.Scanners == nil {
		x, err := config.SrclibPathConfig()
		if err != nil {
			return nil, err
		}
		cfg.Scanners = x.Scanners
	}

	return cfg, nil
}

type ConfigCmd struct {
	Output struct {
		Output string `short:"o" long:"output" description:"output format" default:"text" value-name:"text|json"`
	} `group:"output"`

	Args struct {
		Dir Directory `name:"DIR" default:"." description:"root directory of tree to configure"`
	} `positional-args:"yes"`

	Quiet bool `short:"q" long:"quiet" description:"silence all output"`

	w io.Writer // output stream to print to (defaults to os.Stdout)
}

var configCmd ConfigCmd

func (c *ConfigCmd) Execute(args []string) error {
	if c.w == nil {
		c.w = os.Stdout
	}
	if c.Quiet {
		c.w = nopWriteCloser{}
	}

	cfg, err := getInitialConfig(c.Args.Dir.String())
	if err != nil {
		return err
	}

	if err := scanUnitsIntoConfig(cfg, c.Quiet); err != nil {
		return fmt.Errorf("failed to scan for source units: %s", err)
	}

	localRepo, err := OpenRepo(c.Args.Dir.String())
	if err != nil {
		return fmt.Errorf("failed to open repo: %s", err)
	}
	buildStore, err := buildstore.LocalRepo(localRepo.RootDir)
	if err != nil {
		return err
	}
	commitFS := buildStore.Commit(localRepo.CommitID)

	// Write source units to build cache.
	if err := rwvfs.MkdirAll(commitFS, "."); err != nil {
		return err
	}
	for _, u := range cfg.SourceUnits {
		unitFile := plan.SourceUnitDataFilename(unit.SourceUnit{}, u)
		if err := rwvfs.MkdirAll(commitFS, filepath.Dir(unitFile)); err != nil {
			return err
		}
		f, err := commitFS.Create(unitFile)
		if err != nil {
			return err
		}
		defer func(f io.WriteCloser) {
			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
		}(f)
		if err := json.NewEncoder(f).Encode(u); err != nil {
			return err
		}
	}

	if c.Output.Output == "json" {
		PrintJSON(cfg, "")
	} else {
		fmt.Fprintf(c.w, "SCANNERS (%d)\n", len(cfg.Scanners))
		for _, s := range cfg.Scanners {
			fmt.Fprintf(c.w, " - %s\n", s)
		}
		fmt.Fprintln(c.w)

		fmt.Fprintf(c.w, "SOURCE UNITS (%d)\n", len(cfg.SourceUnits))
		for _, u := range cfg.SourceUnits {
			fmt.Fprintf(c.w, " - %s: %s\n", u.Type, u.Name)
		}
		fmt.Fprintln(c.w)

		fmt.Fprintf(c.w, "CONFIG PROPERTIES (%d)\n", len(cfg.Config))
		for _, kv := range sortedMap(cfg.Config) {
			fmt.Fprintf(c.w, " - %s: %s\n", kv[0], kv[1])
		}
	}

	return nil
}

func sortedMap(m map[string]interface{}) [][2]interface{} {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	sorted := make([][2]interface{}, len(keys))
	for i, k := range keys {
		sorted[i] = [2]interface{}{k, m[k]}
	}
	return sorted
}
