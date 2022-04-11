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

// GithubResponse is the response payload from requesting GET /repos/:author/:repo/commits
type GithubResponse []GithubCommit

// Commit is a singular Git commit to a repo
type Commit struct {
	Sha     string
	Author  string
	Message string
	Date    time.Time
}

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

	if resp.StatusCode > http.StatusOK {
		return commit, errors.Newf("received non-200 status code %v", resp.StatusCode)
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

func getCommitLog(client *http.Client) ([]Commit, error) {
	var commits []Commit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	url := "https://api.github.com/repos/sourcegraph/sourcegraph/commits"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return commits, err
	}

	q := req.URL.Query()
	q.Add("branch", "main")
	q.Add("per_page", "20")
	q.Add("page", "1")

	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return commits, err
	}

	if resp.StatusCode > http.StatusOK {
		return commits, errors.Newf("received non-200 status code %v", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return commits, err
	}

	// fmt.Println(string(body))

	var gh GithubResponse
	err = json.Unmarshal(body, &gh)
	if err != nil {
		return commits, err
	}

	for _, g := range gh {
		lines := strings.Split(g.Commit.Message, "\n")
		message := g.Sha[:7]
		commits = append(commits,
			Commit{Sha: message, Author: g.Commit.Author.Name, Message: lines[0]})
	}

	return commits, nil
}

// checkForCommit checks for the current version in the
// last 20 commits
func checkForCommit(version string, commits []Commit) bool {
	found := false
	for _, c := range commits {
		if c.Sha == version[:7] {
			found = true
		}
	}

	return found
}
