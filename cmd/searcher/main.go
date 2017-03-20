// search is a simple service which exposes an API to text search a repo at
// a specific commit. See the searcher package for more information.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/searcher/search"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/gitcmd"
)

var profBindAddr = env.Get("SRC_PROF_HTTP", "", "net/http/pprof http bind address.")

func main() {
	env.Lock()
	env.HandleHelpFlag()
	log.SetFlags(0)
	traceutil.InitTracer()
	gitserver.DefaultClient.NoCreds = true

	if profBindAddr != "" {
		go debugserver.Start(profBindAddr)
	}

	service := &search.Service{
		Store: &search.Store{
			FetchZip: fetchZip,
			Path:     "/tmp/searcher-archive-store",
		},
	}

	addr := ":3181"
	server := &http.Server{Addr: addr, Handler: service}
	go shutdownOnSIGINT(server)

	log.Println("listening on :3181")
	_ = server.ListenAndServe()
	// TODO enable when we have go 1.8 on CI
	// if err != http.ErrServerClosed {
	// 	log.Fatal(err)
	// }
}

func shutdownOnSIGINT(s *http.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	os.Exit(0)
	// TODO enable when we have go 1.8 on CI. Remove os.Exit
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	// err := s.Shutdown(ctx)
	// if err != nil {
	// 	log.Fatal("graceful server shutdown failed, will exit:", err)
	// }
}

func fetchZip(ctx context.Context, repo, commit string) ([]byte, error) {
	r := gitcmd.Open(&sourcegraph.Repo{URI: repo})
	b, err := r.Archive(ctx, vcs.CommitID(commit))
	// Guess if user error
	if err != nil && (strings.Contains(err.Error(), "invalid git revision") || vcs.IsRepoNotExist(err) || err == vcs.ErrRevisionNotFound) {
		return nil, badRequestError{err.Error()}
	}
	return b, err
}

type badRequestError struct{ msg string }

func (e badRequestError) Error() string    { return e.msg }
func (e badRequestError) BadRequest() bool { return true }
