package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/net/context/ctxhttp"
)

func getCoverageForFile(ctx context.Context, token, repo, rev string, path string) (map[int]lineCoverage, error) {
	const codeHost = "gh" // TODO: support code hosts other than GitHub
	repo = strings.TrimPrefix(repo, "github.com/")

	// TODO: support self-hosted codecov (not just codecov.io)
	url := fmt.Sprintf("https://codecov.io/api/%s/%s/commits/%s?src=extension", codeHost, repo, rev)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	resp, err := ctxhttp.Do(ctx, nil, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(io.LimitReader(resp.Body, 200))
		return nil, fmt.Errorf("Codecov API returned HTTP %d (expected 200) at %s with body: %s", resp.StatusCode, req.URL, string(body))
	}

	type codecovFile struct {
		Lines map[string]lineCoverage `json:"l"`
	}
	type codecovReport struct {
		Files map[string]codecovFile `json:"files"`
	}
	type codecovCommit struct {
		Report codecovReport `json:"report"`
	}
	type codecovResponse struct {
		Commit codecovCommit `json:"commit"`
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response codecovResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, errors.Wrap(err, "unrecognized Codecov response structure")
	}

	hitsByStringLines := response.Commit.Report.Files[path].Lines
	hitsByLines := make(map[int]lineCoverage, len(hitsByStringLines))
	for stringLine, hits := range hitsByStringLines {
		line, err := strconv.Atoi(stringLine)
		if err != nil {
			return nil, err
		}
		hitsByLines[line] = hits
	}

	return hitsByLines, nil
}

// lineCoverage represents the coverage status for a line, described at
// https://docs.codecov.io/v5.0.0/reference#section-codecov-json-report-format.
type lineCoverage struct {
	// Exactly 1 of these groups of fields is set.
	hits               int
	partials, branches int
	skip               bool
}

func (c *lineCoverage) isPartial() bool {
	return c.branches > 0
}

func (c *lineCoverage) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v := v.(type) {
	case float64:
		*c = lineCoverage{hits: int(v)}
	case string:
		fraction := strings.SplitN(v, "/", 2)
		if len(fraction) != 2 {
			return fmt.Errorf("invalid line coverage input %q", data)
		}
		partials, err := strconv.Atoi(fraction[0])
		if err != nil {
			return fmt.Errorf("invalid line coverage hits input %q", fraction[0])
		}
		branches, err := strconv.Atoi(fraction[1])
		if err != nil {
			return fmt.Errorf("invalid line coverage branches input %q", fraction[1])
		}
		*c = lineCoverage{partials: partials, branches: branches}
	case nil:
		*c = lineCoverage{skip: true}
	default:
		return fmt.Errorf("invalid line coverage input %q", data)
	}
	return nil
}
