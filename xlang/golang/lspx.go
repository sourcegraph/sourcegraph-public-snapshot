package golang

import "sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"

// This file contains Go-specific extensions to LSP types for
// communication between the build and language servers.
//
// All custom param types are for build -> language server
// communication. All custom result types are for language -> build
// server communication.
//
// The Go language server MUST NOT rely on these extensions for
// standalone operation on the local file system.

type initializeParams struct {
	lsp.InitializeParams

	// NoOSFileSystemAccess makes the server never access the OS file
	// system. It exclusively uses the file overlay (from
	// textDocument/didOpen) and the LSP proxy's VFS.
	NoOSFileSystemAccess bool

	// BuildContext, if set, configures the language server's default
	// go/build.Context.
	BuildContext *initializeBuildContextParams

	// RootImportPath is the root Go import path for this
	// workspace. For example,
	// "sourcegraph.com/sourcegraph/sourcegraph" is the root import
	// path for "github.com/sourcegraph/sourcegraph".
	RootImportPath string `json:"rootImportPath,omitempty"`
}

type initializeBuildContextParams struct {
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
