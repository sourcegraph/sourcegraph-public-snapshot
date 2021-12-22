package main

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
)

var kubernetes = Corpus{Name: "kubernetes", URL: "https://github.com/kubernetes/kubernetes/archive/refs/tags/v1.22.4.zip"}

type Corpus struct {
	Name string
	URL  string
}

type Input struct {
	Filename string
	Text     string
}

func main() {
	reader, err := kubernetes.openZipReader()
	if err != nil {
		panic(err)
	}
	var goFiles []*Input
	for _, file := range reader.File {
		if !strings.HasSuffix(file.Name, ".go") {
			continue
		}

		input, err := readInput(file, reader)
		if err != nil {
			panic(err)
		}

		goFiles = append(goFiles, input)
	}
	fmt.Println("Number of go files:")
	fmt.Println(len(goFiles))
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
			Text:     "",
		}, nil
	}

	data := make([]byte, stat.Size())
	_, err = io.ReadFull(open, data)
	return &Input{
		Filename: file.Name,
		Text:     string(data),
	}, err
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
