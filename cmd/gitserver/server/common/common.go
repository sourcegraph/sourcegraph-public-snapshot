package common

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// GitDir is an absolute path to a GIT_DIR.
// They will all follow the form:
//
//	${s.ReposDir}/${name}/.git
type GitDir string

// Path is a helper which returns filepath.Join(dir, elem...)
func (dir GitDir) Path(elem ...string) string {
	return filepath.Join(append([]string{string(dir)}, elem...)...)
}

// Set updates cmd so that it will run in dir.
//
// Note: GitDir is always a valid GIT_DIR, so we additionally set the
// environment variable GIT_DIR. This is to avoid git doing discovery in case
// of a bad repo, leading to hard to diagnose error messages.
func (dir GitDir) Set(cmd *exec.Cmd) {
	cmd.Dir = string(dir)
	if cmd.Env == nil {
		// Do not strip out existing env when setting.
		cmd.Env = os.Environ()
	}
	cmd.Env = append(cmd.Env, "GIT_DIR="+string(dir))
}

// GitCommandError is an error of a failed Git command.
type GitCommandError struct {
	// Err is the original error produced by the git command that failed.
	Err error
	// Output is the std error output of the command that failed.
	Output string
}

func (e *GitCommandError) Error() string {
	return fmt.Sprintf("%s - output: %q", e.Err, e.Output)
}

func (e *GitCommandError) Unwrap() error {
	return e.Err
}
