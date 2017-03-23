package search

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkConcurrentFind_large_fixed(b *testing.B) {
	benchConcurrentFind(b, &Params{
		Repo:    "github.com/golang/go",
		Commit:  "0ebaca6ba27534add5930a95acffa9acff182e2b",
		Pattern: "error handler",
	})
}

func BenchmarkConcurrentFind_small_fixed(b *testing.B) {
	benchConcurrentFind(b, &Params{
		Repo:    "github.com/sourcegraph/go-langserver",
		Commit:  "4193810334683f87b8ed5d896aa4753f0dfcdf20",
		Pattern: "object not found",
	})
}

func benchConcurrentFind(b *testing.B, p *Params) {
	if testing.Short() {
		b.Skip("")
	}
	b.ReportAllocs()

	err := validateParams(p)
	if err != nil {
		b.Fatal(err)
	}

	rg, err := compile(p)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	zr := openGithubArchive(b, p.Repo, p.Commit)
	defer zr.Close()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_, err := concurrentFind(ctx, rg, &zr.Reader)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func openGithubArchive(t testing.TB, repo, rev string) *zip.ReadCloser {
	// key is a sha256 hash since we want to use it for the disk name
	h := sha256.Sum256([]byte(repo + " " + rev))
	key := hex.EncodeToString(h[:])
	path := filepath.Join("/tmp/search_test", key+".zip")
	zr, err := zip.OpenReader(path)
	if err == nil {
		return zr
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}

	url := fmt.Sprintf("https://codeload.%s/zip/%s", repo, rev)
	t.Log("fetching", url)

	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("github repo archive: URL %s returned HTTP %d", url, resp.StatusCode)
	}
	f, err := os.OpenFile(path+".part", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { os.Remove(path + ".part") }()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	// We write our own version of the zip to disk, since we want to store
	// it uncompressed.
	r, err := zip.OpenReader(path + ".part")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	err = storeZip(&r.Reader, w)
	if err != nil {
		t.Fatal(err)
	}
	err = w.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile(path, buf.Bytes(), 0600)
	if err != nil {
		t.Fatal(err)
	}

	zr, err = zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	return zr
}

func storeZip(zr *zip.Reader, zw *zip.Writer) error {
	for _, f := range zr.File {
		r, err := f.Open()
		if err != nil {
			return err
		}
		w, err := zw.CreateHeader(&zip.FileHeader{
			Name:   f.Name,
			Method: zip.Store,
		})
		if err != nil {
			return err
		}
		_, err = io.Copy(w, r)
		if err != nil {
			return err
		}
		r.Close()
	}
	return nil
}
