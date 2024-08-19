// Package gojs allows you to run wasm binaries compiled by Go when
// `GOOS=js GOARCH=wasm`. See https://wazero.io/languages/go/ for more.
//
// # Experimental
//
// Go defines js "EXPERIMENTAL... exempt from the Go compatibility promise."
// Accordingly, wazero cannot guarantee this will work from release to release,
// or that usage will be relatively free of bugs. Moreover, `GOOS=wasi` will
// happen, and once that's available in two releases wazero will remove this
// package.
//
// Due to these concerns and the relatively high implementation overhead, most
// will choose TinyGo instead of gojs.
package gojs

import (
	"context"
	"errors"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/internal/gojs"
	internalconfig "github.com/tetratelabs/wazero/internal/gojs/config"
	"github.com/tetratelabs/wazero/internal/gojs/run"
	"github.com/tetratelabs/wazero/internal/wasm"
)

// MustInstantiate calls Instantiate or panics on error.
//
// This is a simpler function for those who know host functions are not already
// instantiated, and don't need to unload them separate from the runtime.
func MustInstantiate(ctx context.Context, r wazero.Runtime, guest wazero.CompiledModule) {
	if _, err := Instantiate(ctx, r, guest); err != nil {
		panic(err)
	}
}

// Instantiate detects and instantiates host functions for wasm compiled with
// `GOOS=js GOARCH=wasm`. `guest` must be a result of `r.CompileModule`.
//
// # Notes
//
//   - Failure cases are documented on wazero.Runtime InstantiateModule.
//   - Closing the wazero.Runtime has the same effect as closing the result.
//   - To add more functions to `goModule`, use FunctionExporter.
func Instantiate(ctx context.Context, r wazero.Runtime, guest wazero.CompiledModule) (api.Closer, error) {
	goModule, err := detectGoModule(guest.ImportedFunctions())
	if err != nil {
		return nil, err
	}
	builder := r.NewHostModuleBuilder(goModule)
	NewFunctionExporter().ExportFunctions(builder)
	return builder.Instantiate(ctx)
}

// detectGoModule is needed because the module name defining host functions for
// `GOOS=js GOARCH=wasm` was renamed from "go" to "gojs" in Go 1.21. We can't
// use the version that compiles wazero because it could be different from what
// compiled the guest.
//
// See https://github.com/golang/go/commit/02411bcd7c8eda9c694a5755aff0a516d4983952
func detectGoModule(imports []api.FunctionDefinition) (string, error) {
	for _, f := range imports {
		moduleName, _, _ := f.Import()
		switch moduleName {
		case "go", "gojs":
			return moduleName, nil
		}
	}
	return "", errors.New("guest wasn't compiled with GOOS=js GOARCH=wasm")
}

// FunctionExporter builds host functions for wasm compiled with
// `GOOS=js GOARCH=wasm`.
type FunctionExporter interface {
	// ExportFunctions builds functions to an existing host module builder.
	//
	// This should be named "go" or "gojs", depending on the version of Go the
	// guest was compiled with. The module name changed from "go" to "gojs" in
	// Go 1.21.
	ExportFunctions(wazero.HostModuleBuilder)
}

// NewFunctionExporter returns a FunctionExporter object.
func NewFunctionExporter() FunctionExporter {
	return &functionExporter{}
}

type functionExporter struct{}

