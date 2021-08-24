package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
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
		log15.Error("failed to generate", "error", err)
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

	// Provide ability to clear specific cache keys (i.e. specific types of GraphQL requests.)
	clearCacheKeys := strings.Fields(os.Getenv("CLEAR_CACHE_KEYS"))
	if len(clearCacheKeys) > 0 {
		for _, key := range clearCacheKeys {
			log15.Info("clearing cache key", "key", key)
			if err := g.db.delete(key); err != nil {
				log15.Info("failed to clear cache key", "key", key, "error", err)
			}
		}
	}
	listCacheKeys, _ := strconv.ParseBool(os.Getenv("LIST_CACHE_KEYS"))
	if listCacheKeys {
		keys, err := g.db.keys()
		if err != nil {
			log15.Info("failed to list cache keys", "error", err)
		}
		for _, key := range keys {
			log15.Info("listing cache keys", "key", key)
		}
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

	// Fetch documentation path info for each chosen repo with LSIF indexes.
	var (
		pagesByRepo    = map[string][]string{}
		totalPages     = 0
		totalStars     uint64
		missingAPIDocs = 0
	)
	queried = 0
	for repoName, indexes := range indexedGoRepos {
		queried++
		log15.Info("discovered API docs pages for repo", "n", queried, "of", len(indexedGoRepos))
		totalStars += indexes[0].ProjectRoot.Repository.Stars
		pathInfo, err := g.fetchDocPathInfo(ctx, gqlDocPathInfoVars{RepoName: repoName})
		if pathInfo == nil || (err != nil && strings.Contains(err.Error(), "page not found")) {
			//log15.Error("no API docs pages found", "repo", repoName, "pathInfo==nil", pathInfo == nil, "error", err)
			if err != nil {
				missingAPIDocs++
			}
			continue
		}
		if err != nil {
			return errors.Wrap(err, "fetchDocPathInfo")
		}
		var walk func(node DocumentationPathInfoResult)
		walk = func(node DocumentationPathInfoResult) {
			pagesByRepo[repoName] = append(pagesByRepo[repoName], node.PathID)
			for _, child := range node.Children {
				walk(child)
			}
		}
		walk(*pathInfo)
		totalPages += len(pagesByRepo[repoName])
	}

	log15.Info("found Go API docs pages", "count", totalPages)
	log15.Info("spanning", "repositories", len(indexedGoRepos), "stars", totalStars)
	log15.Info("Go repos missing API docs", "count", missingAPIDocs)
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

func (g *generator) fetchDocPathInfo(ctx context.Context, vars gqlDocPathInfoVars) (*DocumentationPathInfoResult, error) {
	data, err := g.db.request(requestKey{RequestName: "DocPathInfo", Vars: vars}, func() ([]byte, error) {
		return g.gqlClient.requestGraphQL(ctx, "SitemapDocPathInfo", gqlDocPathInfoQuery, vars)
	})
	if err != nil {
		return nil, err
	}
	var resp gqlDocPathInfoResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, errors.Wrap(err, "Unmarshal GraphQL response")
	}
	payload := resp.Data.Repository.Commit.Tree.LSIF.DocumentationPathInfo
	if payload == "" {
		return nil, nil
	}
	var result DocumentationPathInfoResult
	if err := json.Unmarshal([]byte(payload), &result); err != nil {
		return nil, errors.Wrap(err, "Unmarshal DocumentationPathInfoResult")
	}
	return &result, nil
}
