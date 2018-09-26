package langserver

import (
	"os"
	"runtime"
)

var (
	// GOLSP_WARMUP_ON_INITIALIZE toggles if we typecheck the whole
	// workspace in the background on initialize. This trades off initial
	// CPU and memory to hide perceived latency of the first few
	// requests. If the LSP server is long lived the tradeoff is usually
	// worth it.
	envWarmupOnInitialize = os.Getenv("GOLSP_WARMUP_ON_INITIALIZE")
)

// Config adjusts the behaviour of go-langserver. Please keep in sync with
// InitializationOptions in the README.
type Config struct {
	// FuncSnippetEnabled enables the returning of argument snippets on `func`
	// completions, eg. func(foo string, arg2 bar). Requires code completion
	// to be enabled.
	//
	// Defaults to true if not specified.
	FuncSnippetEnabled bool

	// GocodeCompletionEnabled enables code completion feature (using gocode)
	//
	// Defaults to false if not specified.
	GocodeCompletionEnabled bool

	// FormatTool decides which tool is used to format documents. Supported: goimports and gofmt
	//
	// Defaults to goimports if not specified.
	FormatTool string

	// GoimportsLocalPrefix sets the local prefix (comma-separated string) that goimports will use
	//
	// Defaults to empty string if not specified.
	GoimportsLocalPrefix string

	// DiagnosticsEnabled enables handling of diagnostics
	//
	// Defaults to false if not specified.
	DiagnosticsEnabled bool

	// MaxParallelism controls the maximum number of goroutines that should be used
	// to fulfill requests. This is useful in editor environments where users do
	// not want results ASAP, but rather just semi quickly without eating all of
	// their CPU.
	//
	// Defaults to half of your CPU cores if not specified.
	MaxParallelism int

	// UseBinaryPkgCache controls whether or not $GOPATH/pkg binary .a files should
	// be used.
	//
	// Defaults to true if not specified.
	UseBinaryPkgCache bool
}

// Apply sets the corresponding field in c for each non-nil field in o.
func (c Config) Apply(o *InitializationOptions) Config {
	if o == nil {
		return c
	}
	if o.FuncSnippetEnabled != nil {
		c.FuncSnippetEnabled = *o.FuncSnippetEnabled
	}
	if o.GocodeCompletionEnabled != nil {
		c.GocodeCompletionEnabled = *o.GocodeCompletionEnabled
	}
	if o.FormatTool != nil {
		c.FormatTool = *o.FormatTool
	}
	if o.GoimportsLocalPrefix != nil {
		c.GoimportsLocalPrefix = *o.GoimportsLocalPrefix
	}
	if o.MaxParallelism != nil {
		c.MaxParallelism = *o.MaxParallelism
	}
	if o.UseBinaryPkgCache != nil {
		c.UseBinaryPkgCache = *o.UseBinaryPkgCache
	}
	return c
}

// NewDefaultConfig returns the default config. See the field comments for the
// defaults.
func NewDefaultConfig() Config {
	// Default max parallelism to half the CPU cores, but at least always one.
	maxparallelism := runtime.NumCPU() / 2
	if maxparallelism <= 0 {
		maxparallelism = 1
	}

	return Config{
		FuncSnippetEnabled:      true,
		GocodeCompletionEnabled: false,
		FormatTool:              formatToolGoimports,
		MaxParallelism:          maxparallelism,
		UseBinaryPkgCache:       true,
	}
}
