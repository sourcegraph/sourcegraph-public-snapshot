package sgx

import "golang.org/x/net/context"

var (
	// ServerContextFuncs is a list of funcs that are run to initialize
	// the server's context before handling each request.
	//
	// External packages should append to the list at init time if
	// they need to store information in the server's context.
	ServerContextFuncs []func(context.Context) (context.Context, error)
)
