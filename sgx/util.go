package sgx

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"

	"strings"

	"sourcegraph.com/sourcegraph/srclib"
	"sourcegraph.com/sourcegraph/srclib/flagutil"
	"src.sourcegraph.com/sourcegraph/sgx/sgxcmd"
)

// cmdWithClientArgs prepends --endpoint, etc., and their current
// values to args and then returns a command that executes the
// Sourcegraph program with those options and args.
func cmdWithClientArgs(args ...string) *exec.Cmd {
	endpointArgs, err := flagutil.MarshalArgs(&Endpoint)
	if err != nil {
		log.Fatal(err)
	}

	credsArgs, err := flagutil.MarshalArgs(&Credentials)
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

func getLine() (string, error) {
	var line string
	s := bufio.NewScanner(os.Stdin)
	if s.Scan() {
		line = s.Text()
	}
	return line, s.Err()
}
