package a8n

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
	cby "github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httputil"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"

	log15 "gopkg.in/inconshreveable/log15.v2"
)

// NewCampaignType returns a new CampaignType for the given campaign type name
// and arguments.
func NewCampaignType(campaignTypeName, args string, cf *httpcli.Factory) (CampaignType, error) {
	if strings.ToLower(campaignTypeName) != "comby" {
		return nil, fmt.Errorf("unknown campaign type: %s", campaignTypeName)
	}

	if cf == nil {
		cf = httpcli.NewFactory(
			httpcli.NewMiddleware(httpcli.ContextErrorMiddleware),
			httpcli.TracedTransportOpt,
			httpcli.NewCachedTransportOpt(httputil.Cache, true),
		)
	}

	cli, err := cf.Doer()
	if err != nil {
		return nil, err
	}

	ct := &comby{replacerURL: graphqlbackend.ReplacerURL, httpClient: cli}

	if err := jsonc.Unmarshal(args, &ct.args); err != nil {
		return nil, err
	}

	if ct.args.ScopeQuery == "" {
		return nil, errors.New("missing argument in specification: scopeQuery")
	}

	if ct.args.MatchTemplate == "" {
		return nil, errors.New("missing argument in specification: matchTemplate")
	}

	if ct.args.RewriteTemplate == "" {
		return nil, errors.New("missing argument in specification: rewriteTemplate")
	}

	return ct, nil
}

// A CampaignType provides a search query, argument validation and generates a
// diff in a given repository.
type CampaignType interface {
	searchQuery() string
	generateDiff(context.Context, api.RepoName, api.CommitID) (string, error)
}

type combyArgs struct {
	ScopeQuery      string `json:"scopeQuery"`
	MatchTemplate   string `json:"matchTemplate"`
	RewriteTemplate string `json:"rewriteTemplate"`
}

type comby struct {
	args combyArgs

	replacerURL string
	httpClient  httpcli.Doer
}

func (c *comby) searchQuery() string { return c.args.ScopeQuery }
func (c *comby) generateDiff(ctx context.Context, repo api.RepoName, commit api.CommitID) (string, error) {
	u, err := url.Parse(c.replacerURL)
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

	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response status from replacer service: %q", resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 100), 10*bufio.MaxScanTokenSize)

	var result strings.Builder
	for scanner.Scan() {
		b := scanner.Bytes()
		if err := scanner.Err(); err != nil {
			log15.Info(fmt.Sprintf("Skipping codemod scanner error (line too long?): %s", err.Error()))
			continue
		}

		var raw cby.FileDiff
		if err := json.Unmarshal(b, &raw); err != nil {
			log15.Error("unmarshalling raw diff failed", "err", err)
			continue
		}

		// TODO(a8n): We need to parse the raw diff and inject the `header`
		// below because `go-diff` right now cannot parse multi-file diffs
		// without `diff ...` separator lines between the single file diffs.
		// Once that is fixed in `go-diff` we can simply concatenate
		// `raw.Diff`s without having to parse them.
		parsed, err := diff.ParseFileDiff([]byte(raw.Diff))
		if err != nil {
			log15.Error("parsing diff failed", "err", err)
			continue
		}
		if result.Len() != 0 {
			// We already wrote a diff to the builder, so we need to add
			// a newline
			result.WriteRune('\n')
		}

		header := fmt.Sprintf("diff %s %s\n", parsed.OrigName, parsed.NewName)
		result.WriteString(header)
		result.WriteString(raw.Diff)
	}

	return result.String(), nil
}
