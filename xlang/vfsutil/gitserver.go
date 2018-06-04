package vfsutil

import (
	"context"
	"io"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
)

// NewGitServer returns a VFS to repo at commit. It is backed by an archive
// fetched from gitserver.
func NewGitServer(repo api.RepoURI, commit api.CommitID) *ArchiveFS {
	fetch := func(ctx context.Context) (ar *archiveReader, err error) {
		span, ctx := opentracing.StartSpanFromContext(ctx, "Archive Fetch")
		ext.Component.Set(span, "gitserver")
		span.SetTag("repo", repo)
		span.SetTag("commit", commit)
		defer func() {
			if err != nil {
				ext.Error.Set(span, true)
				span.SetTag("err", err)
			}
			span.Finish()
		}()

		if strings.HasPrefix(string(commit), "-") {
			return nil, errors.New("invalid git revision spec (begins with '-')")
		}

		ff, err := cachedFetch(ctx, "gitserver", string(repo)+"@"+string(commit), func(ctx context.Context) (io.ReadCloser, error) {
			gitserverFetchTotal.Inc()
			return gitserverFetch(ctx, repo, commit)
		})
		if err != nil {
			gitserverFetchFailedTotal.Inc()
			return nil, err
		}
		f := ff.File

		zr, err := zipNewFileReader(f)
		if err != nil {
			f.Close()
			return nil, err
		}

		return &archiveReader{
			Reader:  zr,
			Closer:  f,
			Evicter: ff,
		}, nil
	}
	return &ArchiveFS{fetch: fetch}
}

// gitserverFetch returns a reader of a zip archive of repo at commit.
func gitserverFetch(ctx context.Context, repo api.RepoURI, commit api.CommitID) (r io.ReadCloser, err error) {
	// Compression level of 0 (no compression) seems to perform the
	// best overall on fast network links, but this has not been tuned
	// thoroughly.
	cmd := gitserver.DefaultClient.Command("git", "archive", "--format=zip", "-0", string(commit))
	cmd.Repo = gitserver.Repo{Name: repo}
	r, err = gitserver.StdoutReader(ctx, cmd)
	if err != nil {
		return nil, err
	}
	return r, nil
}

var gitserverFetchTotal = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "xlang",
	Subsystem: "vfs",
	Name:      "gitserver_fetch_total",
	Help:      "Total number of fetches to GitServer.",
})
var gitserverFetchFailedTotal = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "xlang",
	Subsystem: "vfs",
	Name:      "gitserver_fetch_failed_total",
	Help:      "Total number of fetches to GitServer that failed.",
})

func init() {
	prometheus.MustRegister(gitserverFetchTotal)
	prometheus.MustRegister(gitserverFetchFailedTotal)
}
