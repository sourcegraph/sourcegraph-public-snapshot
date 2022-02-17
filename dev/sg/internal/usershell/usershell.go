// Package usershell gathers information on the current user and injects then in a context.Context.
package usershell

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type key struct{}

// userShell stores which shell and which configuration file a user is using.
type userShell struct {
	shellPath       string
	shellConfigPath string
}

// ShellPath returns the path to the shell used by the current unix user.
func ShellPath(ctx context.Context) string {
	v := ctx.Value(key{}).(userShell)
	return v.shellPath
}

// ShellPath returns the path to the shell configuration (bashrc...) used by the current unix user.
func ShellConfigPath(ctx context.Context) string {
	v := ctx.Value(key{}).(userShell)
	return v.shellConfigPath
}

// GuessUserShell inspect the current environment to infer the shell the current user is running
// and which configuration file it depends on.
func GuessUserShell() (string, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	// Look up which shell the user is using, because that's most likely the
	// one that has all the environment correctly setup.
	shell, ok := os.LookupEnv("SHELL")
	var shellrc string
	if !ok {
		// If we can't find the shell in the environment, we fall back to `bash`
		shell = "bash"
	}
	switch {
	case strings.Contains(shell, "bash"):
		shellrc = ".bashrc"
	case strings.Contains(shell, "zsh"):
		if _, err := os.Stat(path.Join(home, ".zshrc")); errors.Is(err, os.ErrNotExist) {
			// A fresh mac installation with standard homebrew will tell the user to append
			// the configuration in .zprofile, not .zshrc.
			shellrc = ".zprofile"
		} else {
			shellrc = ".zshrc"
		}
	}
	return shell, filepath.Join(home, shellrc), nil
}

// Context extends ctx with the UserContext of the current user.
func Context(ctx context.Context) (context.Context, error) {
	shell, shellConfigPath, err := GuessUserShell()
	if err != nil {
		return nil, err
	}
	userCtx := userShell{
		shellPath:       shell,
		shellConfigPath: shellConfigPath,
	}
	return context.WithValue(ctx, key{}, userCtx), nil
}

// Cmd returns a command wrapped in a new shell process, enabling
// changes added by various checks to be run. This negates the new to ask the
// user to restart sg for many checks.
func Cmd(ctx context.Context, cmd string) *exec.Cmd {
	if runtime.GOOS == "linux" {
		// The default Ubuntu bashrc comes with a caveat that prevents the bashrc to be
		// reloaded unless the shell is interactive. Therefore, we need to request for an
		// interactive one.
		//
		// But because we are running an interactive shell, we also need to exit explictly.
		// To avoid messing up with the output checking that depends on this function,
		// we silence the exit commands, which otherwise, prints "exit".
		command := fmt.Sprintf("%s; \nexit $? 2>/dev/null", strings.TrimSpace(cmd))
		return exec.CommandContext(ctx, ShellPath(ctx), "-c", "-i", command)
	} else {
		// The above interactive shell approach fails on OSX because the default shell configuration
		// prints sessions restoration informations that will mess with the output. So we fall back
		// to manually reloading the shell configuration.
		command := fmt.Sprintf("source %s || true; %s", ShellConfigPath(ctx), cmd)
		return exec.CommandContext(ctx, ShellPath(ctx), "-c", command)
	}
}

// CombinedExec runs a command in a fresh shell environment, and returns
// stderr and stdout combined, along with an error.
func CombinedExec(ctx context.Context, cmd string) ([]byte, error) {
	if cmd == "" {
		return nil, errors.Errorf("can't execute empty command")
	}
	return Cmd(ctx, cmd).CombinedOutput()
}
