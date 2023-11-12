package repo

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func getDocumentRanks(ctx context.Context, repoName string) (types.RepoPathRanks, error) {
	root, err := url.Parse(internalapi.Client.URL)
	if err != nil {
		return types.RepoPathRanks{}, err
	}
	// TODO: Compute in worker instead.
	u := root.ResolveReference(&url.URL{
		Path: "/.internal/ranks/" + strings.Trim(repoName, "/") + "/documents",
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

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.RepoPathRanks{}, err
	}

	ranks := types.RepoPathRanks{}
	err = json.Unmarshal(b, &ranks)
	if err != nil {
		return types.RepoPathRanks{}, err
	}

	return ranks, nil
}
