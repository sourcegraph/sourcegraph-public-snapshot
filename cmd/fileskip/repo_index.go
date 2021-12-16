package fileskip

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/cockroachdb/errors"
	"github.com/go-enry/go-enry/v2"
	"github.com/schollz/progressbar/v3"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	Yellow = color("\033[1;33m%s\033[0m")
)

const (
	Version                  = 1
	targetFalsePositiveRatio = 0.01
	maxFileSize              = 1 << 20 // 1_048_576
	maximumQueryNgrams       = 100
)

var IsProgressBarEnabled = true

type QueryBitmask struct {
	Bitmask     *roaring64.Bitmap
	Cardinality uint64
}

type RepoIndex struct {
	Dir   string
	Blobs []BlobIndex
	FS    FileSystem
}
type BlobIndex struct {
	Filter *roaring64.Bitmap
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

var unigramArity = uint64(1) << 62
var bigramArity = uint64(2) << 62
var trigramArity = uint64(3) << 62

func ngramArity(n uint64) int8 {
	if (trigramArity & n) == trigramArity {
		return 3
	}
	if (bigramArity & n) == bigramArity {
		return 2
	}
	return 1
}

func onGrams(text string) *roaring64.Bitmap {
	seen := roaring64.New()
	ch1 := uint64(0)
	ch2 := uint64(0)
	i := 0
	for _, ch0 := range text {
		unigram := uint64(ch0)
		seen.Add(unigram | unigramArity)
		if i > 1 {
			bigram := unigram | (ch1 << 8) | bigramArity
			seen.Add(bigram)
		}
		if i > 2 {
			trigram := unigram | (ch1 << 8) | (ch2 << 16) | trigramArity
			seen.Add(trigram)
		}
		ch2 = ch1
		ch1 = unigram
		i++
	}
	return seen
}

func CollectQueryNgrams(query string) *QueryBitmask {
	grams := onGrams(query)
	return &QueryBitmask{Bitmask: grams, Cardinality: grams.GetCardinality()}
	//result := make([][]byte, len(ngrams))
	//arities := make([]int8, len(ngrams))
	//i := 0
	//for hash := range ngrams {
	//	data := make([]byte, unsafe.Sizeof(hash))
	//	binary.LittleEndian.PutUint64(data, hash)
	//	result[i] = data
	//	arities[i] = ngramArity(hash)
	//	i++
	//}
	//randomNumbers := make([]int, len(ngrams))
	//for i := range randomNumbers {
	//	randomNumbers[i] = rand.Int()
	//}
	//sort.SliceStable(result, func(i, j int) bool {
	//	if arities[i] == arities[j] {
	//		// Shuffle the ordering of n-grams with the same arity to increase entropy
	//		// among the n-grams that appear first in the results.
	//		// For example, the ID number in the query "bugzilla.redhat.com/show_bug.cgi?id=726143"
	//		// appears at the end of the query and we want to move the n-grams from that
	//		// ID to appear early in the results to allow the first bloom filter tests to exit early.
	//		// we want to avoid the case where we test only the start of the query
	//		return randomNumbers[i] < randomNumbers[j]
	//	}
	//	return arities[i] > arities[j]
	//})
	//if len(result) > maximumQueryNgrams {
	//	result = result[:maximumQueryNgrams]
	//}
	//return result
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
	batchSize := 100
	var wg sync.WaitGroup
	for i := 0; i < len(filenames); i += batchSize {
		j := i + batchSize
		if len(filenames) < j {
			j = len(filenames)
		}
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for _, filename := range filenames[start:end] {
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
				ngrams := onGrams(text)
				res <- BlobIndex{Path: filename, Filter: ngrams}
			}
		}(i, j)
	}
	wg.Wait()
	close(res)
	return res
}

func (r *RepoIndex) Grep(query string) {
	start := time.Now()
	matchingPaths := r.FilenamesMatchingQuery(query)
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
	query *QueryBitmask,
	batch []BlobIndex,
	onMatch func(matchingPath string),
) {
	for _, index := range batch {
		if index.Filter == nil {
			continue
		}
		if query.Bitmask.AndCardinality(index.Filter) == query.Cardinality {
			onMatch(index.Path)
		}
	}
}

func (r *RepoIndex) FilenamesMatchingQuerySync(query string) []string {
	grams := CollectQueryNgrams(query)
	var result []string
	r.pathsMatchingQuerySync(grams, r.Blobs, func(matchingPath string) {
		result = append(result, matchingPath)
	})
	return result
}

func (r *RepoIndex) FilenamesMatchingQuery(query string) chan string {
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
