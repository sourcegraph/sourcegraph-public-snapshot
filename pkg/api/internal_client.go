package sourcegraph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
)

var frontendInternal = env.Get("SRC_FRONTEND_INTERNAL", "sourcegraph-frontend-internal", "HTTP address for internal frontend HTTP API.")

type internalClient struct {
	// URL is the root to the internal API frontend server.
	URL string
}

var InternalClient = &internalClient{URL: "http://" + frontendInternal}

func (c *internalClient) DefsRefreshIndex(ctx context.Context, uri, revision string) error {
	req, err := json.Marshal(&DefsRefreshIndexRequest{
		URI:      uri,
		Revision: revision,
	})
	if err != nil {
		return err
	}
	resp, err := c.post("defs/refresh-index", req)
	if err != nil {
		return err
	}
	var inv inventory.Inventory
	err = json.NewDecoder(resp.Body).Decode(&inv)
	if err != nil {
		return err
	}
	return nil
}

func (c *internalClient) ReposCreateIfNotExists(ctx context.Context, uri, description string, fork, private bool) (*Repo, error) {
	req, err := json.Marshal(RepoCreateOrUpdateRequest{
		URI:         uri,
		Description: description,
		Fork:        fork,
		Private:     private,
	})
	if err != nil {
		return nil, err
	}
	resp, err := c.post("repos/create-if-not-exists", req)
	if err != nil {
		return nil, err
	}
	var repo Repo
	err = json.NewDecoder(resp.Body).Decode(&repo)
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

func (c *internalClient) ReposUnindexedDependencies(ctx context.Context, repoID int32, lang string) ([]*DependencyReference, error) {
	req, err := json.Marshal(RepoUnindexedDependenciesRequest{
		RepoID:   repoID,
		Language: lang,
	})
	if err != nil {
		return nil, err
	}
	var unfetchedDeps []*DependencyReference
	resp, err := c.post("repos/unindexed-dependencies", req)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(resp.Body).Decode(&unfetchedDeps)
	if err != nil {
		return nil, err
	}
	return unfetchedDeps, nil
}

func (c *internalClient) ReposUpdateIndex(ctx context.Context, repoID int32, revision, lang string) error {
	req, err := json.Marshal(RepoUpdateIndexRequest{
		RepoID:   repoID,
		Revision: revision,
		Language: lang,
	})
	if err != nil {
		return err
	}
	_, err = c.post("repos/update-index", req)
	if err != nil {
		return err
	}
	return nil
}

func (c *internalClient) ReposGetByURI(ctx context.Context, uri string) (*Repo, error) {
	resp, err := c.get("repos/" + uri)
	if err != nil {
		return nil, err
	}
	var repo Repo
	err = json.NewDecoder(resp.Body).Decode(&repo)
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

func (c *internalClient) ReposGetInventoryUncached(ctx context.Context, revSpec RepoRevSpec) (*inventory.Inventory, error) {
	req, err := json.Marshal(revSpec)
	if err != nil {
		return nil, err
	}
	resp, err := c.post("repos/inventory-uncached", req)
	if err != nil {
		return nil, err
	}
	var inv inventory.Inventory
	err = json.NewDecoder(resp.Body).Decode(&inv)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (c *internalClient) GitoliteUpdateRepos(ctx context.Context) error {
	_, err := c.post("gitolite/update-repos", []byte{})
	if err != nil {
		return err
	}
	return nil
}

func (c *internalClient) PhabricatorRepoCreate(ctx context.Context, uri, callsign, url string) error {
	req, err := json.Marshal(PhabricatorRepoCreateRequest{
		URI:      uri,
		Callsign: callsign,
		URL:      url,
	})
	if err != nil {
		return err
	}
	_, err = c.post("phabricator/repo-create", req)
	if err != nil {
		return err
	}
	return nil
}

func (c *internalClient) post(route string, body []byte) (*http.Response, error) {
	resp, err := http.Post(c.URL+"/.internal/"+route, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	if err := checkAPIResponse(resp); err != nil {
		return nil, err
	}
	return resp, err
}

func (c *internalClient) get(route string) (*http.Response, error) {
	resp, err := http.Get(c.URL + "/.internal/" + route)
	if err != nil {
		return nil, err
	}
	if err := checkAPIResponse(resp); err != nil {
		return nil, err
	}
	return resp, err
}

func checkAPIResponse(resp *http.Response) error {
	if 200 > resp.StatusCode || resp.StatusCode > 299 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		b := buf.Bytes()
		errString := string(b)
		if errString != "" {
			return fmt.Errorf("sourcegraph API response status %d: %s", resp.StatusCode, errString)
		} else {
			return fmt.Errorf("sourcegraph API response status %d", resp.StatusCode)
		}
	}
	return nil
}
