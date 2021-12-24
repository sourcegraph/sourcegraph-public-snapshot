package main

/*
Run instructions:
# current directory is sourcegraph/

# Only needed if you want to run the synhtml-stdin benchmarks
cd ..
git clone --branch vg/add-synhtml-stdin https://github.com/varungandhi-src/syntect.git
cd syntect
cargo build --release --example synhtml-stdin
# Somewhere in $PATH
cp target/release/examples/synhtml-stdin ~/.local/bin/synhtml-stdin

git clone --branch vg/time-nanos https://github.com/sourcegraph/gosyntect.git ../gosyntect

cd ../sourcegraph/docker-images/syntax-highlighter
cargo build --release

# In separate pane
git clone https://github.com/slimsag/http-server-stabilizer.git
cd http-server-stabilizer
go build
# I think ROCKET_WORKERS=N should match concurrency, but not sure.
# Product of workers * concurrency should probably be NCPUs.
./http-server-stabilizer -workers 3 -concurrency 4 -listen ":8000" -- env ROCKET_PORT='{{.PORT}} ROCKET_WORKERS=4 ../sourcegraph/docker-images/syntax-highlighter/target/release/syntect_server

# In separate pane

cd ../sourcegraph/lib
go run github.com/sourcegraph/sourcegraph/lib/codeintel/tree-sitter/bench
*/

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/loov/hrtime"
	"github.com/schollz/progressbar/v3"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/sourcegraph/gosyntect"
)

var (
	kubernetes = &Corpus{Name: "kubernetes", URL: "https://github.com/kubernetes/kubernetes/archive/refs/tags/v1.22.4.zip"}
	megarepo   = &Corpus{
		Name: "megarepo", URL: "https://github.com/sgtest/megarepo/zipball/11c726fd66bb6252cb8e9c0af8933f5ba0fb1e8d",
	}
	all = []*Corpus{kubernetes, megarepo}
)

type Corpus struct {
	Name string
	URL  string
}

type Input struct {
	ZipRelativePath string
	Bytes           []byte
}

const SIZE_LIMIT = 512 * 1024
const SYNTECT_SERVER_URL = "http://0.0.0.0:8000"
const TREE_SITTER = "tree-sitter"
const NPARALLELISM = 10

var extMap = map[string]struct{}{
	".go":   {},
	".c":    {},
	".h":    {},
	".js":   {},
	".jsx":  {},
	".cpp":  {},
	".hpp":  {},
	".ts":   {},
	".tsx":  {},
	".dart": {},
	".rb":   {},
	".rs":   {},
	".java": {},
	".kt":   {},
}

func TryHighlightFileWithExtension(extension string) bool {
	_, exists := extMap[extension]
	return exists
}

type TreeSitter struct{}

func (TreeSitter) name() string                        { return TREE_SITTER }
func (TreeSitter) benchmarkSetup()                     {}
func (TreeSitter) benchmark(input Input) time.Duration { return input.benchmarkTreeSitter() }

var _ Workload = TreeSitter{}

type Synhtml struct{}

func (Synhtml) name() string                        { return "synhtml-stdin" }
func (Synhtml) benchmarkSetup()                     {}
func (Synhtml) benchmark(input Input) time.Duration { return input.benchmarkSynhtml() }

var _ Workload = Synhtml{}

type Syntect struct{ client *gosyntect.Client }

func (*Syntect) name() string                             { return "syntect-server" }
func (self *Syntect) benchmarkSetup()                     { self.client = gosyntect.New(SYNTECT_SERVER_URL) }
func (self *Syntect) benchmark(input Input) time.Duration { return input.benchmarkSyntect(self.client) }

var _ Workload = &Syntect{}

func benchmarkHistogram(w Workload, inputs []Input) {
	fmt.Println(fmt.Sprintf("Benchmarking %s (total)", w.name()))
	bench := hrtime.NewBenchmark(len(inputs))
	highlightTimes := make([]time.Duration, len(inputs))
	w.benchmarkSetup()
	i := 0
	for bench.Next() {
		highlightTimes[i] = w.benchmark(inputs[i])
		i++
	}
	fmt.Println(bench.Histogram(10))
	if w.name() != TREE_SITTER {
		fmt.Printf("%s (highlight only)\n", w.name())
		fmt.Println(hrtime.NewDurationHistogram(highlightTimes, &hrtime.HistogramOptions{BinCount: 10}))
	}
}

func runParallel(functions []func()) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(functions))
	for _, function := range functions {
		go func(doStuff func()) {
			doStuff()
			waitGroup.Done()
		}(function)
	}
	waitGroup.Wait()
}

