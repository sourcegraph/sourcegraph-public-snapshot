package controller

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"

	"github.com/google/go-cmp/cmp"
)

func Run(repo string) error {

	filePaths, err := sample(repo, 5)
	if err != nil {
		return err
	}

	fmt.Println("Got 5 files:")
	for _, path := range filePaths {
		fmt.Println(path)
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		distorted := distort(string(contents))
		diff := cmp.Diff(string(contents), distorted)
		if diff == "" {
			fmt.Println("No distortion")
			continue
		}
		fmt.Println(diff)
		// TODO: Apply diff to repo
		if err := runCody(path); err != nil {
			return err
		}
		if err := validateFile(path); err != nil {
			return err
		}
		// TODO: Roll back the transformation.
	}

	return nil
}

func sample(repo string, count int) ([]string, error) {
	filePaths := make([]string, 0)
	err := filepath.Walk(repo, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".ts" {
			filePaths = append(filePaths, path)
		}
		if info.IsDir() && info.Name() == "node_modules" {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	rand.Shuffle(len(filePaths), func(i, j int) { filePaths[i], filePaths[j] = filePaths[j], filePaths[i] })

	if len(filePaths) < count {
		return nil, fmt.Errorf("Fewer than %d TypeScript files found", count)
	}
	filePaths = filePaths[:count]

	return filePaths, nil
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

func runCody(filePath string) error {
	fmt.Printf("Pretending to run Cody on %q\n", filePath)
	return nil
}

func validateFile(filePath string) error {
	fmt.Println("Pretending to validate file")
	return nil
}
