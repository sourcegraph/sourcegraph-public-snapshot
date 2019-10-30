package run

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"

	log15 "gopkg.in/inconshreveable/log15.v2"
)

type combyArgs struct {
	ScopeQuery      string `json:"scopeQuery"`
	MatchTemplate   string `json:"matchTemplate"`
	RewriteTemplate string `json:"rewriteTemplate"`
}

type comby struct {
	plan *a8n.CampaignPlan

	args combyArgs
}

func (c *comby) validateArgs() error {
	if err := jsonc.Unmarshal(c.plan.Arguments, &c.args); err != nil {
		return err
	}

	if c.args.ScopeQuery == "" {
		return errors.New("missing argument in specification: scopeQuery")
	}

	if c.args.MatchTemplate == "" {
		return errors.New("missing argument in specification: matchTemplate")
	}

	if c.args.RewriteTemplate == "" {
		return errors.New("missing argument in specification: rewriteTemplate")
	}

	return nil
}

func (c *comby) searchQuery() string { return c.args.ScopeQuery }
func (c *comby) generateDiff(repo api.RepoName, commit api.CommitID) (string, error) {
	u, err := url.Parse(graphqlbackend.ReplacerURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("repo", string(repo))
	q.Set("commit", string(commit))
	q.Set("matchtemplate", c.args.MatchTemplate)
	q.Set("rewritetemplate", c.args.RewriteTemplate)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}

	cl := &http.Client{}
	resp, err := cl.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", err
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 100), 10*bufio.MaxScanTokenSize)

	type rawCodemodResult struct {
		URI  string `json:"uri"`
		Diff string
	}

	var diff string
	for scanner.Scan() {
		var raw *rawCodemodResult
		b := scanner.Bytes()
		if err := scanner.Err(); err != nil {
			log15.Info(fmt.Sprintf("Skipping codemod scanner error (line too long?): %s", err.Error()))
			continue
		}
		if err := json.Unmarshal(b, &raw); err != nil {
			continue
		}
		diff += raw.Diff
	}
	return diff, nil
}
