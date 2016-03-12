package python

import (
	"crypto/sha1"
	"encoding/hex"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"sourcegraph.com/sourcegraph/srclib/toolchain"
)

func runCmdLogError(cmd *exec.Cmd) {
	err := runCmdStderr(cmd)
	if err != nil {
		log.Printf("Error running `%s`: %s", strings.Join(cmd.Args, " "), err)
	}
}

func runCmdStderr(cmd *exec.Cmd) error {
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stderr
	return cmd.Run()
}

// getVENVBinPath returns toolchains Python virtualenv path. If toolchain is ran in
// `docker` mode, it will return empty string because there is no virtualenv.
func getVENVBinPath() (string, error) {
	if os.Getenv("IN_DOCKER_CONTAINER") == "" {
		tc, err := toolchain.Lookup("sourcegraph.com/sourcegraph/srclib-python")
		if err != nil {
			return "", err
		}
		return filepath.Join(tc.Dir, ".env", "bin"), nil
	}
	return "", nil
}

func getHash(text string) string {
	hasher := sha1.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))[:8]
}
