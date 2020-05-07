package internal

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// IndexParallel ensures that all of the given revisions have been indexed.
func IndexParallel(cacheDir string, maxConcurrency int, repos []Repo) error {
	var fns []FnPair
	for _, repo := range repos {
		localRepo := repo

		for _, rev := range repo.Revs {
			localRev := rev

			fns = append(fns, FnPair{
				Fn: func() error {
					return index(cacheDir, localRepo.Owner, localRepo.Name, localRev)
				},
				Description: fmt.Sprintf("Indexing %s/%s@%s", localRepo.Owner, localRepo.Name, localRev),
			})
		}
	}

	return RunParallel(maxConcurrency, fns)
}

// index performs an lsif-go index operation on the given repo and revision.
func index(cacheDir, owner, name, rev string) error {
	indexFile, err := filepath.Abs(filepath.Join(cacheDir, "indexes", fmt.Sprintf("%s.%s.dump", name, rev)))
	if err != nil {
		return err
	}
	if exists, err := pathExists(indexFile); err != nil || exists {
		return err
	}

	tempDir, err := ioutil.TempDir(cacheDir, "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	sourceDir := filepath.Join(cacheDir, "repos", name)
	targetDir := filepath.Join(tempDir, name)

	commands := []func() error{
		func() error { return runCommand("", "cp", "-r", sourceDir, tempDir) },
		func() error { return runCommand(targetDir, "git", "checkout", rev) },
		func() error { return runCommand(targetDir, "go", "mod", "vendor") },
		func() error { return runCommand(targetDir, "lsif-go", "-o", indexFile) },
	}

	for _, command := range commands {
		if err := command(); err != nil {
			return err
		}
	}
	return nil
}
