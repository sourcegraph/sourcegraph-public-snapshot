package bitmask

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"github.com/cespare/xxhash/v2"
	"github.com/cockroachdb/errors"
	"github.com/schollz/progressbar/v3"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/go-enry/go-enry/v2"
)

var (
	Yellow = color("\033[1;33m%s\033[0m")
)

const (
	Version                  = 1
	targetFalsePositiveRatio = 0.01
	maxFileSize              = 1 << 20 // 1_048_576
	maximumQueryNgrams       = 100
	MinArity                 = 1
	MaxArity                 = 3
)

var IsProgressBarEnabled = true

type RepoIndex struct {
	Dir   string
	Blobs []BlobIndex
	FS    FileSystem
}
type BlobIndex struct {
	Filter *bloom.BloomFilter
	Path   string
}

func (b *BlobIndex) WriteTo(w io.Writer) (int64, error) {
	var buf bytes.Buffer
	var writtenByteCount int64
	gob.NewEncoder(&buf).Encode(b)
	data := buf.Bytes()
	err := binary.Write(w, binary.BigEndian, uint64(len(data)))
	writtenByteCount = 8
	if err != nil {
		return writtenByteCount, err
	}
	w.Write(data)
	writtenByteCount = writtenByteCount + int64(len(data))
	return writtenByteCount, nil
}

func (b *BlobIndex) ReadFrom(stream io.Reader) (int64, error) {
	var length uint64
	var readByteCount int64
	err := binary.Read(stream, binary.BigEndian, &length)
	if err != nil {
		return readByteCount, err
	}
	readByteCount = 8
	data := make([]byte, length)
	read, err := stream.Read(data)
	if err != nil {
		return readByteCount, err
	}
	if uint64(read) != length {
		return readByteCount, errors.Errorf("read(%v) != length(%v)", read, length)
	}
	readByteCount = readByteCount + int64(len(data))
	other := &BlobIndex{}
	err = gob.NewDecoder(bytes.NewReader(data)).Decode(other)
	if err != nil {
		return readByteCount, err
	}
	b.Path = other.Path
	b.Filter = other.Filter
	return readByteCount, nil
}

type Ngrams struct {
	Int64Bytes []byte
	SeenHashes map[uint64]struct{}
	Unigram    Ngram
	Bigram1    Ngram
	Bigram2    Ngram
	Trigram1   Ngram
	Trigram2   Ngram
	Trigram3   Ngram
	//Quadgram1  Ngram
	//Quadgram2  Ngram
	//Quadgram3  Ngram
	//Quadgram4  Ngram
	//Pentagram1 Ngram
	//Pentagram2 Ngram
	//Pentagram3 Ngram
	//Pentagram4 Ngram
	//Pentagram5 Ngram
}

func NewNgrams() Ngrams {
	return Ngrams{
		Int64Bytes: make([]byte, unsafe.Sizeof(uint64(0))),
		SeenHashes: map[uint64]struct{}{},
		Unigram:    Ngram{Arity: 1, Hash: xxhash.New()},
		Bigram1:    Ngram{Arity: 2, Hash: xxhash.New()},
		Bigram2:    Ngram{Arity: 2, Hash: xxhash.New()},
		Trigram1:   Ngram{Arity: 3, Hash: xxhash.New()},
		Trigram2:   Ngram{Arity: 3, Hash: xxhash.New()},
		Trigram3:   Ngram{Arity: 3, Hash: xxhash.New()},
		//Quadgram1:  Ngram{Arity: 4, Hash: xxhash.New()},
		//Quadgram2:  Ngram{Arity: 4, Hash: xxhash.New()},
		//Quadgram3:  Ngram{Arity: 4, Hash: xxhash.New()},
		//Quadgram4:  Ngram{Arity: 4, Hash: xxhash.New()},
		//Pentagram1: Ngram{Arity: 5, Hash: xxhash.New()},
		//Pentagram2: Ngram{Arity: 5, Hash: xxhash.New()},
		//Pentagram3: Ngram{Arity: 5, Hash: xxhash.New()},
		//Pentagram4: Ngram{Arity: 5, Hash: xxhash.New()},
		//Pentagram5: Ngram{Arity: 5, Hash: xxhash.New()},
	}

}
func (g *Ngrams) Update(b int32) {
	g.Unigram.Update(g, b)

	g.Bigram1.Update(g, b)
	g.Bigram2.Update(g, b)

	g.Trigram1.Update(g, b)
	g.Trigram2.Update(g, b)
	g.Trigram3.Update(g, b)

	//g.Quadgram1.Update(b)
	//g.Quadgram2.Update(b)
	//g.Quadgram3.Update(b)
	//g.Quadgram4.Update(b)
	//
	//g.Pentagram1.Update(b)
	//g.Pentagram2.Update(b)
	//g.Pentagram3.Update(b)
	//g.Pentagram4.Update(b)
	//g.Pentagram5.Update(b)
}

