package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/schema"
)

func main() {
	baseURL := flag.String("url", "https://bitbucket.sgdev.org/", "Bitbucket Server instance base URL")

	flag.Parse()

	log.SetOutput(ioutil.Discard)

	cli, err := bitbucketserver.NewClientWithConfig(&schema.BitbucketServerConnection{
		Url:    *baseURL,
		Token:  os.Getenv("BITBUCKET_SERVER_TOKEN"),
		Plugin: &schema.BitbucketServerPlugin{Permissions: "enabled"},
	})

	if err != nil {
		log.Fatal(err)
	}

	http.ListenAndServe(":6060", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ids []uint32
			err error
		)

		switch r.URL.Path {
		case "/repos":
			ids, err = listRepoIDs(r.Context(), cli, 15000)
		case "/repoids":
			ids, err = cli.RepoIDs(r.Context(), "read")
		}

		sort.Slice(ids, func(i, j int) bool {
			return ids[i] < ids[j]
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			w.WriteHeader(500)
			return
		}

		json.NewEncoder(w).Encode(ids)
	}))
}

func listRepoIDs(ctx context.Context, c *bitbucketserver.Client, pageSize int) (ids []uint32, err error) {
	t := &bitbucketserver.PageToken{Limit: pageSize}

	for t.HasMore() {
		repos, next, err := c.Repos(ctx, t)
		if err != nil {
			return nil, err
		}

		for _, r := range repos {
			ids = append(ids, uint32(r.ID))
		}

		t = next
	}

	return ids, nil
}
