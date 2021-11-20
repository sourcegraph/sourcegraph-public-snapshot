package parser

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"path"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

// maxFileSize is the limit on file size in bytes. Only files smaller than this are processed.
const maxFileSize = 1 << 19 // 512KB

// 32*1024 is the same size used by io.Copy
const BufferSize = 32 * 1024

type parseRequest struct {
	path string
	data []byte
}

type parseRequestOrError struct {
	parseRequest parseRequest
	err          error
}

func fetchRepositoryArchive(ctx context.Context, gitserverClient GitserverClient, fetchSem chan int, repo api.RepoName, commitID api.CommitID, paths []string) <-chan parseRequestOrError {
	fetchQueueSize.Inc()
	fetchSem <- 1 // acquire concurrent fetches semaphore
	fetchQueueSize.Dec()

	fetching.Inc()
	span, ctx := ot.StartSpanFromContext(ctx, "Store.fetch")
	ext.Component.Set(span, "store")
	span.SetTag("repo", repo)
	span.SetTag("commit", commitID)

	requestCh := make(chan parseRequestOrError)

	go func() (err error) {
		defer close(requestCh)
		defer func() { <-fetchSem }() // release concurrent fetches semaphore
		defer fetching.Dec()

		defer func() {
			if err != nil {
				ext.Error.Set(span, true)
				span.SetTag("err", err.Error())
				fetchFailed.Inc()
			}

			span.Finish()
		}()

		defer func() {
			if err != nil {
				requestCh <- parseRequestOrError{err: err}
			}
		}()

		r, err := gitserverClient.FetchTar(ctx, repo, commitID, paths)
		if err != nil {
			return err
		}
		defer r.Close()

		buf := make([]byte, BufferSize)

		tr := tar.NewReader(r)
		for {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			hdr, err := tr.Next()
			if err != nil {
				if err == io.EOF {
					return nil
				}

				return err
			}

			if path.Ext(hdr.Name) == ".json" {
				continue
			}

			// We only care about files
			if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeRegA {
				continue
			}

			// We do not search large files
			if hdr.Size > maxFileSize {
				continue
			}

			n, err := tr.Read(buf)
			if err != nil && err != io.EOF {
				return err
			}
			if n == 0 {
				continue
			}

			// Heuristic: Assume file is binary if first 256 bytes contain a 0x00. Best effort, so ignore err.
			if bytes.IndexByte(buf[:n], 0x00) >= 0 {
				continue
			}

			// Read the file's contents.
			data := make([]byte, int(hdr.Size))
			copy(data, buf[:n])
			if n < int(hdr.Size) {
				if _, err := io.ReadFull(tr, data[n:]); err != nil {
					return err
				}
			}
			requestCh <- parseRequestOrError{parseRequest: parseRequest{path: hdr.Name, data: data}}
		}
	}()

	return requestCh
}

var (
	fetching = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "symbols_store_fetching",
		Help: "The number of fetches currently running.",
	})
	fetchQueueSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "symbols_store_fetch_queue_size",
		Help: "The number of fetch jobs enqueued.",
	})
	fetchFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "symbols_store_fetch_failed",
		Help: "The total number of archive fetches that failed.",
	})
)