func (g *Ngrams) OnIndex(index int, b int32, onBytes OnBytes) {
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
	//switch index % 4 {
	//case 0:
	//	g.Quadgram1.EmitHashAndClear(g, onBytes)
	//case 1:
	//	g.Quadgram2.EmitHashAndClear(g, onBytes)
	//case 2:
	//	g.Quadgram3.EmitHashAndClear(g, onBytes)
	//case 3:
	//	g.Quadgram4.EmitHashAndClear(g, onBytes)
	//}
	//switch index % 5 {
	//case 0:
	//	g.Pentagram1.EmitHashAndClear(g, onBytes)
	//case 1:
	//	g.Pentagram2.EmitHashAndClear(g, onBytes)
	//case 2:
	//	g.Pentagram3.EmitHashAndClear(g, onBytes)
	//case 3:
	//	g.Pentagram4.EmitHashAndClear(g, onBytes)
	//case 4:
	//	g.Pentagram5.EmitHashAndClear(g, onBytes)
	//}
}

type Ngram struct {
	Hash  *xxhash.Digest
	Arity int
}

func (g *Ngram) EmitHashAndClear(gs *Ngrams, onBytes OnBytes) {
	//if g.Arity < MinArity || g.Arity > MaxArity {
	//	return
	//}
	hash := g.Hash.Sum64()
	if _, ok := gs.SeenHashes[hash]; !ok {
		gs.SeenHashes[hash] = struct{}{}
		binary.LittleEndian.PutUint64(gs.Int64Bytes, hash)
		onBytes(gs.Int64Bytes, g.Arity)
	}
	g.Hash.Reset()
}

func (g *Ngram) Update(gs *Ngrams, b int32) {
	//if g.Arity < MinArity || g.Arity > MaxArity {
	//	return
	//}
	binary.BigEndian.PutUint64(gs.Int64Bytes, uint64(b))

	// always returns len(hashedBytes), nil
	_, _ = g.Hash.Write(gs.Int64Bytes)
}

type OnBytes func(b []byte, arity int)

func onGrams(text string, onBytes OnBytes) Ngrams {
	ngrams := NewNgrams()
	for i, b := range text {
		ngrams.OnIndex(i, b, onBytes)
	}
	return ngrams
}

func CollectQueryNgrams(query string) [][]byte {
	var result [][]byte
	var arities []int
	onGrams(query, func(b []byte, arity int) {
		arities = append(arities, arity)
		result = append(result, b)
	})
	randomNumbers := make([]int, len(arities))
	for i := range randomNumbers {
		randomNumbers[i] = rand.Int()
	}
	sort.SliceStable(result, func(i, j int) bool {
		if arities[i] == arities[j] {
			// Shuffle the ordering of n-grams with the same arity to increase entropy
			// among the n-grams that appear first in the results.
			// For example, the ID number in the query "bugzilla.redhat.com/show_bug.cgi?id=726143"
			// appears at the end of the query and we want to move the n-grams from that
			// ID to appear early in the results to allow the first bloom filter tests to exit early.
			// we want to avoid the case where we test only the start of the query
			return randomNumbers[i] < randomNumbers[j]
		}
		return arities[i] > arities[j]
	})
	if len(result) > maximumQueryNgrams {
		result = result[:maximumQueryNgrams]
	}
	return result
}

