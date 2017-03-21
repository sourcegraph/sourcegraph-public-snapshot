package fpath

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	detectCase                    sync.Once
	cachedFileSystemCaseSensitive bool
)

// IsFileSystemCaseSensitive tells if the filesystem (mounted at the temp dir
// location) is case sensitive or not.
func IsFileSystemCaseSensitive() bool {
	detectCase.Do(func() {
		// Create a temporary directory that we'll use for case sensitivity
		// detection.
		dir, err := ioutil.TempDir("", "fpath")
		if err != nil {
			log.Fatal("error detecting FS case sensitivity (TempDir):", err)
		}
		defer os.RemoveAll(dir)

		// Write two files with different case filenames.
		if err := ioutil.WriteFile(filepath.Join(dir, "case"), nil, 0600); err != nil {
			log.Fatal("error detecting FS case sensitivity (WriteFile):", err)
		}
		// ignore err because it's not useful for detection and unclear what
		// error it would return on different platforms in the event of a
		// naming conflict.
		_ = ioutil.WriteFile(filepath.Join(dir, "CaSe"), nil, 0600)

		// If both files exist, the filesystem is case sensitive.
		fi, err := ioutil.ReadDir(dir)
		if err != nil {
			log.Fatal("error detecting FS case sensitivity (ReadDir):", err)
		}
		if len(fi) != 1 && len(fi) != 2 {
			log.Fatal("error detecting FS case sensitivity: found zero or > 2 files")
		}
		cachedFileSystemCaseSensitive = len(fi) == 2
	})
	return cachedFileSystemCaseSensitive
}

type KeyString string

// Key returns a key suitable for representing a filepath in a map. A
// string cannot be used in general due to some filesystems being case
// insensitive (i.e. the key "Foo" should match the key "fOo" in the map).
//
// If the filesystem is case insensitive, a lowercase form of p is returned.
// Otherwise, p is directly returned.
func Key(p string) KeyString {
	if IsFileSystemCaseSensitive() {
		return KeyString(p)
	}
	return KeyString(strings.ToLower(p))
}

func cmpCase(a, b string) (string, string) {
	if !IsFileSystemCaseSensitive() {
		a = strings.ToLower(a)
		b = strings.ToLower(b)
	}
	return a, b
}

// Equal tells if a and b are equal paths. It is identical to string comparison
// of a == b, except that it takes into account case insensitive filesystems.
func Equal(a, b string) bool {
	a, b = cmpCase(a, b)
	return a == b
}

// HasPrefix is like strings.HasPrefix except it properly handles case
// insensitive file paths.
func HasPrefix(s, prefix string) bool { return strings.HasPrefix(cmpCase(s, prefix)) }

// HasSuffix is like strings.HasSuffix except it properly handles case
// insensitive file paths.
func HasSuffix(s, suffix string) bool { return strings.HasSuffix(cmpCase(s, suffix)) }

// TrimPrefix is like strings.TrimPrefix except it first matches the prefix as
// case-insensitive (only on case-insensitive filesystems), then trims the
// prefix if the match is found.
func TrimPrefix(s, prefix string) string {
	if !HasPrefix(s, prefix) {
		return s
	}
	return s[len(prefix):]
}

// CanonicalPathCase returns the filepath input in the canonical case, i.e.
// as seen by a user on the filesystem. This is useful when you have a user
// input file path such as /Foo/baR/bAz which is case insensitive, and an
// (unfortunately) case sensitive code-base to interact with such as
// internal/pkg/notify.
func CanonicalPathCase(input string) (string, error) {
	if IsFileSystemCaseSensitive() {
		return input, nil
	}
	// List the contents of the parent directory of the input.
	dirPath, names, err := readdirnames(filepath.Dir(input))
	if err != nil {
		return "", err
	}
	for _, name := range names {
		name = filepath.Join(dirPath, name)
		if Equal(name, input) {
			// The paths are equal, so `name` is the canonical-case for the
			// path.
			return name, nil
		}
	}
	return "", fmt.Errorf("unable to determine canonical path case for %q", input)
}

// readdirnames reads the names of all files in the specified directory path
// and returns them. The name of the directory, which may not be identical to
// the input path (it may be in canonical case), is returned as the first
// argument.
var readdirnames = func(path string) (string, []string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()
	names, err := f.Readdirnames(-1)
	return f.Name(), names, err
}
