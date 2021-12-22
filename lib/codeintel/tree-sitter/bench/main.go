package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	Filename string
	Bytes    []byte
}

func main() {
	benchmarkSyntect()
}
func benchmarkSyntect() {
	client := gosyntect.New("http://127.0.0.1:9238")
	goInputs, err := kubernetes.goInputs()
	if err != nil {
		panic(err)
	}
	//goInputs = goInputs[0:1000]
	bench := hrtime.NewBenchmark(len(goInputs))
	i := 0
	for bench.Next() {
		goInputs[i].benchmarkSyntect(client)
		i++
	}
	fmt.Println(bench.Histogram(10))
}

func (i *Input) syntectQuery() *gosyntect.Query {
	return &gosyntect.Query{
		Code:             string(i.Bytes),
		Filepath:         i.Filename,
		StabilizeTimeout: 30 * time.Second,
		LineLengthLimit:  2_000,
		CSS:              true,
	}
}
func benchmarkTreeSitter() {
	goFiles, err := kubernetes.goInputs()
	if err != nil {
		panic(err)
	}
	fmt.Println("Number of go files:")
	fmt.Println(len(goFiles))

	bench := hrtime.NewBenchmark(len(goFiles))
	i := 0
	for bench.Next() {
		goFiles[i].benchmarkTreeSitter()
		i++
	}
	histogram := bench.Histogram(20)
	fmt.Println(histogram)
}

func (i *Input) benchmarkSyntect(client *gosyntect.Client) {
	_, err := client.Highlight(context.Background(), i.syntectQuery())
	if err != nil {
		panic(err)
	}
}
func (i *Input) benchmarkTreeSitter() {
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())
	_, err := parser.ParseCtx(context.Background(), nil, i.Bytes)
	if err != nil {
		panic(err)
	}
}

func readInput(file *zip.File, reader *zip.Reader) (*Input, error) {
	open, err := reader.Open(file.Name)
	if err != nil {
		return nil, err
	}
	stat, err := open.Stat()
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return &Input{
			Filename: file.Name,
			Bytes:    []byte{},
		}, nil
	}

	data := make([]byte, stat.Size())
	_, err = io.ReadFull(open, data)
	return &Input{
		Filename: file.Name,
		Bytes:    data,
	}, err
}

func (c *Corpus) goInputs() ([]*Input, error) {
	reader, err := c.openZipReader()
	if err != nil {
		return nil, err
	}
	var goFiles []*Input
	for _, file := range reader.File {
		if !strings.HasSuffix(file.Name, ".go") {
			continue
		}

		input, err := readInput(file, reader)
		if err != nil {
			return nil, err
		}

		goFiles = append(goFiles, input)
	}
	return goFiles, nil
}

func (c *Corpus) openZipReader() (*zip.Reader, error) {
	url, err := c.DownloadUrlAndCache()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(url)
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
