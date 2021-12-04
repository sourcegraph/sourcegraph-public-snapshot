package main

import (
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/go-enry/go-enry/v2"
)

var (
	Yellow = color("\033[1;33m%s\033[0m")
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
	SeenHashes map[uint64]struct{}
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
		SeenHashes: map[uint64]struct{}{},
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
func (g *Ngrams) Update(b int32) {
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

func (g *Ngrams) OnIndex(index int, b int32, onBytes func(b []byte)) {
	g.Update(b)

	g.Unigram.EmitHashAndClear(g, onBytes)

	switch index % 2 {
	case 0:
		g.Bigram1.EmitHashAndClear(g, onBytes)
	case 1:
		g.Bigram2.EmitHashAndClear(g, onBytes)
	}

	switch index % 3 {
	case 0:
		g.Trigram1.EmitHashAndClear(g, onBytes)
	case 1:
		g.Trigram2.EmitHashAndClear(g, onBytes)
	case 2:
		g.Trigram3.EmitHashAndClear(g, onBytes)
	}
	switch index % 4 {
	case 0:
		g.Quadgram1.EmitHashAndClear(g, onBytes)
	case 1:
		g.Quadgram2.EmitHashAndClear(g, onBytes)
	case 2:
		g.Quadgram3.EmitHashAndClear(g, onBytes)
	case 3:
		g.Quadgram4.EmitHashAndClear(g, onBytes)
	}
	switch index % 5 {
	case 0:
		g.Pentagram1.EmitHashAndClear(g, onBytes)
	case 1:
		g.Pentagram2.EmitHashAndClear(g, onBytes)
	case 2:
		g.Pentagram3.EmitHashAndClear(g, onBytes)
	case 3:
		g.Pentagram4.EmitHashAndClear(g, onBytes)
	case 4:
		g.Pentagram5.EmitHashAndClear(g, onBytes)
	}
}

type Ngram struct {
	Hash uint64
}

func (g *Ngram) EmitHashAndClear(gs *Ngrams, onBytes func(b []byte)) {
	if _, ok := gs.SeenHashes[g.Hash]; !ok {
		gs.SeenHashes[g.Hash] = struct{}{}
		hashedBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(hashedBytes, g.Hash)
		onBytes(hashedBytes)
	}
	g.Hash = 0
}

func (g *Ngram) Update(b int32) {
	g.Hash = 31*g.Hash + uint64(b)
}

func onGrams(text string, onBytes func(b []byte)) {
	ngrams := NewNgrams()
	for i, b := range text {
		ngrams.OnIndex(i, b, onBytes)
	}
}

func collectGrams(query string) [][]byte {
	var result [][]byte
	onGrams(query, func(b []byte) {
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

func NewRepoIndex(fs FileSystem) (*RepoIndex, error) {
	filenames, err := fs.ListRelativeFilenames()
	if err != nil {
		return nil, err
	}
	indexes := make([]BlobIndex, len(filenames))
	for i, filename := range filenames {
		if i%100 == 0 {
			fmt.Println(i)
		}
		textBytes, err := fs.ReadRelativeFilename(filename)
		if err != nil {
			continue
		}
		if len(textBytes) > maxFileSize {
			continue
		}
		if enry.IsBinary(textBytes) {
			continue
		}
		text := string(textBytes)
		bloomSize := uint(len(text) * bloomSizePadding)
		if bloomSize < 10_000 {
			bloomSize = 10_000
		}
		filter := bloom.NewWithEstimates(bloomSize, targetFalsePositiveRatio)
		onGrams(text, func(b []byte) {
			filter.Add(b)
		})
		//sizeRatio := float64(filter.ApproximatedSize()) / float64(bloomSize)
		//fmt.Printf("%v %v %v\n", sizeRatio, filter.ApproximatedSize(), bloomSize)
		indexes = append(
			indexes,
			BlobIndex{
				Path:   filename,
				Filter: filter,
			},
		)
	}
	return &RepoIndex{Dir: fs.RootDir(), Blobs: indexes}, nil
}

func (r *RepoIndex) Grep(query string) {
	start := time.Now()
	matchingPaths := r.PathsMatchingQuery(query)
	falsePositive := 0
	truePositive := 0
	totalMatchCount := uint64(0)
	for matchingPath := range matchingPaths {
		textBytes, err := os.ReadFile(filepath.Join(r.Dir, matchingPath))
		if err != nil {
			continue
		}
		text := string(textBytes)
		start := 0
		end := strings.Index(text[start:], "\n")
		matchCount := 0
		for lineNumber, line := range strings.Split(text, "\n") {
			columnNumber := strings.Index(line, query)
			if columnNumber >= 0 {
				matchCount++
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

		totalMatchCount = totalMatchCount + uint64(matchCount)
		if matchCount > 0 {
			truePositive++
		} else {
			//fmt.Println(matchingPath)
			falsePositive++
		}
	}
	end := time.Now()
	elapsed := (end.UnixNano() - start.UnixNano()) / int64(time.Millisecond)
	falsePositiveRatio := float64(falsePositive) / math.Max(1.0, float64(truePositive+falsePositive))
	fmt.Printf(
		"query '%v' matches %v files %v time %vms fpr %v\n",
		query,
		totalMatchCount,
		truePositive,
		elapsed,
		falsePositiveRatio,
	)
}

func color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}

func (r *RepoIndex) pathsMatchingQuerySync(
	grams [][]byte,
	batch []BlobIndex,
	onMatch func(matchingPath string),
) {
	for _, index := range batch {
		if index.Filter == nil {
			continue
		}
		isMatch := len(grams) > 0
		for _, gram := range grams {
			//fmt.Printf("test %v %v\n", index.Filter.Test(gram))
			if !index.Filter.Test(gram) {
				isMatch = false
				break
			}
		}
		if isMatch {
			onMatch(index.Path)
		}
	}
}

func (r *RepoIndex) PathsMatchingQuerySync(query string) []string {
	grams := collectGrams(query)
	var result []string
	r.pathsMatchingQuerySync(grams, r.Blobs, func(matchingPath string) {
		result = append(result, matchingPath)
	})
	return result
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
			r.pathsMatchingQuerySync(grams, batch, func(matchingPath string) {
				res <- matchingPath
			})
		}()
	}
	wg.Wait()
	close(res)
	return res
}
