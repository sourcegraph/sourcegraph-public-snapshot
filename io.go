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
		// data/query_group.txt -> query_group
		g := strings.TrimSuffix(filepath.Base(f.Name()), filepath.Ext(f.Name()))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			queries[g] = append(queries[g], line)
		}
		_ = f.Close()
	}
	return queries, nil
}
