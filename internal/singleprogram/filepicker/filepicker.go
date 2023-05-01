// package filepicker has a best-effort implementation of GUI file dialogs.
//
// The implementation looks for executables that can be run.
package filepicker

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/syncx"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func osascript(ctx context.Context, allowMultiple bool) ([]string, error) {
	var paths []string

	const promptMultiple = `set theFolders to choose folder with prompt "Select repositories or folders with repositories" with multiple selections allowed`
	const promptSingle = `set theFolders to choose folder with prompt "Select a repository or folder with repositories"`
	var prompt string
	if allowMultiple {
		prompt = promptMultiple
	} else {
		prompt = promptSingle
	}

	cmd := exec.CommandContext(ctx,
		"osascript", "-e",
		prompt,
		"-e",
		"set posixPaths to {}",
		"-e",
		"repeat with thisFolder in theFolders",
		"-e",
		"    set end of posixPaths to POSIX path of thisFolder",
		"-e",
		"end repeat",
		"-e",
		"return posixPaths")
	bytePaths, err := cmd.Output()
	if err != nil {
		return paths, err
	}

	// Output looks something like "/path/to/dir1/, /path/to/dir2/\n". We can reliably strip
	// out the trailing new line and then split on the comma+space.
	stringPaths := trimTrailingNewline(bytePaths)
	paths = strings.Split(stringPaths, ", ")
	for i, path := range paths {
		if len(path) > 0 {
			paths[i] = strings.TrimSuffix(path, "/")
		}
	}

	return paths, nil
}

func zenity(ctx context.Context, allowMultiple bool) ([]string, error) {
	// nix-shell -p gnome.zenity
	cmd := exec.CommandContext(ctx, "zenity", "--file-selection", "--directory")
	path, err := cmd.Output()
	if err != nil {
		return []string{}, err
	}

	return []string{trimTrailingNewline(path)}, nil
}

func kdialog(ctx context.Context, allowMultiple bool) ([]string, error) {
	// nix-shell -p kdialog
	cmd := exec.CommandContext(ctx, "kdialog", "--getexistingdirectory")
	pathRaw, err := cmd.Output()
	if err != nil {
		return []string{}, err
	}

	path := trimTrailingNewline(pathRaw)

	// kdialog may return a file, if so pick its parent
	if info, err := os.Lstat(path); err != nil {
		return []string{}, errors.Wrap(err, "failed to Lstat user selected path via kdialog")
	} else if !info.IsDir() {
		return []string{filepath.Dir(path)}, nil
	}

	return []string{path}, nil
}

func trimTrailingNewline(b []byte) string {
	return string(bytes.TrimSuffix(b, []byte{'\n'}))
}

// Picker returns the filepath to a directory a user picked. It should exclude
// the trailing /. If the operation times out or the user cancels, an error is
// returned.
type Picker func(ctx context.Context, allowMultiple bool) ([]string, error)

// Lookup finds a Picker to run. If no Picker can be found, ok is false.
func Lookup(logger log.Logger) (_ Picker, ok bool) {
	pickers := []struct {
		// Cmd must exist on PATH for Run to work.
		Cmd string
		// RequiredEnv is an optional envvar needed for Cmd to work.
		RequiredEnv string
		// Run is the picker implementation.
		Run Picker
	}{{
		Cmd: "osascript",
		Run: osascript,
	}, {
		Cmd:         "zenity",
		RequiredEnv: "DISPLAY",
		Run:         zenity,
	}, {
		Cmd:         "kdialog",
		RequiredEnv: "DISPLAY",
		Run:         kdialog,
	}}

	for _, picker := range pickers {
		if _, err := exec.LookPath(picker.Cmd); err != nil {
			logger.Debug("skipping filepicker due to not being on PATH", log.String("cmd", picker.Cmd), log.Error(err))
			continue
		}

		if picker.RequiredEnv != "" && os.Getenv(picker.RequiredEnv) == "" {
			logger.Debug("skipping filepicker due to missing envvar", log.String("cmd", picker.Cmd), log.String("envvar", picker.RequiredEnv))
			continue
		}

		logger.Debug("found filepicker", log.String("cmd", picker.Cmd))
		return picker.Run, true
	}

	logger.Debug("no filepicker found")
	return nil, false
}

// Available returns true if the filepicker API is available. IE if Lookup returns true for ok.
//
// Note: This is cached on the first call, so may diverge from Lookup if what
// is on PATH changes during the lifetime of the process. It is cached since
// computing it requires several system calls (PATH lookup, filepath.Abs,
// os.Getenv).
var Available = syncx.OnceValue(func() bool {
	_, ok := Lookup(log.Scoped("filepicker", ""))
	return ok
})
