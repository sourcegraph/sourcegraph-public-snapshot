package main

import (
	"encoding/gob"
	"os"
)

func ReadCache(cacheDir string) (*RepoIndex, error) {
	file, err := os.Open(cacheDir)
	if err != nil {
		return nil, err
	}
	decoder := gob.NewDecoder(file)
	r := &RepoIndex{}
	err = decoder.Decode(r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func WriteCache(dir, cacheDir string) error {
	r, err := NewRepoIndex(dir)
	if err != nil {
		return err
	}
	err = r.SerializeToFile(cacheDir)
	if err != nil {
		return err
	}
	return nil
}
