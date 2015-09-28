// Package cli is split from package sgx to avoid import cycles.
package cli

import (
	"net/http"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-flags"
	"sourcegraph.com/sourcegraph/sourcegraph/sgx/sgxcmd"
)

var CLI = flags.NewNamedParser(sgxcmd.Name, flags.PrintErrors|flags.PassDoubleDash)

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

// ServeMuxFuncs are called with the `src serve` subcommand's
// http.ServeMux as the argument. They can attach handlers to the
// ServeMux prior to server startup.
var ServeMuxFuncs []func(*http.ServeMux)

var (
	// ServerContext is a list of funcs that are run to initialize the
	// server's context before each request is handled.
	//
	// External packages should append to the list at init time if
	// they need to store information in the server's context.
	ServerContext []func(context.Context) context.Context
)
