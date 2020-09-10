package main

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type queries map[string][]string

func loadQueries(path string) (queries, error) {
	queries := map[string][]string{}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		f, err := os.Open(filepath.Join(path, file.Name()))
		if err != nil {
			return nil, err
		}

		scanner := bufio.NewScanner(f)
		k := strings.TrimSuffix(f.Name(), ".txt")
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			queries[k] = append(queries[k], line)
		}
		_ = f.Close()
	}
	return queries, nil
}
