package worker

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"sourcegraph.com/sourcegraph/srclib"
	"sourcegraph.com/sourcegraph/srclib/flagutil"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
	"src.sourcegraph.com/sourcegraph/sgx/sgxcmd"
)

// cmdWithClientArgs prepends --endpoint, etc., and their current
// values to args and then returns a command that executes the
// Sourcegraph program with those options and args.
func cmdWithClientArgs(args ...string) *exec.Cmd {
	endpointArgs, err := flagutil.MarshalArgs(&cli.Endpoint)
	if err != nil {
		log.Fatal(err)
	}

	// Refresh the global access token if it is set to expire soon.
	// The token source for cliCtx is set in serve_cmd.go to be
	// a sharedsecret.DefensiveTokenSource, which will update its
	// access token if it is going to expire soon.
	defensiveTokenSource := sourcegraph.CredentialsFromContext(cli.Ctx)
	if defensiveTokenSource != nil {
		// This will update the global token source (i.e. the Credentials object).
		_, err := defensiveTokenSource.Token()
		if err != nil {
			log.Fatal(err)
		}
	}

	credsArgs, err := flagutil.MarshalArgs(&cli.Credentials)
	if err != nil {
		log.Fatal(err)
	}

	clientArgs := append(endpointArgs, credsArgs...)
	return exec.Command(sgxcmd.Path, append(clientArgs, args...)...)
}

func execCmdInDir(dir, prog string, args ...string) error {
	if prog == strings.Split(srclib.CommandName, " ")[0] || prog == sgxcmd.Name || prog == sgxcmd.Path {
		panic("You shouldn't execute srclib or sourcegraph with this helper function, since they need the config flags that cmdWithClientArgs provides. Use that func instead.")
	}

	cmd := exec.Command(prog, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s %v failed: %s\n\nOutput was: %s", prog, args, err, out)
	}
	return nil
}
