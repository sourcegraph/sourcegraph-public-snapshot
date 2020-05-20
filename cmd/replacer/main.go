// Command replacer is an interface to replace and rewrite code. It passes a zipped repo
// to external tools and streams back JSON lines results.
package main

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/replacer/replace"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"

	"github.com/sourcegraph/sourcegraph/internal/tracer"
)

var cacheDir = env.Get("CACHE_DIR", "/tmp", "directory to store cached archives.")
var cacheSizeMB = env.Get("REPLACER_CACHE_SIZE_MB", "100000", "maximum size of the on disk cache in megabytes")

const port = "3185"

func main() {
	env.Lock()
	env.HandleHelpFlag()
	log.SetFlags(0)
	tracer.Init()

	go debugserver.Start()

	var cacheSizeBytes int64
	if i, err := strconv.ParseInt(cacheSizeMB, 10, 64); err != nil {
		log.Fatalf("invalid int %q for REPLACER_CACHE_SIZE_MB: %s", cacheSizeMB, err)
	} else {
		cacheSizeBytes = i * 1000 * 1000
	}

	store := store.Store{
		FetchTar: func(ctx context.Context, repo gitserver.Repo, commit api.CommitID) (io.ReadCloser, error) {
			return gitserver.DefaultClient.Archive(ctx, repo, gitserver.ArchiveOptions{Treeish: string(commit), Format: "tar"})
		},
		Path:              filepath.Join(cacheDir, "replacer-archives"),
		MaxCacheSizeBytes: cacheSizeBytes,
	}
	store.SetMaxConcurrentFetchTar(10)
	store.Start()
	service := &replace.Service{
		Store: &store,
		Log:   log15.Root(),
	}
	handler := ot.Middleware(service)

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, port)
	server := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For cluster liveness and readiness probes
			if r.URL.Path == "/healthz" {
				w.WriteHeader(200)
				_, err := w.Write([]byte("ok"))
				if err != nil {
					log15.Info("Error checking /healthz: " + err.Error())
				}
				return
			}

			handler.ServeHTTP(w, r)
		}),
	}
	go shutdownOnSIGINT(server)

	log15.Info("replacer: listening", "addr", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func shutdownOnSIGINT(s *http.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.Shutdown(ctx)
	if err != nil {
		log.Fatal("graceful server shutdown failed, will exit:", err)
	}
}
