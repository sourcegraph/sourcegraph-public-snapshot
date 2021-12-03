package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/cockroachdb/errors"
	"github.com/go-enry/go-enry/v2"
)

var (
	Yellow      = color("\033[1;33m%s\033[0m")
	hashedBytes = make([]byte, 8)
)

const (
	targetFalsePositiveRatio = 0.01
	maxFileSize              = 1 << 20 // 1_048_576
	bloomSizePadding         = 5
)

type RepoIndex struct {
	Dir   string
	Blobs []BlobIndex
}
type BlobIndex struct {
	Filter *bloom.BloomFilter
	Path   string
}

type Ngrams struct {
	Unigram    Ngram
	Bigram1    Ngram
	Bigram2    Ngram
	Trigram1   Ngram
	Trigram2   Ngram
	Trigram3   Ngram
	Quadgram1  Ngram
	Quadgram2  Ngram
	Quadgram3  Ngram
	Quadgram4  Ngram
	Pentagram1 Ngram
	Pentagram2 Ngram
	Pentagram3 Ngram
	Pentagram4 Ngram
	Pentagram5 Ngram
}

func NewNgrams() Ngrams {
	return Ngrams{
		Unigram:    Ngram{Hash: 0},
		Bigram1:    Ngram{Hash: 0},
		Bigram2:    Ngram{Hash: 0},
		Trigram1:   Ngram{Hash: 0},
		Trigram2:   Ngram{Hash: 0},
		Trigram3:   Ngram{Hash: 0},
		Quadgram1:  Ngram{Hash: 0},
		Quadgram2:  Ngram{Hash: 0},
		Quadgram3:  Ngram{Hash: 0},
		Quadgram4:  Ngram{Hash: 0},
		Pentagram1: Ngram{Hash: 0},
		Pentagram2: Ngram{Hash: 0},
		Pentagram3: Ngram{Hash: 0},
		Pentagram4: Ngram{Hash: 0},
		Pentagram5: Ngram{Hash: 0},
	}

}
func (g *Ngrams) Update(index int, b byte) {
	g.Unigram.Update(b)

	g.Bigram1.Update(b)
	g.Bigram2.Update(b)

	g.Trigram1.Update(b)
	g.Trigram2.Update(b)
	g.Trigram3.Update(b)

	g.Quadgram1.Update(b)
	g.Quadgram2.Update(b)
	g.Quadgram3.Update(b)
	g.Quadgram4.Update(b)

	g.Pentagram1.Update(b)
	g.Pentagram2.Update(b)
	g.Pentagram3.Update(b)
	g.Pentagram4.Update(b)
	g.Pentagram5.Update(b)
}

func (g *Ngrams) OnIndex(index int, b byte, onBytes func(b []byte)) {
	g.Update(index, b)
	g.Unigram.Clear(onBytes)

	switch index % 2 {
	case 0:
		g.Bigram1.Clear(onBytes)
	case 1:
		g.Bigram2.Clear(onBytes)
	}

	switch index % 3 {
	case 0:
		g.Trigram1.Clear(onBytes)
	case 1:
		g.Trigram2.Clear(onBytes)
	case 2:
		g.Trigram3.Clear(onBytes)
	}
	switch index % 4 {
	case 0:
		g.Quadgram1.Clear(onBytes)
	case 1:
		g.Quadgram2.Clear(onBytes)
	case 2:
		g.Quadgram3.Clear(onBytes)
	case 3:
		g.Quadgram4.Clear(onBytes)
	}
	switch index % 5 {
	case 0:
		g.Pentagram1.Clear(onBytes)
	case 1:
		g.Pentagram2.Clear(onBytes)
	case 2:
		g.Pentagram3.Clear(onBytes)
	case 3:
		g.Pentagram4.Clear(onBytes)
	case 4:
		g.Pentagram5.Clear(onBytes)
	}
}

type Ngram struct {
	Hash uint64
}

func (g *Ngram) Clear(onBytes func(b []byte)) {
	binary.LittleEndian.PutUint64(hashedBytes, g.Hash)
	onBytes(hashedBytes)
	g.Hash = 0
}

func (g *Ngram) Update(b byte) {
	g.Hash = 31*g.Hash + uint64(b)
}

func onGrams(textBytes []byte, onBytes func(b []byte)) {
	ngrams := NewNgrams()
	for i, b := range textBytes {
		ngrams.OnIndex(i, b, onBytes)
	}
}

func collectGrams(query string) [][]byte {
	var result [][]byte
	onGrams([]byte(query), func(b []byte) {
		result = append(result, b)
	})
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
	r := &RepoIndex{}
	err := gob.NewDecoder(reader).Decode(r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func NewRepoIndex(dir string) (*RepoIndex, error) {
	var branch bytes.Buffer
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchCmd.Dir = dir
	branchCmd.Stdout = &branch
	err := branchCmd.Run()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to infer the default branch")
	}
	cmd := exec.Command(
		"git",
		"ls-files",
		"-z",
		"--with-tree",
		strings.Trim(branch.String(), "\n"),
	)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()

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
		bloomSize := uint(len(textBytes) * bloomSizePadding)
		//bloomBitCount := uint(math.Ceil(-1 * float64(bloomSize) * math.Log(targetFalsePositiveRatio) / math.Pow(math.Log(2), 2)))
		filter := bloom.NewWithEstimates(bloomSize, targetFalsePositiveRatio)
		if enry.IsBinary(textBytes) {
			continue
		}
		onGrams(textBytes, func(b []byte) {
			filter.Add(b)
		})
		sizeRatio := float64(filter.ApproximatedSize()) / float64(bloomSize)
		if sizeRatio > 0.5 {
			//fmt.Printf("%v %v %v\n", sizeRatio, filter.ApproximatedSize(), bloomSize)
		}
		indexes = append(
			indexes,
			BlobIndex{
				Path:   line,
				Filter: filter,
			},
		)
	}
	return &RepoIndex{Dir: dir, Blobs: indexes}, nil
}

func (r *RepoIndex) Grep(query string) {
	start := time.Now()
	matchingPaths := r.PathsMatchingQuery(query)
	falsePositive := 0
	truePositive := 0
	for matchingPath := range matchingPaths {
		hasMatch := false
		textBytes, err := os.ReadFile(filepath.Join(r.Dir, matchingPath))
		if err != nil {
			continue
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
			//fmt.Println(matchingPath)
			falsePositive++
		}
	}
	end := time.Now()
	elapsed := (end.UnixNano() - start.UnixNano()) / int64(time.Millisecond)
	falsePositiveRatio := float64(falsePositive) / float64(truePositive+falsePositive)
	fmt.Printf("query '%v' time %vms fpr %v\n", query, elapsed, falsePositiveRatio)
}

func color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}

func (r *RepoIndex) PathsMatchingQuery(query string) chan string {
	grams := collectGrams(query)
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
