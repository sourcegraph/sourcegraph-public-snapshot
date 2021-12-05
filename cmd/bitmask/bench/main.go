package main

import (
	"crypto/md5"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/loov/hrtime"
	"github.com/schollz/progressbar/v3"
	"github.com/sourcegraph/sourcegraph/cmd/bitmask"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	flask = Corpus{
		Name: "flask", URL: "https://github.com/pallets/flask/archive/refs/tags/2.0.2.zip",
		Queries: []string{"Z", "62", "204", "text", "Text", "96.944917", "sqlite3", "flask.request.endpoint"},
	}
	sourcegraph = Corpus{Name: "sourcegraph", URL: "https://github.com/sourcegraph/sourcegraph/archive/refs/tags/v3.34.1.zip",
		Queries: []string{
			"ö", "oö", "121", "OLA", "page", "Page", "Repository", "FileTree", "bloomf",
			"COMMENT ON COLUMN lsif", "The identifier of the associated dump",
		},
	}
	kubernetes = Corpus{Name: "kubernetes", URL: "https://github.com/kubernetes/kubernetes/archive/refs/tags/v1.22.4.zip",
		Queries: []string{"OPZ", "Q13", "rrra", "Resolver", "buildServiceResolver", "cache.ResourceEventHandlerFuncs"},
	}
	linux = Corpus{Name: "linux", URL: "https://github.com/torvalds/linux/archive/refs/tags/v5.16-rc3.zip",
		Queries: []string{
			"ø", "AAA", "44441", "soundcard", "#include <sys/socket.h>",
			"new address of the crtc (GPU MC address)",
			"bugzilla.redhat.com/show_bug.cgi?id=726143",
			"Clone map from listener for newly accepted socket",
		},
	}
	chromium = Corpus{
		Name: "chromium", URL: "https://github.com/chromium/chromium/archive/refs/tags/98.0.4747.1.zip",
		Queries: []string{
			"øø",
			"799a",
			"framm",
			"params.has_value()",
			"assert_true(params.has",
			"EXPECT_EQ(kDownloadId, pa",
			"CanShowContextMenuForParams",
			"http://somehost/path?x=id%3Daaaa%26v%3D1.1%26uc&x=id%3Dbbbb%26v%3D2.0%26uc",
		},
	}
	all = []Corpus{flask, sourcegraph, kubernetes, linux, chromium}
)

//func main() {
//	fs, err := flask.LoadFileSystem()
//	if err != nil {
//		panic(err)
//	}
//	err = bitmask.NewOnDiskRepoIndex(fs, flask.indexCachePath())
//	if err != nil {
//		panic(err)
//	}
//}
func main() {
	if len(os.Args) < 2 {
		panic("missing argument for corpus name")
	}
	var corpus Corpus
	for _, c := range all {
		if c.Name == os.Args[2] {
			corpus = c
			break
		}
	}
	if corpus.Name == "" {
		panic("no corpus matching name " + os.Args[1])
	}
	switch os.Args[1] {
	case "bench":
		err := corpus.run()
		if err != nil {
			panic(err)
		}
	case "grep":
		if len(os.Args) < 3 {
			panic("missing grep argument")
		}
		query := os.Args[3]
		index, err := corpus.LoadRepoIndex()
		if err != nil {
			panic(err)
		}
		index.Grep(query)
	}
}

type Query struct {
	Value string
}

type Corpus struct {
	Name    string
	URL     string
	Queries []string
}

func DownloadUrlAndCache(corpus *Corpus) (string, error) {
	path := corpus.zipCachePath()
	stat, err := os.Stat(path)
	if err == nil && !stat.IsDir() && stat.Size() > 0 {
		return path, nil
	}
	fmt.Printf("Downloading... %v\n", corpus.URL)
	resp, err := http.Get(corpus.URL)
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

func (c *Corpus) LoadFileSystem() (*bitmask.ZipFileSystem, error) {
	path, err := DownloadUrlAndCache(c)
	if err != nil {
		return nil, err
	}
	return bitmask.NewZipFileSystem(path)
}

func (c *Corpus) LoadRepoIndex() (*bitmask.RepoIndex, error) {
	fs, err := c.LoadFileSystem()
	if err != nil {
		return nil, err
	}
	cached, err := c.loadCachedRepoIndex()
	if err == nil && cached != nil {
		return cached, nil
	}

	err = bitmask.NewOnDiskRepoIndex(fs, c.indexCachePath())
	if err != nil {
		return nil, err
	}
	return c.loadCachedRepoIndex()
}

func (c *Corpus) loadCachedRepoIndex() (*bitmask.RepoIndex, error) {
	stat, err := os.Stat(c.indexCachePath())
	if err != nil {
		return nil, err
	}
	if !stat.Mode().IsRegular() {
		return nil, errors.Errorf("no such file: %v", c.indexCachePath())
	}
	file, err := os.Open(c.indexCachePath())
	if err != nil {
		return nil, err
	}
	var blobs []bitmask.BlobIndex
	index := &bitmask.BlobIndex{}
	from, err := index.ReadFrom(file)
	for from > 0 && err == nil {
		blobs = append(blobs, *index)
		from, err = index.ReadFrom(file)
	}
	//if err != nil {
	//	return nil, errors
	//}
	result := &bitmask.RepoIndex{Blobs: blobs}
	fs, err := c.LoadFileSystem()
	if err != nil {
		return nil, err
	}
	result.FS = fs
	return result, nil
}

func (c *Corpus) zipCachePath() string {
	return filepath.Join(
		os.TempDir(),
		fmt.Sprintf("%v-%x.zip", c.Name, md5.Sum([]byte(c.URL))),
	)
}

func (c *Corpus) indexCachePath() string {
	return filepath.Join(
		os.TempDir(),
		fmt.Sprintf(
			"%v-%x.repo-index-v%v",
			c.Name,
			md5.Sum([]byte(c.URL)),
			bitmask.Version,
		),
	)
}

func (c *Corpus) run() error {
	index, err := c.LoadRepoIndex()
	if err != nil {
		return err
	}
	for _, query := range c.Queries {
		header := "=========" + strings.Repeat("=", len(query))
		fmt.Println(header)
		fmt.Println("== Query " + query)
		fmt.Println(header)
		bench := hrtime.NewBenchmark(50)
		var matchingPaths []string
		for bench.Next() {
			matchingPaths = []string{}
			for path := range index.PathsMatchingQuery(query) {
				matchingPaths = append(matchingPaths, path)
			}
		}
		bench.Laps()
		hg := bench.Histogram(5)
		fmt.Println(hg)
		if index.FS != nil {
			falsePositives := 0
			for _, p := range matchingPaths {
				bytes, _ := index.FS.ReadRelativeFilename(p)
				text := string(bytes)
				if strings.Index(text, query) < 0 {
					falsePositives++
				}
			}
			falsePositiveRatio := float64(falsePositives) / math.Max(1, float64(len(matchingPaths)))
			fmt.Printf("paths %v fp %v (%.2f%%)\n", len(matchingPaths), falsePositives, falsePositiveRatio*100)
		}
	}
	return nil
}
