package vfsutil

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// NewGitHubRepoVFS creates a new VFS backed by a GitHub downloadable
// repository archive.
func NewGitHubRepoVFS(repo, rev string) (*ArchiveFS, error) {
	if !githubRepoRx.MatchString(repo) {
		return nil, fmt.Errorf(`invalid GitHub repo %q: must be "github.com/user/repo"`, repo)
	}

	fetch := func(ctx context.Context) (ar *archiveReader, err error) {
		span, ctx := opentracing.StartSpanFromContext(ctx, "GitRepoVFS fetch")
		defer func() {
			if err != nil {
				ext.Error.Set(span, true)
				span.LogEvent(fmt.Sprintf("error: %v", err))
			}
			span.Finish()
		}()

		url := fmt.Sprintf("https://codeload.%s/zip/%s", repo, rev)
		span.SetTag("repo", repo)
		span.SetTag("commit", rev)

		urlMu := urlMu(url)
		urlMu.Lock()
		defer urlMu.Unlock()
		span.LogEvent("urlMu acquired")

		var zr *zip.Reader
		var closer io.Closer
		fsPath := filepath.Join("/tmp/xlang-github-cache", repo, rev+".zip")
		var zrc *zip.ReadCloser
		zrc, err = zip.OpenReader(fsPath)
		if err == nil {
			span.LogEvent("read from " + fsPath)
			zr = &zrc.Reader
			closer = zrc
		}
		if os.IsNotExist(err) {
			// https://github.com/a/b/archive/master.zip redirects to
			// codeload.github.com, so let's just use the latter directly and
			// save a roundtrip.
			span.LogEvent("fetching " + url)
			resp, err := http.Get(url)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("github repo archive: URL %s returned HTTP %d", url, resp.StatusCode)
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			span.LogEvent("fetched " + url)
			ghBytes.Add(float64(len(body)))

			// Cache on the file system.
			if err := os.MkdirAll(filepath.Dir(fsPath), 0700); err != nil {
				return nil, err
			}
			if err := ioutil.WriteFile(fsPath, body, 0600); err != nil {
				return nil, err
			}
			span.LogEvent("cached to " + fsPath)

			zr, err = zip.NewReader(bytes.NewReader(body), int64(len(body)))
			if err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}

		// GitHub zip files have a top-level dir "{repobasename}-{sha}/",
		// so we need to remove that. The repobasename is in the canonical
		// casing, which may be different from fs.repo.
		if len(zr.File) == 0 {
			return nil, errors.New("zip archive has no files")
		}
		prefix := zr.File[0].Name
		if strings.Contains(prefix, "/") {
			prefix = path.Dir(prefix)
		}

		return &archiveReader{
			Reader: zr,
			Closer: closer,
			Prefix: prefix + "/",
		}, nil
	}
	return &ArchiveFS{fetch: fetch}, nil
}

var githubRepoRx = regexp.MustCompile(`^github\.com/[\w.-]{1,100}/[\w.-]{1,100}$`)

var ghBytes = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "xlang",
	Subsystem: "vfs",
	Name:      "github_bytes_total",
	Help:      "Total number of bytes read into memory by GitHubRepoVFS.",
})

func init() {
	prometheus.MustRegister(ghBytes)
}
