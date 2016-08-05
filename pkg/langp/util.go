package langp

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

// Cmd is a small helper which logs the command name and parameters and returns
// a command with output going to stdout/stderr.
func Cmd(name string, args ...string) *exec.Cmd {
	s := fmt.Sprintf("exec %s", name)
	for _, arg := range args {
		s = fmt.Sprintf("%s %q", s, arg)
	}
	log.Println(s)
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

// Clone clones the specified repository at the given commit into the specified
// directory. If update is true, this function assumes the git repository
// already exists and can just be fetched / updated.
func Clone(update bool, cloneURI, repoDir, commit string) error {
	if !update {
		c := Cmd("git", "clone", cloneURI, repoDir)
		if err := c.Run(); err != nil {
			return err
		}
	} else {
		// Update our repo to match the remote.
		c := Cmd("git", "remote", "update", "--prune")
		c.Dir = repoDir
		if err := c.Run(); err != nil {
			return err
		}
	}

	// Reset to the specific revision.
	c := Cmd("git", "reset", "--hard", commit)
	c.Dir = repoDir
	return c.Run()
}
