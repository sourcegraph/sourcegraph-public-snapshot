package a8n

import (
	"bufio"
	"context"
	"encoding/json"
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

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	cby "github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httputil"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	schema "github.com/sourcegraph/sourcegraph/schema/campaign-types"
	"github.com/xeipuuv/gojsonschema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// defaultFetchTimeout determines how long we wait for the replacer service to fetch
// zip archives
const defaultFetchTimeout = 30 * time.Second

var schemas = map[string]string{
	"comby":       schema.CombyCampaignTypeSchemaJSON,
	"credentials": schema.CredentialsCampaignTypeSchemaJSON,
}

const patchCampaignType = "patch"

// NewCampaignType returns a new CampaignType for the given campaign type name
// and arguments.
func NewCampaignType(campaignTypeName, args string, cf *httpcli.Factory) (CampaignType, error) {
	campaignTypeName = strings.ToLower(campaignTypeName)

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

	normalizedArgs, err := validateArgs(campaignTypeName, args)
	if err != nil {
		return nil, err
	}

	var ct CampaignType

	switch campaignTypeName {
	case "comby":
		c := &comby{
			replacerURL:  graphqlbackend.ReplacerURL,
			httpClient:   cli,
			fetchTimeout: defaultFetchTimeout,
		}

		if err := json.Unmarshal(normalizedArgs, &c.args); err != nil {
			return nil, err
		}

		ct = c

	case "credentials":
		c := &credentials{newSearch: graphqlbackend.NewSearchImplementer}

		if err := json.Unmarshal(normalizedArgs, &c.args); err != nil {
			return nil, err
		}

		ct = c

	case patchCampaignType:
		// Prefer the more specific createCampaignPlanFromPatches GraphQL API for creating campaigns
		// from patches computed by the caller, to avoid having multiple ways to do the same thing.
		return nil, errors.New("use createCampaignPlanFromPatches for patch campaign types")

	default:
		return nil, fmt.Errorf("unknown campaign type: %s", campaignTypeName)
	}

	return ct, nil
}

func validateArgs(campaignType, args string) ([]byte, error) {
	typeSchema, ok := schemas[campaignType]
	if !ok {
		return nil, fmt.Errorf("unknown campaign type: %s", campaignType)
	}

	sl := gojsonschema.NewSchemaLoader()
	sc, err := sl.Compile(gojsonschema.NewStringLoader(typeSchema))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to compile schema for campaign type %q", campaignType)
	}

	normalized, err := jsonc.Parse(args)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to normalize JSON")
	}

	res, err := sc.Validate(gojsonschema.NewBytesLoader(normalized))
	if err != nil {
		return nil, errors.Wrap(err, "failed to validate specification against schema")
	}

	validationErrs := res.Errors()
	sort.Slice(validationErrs, func(i, j int) bool {
		return validationErrs[i].Field() < validationErrs[j].Field()
	})

	var errs *multierror.Error
	for _, err := range validationErrs {
		e := err.String()
		// Remove `(root): ` from error formatting since these errors are
		// presented to users.
		e = strings.TrimPrefix(e, "(root): ")
		errs = multierror.Append(errs, errors.New(e))
	}
	return normalized, errs.ErrorOrNil()
}

// A CampaignType provides a search query, argument validation and generates a
// diff in a given repository.
type CampaignType interface {
	// searchQuery returns a search query that returns the repositories over
	// which to execute the CampaignType. It changes depending on the arguments
	// with which the CampaignType was initialized and the type of the
	// CampaignType.
	searchQuery() string
	// generateDiff returns a diff (can be blank), a description of the diff in
	// GitHub flavored Markdown (can be blank) and, optionally, an error.
	generateDiff(context.Context, api.RepoName, api.CommitID) (diff, description string, err error)
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
func (c *comby) generateDiff(ctx context.Context, repo api.RepoName, commit api.CommitID) (string, string, error) {
	u, err := url.Parse(c.replacerURL)
	if err != nil {
		return "", "", err
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
		return "", "", err
	}

	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("unexpected response status from replacer service: %q", resp.Status)
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

	return result.String(), "", nil
}

type credentialsMatcher struct {
	MatcherType string `json:"type"`
	ReplaceWith string `json:"replaceWith"`
}

type credentialsArgs struct {
	ScopeQuery string               `json:"scopeQuery"`
	Matchers   []credentialsMatcher `json:"matchers"`
}

var npmTokenRegexp = regexp.MustCompile(`((?:^|:)_(?:auth|authToken|password)\s*=\s*)(.+)$`)
var npmTokenRegexpMultiline = regexp.MustCompile(`(?m)((?:^|:)_(?:auth|authToken|password)\s*=\s*)(.+)$`)

type credentials struct {
	args credentialsArgs

	newSearch func(*graphqlbackend.SearchArgs) (graphqlbackend.SearchImplementer, error)
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

func (c *credentials) generateDiff(ctx context.Context, repo api.RepoName, commit api.CommitID) (string, string, error) {
	t := "regexp"
	search, err := c.newSearch(&graphqlbackend.SearchArgs{
		Version:     "V2",
		PatternType: &t,
		Query:       c.searchQueryForRepo(repo),
	})
	if err != nil {
		return "", "", err
	}

	resultsResolver, err := search.Results(ctx)
	if err != nil {
		return "", "", err
	}

	diffs := []string{}
	tokens := []string{}

	for _, res := range resultsResolver.Results() {
		fm, ok := res.ToFileMatch()
		if !ok {
			continue
		}

		path := fm.File().Path()
		content, err := fm.File().Content(ctx)
		if err != nil {
			return "", "", err
		}

		submatches := npmTokenRegexpMultiline.FindAllStringSubmatch(content, -1)
		for _, match := range submatches {
			tokens = append(tokens, match[len(match)-1])
		}

		replacement := fmt.Sprintf("${1}%s", c.args.Matchers[0].ReplaceWith)
		newContent := npmTokenRegexpMultiline.ReplaceAllString(content, replacement)

		diff, err := tmpfileDiff(path, content, newContent)
		if err != nil {
			return "", "", err
		}

		withHeader := fmt.Sprintf("diff %s %s\n%s", path, path, diff)

		diffs = append(diffs, withHeader)
	}

	var description strings.Builder
	if len(tokens) > 0 {
		description.WriteString("Tokens found:\n\n")
		for _, tok := range tokens {
			description.WriteString(fmt.Sprintf("- [ ] `%s`\n", tok))
		}
	}

	return strings.Join(diffs, "\n"), description.String(), nil
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
