// Package usershell gathers information on the current user and injects then in a context.Context.
package usershell

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

type Shell string

const (
	BashShell Shell = "bash"
	ZshShell  Shell = "zsh"
	FishShell Shell = "fish"
)

// key is used to store userShell in context.
type key struct{}

// userShell stores which shell and which configuration file a user is using.
type userShell struct {
	shell           Shell
	shellPath       string
	shellConfigPath string
}

// fromContext retrieves userShell from context, and may panic if not set, intentionally
// so - unset means usershell.Context must have not been called (a detection failure
// scenario should be handled from the error provided by usershell.Context)
func fromContext(ctx context.Context) userShell {
	return ctx.Value(key{}).(userShell)
}

// ShellPath returns the path to the shell used by the current unix user.
func ShellPath(ctx context.Context) string {
	return fromContext(ctx).shellPath
}

// ShellPath returns the path to the shell configuration (bashrc...) used by the current unix user.
func ShellConfigPath(ctx context.Context) string {
	return fromContext(ctx).shellConfigPath
}

// Shell returns the current shell type used by the current unix user.
func ShellType(ctx context.Context) Shell {
	return fromContext(ctx).shell
}

// IsSupportedShell returns true if the given shell is supported by sg-cli
func IsSupportedShell(ctx context.Context) bool {
	shell := ShellType(ctx)
	return shell == BashShell || shell == ZshShell
}

// GuessUserShell inspect the current environment to infer the shell the current user is running
// and which configuration file it depends on.
func GuessUserShell() (shellPath string, shellConfigPath string, shell Shell, error error) {
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
		shellrc := ".bashrc"
		shell = BashShell
		shellConfigPath = filepath.Join(home, shellrc)
	case strings.Contains(shellPath, "zsh"):
		shellrc := ".zshrc"
		shell = ZshShell
		basePath, ok := os.LookupEnv("ZDOTDIR")
		if !ok {
			basePath = home
		}
		shellConfigPath = filepath.Join(basePath, shellrc)
	case strings.Contains(shellPath, "fish"):
		shellrc := ".config/fish/config.fish"
		shell = FishShell
		shellConfigPath = filepath.Join(home, shellrc)
	}
	return shellPath, shellConfigPath, shell, nil
}

// Context extends ctx with the UserContext of the current user.
func Context(ctx context.Context) (context.Context, error) {
	shell, shellConfigPath, shellType, err := GuessUserShell()
	if err != nil {
		return ctx, err
	}
	userCtx := userShell{
		shell:           shellType,
		shellPath:       shell,
		shellConfigPath: shellConfigPath,
	}
	return context.WithValue(ctx, key{}, userCtx), nil
}