// ExportFunctions implements FunctionExporter.ExportFunctions
func (e *functionExporter) ExportFunctions(builder wazero.HostModuleBuilder) {
	hfExporter := builder.(wasm.HostFuncExporter)

	hfExporter.ExportHostFunc(gojs.GetRandomData)
	hfExporter.ExportHostFunc(gojs.Nanotime1)
	hfExporter.ExportHostFunc(gojs.WasmExit)
	hfExporter.ExportHostFunc(gojs.CopyBytesToJS)
	hfExporter.ExportHostFunc(gojs.ValueCall)
	hfExporter.ExportHostFunc(gojs.ValueGet)
	hfExporter.ExportHostFunc(gojs.ValueIndex)
	hfExporter.ExportHostFunc(gojs.ValueLength)
	hfExporter.ExportHostFunc(gojs.ValueNew)
	hfExporter.ExportHostFunc(gojs.ValueSet)
	hfExporter.ExportHostFunc(gojs.WasmWrite)
	hfExporter.ExportHostFunc(gojs.ResetMemoryDataView)
	hfExporter.ExportHostFunc(gojs.Walltime)
	hfExporter.ExportHostFunc(gojs.ScheduleTimeoutEvent)
	hfExporter.ExportHostFunc(gojs.ClearTimeoutEvent)
	hfExporter.ExportHostFunc(gojs.FinalizeRef)
	hfExporter.ExportHostFunc(gojs.StringVal)
	hfExporter.ExportHostFunc(gojs.ValueDelete)
	hfExporter.ExportHostFunc(gojs.ValueSetIndex)
	hfExporter.ExportHostFunc(gojs.ValueInvoke)
	hfExporter.ExportHostFunc(gojs.ValuePrepareString)
	hfExporter.ExportHostFunc(gojs.ValueInstanceOf)
	hfExporter.ExportHostFunc(gojs.ValueLoadString)
	hfExporter.ExportHostFunc(gojs.CopyBytesToGo)
	hfExporter.ExportHostFunc(gojs.Debug)
}

// Config extends wazero.ModuleConfig with GOOS=js specific extensions.
// Use NewConfig to create an instance.
type Config interface {
	// WithOSWorkdir sets the initial working directory used to Run Wasm to
	// the value of os.Getwd instead of the default of root "/".
	//
	// Here's an example that overrides this to the current directory:
	//
	//	err = gojs.Run(ctx, r, compiled, gojs.NewConfig(moduleConfig).
	//			WithOSWorkdir())
	//
	// Note: To use this feature requires mounting the real root directory via
	// wazero.FSConfig `WithDirMount`. On windows, this root must be the same drive
	// as the value of os.Getwd. For example, it would be an error to mount `C:\`
	// as the guest path "", while the current directory is inside `D:\`.
	WithOSWorkdir() Config
}

// NewConfig returns a Config that can be used for configuring module instantiation.
func NewConfig(moduleConfig wazero.ModuleConfig) Config {
	return &cfg{moduleConfig: moduleConfig, internal: internalconfig.NewConfig()}
}

type cfg struct {
	moduleConfig wazero.ModuleConfig
	internal     *internalconfig.Config
}

func (c *cfg) clone() *cfg {
	return &cfg{moduleConfig: c.moduleConfig, internal: c.internal.Clone()}
}

// WithOSWorkdir implements Config.WithOSWorkdir
func (c *cfg) WithOSWorkdir() Config {
	ret := c.clone()
	ret.internal.OsWorkdir = true
	return ret
}

// Run instantiates a new module and calls "run" with the given config.
//
// # Parameters
//
//   - ctx: context to use when instantiating the module and calling "run".
//   - r: runtime to instantiate both the host and guest (compiled) module in.
//   - compiled: guest binary compiled with `GOOS=js GOARCH=wasm`
//   - config: the Config to use including wazero.ModuleConfig or extensions of
//     it.
//
// # Example
//
// After compiling your Wasm binary with wazero.Runtime's `CompileModule`, run
// it like below:
//
//	// Instantiate host functions needed by gojs
//	gojs.MustInstantiate(ctx, r)
//
//	// Assign any configuration relevant for your compiled wasm.
//	config := gojs.NewConfig(wazero.NewConfig())
//
//	// Run your wasm, notably handing any ExitError
//	err = gojs.Run(ctx, r, compiled, config)
//	if exitErr, ok := err.(*sys.ExitError); ok && exitErr.ExitCode() != 0 {
//		log.Panicln(err)
//	} else if !ok {
//		log.Panicln(err)
//	}
//
// # Notes
//
//   - Wasm generated by `GOOS=js GOARCH=wasm` is very slow to compile: Use
//     wazero.RuntimeConfig with wazero.CompilationCache when re-running the
//     same binary.
//   - The guest module is closed after being run.
func Run(ctx context.Context, r wazero.Runtime, compiled wazero.CompiledModule, moduleConfig Config) error {
	c := moduleConfig.(*cfg)
	return run.Run(ctx, r, compiled, c.moduleConfig, c.internal)
}
