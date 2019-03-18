package vfsutil

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"strings"

	"golang.org/x/net/context/ctxhttp"

	"github.com/pkg/errors"
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
		span, ctx := opentracing.StartSpanFromContext(ctx, "Archive Fetch")
		ext.Component.Set(span, "githubvfs")
		span.SetTag("repo", repo)
		span.SetTag("commit", rev)
		defer func() {
			if err != nil {
				ext.Error.Set(span, true)
				span.SetTag("err", err)
			}
			span.Finish()
		}()

		ff, err := cachedFetch(ctx, "githubvfs", repo+"@"+rev, func(ctx context.Context) (io.ReadCloser, error) {
			ghFetch.Inc()
			url := fmt.Sprintf("https://codeload.%s/zip/%s", repo, rev)
			resp, err := ctxhttp.Get(ctx, nil, url)
			if err != nil {
				return nil, err
			}
			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				return nil, errors.Errorf("github repo archive: URL %s returned HTTP %d", url, resp.StatusCode)
			}
			return resp.Body, nil
		})
		if err != nil {
			ghFetchFailed.Inc()
			return nil, err
		}
		f := ff.File

		zr, err := zipNewFileReader(f)
		if err != nil {
			f.Close()
			return nil, err
		}

		// GitHub zip files have a top-level dir "{repobasename}-{sha}/",
		// so we need to remove that. The repobasename is in the canonical
		// casing, which may be different from fs.repo.
		if len(zr.File) == 0 {
			f.Close()
			return nil, errors.New("zip archive has no files")
		}
		prefix := zr.File[0].Name
		if strings.Contains(prefix, "/") {
			prefix = path.Dir(prefix)
		}

		return &archiveReader{
			Reader: zr,
			Closer: f,
			Prefix: prefix + "/",
		}, nil
	}
	return &ArchiveFS{fetch: fetch}, nil
}

var githubRepoRx = regexp.MustCompile(`^github\.com/[\w.-]{1,100}/[\w.-]{1,100}$`)

var ghFetch = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "xlang",
	Subsystem: "vfs",
	Name:      "github_fetch_total",
	Help:      "Total number of fetches by GitHubRepoVFS.",
})
var ghFetchFailed = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "xlang",
	Subsystem: "vfs",
	Name:      "github_fetch_failed_total",
	Help:      "Total number of fetches by GitHubRepoVFS that failed.",
})

func init() {
	prometheus.MustRegister(ghFetch)
	prometheus.MustRegister(ghFetchFailed)
}