func benchmarkPure(w Workload, inputs []Input, outputs []Output) {
	w.benchmarkSetup()

	runFunc := func(startIndex int, endIndex int) func() {
		return func() {
			for i := startIndex; i < endIndex; i++ {
				input := inputs[i]
				before := time.Now()
				highlightDuration := w.benchmark(input)
				outputs[i].TotalDuration = time.Now().Sub(before)
				if highlightDuration == -1 {
					outputs[i].HighlightDuration = outputs[i].TotalDuration
				} else {
					outputs[i].HighlightDuration = highlightDuration
				}
			}
		}
	}
	chunkLen := len(inputs) / NPARALLELISM
	funcs := []func(){}
	for startIdx := 0; startIdx < len(inputs); startIdx += chunkLen {
		funcs = append(funcs, runFunc(startIdx, minInt(startIdx+chunkLen, len(inputs))))
	}
	runParallel(funcs)
}

func minInt(a int, b int) int { // Gimme generics
	if a < b {
		return a
	}
	return b
}

func createOutputCSV() (*os.File, string) {
	file, err := os.CreateTemp(os.TempDir(), "parse-benchmark")
	if err != nil {
		panic(fmt.Sprintf("failed to create temp file for output: %+v", err))
	}
	oldPath := file.Name()
	newPath := oldPath + ".csv"
	if err = file.Close(); err != nil {
		panic(fmt.Sprintf("failed to close temp file: %+v", err))
	}
	if err = os.Rename(oldPath, newPath); err != nil {
		panic(fmt.Sprintf("failed to rename file to CSV: %+v", err))
	}
	file, err = os.OpenFile(newPath, os.O_WRONLY, 0755)
	if err != nil {
		panic(fmt.Sprintf("failed to open newly created CSV: %+v", err))
	}
	return file, newPath
}

func histogramMain(testCorpora []*Corpus, workloads []Workload) {
	for _, corpus := range testCorpora {
		inputs, err := corpus.testInputs()
		if err != nil {
			panic(fmt.Sprintf("failed to get inputs for corpus %s: %+v", corpus.Name, err))
		}
		for _, w := range workloads {
			benchmarkHistogram(w, inputs)
		}
	}
}

func csvMain(testCorpora []*Corpus, workloads []Workload) {
	outputCSV, csvPath := createOutputCSV()
	defer outputCSV.Close()
	for _, corpus := range testCorpora {
		inputs, err := corpus.testInputs()
		if err != nil {
			panic(fmt.Sprintf("failed to get inputs for corpus %s: %+v", corpus.Name, err))
		}
		fullOutput := FullOutput{inputs: inputs}
		for _, w := range workloads {
			outputs := make([]Output, len(inputs))
			benchmarkPure(w, inputs, outputs)
			fullOutput.timings = append(fullOutput.timings, outputs)
			fullOutput.parsers = append(fullOutput.parsers, w.name())
		}
		fullOutput.appendTo(outputCSV)
	}
	fmt.Printf("Recorded outputs in\n%s\n", csvPath)
	names := []string{}
	for _, w := range workloads {
		names = append(names, w.name())
	}
}

func main() {
	// testCorpora := []*Corpus{megarepo}
	testCorpora := []*Corpus{kubernetes}
	workloads := []Workload{TreeSitter{}, &Syntect{} /*&Synhtml{}*/}
	if len(os.Args) >= 2 && os.Args[1] == "--histogram" {
		histogramMain(testCorpora, workloads)
	} else {
		csvMain(testCorpora, workloads)
	}
}

func (self *FullOutput) appendTo(outputCSV *os.File) {
	w := csv.NewWriter(outputCSV)
	header := []string{"filesize (bytes)", "extension"}
	for _, parser := range self.parsers {
		if parser != TREE_SITTER {
			header = append(header, fmt.Sprintf("%s total time (ns)", parser))
			header = append(header, fmt.Sprintf("%s highlight time (ns)", parser))
		} else {
			header = append(header, fmt.Sprintf("%s parse time (ns)", parser))
		}
	}
	pathIdx := len(header)
	header = append(header, "path")
	if err := w.Write(header); err != nil {
		panic(fmt.Sprintf("Failed to write header to CSV: %v", err))
	}
	for i, input := range self.inputs {
		row := make([]string, len(header))
		row[0] = fmt.Sprintf("%d", len(input.Bytes))
		row[1] = filepath.Ext(input.ZipRelativePath)
		idx := 2
		for j, name := range self.parsers {
			row[idx] = fmt.Sprintf("%d", self.timings[j][i].TotalDuration)
			idx++
			if name != TREE_SITTER {
				row[idx] = fmt.Sprintf("%d", self.timings[j][i].HighlightDuration)
				idx++
			}
		}
		row[pathIdx] = input.ZipRelativePath
		if err := w.Write(row); err != nil {
			panic(fmt.Sprintf("Failed to write header to CSV: %v", err))
		}
		if i%1000 == 0 {
			w.Flush()
		}
	}
	w.Flush()
}

