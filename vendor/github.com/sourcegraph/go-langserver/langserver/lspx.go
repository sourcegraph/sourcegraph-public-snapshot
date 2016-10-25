package langserver

import "github.com/sourcegraph/go-langserver/pkg/lsp"

// This file contains Go-specific extensions to LSP types.
//
// The Go language server MUST NOT rely on these extensions for
// standalone operation on the local file system. (VSCode has no way
// of including these fields.)

type InitializeParams struct {
	lsp.InitializeParams

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
