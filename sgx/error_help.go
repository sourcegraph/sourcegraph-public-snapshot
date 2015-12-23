package sgx

import (
	"log"
	"strings"

	"src.sourcegraph.com/sourcegraph/sgx/cli"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// printErrorHelp prints an informational message describing how to
// resolve certain types of errors returned by CLI operations.
func printErrorHelp(err error) {
	// Present the user with instructions on what to do in the event of
	// authentication or connectivity issues.
	code := grpc.Code(err)
	if code == codes.Unauthenticated {
		endpoint := cli.Endpoint.URL
		if endpoint == "" {
			endpoint = "<url-to-sourcegraph-server>"
		}
		log.Printf(`
================================================================================
======== Your 'src' client is not authenticated with the remote server. ========
================================================================================

To authenticate with the server, run:

  src --endpoint=%s login

`, endpoint)
	} else if code == codes.Unknown && strings.Contains(err.Error(), "grpc: the client connection is closing") {
		// TODO(slimsag): determine why the error code for the above Unknown
		// instead of something more concrete that we can rely on (instead of a
		// string containment check).

		endpoint := cli.Endpoint.URLOrDefault()
		log.Printf(`
================================================================================
======== The server at %v is not available.
================================================================================

Check that the server at %v is online and accessible,
or specify a different server using the endpoint flag like so:

  src --endpoint=http://sourcegraph.myteam.org repo create my/repo

`, endpoint, endpoint)
	}
}
