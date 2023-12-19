package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"

	"github.com/google/go-cmp/cmp"
)

func main() {
	repo := flag.String("repo", "", "repository root")
	flag.Parse()

	if *repo == "" {
		fmt.Println("Please specify a repository root")
		return
	}

	filePaths, err := sample(*repo, 5)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Got 5 files:")
	for _, path := range filePaths {
		fmt.Println(path)
		contents, err := os.ReadFile(path)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		distorted := distort(string(contents))
		diff := cmp.Diff(string(contents), distorted)
		if diff == "" {
			fmt.Println("No distortion")
			continue
		}
		fmt.Println(diff)
	}
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
func distort(contents string) string {
	typeAnnotation := regexp.MustCompile(`:\s*([a-zA-Z\[\]<>]+)`)
	matches := typeAnnotation.FindStringSubmatch(contents)
	if len(matches) > 0 {
		if matches[1] != ": string" {
			return typeAnnotation.ReplaceAllString(contents, ": string")
		}
	}
	return contents
}
