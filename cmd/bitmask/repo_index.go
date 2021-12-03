package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

var (
	Yellow = color("\033[1;33m%s\033[0m")
)

type RepoIndex struct {
	Blobs []BlobIndex
}

func (r *RepoIndex) Grep(query string) {
	paths := r.PathsMatchingQuery([]byte(query))
	for path := range paths {
		textBytes, err := os.ReadFile(path)
		if err != nil {
			return
		}
		text := string(textBytes)
		start := 0
		end := strings.Index(text[start:], "\n")
		for end >= 0 && end < len(text)-1 {
			line := text[start:end]
			m := strings.Index(line, query)
			if m >= 0 {
				prefix := line[0:m]
				suffix := line[m+len(query):]
				fmt.Printf(prefix + Yellow(query) + suffix + "\n")
			}
			start = end
			end = strings.Index(text[end+1:], "\n")
		}

	}
}

func color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}

func (r *RepoIndex) PathsMatchingQuery(query []byte) chan string {
	res := make(chan string, len(r.Blobs))
	batchSize := 5_000
	var wg sync.WaitGroup
	for i := 0; i < len(r.Blobs); i += batchSize {
		j := i + batchSize
		if j > len(r.Blobs) {
			j = len(r.Blobs)
		}
		batch := r.Blobs[i:j]
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, index := range batch {
				if index.Filter != nil && index.Filter.Test(query) {
					res <- index.Path
				}
			}
		}()
	}
	wg.Wait()
	close(res)
	return res
}
