package backend

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/die-net/lrucache"
	"github.com/gregjones/httpcache"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// TODO!(sqs): This file contains backcompat stubs for definitions that were removed in the
// migration to using Sourcegraph extensions for language support.

var MockBackcompatBackendDefsTotalRefs func(ctx context.Context, repo api.RepoName) (int, error)

// 100 MiB cache, no age-based eviction
var httpClient = &http.Client{Transport: httpcache.NewTransport(lrucache.New(100*1024*1024, 0))}

type GDDOResponse struct {
	Results []GDDOResult `json:"results"`
}

type GDDOResult struct {
	Path string `json:"path"`
}

func BackcompatBackendDefsTotalRefs(ctx context.Context, repo api.RepoName) (int, error) {
	if MockBackcompatBackendDefsTotalRefs != nil {
		return MockBackcompatBackendDefsTotalRefs(ctx, repo)
	}
	// Assumes the import path is the same as the repo name - not always true!
	response, err := httpClient.Get("https://api.godoc.org/importers/" + string(repo))
	if err != nil {
		return 0, err
	}
	var result *GDDOResponse
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return 0, err
	}
	return len(result.Results), nil
}
