package root

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var once sync.Once
var repositoryRootValue string
var repositoryRootError error

var ErrNotInsideSourcegraph = errors.New("not running inside sourcegraph/sourcegraph")

// RepositoryRoot caches and returns the value of findRoot.
func RepositoryRoot() (string, error) {
	once.Do(func() { repositoryRootValue, repositoryRootError = findRootFromCwd() })
	return repositoryRootValue, repositoryRootError
}

// findRootFromCwd finds root path of the sourcegraph/sourcegraph repository from
// the current working directory. Is it an error to run this binary outside
// of the repository.
func findRootFromCwd() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return findRoot(wd)
}

// findRoot finds the root path of sourcegraph/sourcegraph from wd
func findRoot(wd string) (string, error) {
	for {
		contents, err := os.ReadFile(filepath.Join(wd, "go.mod"))
		if err == nil {
			for _, line := range strings.Split(string(contents), "\n") {
				if line == "module github.com/sourcegraph/sourcegraph" {
					return wd, nil
				}
			}
		} else if !os.IsNotExist(err) {
			return "", err
		}

		if parent := filepath.Dir(wd); parent != wd {
			wd = parent
			continue
		}

		return "", ErrNotInsideSourcegraph
	}
}

func GetSGHomePath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(homedir, ".sourcegraph")
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return "", err
	}
	return path, nil
}
