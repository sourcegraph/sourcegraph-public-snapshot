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

var MockCountGoImporters func(ctx context.Context, repo api.RepoName) (int, error)

// 100 MiB cache, no age-based eviction
var httpClient = &http.Client{Transport: httpcache.NewTransport(lrucache.New(100*1024*1024, 0))}

// CountGoImporters returns the number of Go importers for the repository. This is a special case
// used only on Sourcegraph.com for repository badges.
//
// TODO: The import path is not always the same as the repository name.
func CountGoImporters(ctx context.Context, repo api.RepoName) (int, error) {
	if MockCountGoImporters != nil {
		return MockCountGoImporters(ctx, repo)
	}
	// Assumes the import path is the same as the repo name - not always true!
	response, err := httpClient.Get("https://api.godoc.org/importers/" + string(repo))
	if err != nil {
		return 0, err
	}
	var result struct {
		Results []struct {
			Path string
		}
	}
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
