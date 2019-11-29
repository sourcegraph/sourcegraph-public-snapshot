package a8n

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	cby "github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httputil"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// defaultFetchTimeout determines how long we wait for the replacer service to fetch
// zip archives
const defaultFetchTimeout = 2 * time.Second

// NewCampaignType returns a new CampaignType for the given campaign type name
// and arguments.
func NewCampaignType(campaignTypeName, args string, cf *httpcli.Factory) (CampaignType, error) {
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

	var ct CampaignType

	switch strings.ToLower(campaignTypeName) {
	case "comby":
		c := &comby{
			replacerURL:  graphqlbackend.ReplacerURL,
			httpClient:   cli,
			fetchTimeout: defaultFetchTimeout,
		}

		if err := jsonc.Unmarshal(args, &c.args); err != nil {
			return nil, err
		}

		if c.args.ScopeQuery == "" {
			return nil, errors.New("missing argument in specification: scopeQuery")
		}

		if c.args.MatchTemplate == "" {
			return nil, errors.New("missing argument in specification: matchTemplate")
		}

		if c.args.RewriteTemplate == "" {
			return nil, errors.New("missing argument in specification: rewriteTemplate")
		}

		ct = c

	case "credentials":
		c := &credentials{}

		if err := jsonc.Unmarshal(args, &c.args); err != nil {
			return nil, err
		}

		if c.args.ScopeQuery == "" {
			return nil, errors.New("missing argument in specification: scopeQuery")
		}

		if len(c.args.Matchers) != 1 {
			return nil, errors.New("missing argument in specification: matchers")
		}

		if strings.ToLower(c.args.Matchers[0].MatcherType) != "npm" {
			t := c.args.Matchers[0].MatcherType
			return nil, fmt.Errorf("wrong matcher type in specification: %q", t)
		}

		ct = c

	default:
		return nil, fmt.Errorf("unknown campaign type: %s", campaignTypeName)
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

	replacerURL  string
	httpClient   httpcli.Doer
	fetchTimeout time.Duration
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
	if c.fetchTimeout > 0 {
		q.Set("fetchtimeout", c.fetchTimeout.String())
	}
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

	type fileDiffRaw struct {
		diff *diff.FileDiff
		raw  *string
	}

	var diffs []fileDiffRaw

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
		diffs = append(diffs, fileDiffRaw{diff: parsed, raw: &raw.Diff})
	}

	sort.Slice(diffs, func(i, j int) bool {
		if diffs[i].diff.OrigName != diffs[j].diff.OrigName {
			return diffs[i].diff.OrigName < diffs[j].diff.OrigName
		}
		return diffs[i].diff.NewName < diffs[j].diff.NewName
	})

	var result strings.Builder
	for _, fdr := range diffs {
		if result.Len() != 0 {
			// We already wrote a diff to the builder, so we need to add
			// a newline
			result.WriteRune('\n')
		}

		header := fmt.Sprintf("diff %s %s\n", fdr.diff.OrigName, fdr.diff.NewName)
		result.WriteString(header)
		result.WriteString(*fdr.raw)
	}

	return result.String(), nil
}

type credentialsMatcher struct {
	MatcherType string `json:"type"`
}

type credentialsArgs struct {
	ScopeQuery string               `json:"scopeQuery"`
	Matchers   []credentialsMatcher `json:"matchers"`
}

var npmTokenRegexp = regexp.MustCompile(`((?:^|:)_(?:auth|authToken|password)\s*=\s*)(.+)$`)
var npmTokenRegexpMultiline = regexp.MustCompile(`(?m)((?:^|:)_(?:auth|authToken|password)\s*=\s*)(.+)$`)

type credentials struct {
	args credentialsArgs
}

func (c *credentials) searchQuery() string {
	return c.args.ScopeQuery + " " + npmTokenRegexp.String() + " file:.npmrc"
}

func (c *credentials) searchQueryForRepo(n api.RepoName) string {
	return fmt.Sprintf(
		"file:.npmrc repo:%s %s",
		regexp.QuoteMeta(string(n)),
		npmTokenRegexp.String(),
	)
}

func (c *credentials) generateDiff(ctx context.Context, repo api.RepoName, commit api.CommitID) (string, error) {
	t := "regexp"
	search, err := graphqlbackend.NewSearchImplementer(&graphqlbackend.SearchArgs{
		Version:     "V2",
		PatternType: &t,
		Query:       c.searchQueryForRepo(repo),
	})
	if err != nil {
		return "", err
	}

	resultsResolver, err := search.Results(ctx)
	if err != nil {
		return "", err
	}

	diffs := []string{}

	for _, res := range resultsResolver.Results() {
		fm, ok := res.ToFileMatch()
		if !ok {
			continue
		}

		path := fm.File().Path()
		content, err := fm.File().Content(ctx)
		if err != nil {
			return "", err
		}
		newContent := npmTokenRegexpMultiline.ReplaceAllString(content, `${1}REMOVED_TOKEN`)

		diff, err := tmpfileDiff(path, content, newContent)
		if err != nil {
			return "", err
		}

		withHeader := fmt.Sprintf("diff %s %s\n%s", path, path, diff)

		diffs = append(diffs, withHeader)
	}

	return strings.Join(diffs, "\n"), nil
}

func tmpfileDiff(filename, a, b string) (string, error) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("diffing-%s", filename))
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	tmpFile := func(name, content string) (string, error) {
		p := filepath.Join(dir, name)
		if err := ioutil.WriteFile(p, []byte(content), 0666); err != nil {
			return "", err
		}
		return p, nil
	}

	fileA, err := tmpFile("a", a)
	if err != nil {
		return "", err
	}

	fileB, err := tmpFile("b", b)
	if err != nil {
		return "", err
	}

	cmd := exec.Command(
		"diff",
		"-u",
		"--label", filename,
		"--label", filename,
		fileA,
		fileB,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		// diff's exit code is 0 if no differences were found, 1 if differences
		// were found and 2+ if it ran into an error.
		if cmd.ProcessState.ExitCode() >= 2 {
			return string(out), err
		}
	}
	return string(out), nil
}
