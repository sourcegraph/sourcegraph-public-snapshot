package vfsutil

import (
	"context"
	"io"
	"os"
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
func NewGitServer(repo api.RepoName, commit api.CommitID) *ArchiveFS {
	fetch := func(ctx context.Context) (ar *archiveReader, err error) {
		f, evictor, err := GitServerFetchArchive(ctx, ArchiveOpts{Repo: repo, Commit: commit})
		if err != nil {
			return nil, err
		}

		zr, err := zipNewFileReader(f)
		if err != nil {
			f.Close()
			return nil, err
		}

		return &archiveReader{
			Reader:  zr,
			Closer:  f,
			Evicter: evictor,
		}, nil
	}
	return &ArchiveFS{fetch: fetch}
}

// ArchiveOpts describes options for fetching a repository archive.
type ArchiveOpts struct {
	// Repo is the repository whose contents should be fetched.
	Repo api.RepoName

	// Commit is the commit whose contents should be fetched.
	Commit api.CommitID
}

// GitServerFetchArchive fetches an archive of a repositories contents from gitserver.
func GitServerFetchArchive(ctx context.Context, opts ArchiveOpts) (archive *os.File, cacheEvicter Evicter, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Archive Fetch")
	ext.Component.Set(span, "gitserver")
	span.SetTag("repo", opts.Repo)
	span.SetTag("commit", opts.Commit)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err)
		}
		span.Finish()
	}()

	if strings.HasPrefix(string(opts.Commit), "-") {
		return nil, nil, errors.New("invalid git revision spec (begins with '-')")
	}

	ff, err := cachedFetch(ctx, "gitserver", string(opts.Repo)+"@"+string(opts.Commit), func(ctx context.Context) (io.ReadCloser, error) {
		gitserverFetchTotal.Inc()

		// Compression level of 0 (no compression) seems to perform the
		// best overall on fast network links, but this has not been tuned
		// thoroughly.
		cmd := gitserver.DefaultClient.Command("git", "archive", "--format=zip", "-0", string(opts.Commit))
		cmd.Repo = gitserver.Repo{Name: opts.Repo}
		r, err := gitserver.StdoutReader(ctx, cmd)
		if err != nil {
			return nil, err
		}
		return r, nil
	})
	if err != nil {
		gitserverFetchFailedTotal.Inc()
		return nil, nil, err
	}
	return ff.File, ff, nil
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
