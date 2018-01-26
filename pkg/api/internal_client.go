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

// A ConfigurationSubject is something that can have settings. A subject with no
// fields set represents the global site settings subject.
type ConfigurationSubject struct {
	Site *string // the site's ID
	Org  *int32  // the org's ID
	User *int32  // the user's ID
}

func (s ConfigurationSubject) String() string {
	switch {
	case s.Site != nil:
		return fmt.Sprintf("site %q", *s.Site)
	case s.Org != nil:
		return fmt.Sprintf("org %d", *s.Org)
	case s.User != nil:
		return fmt.Sprintf("user %d", *s.User)
	default:
		return "unknown configuration subject"
	}
}

type SavedQueryIDSpec struct {
	Subject ConfigurationSubject
	Key     string
}

// ConfigSavedQuery is the JSON shape of a saved query entry in the JSON configuration
// (i.e., an entry in the {"search.savedQueries": [...]} array).
type ConfigSavedQuery struct {
	Key                 string   `json:"key,omitempty"`
	Description         string   `json:"description"`
	Query               string   `json:"query"`
	ScopeQuery          string   `json:",omitempty"`
	ShowOnHomepage      bool     `json:"showOnHomepage"`
	Notify              bool     `json:"notify,omitempty"`
	NotifySlack         bool     `json:"notifySlack,omitempty"`
	NotifyUsers         []string `json:"notifyUsers,omitempty"`
	NotifyOrganizations []string `json:"notifyOrganizations,omitempty"`
}

func (sq ConfigSavedQuery) Equals(other ConfigSavedQuery) bool {
	a, _ := json.Marshal(sq)
	b, _ := json.Marshal(other)
	return bytes.Equal(a, b)
}

// PartialConfigSavedQueries is the JSON configuration shape, including only the
// search.savedQueries section.
type PartialConfigSavedQueries struct {
	SavedQueries []ConfigSavedQuery `json:"search.savedQueries"`
}

// SavedQuerySpecAndConfig represents a saved query configuration its unique ID.
type SavedQuerySpecAndConfig struct {
	Spec   SavedQueryIDSpec
	Config ConfigSavedQuery
}

// SavedQueriesListAll lists all saved queries, from every user, org, etc.
func (c *internalClient) SavedQueriesListAll(ctx context.Context) (map[SavedQueryIDSpec]ConfigSavedQuery, error) {
	resp, err := c.post("saved-queries/list-all", nil)
	if err != nil {
		return nil, err
	}
	var result []SavedQuerySpecAndConfig
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	m := map[SavedQueryIDSpec]ConfigSavedQuery{}
	for _, r := range result {
		m[r.Spec] = r.Config
	}
	return m, nil
}

// SavedQueryInfo represents information about a saved query that was executed.
type SavedQueryInfo struct {
	// Query is the search query in question.
	Query string

	// LastExecuted is the timestamp of the last time that the search query was
	// executed.
	LastExecuted time.Time

	// LatestResult is the timestamp of the latest-known result for the search
	// query. Therefore, searching `after:<LatestResult>` will return the new
	// search results not yet seen.
	LatestResult time.Time

	// ExecDuration is the amount of time it took for the query to execute.
	ExecDuration time.Duration
}

// SavedQueriesGetInfo gets the info from the DB for the given saved query. nil
// is returned if there is no existing info for the saved query.
func (c *internalClient) SavedQueriesGetInfo(ctx context.Context, query string) (*SavedQueryInfo, error) {
	args, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	resp, err := c.post("saved-queries/get-info", args)
	if err != nil {
		return nil, err
	}
	var result *SavedQueryInfo
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SavedQueriesSetInfo sets the info in the DB for the given query.
func (c *internalClient) SavedQueriesSetInfo(ctx context.Context, info *SavedQueryInfo) error {
	args, err := json.Marshal(info)
	if err != nil {
		return err
	}
	_, err = c.post("saved-queries/set-info", args)
	return err
}

func (c *internalClient) SavedQueriesDeleteInfo(ctx context.Context, query string) error {
	args, err := json.Marshal(query)
	if err != nil {
		return err
	}
	_, err = c.post("saved-queries/delete-info", args)
	return err
}

func (c *internalClient) OrgsListUsers(ctx context.Context, orgID int32) (users []int32, err error) {
	args, err := json.Marshal(orgID)
	if err != nil {
		return nil, err
	}
	resp, err := c.post("orgs/list-users", args)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (c *internalClient) OrgsGetByName(ctx context.Context, orgName string) (orgID *int32, err error) {
	args, err := json.Marshal(orgName)
	if err != nil {
		return nil, err
	}
	resp, err := c.post("orgs/get-by-name", args)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(resp.Body).Decode(&orgID)
	if err != nil {
		return nil, err
	}
	return orgID, nil
}

func (c *internalClient) OrgsGetSlackWebhooks(ctx context.Context, orgIDs []int32) (webhooks []*string, err error) {
	args, err := json.Marshal(orgIDs)
	if err != nil {
		return nil, err
	}
	resp, err := c.post("orgs/get-slack-webhooks", args)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(resp.Body).Decode(&webhooks)
	if err != nil {
		return nil, err
	}
	return webhooks, nil
}

func (c *internalClient) UsersGetByUsername(ctx context.Context, username string) (user *int32, err error) {
	args, err := json.Marshal(username)
	if err != nil {
		return nil, err
	}
	resp, err := c.post("users/get-by-username", args)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (c *internalClient) UserEmailsGetEmail(ctx context.Context, userID int32) (email *string, err error) {
	args, err := json.Marshal(userID)
	if err != nil {
		return nil, err
	}
	resp, err := c.post("user-emails/get-email", args)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(resp.Body).Decode(&email)
	if err != nil {
		return nil, err
	}
	return email, nil
}

// TODO(slimsag): In the future, once we're no longer using environment
// variables to build AppURL, remove this in favor of services just reading it
// directly from the configuration file.
func (c *internalClient) AppURL(ctx context.Context) (string, error) {
	resp, err := c.post("app-url", nil)
	if err != nil {
		return "", err
	}
	var appURL string
	err = json.NewDecoder(resp.Body).Decode(&appURL)
	if err != nil {
		return "", err
	}
	return appURL, nil
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

func (c *internalClient) ReposCreateIfNotExists(ctx context.Context, op RepoCreateOrUpdateRequest) (*Repo, error) {
	req, err := json.Marshal(op)
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
