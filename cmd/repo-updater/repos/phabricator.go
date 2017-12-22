package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	log15 "gopkg.in/inconshreveable/log15.v2"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

type phabRepo struct {
	Fields *struct {
		// e.g. "git"
		VCS string `json:"vcs"`
		// phab unique short name
		Callsign string `json:"callsign"`
		// "active" || "inactive"
		Status string `json:"status"`
	} `json:"fields"`
	Attachments *struct {
		URIs *struct {
			URIs []*struct {
				Fields *struct {
					URI *struct {
						Raw        string `json:"raw"`
						Normalized string `json:"normalized"`
					} `json:"uri"`
					Builtin *struct {
						Protocol   *string `json:"protocol"`
						Identifier *string `json:"identifier"`
					} `json:"builtin"`
				} `json:"fields"`
			} `json:"uris"`
		} `json:"uris"`
	} `json:"attachments"`
}

type phabRepoLookupResponse struct {
	Data   []*phabRepo `json:"data"`
	Cursor *struct {
		Limit  int     `json:"limit"`
		After  *string `json:"after"`
		Before *string `json:"before"`
	} `json:"cursor"`
}

type phabAPIResponse struct {
	Result    *phabRepoLookupResponse `json:"result"`
	Error     *string                 `json:"error_code"`
	ErrorInfo *string                 `json:"error_info"`
}

var (
	phabConf               = conf.Get().Phabricator
	repoUpdateIntervalConf = conf.Get().RepoListUpdateInterval
)

// RunPhabricatorRepositorySyncWorker runs the worker that syncs repositories from Phabricator to Sourcegraph
func RunPhabricatorRepositorySyncWorker(ctx context.Context) error {
	if repoUpdateIntervalConf == 0 {
		return errors.New("Update interval is 0 (set repoListUpdateInterval to a non-zero value or omit it)")
	}
	updateInterval := time.Duration(repoUpdateIntervalConf) * time.Second

	for {
		for _, c := range phabConf {
			if c.Token == "" {
				continue
			}

			after := ""
			for {
				res, err := fetchPhabRepos(ctx, c, after)
				if err != nil {
					log15.Error("Error fetching Phabricator repos", "err", err)
					break
				}
				err = updatePhabRepos(ctx, &c, res.Data)
				if err != nil {
					log15.Error("Error updating Phabricator repos", "err", err)
				}

				if res.Cursor.After == nil {
					break
				}
				after = *res.Cursor.After
			}

		}
		time.Sleep(updateInterval)
	}
}

func fetchPhabRepos(ctx context.Context, cfg schema.Phabricator, after string) (*phabRepoLookupResponse, error) {
	form := url.Values{}
	form.Add("output", "json")
	form.Add("params[__conduit__]", `{"token": "`+cfg.Token+`"}`)
	form.Add("params[attachments]", `{"uris": true}`)
	if after != "" {
		form.Add("params[after]", after)
	}
	req, err := http.NewRequest("POST", strings.TrimSuffix(cfg.Url, "/")+"/api/diffusion.repository.search", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	res := phabAPIResponse{}
	err = json.Unmarshal(respBody, &res)
	if err != nil {
		return nil, err
	}

	if res.Error != nil {
		return nil, fmt.Errorf("phab error %s: %s", *res.Error, *res.ErrorInfo)
	}
	return res.Result, nil
}

// updatePhabRepos ensures that all provided repositories exist in the phabricator_repos table.
func updatePhabRepos(ctx context.Context, cfg *schema.Phabricator, repos []*phabRepo) error {
	for _, repo := range repos {
		if repo.Fields.VCS != "git" {
			continue
		}
		if repo.Fields.Status == "inactive" {
			continue
		}
		var uri string
		for _, u := range repo.Attachments.URIs.URIs {
			// Phabricator may list multiple URIs for a repo, some of which are internal Phabricator resources.
			// We select the first URI which doesn't have `builtin` fields (as those only come from internal Phab
			// resources).
			if u.Fields.Builtin != nil && u.Fields.Builtin.Identifier != nil {
				continue
			}
			uri = u.Fields.URI.Normalized
			break
		}
		if uri == "" {
			// some repos have no attachments
			return nil
		}

		err := sourcegraph.InternalClient.PhabricatorRepoCreate(ctx, uri, repo.Fields.Callsign, cfg.Url)
		if err != nil {
			return err
		}
	}

	return nil
}
