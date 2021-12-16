package main

import (
	"crypto/md5"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/loov/hrtime"
	"github.com/schollz/progressbar/v3"
	"github.com/sourcegraph/sourcegraph/cmd/fileskip"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	flask = Corpus{
		Name: "flask", URL: "https://github.com/pallets/flask/archive/refs/tags/2.0.2.zip",
		Queries: []string{"Z", "62", "204", "text", "Text", "96.944917", "sqlite3", "flask.request.endpoint"},
	}
	sourcegraph = Corpus{Name: "sourcegraph", URL: "https://github.com/sourcegraph/sourcegraph/archive/refs/tags/v3.34.1.zip",
		Queries: []string{
			"add",
			"121",
			"incre",
			"thread",
			"pageres",
			"exec.Command",
			"page", "Page", "Repository", "FileTree", "bloomf",
			"COMMENT ON COLUMN lsif", "The identifier of the associated dump",
		},
	}
	kubernetes = Corpus{Name: "kubernetes", URL: "https://github.com/kubernetes/kubernetes/archive/refs/tags/v1.22.4.zip",
		Queries: []string{"OPZ", "Q13", "rrra", "Resolver", "buildServiceResolver", "cache.ResourceEventHandlerFuncs"},
	}
	linux = Corpus{Name: "linux", URL: "https://github.com/torvalds/linux/archive/refs/tags/v5.16-rc3.zip",
		Queries: []string{
			"add", "gnttab", "unmap", "phys_add", "phys_add", "phys-addr", "phys_addr", "44441", "soundcard", "#include <sys/socket.h>",
			"new address of the crtc (GPU MC address)",
			"bugzilla.redhat.com/show_bug.cgi?id=726143",
			"Clone map from listener for newly accepted socket",
		},
	}
	chromium = Corpus{
		Name: "chromium", URL: "https://github.com/chromium/chromium/archive/refs/tags/98.0.4747.1.zip",
		Queries: []string{
			"nest",
			"nested",
			"address",
			"folded",
			"Instru",
			"messager",
			"messenger",
			"params.has_value()",
			"assert_true(params.has",
			"EXPECT_EQ(kDownloadId, pa",
			"CanShowContextMenuForParams",
			"http://somehost/path?x=id%3Daaaa%26v%3D1.1%26uc&x=id%3Dbbbb%26v%3D2.0%26uc",
		},
	}
	megarepo = Corpus{
		Name: "megarepo", URL: "https://github.com/sgtest/megarepo/zipball/11c726fd66bb6252cb8e9c0af8933f5ba0fb1e8d",
		Queries: []string{
			"44a1",
			"*hl",
			"*hl_",
			"if (bflo",
			"OutputPre",
			"FolderStru",
			"FolderStru",
			"TEST_F(WindowDatasetOpTes",
			"IteratorOutputPrefixTestCases",
			"EffectiveOperandPrecisionIsBF16",
			"github.com/Azure/go-autorest/autorest",
		},
	}

	all = []Corpus{flask, sourcegraph, kubernetes, linux, chromium, megarepo}
)

func main() {
	if len(os.Args) < 3 {
		panic("missing argument for corpus name")
	}
	var corpora []Corpus
	if os.Args[2] == "all" {
		corpora = all
	}
	for _, c := range all {
		if c.Name == os.Args[2] {
			corpora = []Corpus{c}
			break
		}
	}
	if len(corpora) == 0 {
		panic("no corpus matching name " + os.Args[2])
	}
	switch os.Args[1] {
	case "download":
		for _, corpus := range corpora {
			corpus.DownloadUrlAndCache()
		}
	case "bench":
		for _, corpus := range corpora {
			err := corpus.run()
			if err != nil {
			}
			panic(err)
		}
	case "grep":
		for _, corpus := range corpora {
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
}

type Query struct {
	Value string
}

type Corpus struct {
	Name    string
	URL     string
	Queries []string
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

func (c *Corpus) LoadFileSystem() (*fileskip.ZipFileSystem, error) {
	path, err := c.DownloadUrlAndCache()
	if err != nil {
		return nil, err
	}
	return fileskip.NewZipFileSystem(path)
}

func (c *Corpus) LoadRepoIndex() (*fileskip.RepoIndex, error) {
	fs, err := c.LoadFileSystem()
	if err != nil {
		return nil, err
	}
	//cached, err := c.loadCachedRepoIndex()
	//if err == nil && cached != nil {
	//	return cached, nil
	//}

	return fileskip.NewInMemoryRepoIndex(fs)
}

func (c *Corpus) loadCachedRepoIndex() (*fileskip.RepoIndex, error) {
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
	var blobs []fileskip.BlobIndex
	index := &fileskip.BlobIndex{}
	from, err := index.ReadFrom(file)
	for from > 0 && err == nil {
		blobs = append(blobs, *index)
		from, err = index.ReadFrom(file)
	}
	//if err != nil {
	//	return nil, errors
	//}
	result := &fileskip.RepoIndex{Blobs: blobs}
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
			fileskip.Version,
		),
	)
}

func (c *Corpus) run() error {
	index, err := c.LoadRepoIndex()
	if err != nil {
		return err
	}
	isMatch := map[string]map[string]struct{}{}
	bar := progressbar.DefaultBytes(
		int64(len(index.Blobs)),
		"testing",
	)
	for _, query := range c.Queries {
		isMatch[query] = map[string]struct{}{}
	}
	var wg sync.WaitGroup
	batchSize := 100
	for i := 0; i < len(index.Blobs); i += batchSize {
		j := i + batchSize
		if len(index.Blobs) < j {
			j = len(index.Blobs)
		}
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for _, p := range index.Blobs[start:end] {
				bar.Add(1)
				bytes, _ := index.FS.ReadRelativeFilename(p.Path)
				text := string(bytes)
				for _, query := range c.Queries {
					if strings.Index(text, query) >= 0 {
						isMatch[query][p.Path] = struct{}{}
					}
				}
			}
		}(i, j)
	}
	wg.Wait()
	for _, query := range c.Queries {
		header := "=========" + strings.Repeat("=", len(query))
		fmt.Println(header)
		fmt.Println("== Query " + query)
		fmt.Println(header)
		bench := hrtime.NewBenchmark(50)
		matchingPaths := map[string]struct{}{}
		for bench.Next() {
			for path := range index.FilenamesMatchingQuery(query) {
				matchingPaths[path] = struct{}{}
			}
		}
		bench.Laps()
		hg := bench.Histogram(5)
		fmt.Println(hg)
		if index.FS != nil {
			falsePositives := 0
			var falseNegatives []string
			for _, p := range index.Blobs {
				_, found := isMatch[query][p.Path]
				_, isBloomFound := matchingPaths[p.Path]
				if isBloomFound && !found {
					falsePositives++
				} else if found && !isBloomFound {
					falseNegatives = append(falseNegatives, p.Path)
				}
			}
			if len(falseNegatives) > 0 {
				panic(fmt.Sprintf("false negatives %v", falseNegatives))
			}
			falsePositiveRatio := float64(falsePositives) / math.Max(1, float64(len(matchingPaths)))
			fmt.Printf(
				"total %v tp %v fp %v (%.2f%%)\n",
				len(matchingPaths),
				len(matchingPaths)-falsePositives,
				falsePositives,
				falsePositiveRatio*100,
			)
		}
	}
	return nil
}
