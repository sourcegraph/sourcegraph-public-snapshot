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

func trigrams(query string) [][]byte {
	result := [][]byte{}
	for i := 0; i < len(query)-3; i++ {
		result = append(result, []byte(query[i:i+3]))
	}
	return result
}

func (r *RepoIndex) Grep(query string) {
	paths := r.PathsMatchingQuery(query)
	for path := range paths {
		hasMatch := false
		textBytes, err := os.ReadFile(path)
		if err != nil {
			return
		}
		text := string(textBytes)
		start := 0
		end := strings.Index(text[start:], "\n")
		lineNumber := -1
		for end > start && end >= 0 && end < len(text)-1 {
			lineNumber++
			line := text[start:end]
			columnNumber := strings.Index(line, query)
			if columnNumber >= 0 {
				hasMatch = true
				prefix := line[1:columnNumber]
				suffix := line[columnNumber+len(query):]
				fmt.Printf(
					"%v:%v:%v %v%v%v\n",
					path,
					lineNumber,
					columnNumber,
					prefix,
					Yellow(query),
					suffix,
				)
			}
			start = end + 1
			end = strings.Index(text[end+1:], "\n")
		}

		if !hasMatch {
			//fmt.Println("false positive " + path)
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

func (r *RepoIndex) PathsMatchingQuery(query string) chan string {
	grams := trigrams(query)
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
				if index.Filter == nil {
					continue
				}
				isMatch := true
				for _, gram := range grams {
					if !index.Filter.Test(gram) {
						isMatch = false
						break
					}
				}
				if isMatch {
					res <- index.Path
				}
			}
		}()
	}
	wg.Wait()
	close(res)
	return res
}
