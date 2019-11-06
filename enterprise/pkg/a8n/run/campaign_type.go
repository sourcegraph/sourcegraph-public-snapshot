package run

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"

	log15 "gopkg.in/inconshreveable/log15.v2"
)

// NewCampaignType returns a new Campaign for the given CampaignType and
// arguments.
// Before the returned CampaignType can be passed to a Runner its Valid method
// needs to be called.
func NewCampaignType(campaignTypeName, args string) (CampaignType, error) {
	switch strings.ToLower(campaignTypeName) {
	case "comby":
		return &comby{rawArgs: args}, nil
	default:
		return nil, fmt.Errorf("unknown campaign type: %s", campaignTypeName)
	}
}

// A CampaignType provides a search query, argument validation and generates a
// diff in a given repository.
type CampaignType interface {
	Valid() error

	searchQuery() string
	generateDiff(context.Context, api.RepoName, api.CommitID) (string, error)
}

type combyArgs struct {
	ScopeQuery      string `json:"scopeQuery"`
	MatchTemplate   string `json:"matchTemplate"`
	RewriteTemplate string `json:"rewriteTemplate"`
}

type comby struct {
	rawArgs string
	args    combyArgs
}

func (c *comby) Valid() error {
	if err := jsonc.Unmarshal(c.rawArgs, &c.args); err != nil {
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
func (c *comby) generateDiff(ctx context.Context, repo api.RepoName, commit api.CommitID) (string, error) {
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

	var diffs []*diff.FileDiff
	for scanner.Scan() {
		var raw *rawCodemodResult
		b := scanner.Bytes()
		if err := scanner.Err(); err != nil {
			log15.Info(fmt.Sprintf("Skipping codemod scanner error (line too long?): %s", err.Error()))
			continue
		}
		if err := json.Unmarshal(b, &raw); err != nil {
			log15.Error("unmarshalling raw diff failed", "err", err)
			continue
		}
		// TODO(a8n): Do we need to use `diff.ParseFileDiff` or can we just concatenate?
		parsed, err := diff.ParseFileDiff([]byte(raw.Diff))
		if err != nil {
			log15.Error("parsing diff failed", "err", err)
			continue
		}
		diffs = append(diffs, parsed)
	}

	// TODO(a8n): The diffs returned by Comby and then produced by this are
	// missing the "extended" fields in a `diff.FileDiff` that would allow
	// `diff.Parse*` to parse the diffs properly again. We need to change the
	// comby API so that it returns proper git diffs.
	multiDiff, err := diff.PrintMultiFileDiff(diffs)
	if err != nil {
		return "", err
	}

	return string(multiDiff), nil
}
