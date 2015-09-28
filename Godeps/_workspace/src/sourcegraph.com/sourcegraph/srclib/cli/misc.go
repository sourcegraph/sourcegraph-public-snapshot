package cli

import (
	"log"
	"os"
	"path/filepath"

	"sourcegraph.com/sourcegraph/go-flags"
)

// Directory is flags.Completer that provides directory name
// completion. Do not convert Directory to a string type manually,
// always use Directory.String(). Only use Directory for
// go-flags.Command fields, not for internal functions.
//
// TODO(sqs): this is annoying. it only completes the dir name and doesn't let
// you keep typing the arg.
type Directory string

// Complete implements flags.Completer and returns a list of existing
// directories with the given prefix.
func (d Directory) Complete(match string) []flags.Completion {
	names, err := filepath.Glob(match + "*")
	if err != nil {
		log.Println(err)
		return nil
	}

	var dirs []flags.Completion
	for _, name := range names {
		if fi, err := os.Stat(name); err == nil && fi.Mode().IsDir() {
			dirs = append(dirs, flags.Completion{Item: name + "/"})
		}
	}
	return dirs
}

// String returns the uncleaned string representation of d. If d is
// empty, "." is returned. Never convert Directories to strings
// manually, always call String.
func (d Directory) String() string {
	if d == "" {
		return "."
	}
	dir, file := filepath.Split(string(d))
	if file == "" {
		return dir
	}
	if file == "." || file == ".." {
		return dir + file
	}
	return dir + file + string(os.PathSeparator)
}
