package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
)

func main() {
	gen := &generator{
		graphQLURL:    "https://sourcegraph.com/.api/graphql",
		token:         os.Getenv("SRC_ACCESS_TOKEN"),
		outDir:        "sitemap/",
		queryDatabase: "sitemap_query.db",
	}
	if err := gen.generate(context.Background()); err != nil {
		log15.Error("failed to generate", err)
		os.Exit(-1)
	}
	log15.Info("generated sitemap", "out", gen.outDir)
}

type generator struct {
	graphQLURL    string
	token         string
	outDir        string
	queryDatabase string

	db        *queryDatabase
	gqlClient *graphQLClient
}

// generate generates the sitemap files to the specified directory.
func (g *generator) generate(ctx context.Context) error {
	if err := os.MkdirAll(g.outDir, 0700); err != nil {
		return errors.Wrap(err, "MkdirAll")
	}
	if err := os.MkdirAll(filepath.Dir(g.queryDatabase), 0700); err != nil {
		return errors.Wrap(err, "MkdirAll")
	}

	// The query database caches our GraphQL queries across multiple runs, as well as allows us to
	// update the sitemap to include new repositories / pages without re-querying everything which
	// would be very expensive. It's a simple on-disk key-vaue store (bbolt).
	var err error
	g.db, err = openQueryDatabase(g.queryDatabase)
	if err != nil {
		return errors.Wrap(err, "openQueryDatabase")
	}
	defer g.db.close()

	g.gqlClient = &graphQLClient{
		URL:   g.graphQLURL,
		Token: g.token,
	}

	// Build a set of Go repos that have LSIF indexes.
	indexedGoRepos := map[string][]gqlLSIFIndex{}
	queried := 0
	if err := g.eachLsifIndex(ctx, func(each gqlLSIFIndex, total uint64) error {
		queried++
		if queried%1000 == 0 {
			log15.Info("discovered LSIF indexes", "n", queried, "of", total)
		}
		if strings.Contains(each.InputIndexer, "lsif-go") {
			repoName := each.ProjectRoot.Repository.Name
			indexedGoRepos[repoName] = append(indexedGoRepos[repoName], each)
		}
		return nil
	}); err != nil {
		return err
	}

	log15.Info("found indexed Go repositories", "count", len(indexedGoRepos))
	return nil
}

func (g *generator) eachLsifIndex(ctx context.Context, each func(index gqlLSIFIndex, total uint64) error) error {
	var (
		hasNextPage = true
		cursor      *string
	)
	for hasNextPage {
		retries := 0
	retry:
		lsifIndexes, err := g.fetchLsifIndexes(ctx, gqlLSIFIndexesVars{
			State: strPtr("COMPLETED"),
			First: intPtr(5000),
			After: cursor,
		})
		if err != nil {
			retries++
			if maxRetries := 10; retries < maxRetries {
				log15.Error("error listing LSIF indexes", "retry", retries, "of", maxRetries)
				goto retry
			}
			return err
		}

		for _, index := range lsifIndexes.Data.LsifIndexes.Nodes {
			if err := each(index, lsifIndexes.Data.LsifIndexes.TotalCount); err != nil {
				return err
			}
		}
		hasNextPage = lsifIndexes.Data.LsifIndexes.PageInfo.HasNextPage
		cursor = lsifIndexes.Data.LsifIndexes.PageInfo.EndCursor
	}
	return nil
}

func (g *generator) fetchLsifIndexes(ctx context.Context, vars gqlLSIFIndexesVars) (*gqlLSIFIndexesResponse, error) {
	data, err := g.db.request(requestKey{RequestName: "LsifIndexes", Vars: vars}, func() ([]byte, error) {
		return g.gqlClient.requestGraphQL(ctx, "SitemapLsifIndexes", gqlLSIFIndexesQuery, vars)
	})
	if err != nil {
		return nil, err
	}
	var resp gqlLSIFIndexesResponse
	return &resp, json.Unmarshal(data, &resp)
}
