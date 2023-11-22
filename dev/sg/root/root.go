package root

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sourcegraph/run"

	gitignore "github.com/sabhiram/go-gitignore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var once sync.Once
var repositoryRootValue string
var repositoryRootError error

var ErrNotInsideSourcegraph = errors.New("not running inside sourcegraph/sourcegraph")

// RepositoryRoot caches and returns the value of findRoot.
func RepositoryRoot() (string, error) {
	once.Do(func() {
		// This effectively disables automatic repo detection. This is useful in select automation
		// cases where we really do not need to be sourcegraph/sourcegraph repo ie. generate help docs.
		// Some commands call RepositoryRoot at init time. So we use the environment variable here to allow us
		// to set the repo root as early as possible.
		if forcedRoot := os.Getenv("SG_FORCE_REPO_ROOT"); forcedRoot != "" {
			repositoryRootValue = forcedRoot
		} else {
			repositoryRootValue, repositoryRootError = findRootFromCwd()
		}
	})
	return repositoryRootValue, repositoryRootError
}

// Run executes the given command in repository root. Optionally, path segments relative
// to the repository root can also be provided.
func Run(cmd *run.Command, path ...string) run.Output {
	root, err := RepositoryRoot()
	if err != nil {
		return run.NewErrorOutput(err)
	}
	if len(path) > 0 {
		dir := filepath.Join(append([]string{root}, path...)...)
		return cmd.Dir(dir).Run()
	}
	return cmd.Dir(root).Run()
}

// SkipGitIgnoreWalkFunc wraps the provided walkFn with a function that skips over:
// - files and folders that are ignored by the repository's .gitignore file
// - the contents of the .git directory itself
func SkipGitIgnoreWalkFunc(walkFn fs.WalkDirFunc) fs.WalkDirFunc {
	root, err := RepositoryRoot()
	if err != nil {
		return func(_ string, _ fs.DirEntry, _ error) error {
			return errors.Wrap(err, "getting repository root")
		}
	}

	ignoreFile := filepath.Join(root, ".gitignore")
	additionalLines := []string{
		// We also don't want to traverse the .git directory itself, but it's not going to be
		// specified in the .gitignore file, so we need to provide an extra rule here.
		".git/",
	}

	return skipGitIgnoreWalkFunc(walkFn, ignoreFile, additionalLines...)
}

func skipGitIgnoreWalkFunc(walkFn fs.WalkDirFunc, gitignorePath string, additionalGitIgnoreLines ...string) fs.WalkDirFunc {
	ignore, err := gitignore.CompileIgnoreFileAndLines(gitignorePath, additionalGitIgnoreLines...)
	if err != nil {
		return func(_ string, _ fs.DirEntry, _ error) error {
			return errors.Wrap(err, "compiling .gitignore configuration")
		}
	}

	root := filepath.Dir(gitignorePath)
	wrappedWalkFunc := func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return errors.Wrapf(err, "calculating relative path for %q (root: %q)", path, root)
		}

		if ignore.MatchesPath(relPath) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		return walkFn(path, entry, err)
	}

	return wrappedWalkFunc
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return createSGHome(homeDir)
}

func createSGHome(home string) (string, error) {
	path := filepath.Join(home, ".sourcegraph")
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return "", err
	}
	return path, nil
}
