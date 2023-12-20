package feature

import (
	"path/filepath"

	"github.com/grafana/regexp"
)

type Walkable interface {
	// Walk the repo by file node.
	// Callback should use filepath.SkipDir to skip directory, and StopWalking to finish early.
	// Errors are propagated except these two.
	Walk(func(isDir bool, name string) error) error
}
type StopWalking struct{}

func (err StopWalking) Error() string { return "stop walking" }

type TypeScriptTypeBreak struct{}

func (f TypeScriptTypeBreak) String() string {
	return "TypeScriptTypeBreak"
}

// Distort changes file contents to replace a type annotation with `: string`
// Does not work that well, for instance will replace // TODO: foo with // TODO: string.
func (f TypeScriptTypeBreak) Distort(contents string) string {
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

func (f TypeScriptTypeBreak) ValidateFile(got, want string) bool {
	return got == want
}

func (f TypeScriptTypeBreak) Sample(repo Walkable, callback func(path string) (wantNext bool, err error)) error {
	err := repo.Walk(func(isDir bool, path string) error {
		if !isDir && filepath.Ext(path) == ".ts" {
			wantNext, err := callback(path)
			if err != nil {
				return err
			}
			if !wantNext {
				return StopWalking{}
			}
		}
		if isDir && filepath.Base(path) == "node_modules" {
			return filepath.SkipDir
		}
		return nil

	})
	if err == (StopWalking{}) {
		return nil
	}
	return err
}
