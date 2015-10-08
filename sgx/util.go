package sgx

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"strings"

	"sourcegraph.com/sourcegraph/srclib"
	"sourcegraph.com/sourcegraph/srclib/flagutil"
	"src.sourcegraph.com/sourcegraph/sgx/sgxcmd"
)

// cmdWithClientArgs prepends --endpoint, --grpc-endpoint, etc., and
// their current values to args and then returns a command that
// executes the Sourcegraph program with those options and args.
func cmdWithClientArgs(args ...string) *exec.Cmd {
	endpointArgs, err := flagutil.MarshalArgs(&Endpoints)
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

// openTempFileInEditor runs $EDITOR with a temp file that contains
// contents. It returns the final contents of the file after editing.
func openTempFileInEditor(contents []byte) ([]byte, error) {
	f, err := ioutil.TempFile("", "src")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	defer os.Remove(f.Name())
	if _, err := f.Write(contents); err != nil {
		return nil, err
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		return nil, errors.New("no EDITOR environment variable set")
	}

	cmd := exec.Command(editor, f.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	// Seeking to the beginning and reading the file's contents does
	// not work reliably, for some reason. For example, if EDITOR=vi,
	// then it sees an empty file. So, just call ReadFile.
	return ioutil.ReadFile(f.Name())
}
