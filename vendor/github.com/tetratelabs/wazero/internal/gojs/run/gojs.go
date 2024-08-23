// Package run exists to avoid dependency cycles when keeping most of gojs
// code internal.
package run

import (
	"context"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/internal/gojs"
	"github.com/tetratelabs/wazero/internal/gojs/config"
	"github.com/tetratelabs/wazero/sys"
)

func Run(ctx context.Context, r wazero.Runtime, compiled wazero.CompiledModule, moduleConfig wazero.ModuleConfig, config *config.Config) error {
	if err := config.Init(); err != nil {
		return err
	}

	// Instantiate the module compiled by go, noting it has no init function.
	mod, err := r.InstantiateModule(ctx, compiled, moduleConfig)
	if err != nil {
		return err
	}
	defer mod.Close(ctx)

	// Extract the args and env from the module Config and write it to memory.
	argc, argv, err := gojs.WriteArgsAndEnviron(mod)
	if err != nil {
		return err
	}

	// Create host-side state for JavaScript values and events.
	ctx = context.WithValue(ctx, gojs.StateKey{}, gojs.NewState(config))

	// Invoke the run function.
	_, err = mod.ExportedFunction("run").Call(ctx, uint64(argc), uint64(argv))
	if se, ok := err.(*sys.ExitError); ok {
		if se.ExitCode() == 0 { // Don't err on success.
			err = nil
		}
	}
	return err
}
