package resolvers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	pathpkg "path"
	"regexp"
	"strings"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func (r *componentResolver) Labels(ctx context.Context) ([]gql.ComponentLabelResolver, error) {
	data, err := r.labelData(ctx)
	if err != nil {
		return nil, err
	}

	rs := make([]gql.ComponentLabelResolver, len(data))
	for i, data := range data {
		rs[i] = &componentLabelResolver{data: data, db: r.db}
	}
	return rs, nil
}

func (r *componentResolver) labelData(ctx context.Context) ([]componentLabelData, error) {
	computedLabels := catalog.ComputedLabels()
	var data []componentLabelData
	for _, cl := range computedLabels {
		var values []string
		for value, query := range cl.ValueQueries {
			isNegated := strings.HasPrefix(value, "!")
			if isNegated {
				value = strings.TrimPrefix(value, "!")
			}

			ok, err := r.checkIfComponentMatchesLabelValueQuery(ctx, query)
			if err != nil {
				return nil, err
			}
			if ok != isNegated {
				values = append(values, value)
			}
		}
		if len(values) > 0 {
			data = append(data, componentLabelData{
				Key:    cl.Key,
				Values: values,
			})
		}
	}
	return data, nil
}

func (r *componentResolver) checkIfComponentMatchesLabelValueQuery(ctx context.Context, labelValueQuery string) (ok bool, err error) {
	type cacheEntry bool
	cachePath := func(componentName, labelValueQuery string) string {
		const dir = "/tmp/sqs-wip-cache/labelData"
		_ = os.MkdirAll(dir, 0700)

		b, err := json.Marshal([]interface{}{componentName, labelValueQuery})
		if err != nil {
			panic(err)
		}
		h := sha256.Sum256(b)
		name := hex.EncodeToString(h[:])

		return pathpkg.Join(dir, name)
	}
	get := func(path string) (cacheEntry, bool) {
		b, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			return cacheEntry(false), false
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
	set := func(path string, data cacheEntry) {
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(data); err != nil {
			panic(err)
		}
		if err := ioutil.WriteFile(path, buf.Bytes(), 0600); err != nil {
			panic(err)
		}
	}

	doUncached := func(ctx context.Context, labelValueQuery string) (ok bool, err error) {
		slocs, err := r.sourceLocations(ctx)
		if err != nil {
			return false, err
		}

		var queryParts []string
		for sloc, paths := range groupSourceLocationsByRepo(slocs) {
			queryParts = append(queryParts, fmt.Sprintf("repo:%s file:%s %s count:1",
				"^"+regexp.QuoteMeta(string(sloc.repoName))+"$",
				joinPathPrefixRegexps(paths),
				labelValueQuery,
			))
		}

		query := joinQueryParts(queryParts)
		search, err := gql.NewSearchImplementer(ctx, r.db, &gql.SearchArgs{
			Version: "V2",
			Query:   query,
		})
		if err != nil {
			return false, err
		}
		results, err := search.Results(ctx)
		if err != nil {
			return false, err
		}
		return results.MatchCount() > 0, nil
	}

	p := cachePath(r.component.Name, labelValueQuery)
	v, ok := get(p)
	if ok {
		return bool(v), nil
	}

	ok, err = doUncached(ctx, labelValueQuery)
	if err == nil {
		set(p, cacheEntry(ok))
	}
	return ok, err
}

type componentLabelData struct {
	Key    string
	Values []string
}

type componentLabelResolver struct {
	data componentLabelData
	db   database.DB
}

func (r *componentLabelResolver) Key() string      { return r.data.Key }
func (r *componentLabelResolver) Values() []string { return r.data.Values }
