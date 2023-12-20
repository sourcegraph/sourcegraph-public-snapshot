package controller

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/dev/chatevaluation/feature"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Dependencies of evaluation run universal across languages.
var (
	// TODO: Extract as sampling implementation is actually determined for TypeScript.
	// TODO: We should keep sampling in case we can't distort files.
	sample = func(repo Repo, count int) ([]string, error) {
		filePaths := make([]string, 0)
		err := repo.Walk(func(isDir bool, path string) error {
			if !isDir && filepath.Ext(path) == ".ts" {
				filePaths = append(filePaths, path)
			}
			if isDir && filepath.Base(path) == "node_modules" {
				return filepath.SkipDir
			}
			return nil

		})
		if err != nil {
			return nil, err
		}
		rand.Shuffle(len(filePaths), func(i, j int) { filePaths[i], filePaths[j] = filePaths[j], filePaths[i] })
		if len(filePaths) < count {
			return nil, errors.Newf("Fewer than %d TypeScript files found", count)
		}
		filePaths = filePaths[:count]
		return filePaths, nil
	}
	diagnosef = func(line string, args ...any) {
		fmt.Printf(line+"\n", args...)
	}
)

// Repo abstracts access to a repository. It assumes that Cody also has access
// to read and modify files in that repository.
type Repo interface {
	feature.Walkable
	// Read returns the contents of a file in the repo given a path.
	Read(path string) (string, error)
	// Update overwrites contents of a file at given path with given contents.
	Update(path string, newContents string) error
}

// TestCase represents a single code generating chat interaction that we are testing against.
// This chat interaction can be applied to multiple files.
type TestCase interface {
	// Sample iterates through the repo and selects files relevant for the test case.
	// Once the callback captures a file, it can request more with wantNext. Errors are propagated.
	Sample(repo feature.Walkable, callback func(path string) (wantNext bool, err error)) error
	// Distort breaks a file contents. Input is a valid file from a repo that builds OK.
	// A test case is then modifying the file to give Cody opportunity to fix it.
	Distort(contents string) string
	// ValidateFile returns true if the file we got after Cody's changes is valid.
	// We will need to expand that API in order to capture different degrees of validity.
	ValidateFile(got, want string) bool
}

// TODO: Extract configuration.
func Run(repo string) error {
	// TODO: Move repo to a parameter
	r := localRepo(repo)
	f := feature.TypeScriptTypeBreak{}
	var (
		countTotal int = 0
		countPass  int = 0
	)
	f.Sample(r, func(path string) (wantNext bool, err error) {
		diagnosef("Testing %q", path)
		original, err := r.Read(path)
		if err != nil {
			return false, err
		}
		distorted := f.Distort(original)
		diff := cmp.Diff(distorted, original)
		// Continue to next file if no diff after distorting:
		if diff == "" {
			return true, nil
		}
		diagnosef("Diff:\n%s", diff)
		if err := r.Update(path, distorted); err != nil {
			return false, err
		}
		if err := runCody(path); err != nil {
			return false, err
		}
		fixed, err := r.Read(path)
		if err != nil {
			return false, err
		}
		countTotal++
		if f.ValidateFile(fixed, original) {
			countPass++
		}
		if err := r.Update(path, original); err != nil {
			return false, err
		}
		return countTotal < 5, nil
	})
	diagnosef("Total %d cases, %d passed", countTotal, countPass)
	return nil
}

type localRepo string

func (r localRepo) Read(path string) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}
func (r localRepo) Update(path string, newContents string) error {
	return os.WriteFile(path, []byte(newContents), 0644)
}
func (r localRepo) Walk(f func(isDir bool, path string) error) error {
	return filepath.Walk(string(r), func(path string, info os.FileInfo, err error) error {
		return f(info.IsDir(), path)
	})
}

func runCody(filePath string) error {
	fmt.Printf("Pretending to run Cody on %q\n", filePath)
	return nil
}
