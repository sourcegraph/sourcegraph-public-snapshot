package main

import (
	"os"
)

func ReadCache(cacheDir string) (*RepoIndex, error) {
	file, err := os.Open(cacheDir)
	if err != nil {
		return nil, err
	}
	return DeserializeRepoIndex(file)
}

func WriteCache(dir, cacheDir string) error {
	r, err := NewRepoIndex(dir)
	if err != nil {
		return err
	}
	return r.SerializeToFile(cacheDir)
}
