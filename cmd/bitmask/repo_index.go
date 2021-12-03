package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/bits-and-blooms/bloom/v3"
	"github.com/go-enry/go-enry/v2"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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
	var result [][]byte
	queryBytes := []byte(query)
	for i := 0; i < len(queryBytes)-3; i++ {
		result = append(result, queryBytes[i:i+3])
	}
	return result
}

func (r *RepoIndex) SerializeToFile(cacheDir string) (err error) {
	_ = os.Remove(cacheDir)
	err = os.MkdirAll(filepath.Dir(cacheDir), 0755)
	if err != nil {
		return err
	}
	cacheOut, err := os.Create(cacheDir)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := cacheOut.Close()
		if err != nil {
			err = closeErr
		}
	}()
	err = r.Serialize(cacheOut)
	return
}

func (r *RepoIndex) Serialize(w io.Writer) error {
	return gob.NewEncoder(w).Encode(r)
}

func DeserializeRepoIndex(reader io.Reader) (*RepoIndex, error) {
	var r *RepoIndex
	err := gob.NewDecoder(reader).Decode(r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func NewRepoIndex(dir string) (*RepoIndex, error) {
	cmd := exec.Command("git", "ls-files", "-z", "--with-tree=main")
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	if err != nil {
		return nil, err
	}
	stdout := string(out.Bytes())
	NUL := string([]byte{0})
	lines := strings.Split(stdout, NUL)
	indexes := make([]BlobIndex, len(lines))
	for i, line := range lines {
		if i%100 == 0 {
			fmt.Println(i)
		}
		abspath := path.Join(dir, line)
		textBytes, err := os.ReadFile(abspath)
		if err != nil {
			continue
		}
		if len(textBytes) > maxFileSize {
			continue
		}
		filter := bloom.NewWithEstimates(uint(len(textBytes)*bloomSizePadding), estimate)
		if enry.IsBinary(textBytes) {
			continue
		}
		for i = 0; i < len(textBytes)-3; i++ {
			trigram := textBytes[i : i+3]
			filter.Add(trigram)
		}
		indexes = append(
			indexes,
			BlobIndex{
				Path:   line,
				Filter: filter,
			},
		)
	}
	return &RepoIndex{indexes}, nil
}

func (r *RepoIndex) Grep(query string) {
	matchingPaths := r.PathsMatchingQuery(query)
	falsePositive := 0
	truePositive := 0
	for matchingPath := range matchingPaths {
		hasMatch := false
		textBytes, err := os.ReadFile(matchingPath)
		if err != nil {
			return
		}
		text := string(textBytes)
		start := 0
		end := strings.Index(text[start:], "\n")
		for lineNumber, line := range strings.Split(text, "\n") {
			columnNumber := strings.Index(line, query)
			if columnNumber >= 0 {
				hasMatch = true
				prefix := line[0:columnNumber]
				suffix := line[columnNumber+len(query):]
				fmt.Printf(
					"%v:%v:%v %v%v%v\n",
					matchingPath,
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

		if hasMatch {
			truePositive++
		} else {
			fmt.Println(matchingPath)
			falsePositive++
		}
	}
	fmt.Printf("fpr %v", float64(falsePositive)/float64(truePositive+falsePositive))
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
