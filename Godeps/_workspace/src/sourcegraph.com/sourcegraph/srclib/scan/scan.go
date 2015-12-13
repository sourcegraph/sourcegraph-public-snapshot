package scan

import (
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"sync"

	"github.com/rogpeppe/rog-go/parallel"
	"sourcegraph.com/sourcegraph/srclib/config"
	"sourcegraph.com/sourcegraph/srclib/flagutil"
	"sourcegraph.com/sourcegraph/srclib/toolchain"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

type Options struct {
	config.Options
	// Quiet silences all output.
	Quiet bool
}

// ScanMulti runs multiple scanner tools in parallel. It passes command-line
// options from opt to each one, and it sends the JSON representation of cfg
// (the repo/tree's Config) to each tool's stdin.
func ScanMulti(scanners []toolchain.Tool, opt Options, treeConfig map[string]interface{}) ([]*unit.SourceUnit, error) {
	if treeConfig == nil {
		treeConfig = map[string]interface{}{}
	}

	var (
		units []*unit.SourceUnit
		mu    sync.Mutex
	)

	run := parallel.NewRun(runtime.GOMAXPROCS(0))
	for _, scanner_ := range scanners {
		scanner := scanner_
		run.Do(func() error {
			units2, err := Scan(scanner, opt, treeConfig)
			if err != nil {
				cmd, newErr := scanner.Command()
				if newErr != nil {
					return fmt.Errorf("cmd error: %s", newErr)
				}
				return fmt.Errorf("scanner %v: %s", cmd.Args, err)
			}

			mu.Lock()
			defer mu.Unlock()
			units = append(units, units2...)
			return nil
		})
	}
	err := run.Wait()
	// Return error only if none of the commands succeeded.
	if len(units) == 0 {
		return nil, err
	}
	return units, nil
}

func Scan(scanner toolchain.Tool, opt Options, treeConfig map[string]interface{}) ([]*unit.SourceUnit, error) {
	args, err := flagutil.MarshalArgs(&opt)
	if err != nil {
		return nil, err
	}

	if opt.Quiet {
		scanner.SetLogger(log.New(ioutil.Discard, "", 0))
	}
	var units []*unit.SourceUnit
	if err := scanner.Run(args, treeConfig, &units); err != nil {
		return nil, err
	}

	for _, u := range units {
		u.Repo = opt.Repo
	}

	return units, nil
}
