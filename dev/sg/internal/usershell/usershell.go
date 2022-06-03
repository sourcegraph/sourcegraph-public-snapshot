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

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type key struct{}

type Shell string

const (
	BashShell Shell = "bash"
	ZshShell  Shell = "zsh"
	FishShell Shell = "fish"
)

// userShell stores which shell and which configuration file a user is using.
type userShell struct {
	shell           Shell
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

// Shell returns the current shell type used by the current unix user.
func ShellType(ctx context.Context) Shell {
	v := ctx.Value(key{}).(userShell)
	return v.shell
}

// GuessUserShell inspect the current environment to infer the shell the current user is running
// and which configuration file it depends on.
func GuessUserShell() (shellPath string, shellrc string, shell Shell, error error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", "", err
	}
	// Look up which shell the user is using, because that's most likely the
	// one that has all the environment correctly setup.
	shellPath, ok := os.LookupEnv("SHELL")
	if !ok {
		// If we can't find the shell in the environment, we fall back to `bash`
		shellPath = "bash"
		shell = BashShell
	}
	switch {
	case strings.Contains(shellPath, "bash"):
		shellrc = ".bashrc"
		shell = BashShell
	case strings.Contains(shellPath, "zsh"):
		if _, err := os.Stat(path.Join(home, ".zshrc")); errors.Is(err, os.ErrNotExist) {
			// A fresh mac installation with standard homebrew will tell the user to append
			// the configuration in .zprofile, not .zshrc.
			shellrc = ".zprofile"
		} else {
			shellrc = ".zshrc"
		}
		shell = ZshShell
	case strings.Contains(shellPath, "fish"):
		shellrc = ".config/fish/config.fish"
		shell = FishShell
	}
	return shellPath, filepath.Join(home, shellrc), shell, nil
}

// Context extends ctx with the UserContext of the current user.
func Context(ctx context.Context) (context.Context, error) {
	shell, shellConfigPath, shellType, err := GuessUserShell()
	if err != nil {
		return nil, err
	}
	userCtx := userShell{
		shell:           shellType,
		shellPath:       shell,
		shellConfigPath: shellConfigPath,
	}
	return context.WithValue(ctx, key{}, userCtx), nil
}

type wrapped struct {
	ShellPath  string
	ShellFlags []string
	Command    string
	Environ    []string
}

// wrap builds a wrapping for executing the given command in a new shell process.
func wrap(ctx context.Context, cmd string) wrapped {
	// Defaults
	w := wrapped{
		ShellPath:  ShellPath(ctx),
		ShellFlags: []string{"-c"},
		Command:    cmd,
		Environ:    os.Environ(),
	}

	switch {
	case os.Getenv("SG_DEV_NO_RELOAD_ENV") != "":
		// If the user does not want the auto env reloading mechanism, just
		// perform a standard command.

	case ShellType(ctx) == FishShell:
		w.Command = fmt.Sprintf("fish || true; %s", cmd)

	case runtime.GOOS == "linux":
		// The default Ubuntu bashrc comes with a caveat that prevents the bashrc to be
		// reloaded unless the shell is interactive. Therefore, we need to request for an
		// interactive one.
		//
		// But because we are running an interactive shell, we also need to exit explictly.
		// To avoid messing up with the output checking that depends on this function,
		// we silence the exit commands, which otherwise, prints "exit".
		w.Command = fmt.Sprintf("%s; \nexit $? 2>/dev/null", strings.TrimSpace(cmd))
		w.ShellFlags = append(w.ShellFlags, "-i")

	default:
		// The above interactive shell approach fails on OSX because the default shell configuration
		// prints sessions restoration informations that will mess with the output. So we fall back
		// to manually reloading the shell configuration.
		w.Command = fmt.Sprintf("source %s || true; %s", ShellConfigPath(ctx), cmd)
	}

	if ShellType(ctx) == ZshShell {
		// Set this env var for oh-my-zsh users so that oh-my-zsh does not try to
		// auto-update itself when we're restarting the shell.
		w.Environ = append(w.Environ, "DISABLE_AUTO_UPDATE=true")
	}

	return w
}

// Cmd returns a command wrapped in a new shell process, enabling
// changes added by various checks to be run. This negates the new to ask the
// user to restart sg for many checks.
func Cmd(ctx context.Context, cmd string) *exec.Cmd {
	w := wrap(ctx, cmd)

	wrappedCmd := exec.CommandContext(ctx, w.ShellPath, append(w.ShellFlags, w.Command)...)
	wrappedCmd.Env = w.Environ
	return wrappedCmd
}

// CombinedExec runs a command in a fresh shell environment, and returns
// stderr and stdout combined, along with an error.
func CombinedExec(ctx context.Context, cmd string) ([]byte, error) {
	if cmd == "" {
		return nil, errors.Errorf("can't execute empty command")
	}
	return Cmd(ctx, cmd).CombinedOutput()
}

// Run runs a command in a fresh shell environment, and returns run.Output for easy usage.
func Run(ctx context.Context, cmd string) run.Output {
	w := wrap(ctx, cmd)
	return run.Cmd(ctx, w.ShellPath, strings.Join(w.ShellFlags, " "), run.Arg(w.Command)).
		Environ(w.Environ).
		Run()
}

// IsSupportedShell returns true if the given shell is supported by sg-cli
func IsSupportedShell(ctx context.Context) bool {
	shell := ShellType(ctx)
	return shell == BashShell || shell == ZshShell
}
