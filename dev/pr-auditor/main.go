package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
)

type Flags struct {
	GitHubPayloadPath string
	GitHubToken       string
	GitHubRunURL      string

	IssuesRepoOwner string
	IssuesRepoName  string
}

func (f *Flags) Parse() {
	flag.StringVar(&f.GitHubPayloadPath, "github.payload-path", "", "path to JSON file with GitHub event payload")
	flag.StringVar(&f.GitHubToken, "github.token", "", "GitHub token")
	flag.StringVar(&f.GitHubRunURL, "github.run-url", "", "URL to GitHub actions run")
	flag.StringVar(&f.IssuesRepoOwner, "issues.repo-owner", "sourcegraph", "owner of repo to create issues in")
	flag.StringVar(&f.IssuesRepoName, "issues.repo-name", "sec-pr-audit-trail", "name of repo to create issues in")
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
	var payload *EventPayload
	if err := json.Unmarshal(payloadData, &payload); err != nil {
		log.Fatal("Unmarshal: ", err)
	}
	log.Printf("handling event for pull request %s, payload: %+v\n", payload.PullRequest.URL, payload.Dump())

	// Discard unwanted events
	if payload.PullRequest.Base.Ref != "main" {
		log.Printf("unknown pull request base %q - discarding\n", payload.PullRequest.Base.Ref)
		return
	}
	if payload.PullRequest.Draft {
		log.Printf("skipping event on draft PR")
		return
	}
	if payload.Action == "closed" && !payload.PullRequest.Merged {
		log.Println("ignoring closure of un-merged pull request")
		return
	}
	if payload.Action == "edited" && payload.PullRequest.Merged {
		log.Println("ignoring edit of already-merged pull request")
		return
	}

	// Do checks
	if payload.PullRequest.Merged {
		if err := postMergeAudit(ctx, ghc, payload, flags); err != nil {
			log.Fatalf("postMergeAudit: %s", err)
		}
	} else {
		if err := preMergeAudit(ctx, ghc, payload, flags); err != nil {
			log.Fatalf("preMergeAudit: %s", err)
		}
	}
}

const (
	commitStatusPostMerge = "pr-auditor / post-merge"
	commitStatusPreMerge  = "pr-auditor / pre-merge"
)

func postMergeAudit(ctx context.Context, ghc *github.Client, payload *EventPayload, flags *Flags) error {
	result := checkPR(ctx, ghc, payload, checkOpts{
		ValidateReviews: true,
	})
	log.Printf("checkPR: %+v\n", result)

	if result.HasTestPlan() && result.Reviewed {
		log.Println("Acceptance checked and PR reviewed, done")
		// Don't create status that likely nobody will check anyway
		return nil
	}

	owner, repo := payload.Repository.GetOwnerAndName()
	if result.Error != nil {
		_, _, statusErr := ghc.Repositories.CreateStatus(ctx, owner, repo, payload.PullRequest.Head.SHA, &github.RepoStatus{
			Context:     github.String(commitStatusPostMerge),
			State:       github.String("error"),
			Description: github.String(fmt.Sprintf("checkPR: %s", result.Error.Error())),
			TargetURL:   github.String(flags.GitHubRunURL),
		})
		if statusErr != nil {
			return errors.Newf("result.Error != nil (%w), statusErr: %w", result.Error, statusErr)
		}
		return nil
	}

	issue := generateExceptionIssue(payload, &result)

	log.Printf("Creating issue for exception: %+v\n", issue)
	created, _, err := ghc.Issues.Create(ctx, flags.IssuesRepoOwner, flags.IssuesRepoName, issue)
	if err != nil {
		// Let run fail, don't include special status
		return errors.Newf("Issues.Create: %w", err)
	}

	log.Println("Created issue: ", created.GetHTMLURL())

	// Let run succeed, create separate status indicating an exception was created
	_, _, err = ghc.Repositories.CreateStatus(ctx, owner, repo, payload.PullRequest.Head.SHA, &github.RepoStatus{
		Context:     github.String(commitStatusPostMerge),
		State:       github.String("failure"),
		Description: github.String("Exception detected and audit trail issue created"),
		TargetURL:   github.String(created.GetHTMLURL()),
	})
	if err != nil {
		return errors.Newf("CreateStatus: %w", err)
	}

	return nil
}

func preMergeAudit(ctx context.Context, ghc *github.Client, payload *EventPayload, flags *Flags) error {
	result := checkPR(ctx, ghc, payload, checkOpts{
		ValidateReviews: true,
	})
	log.Printf("checkPR: %+v\n", result)

	var prState, stateDescription string
	stateURL := flags.GitHubRunURL
	switch {
	case result.Error != nil:
		prState = "error"
		stateDescription = fmt.Sprintf("checkPR: %s", result.Error.Error())
	case !result.HasTestPlan():
		prState = "failure"
		stateDescription = "No test plan detected - please provide one!"
		stateURL = "https://docs.sourcegraph.com/dev/background-information/testing_principles#test-plans"
	default:
		// No need to set a status
		return nil
	}

	owner, repo := payload.Repository.GetOwnerAndName()
	_, _, err := ghc.Repositories.CreateStatus(ctx, owner, repo, payload.PullRequest.Head.SHA, &github.RepoStatus{
		Context:     github.String(commitStatusPreMerge),
		State:       github.String(prState),
		Description: github.String(stateDescription),
		TargetURL:   github.String(stateURL),
	})
	if err != nil {
		return errors.Newf("CreateStatus: %w", err)
	}
	return nil
}
