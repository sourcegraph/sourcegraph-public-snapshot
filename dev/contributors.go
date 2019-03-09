// Version that excludes Sourcegraphers and prints a markdown list with hyperlinks.

// Commented out code was experimenting with printing contributors for each relevant repository but as some repositories don't have issues, it wasn't used.

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/go-github/github" // with go modules enabled (GO111MODULE=on or outside GOPATH)
	"golang.org/x/oauth2"
)

type IssueAuthorDetails struct {
	handle string
	url    string
}

func main() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "9de0403e531d51cd6a237a2c56c667567f32ab4e"},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	repos := [...]string{"deploy-sourcegraph", "lang-typescript", "sourcegraph", "about", "go-langserver", "lang-go", "lang-python", "sourcegraph-basic-code-intel", "python-language-server"}
	sourcegraphers := map[string]string{
		"ryan-blunden":    "",
		"felixfbecker":    "",
		"attfarhan":       "",
		"vanesa":          "",
		"nicksnyder":      "",
		"lguychard":       "",
		"ijsnow":          "",
		"sqs":             "",
		"renovate[bot]":   "",
		"keegancsmith":    "",
		"beyang":          "",
		"chrismwendt":     "",
		"dadlerj":         "",
		"francisschmaltz": "",
		"ggilmore":        "",
		"KattMingMing":    "",
		"slimsag":         "",
		"tsenart":         "",
		"kevinzliu":       "",
	}

	// Update this to the createDate you're looking for.
	issueDate, err := time.Parse(time.RFC3339, "2019-02-08T12:00:00.000Z")
	if err != nil {
		log.Fatal("error parsing date", err)
		return
	}
	contributors := make(map[string]IssueAuthorDetails)
	page := 0

	// First print via repo
	for _, repo := range repos {
		// repoContributors := make(map[string]IssueAuthorDetails)

		for page < 15 {
			opts := &github.IssueListByRepoOptions{Since: issueDate, ListOptions: github.ListOptions{PerPage: 100, Page: page}}
			// list issues for the specified repository. You could fetch a list of all repos in an org and iterate over that
			// to find all issues created in our org.
			issues, _, err := client.Issues.ListByRepo(ctx, "sourcegraph", repo, opts)
			if err != nil {
				log.Fatal("error fetching repo", err)
				return
			}
			// If there's no more results returned don't go to next page.
			if len(issues) == 0 {
				break
			}

			for _, issue := range issues {
				if _, ok := sourcegraphers[*issue.User.Login]; ok {
					continue
				}

				contributors[*issue.User.Login] = IssueAuthorDetails{handle: *issue.User.Login, url: *issue.User.HTMLURL}
				// repoContributors[*issue.User.Login] = IssueAuthorDetails{handle: *issue.User.Login, url: *issue.User.HTMLURL}
			}
			page++
		}

		// List contributors by repo

		// fmt.Println(fmt.Sprintf("\n\n**[%s](https://github.com/sourcegraph/%s)**:\n", repo, repo))

		// for _, repoContributor := range repoContributors {
		// 	fmt.Println(fmt.Sprintf("- [@%s](%s)", repoContributor.handle, repoContributor.url))
		// }
	}

	for _, contributor := range contributors {
		fmt.Println(fmt.Sprintf("- [@%s](%s)", contributor.handle, contributor.url))
	}
}
