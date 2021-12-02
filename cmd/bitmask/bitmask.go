package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/bits-and-blooms/bloom/v3"
	"os"
	"os/exec"
	"path"
	"strings"
)

var cache = os.Getenv("HOME") + "/dev/sourcegraph/bitmask-cache"
var estimate = 0.01
var query = "COM"

type TrigramIndex struct {
	Filter *bloom.BloomFilter
	Path   string
}

func QueryCache() {
	indexes := ReadCache()
	fmt.Printf("INDEXES %v\n", len(indexes))
	for _, index := range indexes {
		if index.Filter != nil {
			if index.Filter.TestString(query) {
				fmt.Printf("path %v\n", index.Path)
			}
		}
	}
}

func ReadCache() []TrigramIndex {
	file, err := os.Open(cache)
	if err != nil {
		fmt.Printf("err %v\n", err)
		panic(err)
	}
	decoder := gob.NewDecoder(file)
	var indexes []TrigramIndex
	err = decoder.Decode(&indexes)
	if err != nil {
		panic(err)
	}
	return indexes
}

func WriteCache() {
	//path := os.Getenv("HOME") + "/dev/sgtest/megarepo"
	dir := os.Getenv("HOME") + "/dev/sgtest/megarepo"
	//fmt.Println(path)

	cmd := exec.Command("git", "ls-files", "-z", "--with-tree=main")
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	if err != nil {
		panic(err)
	}
	stdout := string(out.Bytes())
	NUL := string([]byte{0})
	lines := strings.Split(stdout, NUL)
	indexes := make([]TrigramIndex, len(lines))
	trigrams := make(map[string]struct{})
	for i, line := range lines {
		if i%100 == 0 {
			fmt.Println(i)
		}
		abspath := path.Join(dir, line)
		stat, err := os.Stat(abspath)
		if err != nil {
			panic(err)
		}
		if stat.Size() > 1_000_000 {
			continue
		}
		textBytes, _ := os.ReadFile(abspath)
		text := string(textBytes)
		for i = 0; i < len(text)-3; i++ {
			trigram := text[i : i+3]
			trigrams[trigram] = struct{}{}
		}
		size := len(trigrams)
		filter := bloom.NewWithEstimates(uint(size), estimate)
		for trigram, _ := range trigrams {
			filter.AddString(trigram)
		}
		indexes = append(
			indexes,
			TrigramIndex{
				Path:   line,
				Filter: filter,
			},
		)
	}
	_ = os.Remove(cache)
	cacheOut, err := os.Create(cache)
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	encoder := gob.NewEncoder(cacheOut)
	err = encoder.Encode(indexes)
	if err != nil {
		panic(err)
	}
	fmt.Printf("out: %v\n", cache)
	err = cacheOut.Close()
	//fmt.Println("Hello world!")
	//filter := bloom.NewWithEstimates(1000, 0.1)
	//filter.Add([]byte("Love"))
	////filter.Test([]byte("Love"))
	//fmt.Printf("contains(a) %v", filter.Test([]byte("Love")))
}
