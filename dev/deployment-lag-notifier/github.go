package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GithubCommit represents the "commit" member of a response object
type GithubCommit struct {
	Sha    string `json:"sha"`
	Commit struct {
		Author struct {
			Name string    `json:"name"`
			Date time.Time `json:"date"`
		} `json:"author"`
		Message string `json:"message"`
	} `json:"commit"`
}

// GithubResponse is the response payload from requesting GET /repos/:author/:repo/commits, made up
// of a slice of GithubCommit's
type GithubResponse []GithubCommit

// Commit is a singular Git commit to a repo
type Commit struct {
	Sha     string
	Author  string
	Message string
	Date    time.Time
}

// getCommit hits the Github API to fetch information on a singular commit
func getCommit(client *http.Client, sha string) (Commit, error) {
	var commit Commit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	url := fmt.Sprintf("https://api.github.com/repos/sourcegraph/sourcegraph/commits/%v", sha)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return commit, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return commit, err
	}

	if resp.StatusCode != http.StatusOK {
		return commit, errors.Newf("received non-200 status code %v: %s", resp.StatusCode, err.Error())
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return commit, err
	}

	var gh GithubCommit
	err = json.Unmarshal(body, &gh)
	if err != nil {
		return commit, err
	}

	commit = Commit{Sha: gh.Sha, Author: gh.Commit.Author.Name, Message: gh.Commit.Message, Date: gh.Commit.Author.Date}

	return commit, nil
}

// getCommitLog fetches the last numCommits commits of sourcegraph/sourcegraph@main from the Github API
func getCommitLog(client *http.Client, numCommits int) ([]Commit, error) {
	var commits []Commit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	url := "https://api.github.com/repos/sourcegraph/sourcegraph/commits"

	var gh GithubResponse

	page := 1
	for len(gh) < numCommits {
		temp, err := func() (GithubResponse, error) {
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				return nil, err
			}

			commitsToGet := numCommits - (page-1)*100
			if commitsToGet > 100 {
				commitsToGet = 100
			}

			q := req.URL.Query()
			q.Add("branch", "main")
			q.Add("per_page", fmt.Sprintf("%v", commitsToGet))
			q.Add("page", fmt.Sprintf("%v", page))

			req.URL.RawQuery = q.Encode()

			resp, err := client.Do(req)
			if err != nil {
				return nil, err
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return nil, errors.Newf("received non-200 status code %v: %s", resp.StatusCode, err.Error())
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			var temp GithubResponse
			err = json.Unmarshal(body, &temp)
			if err != nil {
				return nil, err
			}
			gh = append(gh, temp...)

			page += 1
			return temp, nil
		}()
		if err != nil {
			return commits, err
		}

		// numCommits is greater than total amount of commits so stop querying
		if len(temp) < 100 {
			break
		}
	}

	if len(gh) != numCommits {
		return commits, errors.Newf("did not receive the expected number of commits. got: %v", len(gh))
	}

	for _, g := range gh {
		lines := strings.Split(g.Commit.Message, "\n")
		message := g.Sha[:7]
		commits = append(commits,
			Commit{Sha: message, Author: g.Commit.Author.Name, Message: lines[0], Date: g.Commit.Author.Date})
	}

	return commits, nil
}
