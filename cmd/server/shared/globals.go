package shared

// This file contains global variables that can be modified in a limited fashion by an external
// package (e.g., the enterprise package).

// SrcProfServices defines the default value for SRC_PROF_SERVICES.
//
// If it is modified by an external package, it must be modified immediately on startup, before
// `shared.Main` is called.
//
// This should be kept in sync with dev/src-prof-services.json.
var SrcProfServices = []map[string]string{
	{"Name": "frontend", "Host": "127.0.0.1:6063"},
	{"Name": "gitserver", "Host": "127.0.0.1:6068"},
	{"Name": "searcher", "Host": "127.0.0.1:6069"},
	{"Name": "lsp-proxy", "Host": "127.0.0.1:6061"},
	{"Name": "symbols", "Host": "127.0.0.1:6071"},
	{"Name": "repo-updater", "Host": "127.0.0.1:6074"},
	{"Name": "indexer", "Host": "127.0.0.1:6073"},
	{"Name": "query-runner", "Host": "127.0.0.1:6067"},
}

// ProcfileAdditions is a list of Procfile lines that should be added to the emitted Procfile that
// defines the services configuration.
//
// If it is modified by an external package, it must be modified immediately on startup, before
// `shared.Main` is called.
var ProcfileAdditions []string

// DataDir is the root directory for storing persistent data. It should NOT be modified by any
// external package.
var DataDir = SetDefaultEnv("DATA_DIR", "/var/opt/sourcegraph")
