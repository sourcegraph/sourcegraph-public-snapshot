package root

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var once sync.Once
var repositoryRootValue string
var repositoryRootError error

// RepositoryRoot caches and returns the value of findRoot.
func RepositoryRoot() (string, error) {
	once.Do(func() { repositoryRootValue, repositoryRootError = findRoot() })
	return repositoryRootValue, repositoryRootError
}

// findRoot finds root path of the sourcegraph/sourcegraph repository from
// the current working directory. Is it an error to run this binary outside
// of the repository.
func findRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		contents, err := ioutil.ReadFile(filepath.Join(wd, "go.mod"))
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

		return "", fmt.Errorf("not running inside sourcegraph/sourcegraph")
	}
}
