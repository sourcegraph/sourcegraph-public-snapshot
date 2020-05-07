package internal

import (
	"fmt"
	"path/filepath"
)

// CloneParallel ensures that all of the given repos have been cloned into the cache dir.
func CloneParallel(cacheDir string, maxConcurrency int, repos []Repo) error {
	var fns []FnPair
	for _, repo := range repos {
		localRepo := repo

		fns = append(fns, FnPair{
			Fn: func() error {
				return clone(cacheDir, localRepo)
			},
			Description: fmt.Sprintf("Cloning %s/%s", localRepo.Owner, localRepo.Name),
		})
	}

	return RunParallel(maxConcurrency, fns)
}

// clone clones the given repo into the cache dir if it doesn't already exist.
func clone(cacheDir string, repo Repo) error {
	targetDir := filepath.Join(cacheDir, "repos")
	if exists, err := pathExists(filepath.Join(targetDir, repo.Name)); err != nil || exists {
		return err
	}

	cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", repo.Owner, repo.Name)
	return runCommand(targetDir, "git", "clone", cloneURL)
}
