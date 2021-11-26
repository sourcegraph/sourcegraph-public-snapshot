package resolvers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"io/ioutil"
	"os"
	pathpkg "path"
	"sync"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func (r *componentResolver) Usage(ctx context.Context, args *gql.ComponentUsageArgs) (gql.ComponentUsageResolver, error) {
	if len(r.component.UsagePatterns) == 0 {
		return nil, nil
	}

	var queries []string
	for _, p := range r.component.UsagePatterns {
		queries = append(queries, p.Query)
	}

	search, err := gql.NewSearchImplementer(ctx, r.db, &gql.SearchArgs{
		Version: "V2",
		Query:   joinQueryParts(queries),
	})
	if err != nil {
		return nil, err
	}
	return &componentUsageResolver{
		search:    search,
		component: r,
		db:        r.db,
	}, nil
}

type componentUsageResolver struct {
	search    gql.SearchImplementer
	component *componentResolver
	db        database.DB

	resultsOnce sync.Once
	results     *gql.SearchResultsResolver
	resultsErr  error
}

func (r *componentUsageResolver) cachedResults(ctx context.Context) (*gql.SearchResultsResolver, error) {
	type cacheEntry struct {
		SearchResults []result.Match
	}
	cachePath := func(search gql.SearchImplementer) string {
		const dir = "/tmp/sqs-wip-cache/componentUsage"
		_ = os.MkdirAll(dir, 0700)

		h := sha256.Sum256([]byte(r.search.Inputs().OriginalQuery))
		name := hex.EncodeToString(h[:])

		return pathpkg.Join(dir, name)
	}
	get := func(search gql.SearchImplementer) (cacheEntry, bool) {
		b, err := ioutil.ReadFile(cachePath(search))
		if os.IsNotExist(err) {
			return cacheEntry{}, false
		}
		if err != nil {
			panic(err)
		}
		var v cacheEntry
		if err := gob.NewDecoder(bytes.NewReader(b)).Decode(&v); err != nil {
			panic(err)
		}
		return v, true
	}
	set := func(search gql.SearchImplementer, data cacheEntry) {
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(data); err != nil {
			panic(err)
		}
		if err := ioutil.WriteFile(cachePath(search), buf.Bytes(), 0600); err != nil {
			panic(err)
		}
	}
	resultsCached := func(ctx context.Context) (*gql.SearchResultsResolver, error) {
		v, ok := get(r.search)
		if ok {
			// log.Println("HIT")
			return &gql.SearchResultsResolver{
				SearchResults: &gql.SearchResults{Matches: v.SearchResults},
			}, nil
		}
		// log.Println("MISS")

		results, err := r.search.Results(ctx)
		if err == nil {
			set(r.search, cacheEntry{SearchResults: results.SearchResults.Matches})
		}
		return results, err
	}

	r.resultsOnce.Do(func() {
		r.results, r.resultsErr = resultsCached(ctx)
	})
	return r.results, r.resultsErr
}

func init() {
	gob.Register(&result.FileMatch{})
}