func (r *RepoIndex) SerializeToFile(cacheDir string) (err error) {
	//_ = os.Remove(cacheDir)
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

func NewOnDiskRepoIndex(fs FileSystem, outputPath string) error {
	file, err := os.CreateTemp("", "repo-index")
	if err != nil {
		return errors.Wrapf(err, "NewOnDiskRepoIndex - failed to create temporary directory")
	}
	tmpName := file.Name()
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	filenames, err := fs.ListRelativeFilenames()
	if err != nil {
		return errors.Wrapf(err, "NewOnDiskRepoIndex - failed fs.ListRelativeFilenames")
	}
	for index := range repoIndexes(fs, filenames) {
		_, err = index.WriteTo(file)
		if err != nil {
			break
		}
	}
	if err != nil {
		return errors.Wrapf(err, "NewOnDiskRepoIndex - failed to write repo indexes")
	}
	err = file.Close()
	file = nil
	if err != nil {
		return errors.Wrapf(err, "NewOnDiskRepoIndex - failed to close tmp file")
	}
	stat, err := os.Stat(outputPath)
	if err == nil {
		if stat.IsDir() {
			return errors.Errorf("can't write to directory %v", outputPath)
		}
		err = os.Remove(outputPath)
		if err != nil {
			return errors.Wrapf(err, "NewOnDiskRepoIndex - failed to remove output path")
		}
	} else {
		err = os.MkdirAll(filepath.Dir(outputPath), 0755)
		if err != nil {
			return errors.Wrapf(err, "NewOnDiskRepoIndex - failed to MkdirAll")
		}
	}
	destination, err := os.Create(outputPath)
	if err != nil {
		return errors.Wrapf(err, "NewOnDiskRepoIndex - failed to create tmp file (after closing it)")
	}
	defer destination.Close()

	source, err := os.Open(tmpName)
	if err != nil {
		return errors.Wrapf(err, "NewOnDiskRepoIndex - failed to re-open tmp file")
	}
	_, err = io.Copy(destination, source)
	if err != nil {
		return errors.Wrapf(err, "NewOnDiskRepoIndex - failed to copy from tmp file to destination path")
	}
	return err
}

func NewInMemoryRepoIndex(fs FileSystem) (*RepoIndex, error) {
	filenames, err := fs.ListRelativeFilenames()
	if err != nil {
		return nil, err
	}
	var indexes []BlobIndex
	for index := range repoIndexes(fs, filenames) {
		indexes = append(indexes, index)
	}
	return &RepoIndex{Blobs: indexes, FS: fs}, nil
}

func repoIndexes(fs FileSystem, filenames []string) chan BlobIndex {
	res := make(chan BlobIndex, len(filenames))
	var bar *progressbar.ProgressBar
	if IsProgressBarEnabled {
		bar = progressbar.Default(int64(len(filenames)))
	}
	var wg sync.WaitGroup
	for _, filename := range filenames {
		if IsProgressBarEnabled {
			bar.Add(1)
		}
		textBytes, err := fs.ReadRelativeFilename(filename)
		if err != nil {
			fmt.Printf("err %v\n", err)
			continue
		}
		if len(textBytes) == 0 {
			continue
		}
		if len(textBytes) > maxFileSize {
			continue
		}
		if enry.IsBinary(textBytes) {
			continue
		}
		text := string(textBytes)
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			for i := range text {

			}
			bloomSize := uint(len(ngrams.SeenHashes))
			filter := bloom.NewWithEstimates(bloomSize, targetFalsePositiveRatio)
			data := make([]byte, unsafe.Sizeof(uint64(1)))
			for hash := range ngrams.SeenHashes {
				binary.LittleEndian.PutUint64(data, hash)
				filter.Add(data)
			}
			res <- BlobIndex{Path: path, Filter: filter}
		}(filename)
	}
	wg.Wait()
	close(res)
	return res
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
		for _, line := range strings.Split(text, "\n") {
			columnNumber := strings.Index(line, query)
			if columnNumber >= 0 {
				matchCount++
				//prefix := line[0:columnNumber]
				//suffix := line[columnNumber+len(query):]
				//fmt.Printf(
				//	"%v:%v:%v %v%v%v\n",
				//	matchingPath,
				//	lineNumber,
				//	columnNumber,
				//	prefix,
				//	Yellow(query),
				//	suffix,
				//)
			}
			start = end + 1
			end = strings.Index(text[end+1:], "\n")
		}

		totalMatchCount = totalMatchCount + uint64(matchCount)
		if matchCount > 0 {
			truePositive++
		} else {
			if falsePositive == 1 {
				fmt.Printf("FALSE POSITIVE %v\n", matchingPath)
			}
			fmt.Printf("FALSE POSITIVE %v\n", matchingPath)
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
	grams := CollectQueryNgrams(query)
	var result []string
	r.pathsMatchingQuerySync(grams, r.Blobs, func(matchingPath string) {
		result = append(result, matchingPath)
	})
	return result
}

func (r *RepoIndex) PathsMatchingQuery(query string) chan string {
	grams := CollectQueryNgrams(query)
	res := make(chan string, len(r.Blobs))
	batchSize := 10_000
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
