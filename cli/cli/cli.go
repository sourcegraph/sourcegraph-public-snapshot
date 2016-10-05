// Package cli is split from package sgx to avoid import cycles.
package cli

import (
	"sourcegraph.com/sourcegraph/go-flags"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/srccmd"
)

var CLI = flags.NewNamedParser(srccmd.Name, flags.PrintErrors|flags.PassDoubleDash)

// PostInit funcs are executed after all sgx init funcs have been run.
var PostInit []func()

// CustomHelpCmds is a list of registered command names which should not have
// the default help group registered for them.
var CustomHelpCmds []string

// Serve is the "sourcegraph serve" command group.
var Serve *flags.Command

// Internal is the "sourcegraph internal" command group.
var Internal *flags.Command

// ServeInit funcs are executed when the "sourcegraph serve" command handler
// begins execution.
var ServeInit []func()
