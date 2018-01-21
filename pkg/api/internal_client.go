package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
)

var frontendInternal = env.Get("SRC_FRONTEND_INTERNAL", "sourcegraph-frontend-internal", "HTTP address for internal frontend HTTP API.")

type internalClient struct {
	// URL is the root to the internal API frontend server.
	URL string
}

var InternalClient = &internalClient{URL: "http://" + frontendInternal}

// RetryPingUntilAvailable retries a noop request to the internal API until it is able to reach
// the endpoint, indicating that the endpoint is available.
func (c *internalClient) RetryPingUntilAvailable(ctx context.Context) error {
	ping := func(ctx context.Context) error {
		resp, err := http.Get(c.URL + "/.internal/ping")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("ping: bad HTTP response status %d", resp.StatusCode)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if want := "pong"; string(body) != want {
			const max = 15
			if len(body) > max {
				body = body[:max]
			}
			return fmt.Errorf("ping: bad HTTP response body %q (want %q)", body, want)
		}
		return nil
	}

	var lastErr error
	for {
		err := ping(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return fmt.Errorf("frontend API not reachable: %s (last error: %v)", err, lastErr)
			}

			// Keep trying.
			lastErr = err
			time.Sleep(250 * time.Millisecond)
			continue
		}
		break
	}
	return nil
}

func (c *internalClient) DefsRefreshIndex(ctx context.Context, uri RepoURI, commitID CommitID) error {
	req, err := json.Marshal(&DefsRefreshIndexRequest{
		RepoURI:  uri,
		CommitID: commitID,
	})
	if err != nil {
		return err
	}
	_, err = c.post("defs/refresh-index", req)
	return err
}

func (c *internalClient) PkgsRefreshIndex(ctx context.Context, uri RepoURI, commitID CommitID) error {
	req, err := json.Marshal(&PkgsRefreshIndexRequest{
		RepoURI:  uri,
		CommitID: commitID,
	})
	if err != nil {
		return err
	}
	_, err = c.post("pkgs/refresh-index", req)
	return err
}

func (c *internalClient) ReposCreateIfNotExists(ctx context.Context, uri RepoURI, description string, fork, enabled bool) (*Repo, error) {
	req, err := json.Marshal(RepoCreateOrUpdateRequest{
		RepoURI:     uri,
		Description: description,
		Fork:        fork,
		Enabled:     enabled,
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

func (c *internalClient) ReposUnindexedDependencies(ctx context.Context, repo RepoID, lang string) ([]*DependencyReference, error) {
	req, err := json.Marshal(RepoUnindexedDependenciesRequest{
		RepoID:   repo,
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

func (c *internalClient) ReposUpdateIndex(ctx context.Context, repo RepoID, commitID CommitID, lang string) error {
	req, err := json.Marshal(RepoUpdateIndexRequest{
		RepoID:   repo,
		CommitID: commitID,
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

func (c *internalClient) ReposGetByURI(ctx context.Context, uri RepoURI) (*Repo, error) {
	resp, err := c.get("repos/" + string(uri))
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

func (c *internalClient) ReposGetInventoryUncached(ctx context.Context, repo RepoID, commitID CommitID) (*inventory.Inventory, error) {
	req, err := json.Marshal(ReposGetInventoryUncachedRequest{Repo: repo, CommitID: commitID})
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

func (c *internalClient) PhabricatorRepoCreate(ctx context.Context, uri RepoURI, callsign, url string) error {
	req, err := json.Marshal(PhabricatorRepoCreateRequest{
		RepoURI:  uri,
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
		}
		return fmt.Errorf("sourcegraph API response status %d", resp.StatusCode)
	}
	return nil
}
