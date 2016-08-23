package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/lang"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
)

var (
	httpAddr = flag.String("http", ":4142", "HTTP address to listen on")
	lpAddr   = flag.String("lp", "http://127.0.0.1:4143", "Java LP server address to proxy requests to")
	profbind = flag.String("prof-http", ":6060", "net/http/pprof http bind address")
	workDir  = flag.String("workspace", "$SGPATH/workspace/java", "where to create workspace directories")
)

func prepareRepo(update bool, workspace, repo, commit string) error {

	_, cloneURI := langp.ResolveRepoAlias(repo)

	// Clone the repository.
	return langp.Clone(update, cloneURI, workspace, commit)
}

func main() {
	flag.Parse()

	if *profbind != "" {
		go debugserver.Start(*profbind)
	}
	langp.InitMetrics("java")

	lang.PrepareKeys()

	workDir, err := langp.ExpandSGPath(*workDir)
	if err != nil {
		log.Fatal(err)
	}

	client, err := langp.NewClient(*lpAddr)
	if err != nil {
		log.Fatal(err)
	}

	handler, err := langp.NewProxy(client,
		&langp.Preparer{
			WorkDir:     workDir,
			PrepareRepo: prepareRepo,
			PrepareDeps: func(update bool, workspace, repo, commit string) error {
				return client.Prepare(context.Background(), &langp.RepoRev{
					Repo:   repo,
					Commit: commit,
				})
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Translating HTTP", *httpAddr, "to Java LP", *lpAddr)
	http.Handle("/", handler)
	http.ListenAndServe(*httpAddr, nil)
}
