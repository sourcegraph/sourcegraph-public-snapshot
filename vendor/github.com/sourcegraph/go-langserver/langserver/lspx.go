package langserver

import "github.com/sourcegraph/go-langserver/pkg/lsp"

// This file contains Go-specific extensions to LSP types.
//
// The Go language server MUST NOT rely on these extensions for
// standalone operation on the local file system. (VSCode has no way
// of including these fields.)

// InitializationOptions are the options supported by go-langserver. It is the
// Config struct, but each field is optional.
type InitializationOptions struct {
	// FuncSnippetEnabled is an optional version of Config.FuncSnippetEnabled
	FuncSnippetEnabled *bool `json:"funcSnippetEnabled"`

	// GocodeCompletionEnabled is an optional version of
	// Config.GocodeCompletionEnabled
	GocodeCompletionEnabled *bool `json:"gocodeCompletionEnabled"`

	// MaxParallelism is an optional version of Config.MaxParallelism
	MaxParallelism *int `json:"maxParallelism"`

	// UseBinaryPkgCache is an optional version of Config.UseBinaryPkgCache
	UseBinaryPkgCache *bool `json:"useBinaryPkgCache"`
}

type InitializeParams struct {
	lsp.InitializeParams

	InitializationOptions *InitializationOptions `json:"initializationOptions,omitempty"`

	// TODO these should be InitializationOptions

	// NoOSFileSystemAccess makes the server never access the OS file
	// system. It exclusively uses the file overlay (from
	// textDocument/didOpen) and the LSP proxy's VFS.
	NoOSFileSystemAccess bool

	// BuildContext, if set, configures the language server's default
	// go/build.Context.
	BuildContext *InitializeBuildContextParams

	// RootImportPath is the root Go import path for this
	// workspace. For example,
	// "golang.org/x/tools" is the root import
	// path for "github.com/golang/tools".
	RootImportPath string
}

type InitializeBuildContextParams struct {
	// These fields correspond to the fields of the same name from
	// go/build.Context.

	GOOS        string
	GOARCH      string
	GOPATH      string
	GOROOT      string
	CgoEnabled  bool
	UseAllFiles bool
	Compiler    string
	BuildTags   []string

	// Irrelevant fields: ReleaseTags, InstallSuffix.
}
