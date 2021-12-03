package main

import (
	"encoding/gob"
	"fmt"
	"github.com/bits-and-blooms/bloom/v3"
	"os"
)

const (
	estimate         = 0.01
	maxFileSize      = 1_000_000
	bloomSizePadding = 2.0
)

var query = "COM"

type BlobIndex struct {
	Filter *bloom.BloomFilter
	Path   string
}

func ReadCache(cacheDir string) RepoIndex {
	file, err := os.Open(cacheDir)
	if err != nil {
		fmt.Printf("err %v\n", err)
		panic(err)
	}
	decoder := gob.NewDecoder(file)
	var r RepoIndex
	err = decoder.Decode(&r)
	if err != nil {
		panic(err)
	}
	return r
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
