package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
)

type Flags struct {
	GitHubPayloadPath string
	GitHubToken       string

	IssuesRepoOwner string
	IssuesRepoName  string
}

func (f *Flags) Parse() {
	flag.StringVar(&f.GitHubPayloadPath, "github.payload-path", "", "path to JSON file with GitHub event payload")
	flag.StringVar(&f.GitHubToken, "github.token", "", "GitHub token")
	flag.StringVar(&f.IssuesRepoOwner, "issues.repo-owner", "sourcegraph", "owner of repo to create issues in")
	flag.StringVar(&f.IssuesRepoName, "issues.repo-name", "sec-audit-trail", "name of repo to create issues in")
	flag.Parse()
}

func main() {
	flags := &Flags{}
	flags.Parse()

	ctx := context.Background()
	ghc := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: flags.GitHubToken},
	)))

	payloadData, err := os.ReadFile(flags.GitHubPayloadPath)
	if err != nil {
		log.Fatal("ReadFile: ", err)
	}
	var payload *Payload
	if err := json.Unmarshal(payloadData, &payload); err != nil {
		log.Fatal("Unmarshal: ", err)
	}
	log.Printf("payload: %+v\n", payload)

	if !payload.PullRequest.Merged {
		log.Printf("pull request %s not merged, discarding\n", payload.PullRequest.URL)
		return
	}

	acceptance := checkAcceptance(payload.PullRequest.Body)
	log.Printf("checkAcceptance: %+v\n", acceptance)
	if acceptance.Checked {
		log.Println("Acceptance checked, done")
		return
	}

	var (
		issueTitle     = fmt.Sprintf("acceptance checklist exception: PR %s#%d", payload.Repository.FullName, payload.PullRequest.Number)
		issueBody      = fmt.Sprintf("%s did not go through the acceptance checklist.", payload.PullRequest.URL)
		issueAssignees = []string{}
		closeIssue     bool
	)
	if acceptance.Explanation == "" {
		user := payload.PullRequest.MergedBy.Login
		issueAssignees = append(issueAssignees, user)
		issueBody += fmt.Sprintf("\n\nNo explanation was provided - @%s please comment in this issue with an explanation for this exception and close this issue.", user)
	} else {
		closeIssue = true
		issueBody += fmt.Sprintf("\n\nProvided explanation:\n\n%s", acceptance.Explanation)
	}

	log.Println("Creating issue for exception")
	created, _, err := ghc.Issues.Create(ctx, flags.IssuesRepoOwner, flags.IssuesRepoName, &github.IssueRequest{
		Title:     github.String(issueTitle),
		Body:      github.String(issueBody),
		Assignees: &issueAssignees,
	})
	if err != nil {
		log.Fatal("Issues.Create: ", err)
	}
	log.Println("Created issue: ", created.GetHTMLURL())

	if closeIssue {
		_, _, err := ghc.Issues.Edit(ctx, flags.IssuesRepoOwner, flags.IssuesRepoName, created.GetNumber(), &github.IssueRequest{
			State: github.String("closed"),
		})
		if err != nil {
			log.Fatal("Issues.Edit: ", err)
		}
		log.Println("Created issue was closed")
	}
}
