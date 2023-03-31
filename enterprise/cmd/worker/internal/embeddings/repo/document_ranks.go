package repo

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func getDocumentRanks(ctx context.Context, repoName api.RepoName) (types.RepoPathRanks, error) {
	root, err := url.Parse(internalapi.Client.URL)
	if err != nil {
		return types.RepoPathRanks{}, err
	}
	u := root.ResolveReference(&url.URL{
		Path: "/.internal/ranks/" + strings.Trim(string(repoName), "/") + "/documents",
	})

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return types.RepoPathRanks{}, err
	}

	resp, err := httpcli.InternalDoer.Do(req)
	if err != nil {
		return types.RepoPathRanks{}, err
	}

	if resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
		_ = resp.Body.Close()
		if err != nil {
			return types.RepoPathRanks{}, err
		}
		return types.RepoPathRanks{}, &url.Error{
			Op:  "Get",
			URL: u.String(),
			Err: errors.Errorf("%s: %s", resp.Status, string(b)),
		}
	}

	var ranks types.RepoPathRanks
	return ranks, json.NewDecoder(resp.Body).Decode(&ranks)
}
