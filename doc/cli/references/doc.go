package main

//go:generate go run ./doc.go

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func clean(base string) error {
	// Delete every Markdown file that we find, and track the directories that
	// exist.
	dirs := []string{}
	if err := filepath.Walk(base, func(fp string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			dirs = append(dirs, fp)
		} else if path.Ext(fp) == ".md" {
			return os.Remove(fp)
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "error walking Markdown files")
	}

	// Now iterate over the directories depth-first, removing the ones that are
	// empty.
	sort.Slice(dirs, func(i, j int) bool {
		return len(dirs[j]) < len(dirs[i])
	})
	for _, dir := range dirs {
		d, err := os.ReadDir(dir)
		if err != nil {
			log.Fatal(err)
		}

		if len(d) == 0 {
			if err := os.Remove(dir); err != nil {
				return errors.Wrapf(err, "error removing directory %q", dir)
			}
		}
	}

	return nil
}

func get(url string, v any) error {
	resp, err := http.Get(url)
	if err != nil {
		return errors.Wrapf(err, "http get: %s", url)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "http read: %s", url)
	}

	err = json.Unmarshal(b, v)
	if err != nil {
		return errors.Wrapf(err, "http json unmarshal: %s", url)
	}
	return nil
}

func build() error {
	dir, err := os.MkdirTemp("", "src-cli-doc-gen")
	if err != nil {
		return errors.Wrap(err, "creating temporary directory")
	}
	defer os.RemoveAll(dir)

	release := struct {
		Name   string
		Assets []struct {
			Name string
			URL  string `json:"browser_download_url"`
		}
	}{}
	if err := get("https://api.github.com/repos/sourcegraph/src-cli/releases/latest", &release); err != nil {
		return errors.Wrap(err, "src-cli release metadata")
	}

	bin := fmt.Sprintf("src_%s_%s", runtime.GOOS, runtime.GOARCH)
	url := ""
	for _, asset := range release.Assets {
		if bin == asset.Name {
			url = asset.URL
			break
		}
	}

	if url == "" {
		return errors.Newf("failed to find %s for src-cli release %s", bin, release.Name)
	}

	// more succinct to use curl than pipe http.Get into file
	src := filepath.Join(dir, bin)
	srcGet := exec.Command("curl", "-L", "-o", src, url)
	if _, err := srcGet.Output(); err != nil {
		return errors.Wrap(err, "src-cli download")
	}

	if err := os.Chmod(src, 0700); err != nil {
		return errors.Wrap(err, "src-cli mark executable")
	}

	srcDoc := exec.Command(src, "doc", "-o", ".")
	srcDoc.Env = os.Environ()
	// Always set this to 8 so the docs don't change when generated on
	// different machines.
	srcDoc.Env = append(srcDoc.Env, "GOMAXPROCS=8")
	if out, err := srcDoc.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "running src doc:\n%s\n", string(out))
	}

	return nil
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("error getting working directory: %v", err)
	}

	if err := clean(wd); err != nil {
		log.Fatalf("error cleaning working directory: %v", err)
	}

	if err := build(); err != nil {
		log.Fatalf("error building documentation: %v", err)
	}
}
