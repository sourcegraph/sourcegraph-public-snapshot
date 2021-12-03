package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/bits-and-blooms/bloom/v3"
	"github.com/go-enry/go-enry/v2"
	"os"
	"os/exec"
	"path"
	"strings"
)

var cache = os.Getenv("HOME") + "/dev/sourcegraph/bitmask-cache"
var estimate = 0.01
var query = "COM"

type BlobIndex struct {
	Filter *bloom.BloomFilter
	Path   string
}

func ReadCache() RepoIndex {
	file, err := os.Open(cache)
	if err != nil {
		fmt.Printf("err %v\n", err)
		panic(err)
	}
	decoder := gob.NewDecoder(file)
	var indexes []BlobIndex
	err = decoder.Decode(&indexes)
	if err != nil {
		panic(err)
	}
	return RepoIndex{indexes}
}

func WriteCache(dir string, maxLines int) {
	//path := os.Getenv("HOME") + "/dev/sgtest/megarepo"
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
	indexes := make([]BlobIndex, len(lines))
	trigrams := make(map[string]struct{})
	for i, line := range lines {
		if i%100 == 0 {
			fmt.Println(i)
		}
		if i > maxLines {
			break
		}
		abspath := path.Join(dir, line)
		stat, err := os.Stat(abspath)
		if err != nil {
			panic(err)
		}
		if stat.IsDir() {
			continue
		}
		if stat.Size() > 1_000_000 {
			continue
		}
		textBytes, err := os.ReadFile(abspath)
		if err != nil {
			panic(err)
		}
		if enry.IsBinary(textBytes) {
			fmt.Printf("isBinary %v\n", abspath)
			continue
		}
		text := string(textBytes)
		for i = 0; i < len(text)-3; i++ {
			trigram := text[i : i+3]
			trigrams[trigram] = struct{}{}
		}
		size := len(trigrams)
		//fmt.Printf("size %v %v\n", size, abspath)
		filter := bloom.NewWithEstimates(uint(size), estimate)
		for trigram := range trigrams {
			filter.AddString(trigram)
		}
		indexes = append(
			indexes,
			BlobIndex{
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
