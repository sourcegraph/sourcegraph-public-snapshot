package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
)

var (
	httpAddr = flag.String("http", ":4142", "HTTP address to listen on")
	lpAddr   = flag.String("lp", "http://127.0.0.1:4143", "Java LP server address to proxy requests to")
	profbind = flag.String("prof-http", ":6060", "net/http/pprof http bind address")
	workDir  = flag.String("workspace", "$SGPATH/workspace/java", "where to create workspace directories")
)

func prepareRepo(ctx context.Context, update bool, workspace, repo, commit string) error {
	cloneURI := langp.RepoCloneURL(ctx, repo)
	repo = langp.ResolveRepoAlias(repo)

	// Clone the repository.
	return langp.Clone(ctx, update, cloneURI, workspace, commit)
}

func main() {
	flag.Parse()

	if *profbind != "" {
		go debugserver.Start(*profbind)
	}
	langp.InitMetrics("java")

	workDir, err := langp.ExpandSGPath(*workDir)
	if err != nil {
		log.Fatal(err)
	}

	client, err := langp.NewClient(map[string][]string{"Java": []string{*lpAddr}})
	if err != nil {
		log.Fatal(err)
	}

	srcEndpoint := os.Getenv("SRC_ENDPOINT")
	if srcEndpoint == "" {
		srcEndpoint = "http://localhost:3080"
	}

	handler, err := langp.NewProxy(client,
		langp.NewPreparer(&langp.PreparerOpts{
			SrcEndpoint: srcEndpoint,
			WorkDir:     workDir,
			PrepareRepo: prepareRepo,
			PrepareDeps: func(ctx context.Context, update bool, workspace, repo, commit string) error {
				return client.Prepare(context.Background(), &langp.RepoRev{
					Repo:   repo,
					Commit: commit,
				})
			},
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Translating HTTP", *httpAddr, "to Java LP", *lpAddr)
	http.Handle("/", handler)
	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}
