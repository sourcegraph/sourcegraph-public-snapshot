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

func main() {
	createCache()
	queryCache()

}

type TrigramIndexes struct {
	Indexes []TrigramIndex
}
type TrigramIndex struct {
	Path   string
	Filter []byte
	Size   uint
}
type DecodedTrigramIndex struct {
	Filter *bloom.BloomFilter
	Path   string
}

func queryCache() {
	indexes := readCache()
	query := []byte("Pag")
	for _, index := range indexes {
		if index.Filter.Test(query) {
			fmt.Println(index.Path)
		}
	}
}

func readCache() []DecodedTrigramIndex {
	file, err := os.Open(cache)
	if err != nil {
		panic(err)
	}
	decoder := gob.NewDecoder(file)
	indexes := TrigramIndexes{}
	err = decoder.Decode(&indexes)
	if err != nil {
		panic(err)
	}
	result := make([]DecodedTrigramIndex, len(indexes.Indexes))
	for _, index := range indexes.Indexes {
		filter := bloom.NewWithEstimates(index.Size, estimate)
		result = append(result, DecodedTrigramIndex{Filter: filter, Path: index.Path})
	}
	return result
}

func createCache() {
	//path := os.Getenv("HOME") + "/dev/sgtest/megarepo"
	dir := os.Getenv("HOME") + "/dev/sourcegraph/sourcegraph"
	//fmt.Println(path)

	cmd := exec.Command("git", "ls-files", "-z")
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
	for i, line := range lines {
		if i > 10 {
			break
		}
		abspath := path.Join(dir, line)
		textBytes, _ := os.ReadFile(abspath)
		text := string(textBytes)
		trigrams := make(map[string]struct{})
		for i = 0; i < len(text)-3; i++ {
			trigram := text[i : i+2]
			trigrams[trigram] = struct{}{}
		}
		size := len(trigrams)
		filter := bloom.NewWithEstimates(uint(size), estimate)
		for trigram, _ := range trigrams {
			filter.AddString(trigram)
		}
		cacheOut := path.Join(cache, line)
		var buf bytes.Buffer
		_, err = filter.WriteTo(&buf)
		indexes = append(
			indexes,
			TrigramIndex{
				Path:   line,
				Filter: buf.Bytes(),
				Size:   uint(size),
			},
		)
		fmt.Println(cacheOut)
	}
	_ = os.Remove(cache)
	cacheOut, err := os.Create(cache)
	if err != nil {
		panic(err)
	}
	err = cacheOut.Close()
	if err != nil {
		panic(err)
	}
	encoder := gob.NewEncoder(cacheOut)
	_ = encoder.Encode(TrigramIndexes{Indexes: indexes})
	fmt.Printf("out: %v", cache)
	//fmt.Println("Hello world!")
	//filter := bloom.NewWithEstimates(1000, 0.1)
	//filter.Add([]byte("Love"))
	////filter.Test([]byte("Love"))
	//fmt.Printf("contains(a) %v", filter.Test([]byte("Love")))
}
