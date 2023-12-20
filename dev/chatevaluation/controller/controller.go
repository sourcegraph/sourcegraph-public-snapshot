package controller

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/regexp"

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

type Repo interface {
	Read(path string) (string, error)
	Update(path string, newContents string) error
	// Walk the repo by file node. filepath.SkipDir works.
	Walk(func(isDir bool, name string) error) error
}

// TODO: Extract configuration.
func Run(repo string) error {
	// TODO: Move repo to a parameter
	r := localRepo(repo)
	filePaths, err := sample(r, 5)
	if err != nil {
		return err
	}
	for _, path := range filePaths {
		contents, err := r.Read(path)
		if err != nil {
			return err
		}
		distorted := distort(string(contents))
		if err := r.Update(path, distorted); err != nil {
			return err
		}
		diagnosef("Diff:\n%s", cmp.Diff(distorted, contents))
		if err := runCody(path); err != nil {
			return err
		}
		if err := validateFile(path); err != nil {
			return err
		}
		// Roll back the file change.
		if err := r.Update(path, contents); err != nil {
			return err
		}
	}
	return nil
}

// Distort works on TypeScript files and changes a non-string type declaration it finds to : string.
// Does not work that well, for instance will replace // TODO: foo with // TODO: string.
func distort(contents string) string {
	typeAnnotation := regexp.MustCompile(`:\s*([a-zA-Z\[\]<>.]+)`)
	matches := typeAnnotation.FindStringSubmatch(contents)
	if len(matches) > 0 {
		if matches[1] != ": string" {
			var replaced bool
			return typeAnnotation.ReplaceAllStringFunc(contents, func(typ string) string {
				if replaced {
					return typ
				}
				if typ != ": string" {
					replaced = true
				}
				return ": string"
			})
		}
	}
	return contents
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

func validateFile(filePath string) error {
	fmt.Println("Pretending to validate file")
	return nil
}
