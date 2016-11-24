// Package cli is split from package sgx to avoid import cycles.
package cli

// ServeInit funcs are executed when the "sourcegraph serve" command handler
// begins execution.
var ServeInit []func()