type FullOutput struct {
	inputs  []Input
	timings [][]Output
	parsers []string
}

type Output struct {
	TotalDuration     time.Duration
	HighlightDuration time.Duration
}

type Workload interface {
	name() string
	benchmarkSetup()
	benchmark(input Input) time.Duration
}

func (i *Input) syntectQuery() *gosyntect.Query {
	return &gosyntect.Query{
		Code:             string(i.Bytes),
		Filepath:         i.ZipRelativePath,
		StabilizeTimeout: 30 * time.Second,
		LineLengthLimit:  2_000,
		CSS:              true,
	}
}

func (i *Input) benchmarkSyntect(client *gosyntect.Client) time.Duration {
	query := i.syntectQuery()
	resp, err := client.Highlight(context.Background(), query)
	if errors.Is(err, gosyntect.ErrHSSWorkerTimeout) {
		return query.StabilizeTimeout
	}
	if err != nil {
		panic(fmt.Sprintf("syntect server failed to highlight code with err = %v", err))
	}
	return time.Duration(resp.TimeNanos)
}

func (i *Input) benchmarkSynhtml() time.Duration {
	cmd := exec.Command("synhtml-stdin", filepath.Join(os.TempDir(), i.ZipRelativePath))
	writer, err := cmd.StdinPipe()
	if err != nil {
		panic(fmt.Sprintf("failed to open stdin pipe to synhtml-stdin err = %v", "", err))
	}
	if err != nil {
		panic(fmt.Sprintf("failed to open pipe for reading synhtml-stdin's stderr: %v", err))
	}
	var errBuf strings.Builder
	cmd.Stderr = &errBuf
	defer writer.Close()
	if nwrote, err := writer.Write(i.Bytes); err != nil {
		panic(fmt.Sprintf("failed to write bytes to stdin for synhtml-stdin err = %v (wrote %d bytes)", err, nwrote))
	}
	if err = cmd.Run(); err != nil {
		panic(fmt.Sprintf("synhtml-stdin failed to exit %v", err))
	}
	highlightTime, err := strconv.ParseInt(strings.TrimSpace(errBuf.String()), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to parse stderr as int64, err = %v", err))
	}
	return time.Duration(highlightTime)
}

func (i *Input) benchmarkTreeSitter() time.Duration {
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())
	_, err := parser.ParseCtx(context.Background(), nil, i.Bytes)
	if err != nil {
		panic(err)
	}
	return -1
}

func readInput(file *zip.File, reader *zip.Reader) (Input, bool, error) {
	if ext := filepath.Ext(file.Name); !TryHighlightFileWithExtension(ext) {
		return Input{}, true, nil
	}
	open, err := reader.Open(file.Name)
	if err != nil {
		return Input{}, false, err
	}
	stat, err := open.Stat()
	if err != nil {
		return Input{}, false, err
	}
	if stat.IsDir() || stat.Size() >= SIZE_LIMIT {
		return Input{}, true, nil
	}

	data := make([]byte, stat.Size())
	_, err = io.ReadFull(open, data)
	return Input{
		ZipRelativePath: file.Name,
		Bytes:           data,
	}, false, err
}

func (c *Corpus) testInputs() ([]Input, error) {
	reader, err := c.openZipReader()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open zip reader for corpus: %+v", c)
	}
	var inputs []Input
	for _, file := range reader.File {
		input, skip, err := readInput(file, reader)
		if err != nil {
			return nil, err
		}
		if skip {
			continue
		}
		inputs = append(inputs, input)
	}
	// inputs = inputs[0:100] // testing
	return inputs, nil
}

func (c *Corpus) openZipReader() (*zip.Reader, error) {
	zipPath, err := c.DownloadUrlAndCache()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(zipPath)
	if err != nil {
		return nil, err
	}
	return zip.NewReader(bytes.NewReader(data), int64(len(data)))
}

func (c *Corpus) zipCachePath() string {
	return filepath.Join(
		os.TempDir(),
		fmt.Sprintf("%v-%x.zip", c.Name, md5.Sum([]byte(c.URL))),
	)
}

func (c *Corpus) DownloadUrlAndCache() (string, error) {
	path := c.zipCachePath()
	stat, err := os.Stat(path)
	if err == nil && !stat.IsDir() && stat.Size() > 0 {
		return path, nil
	}
	fmt.Printf("Downloading... %v\n", c.URL)
	resp, err := http.Get(c.URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer out.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"downloading",
	)

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)

	if err != nil {
		return "", err
	}
	return path, nil
}
