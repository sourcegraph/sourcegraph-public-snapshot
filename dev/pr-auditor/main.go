package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
)

type Flags struct {
	GitHubPayloadPath string
	GitHubToken       string
	GitHubRunURL      string

	IssuesRepoOwner string
	IssuesRepoName  string

	// ProtectedBranch designates a branch name that should always record an exception when a PR is opened
	// against it. It's primary use case is to discourage PRs againt the release branch on sourcegraph/deploy-sourcegraph-cloud.
	ProtectedBranch string

	// AdditionalContext contains a paragraph that will be appended at the end of the created exception. It enables
	// repositories to further explain why an exception has been recorded.
	AdditionalContext string

	// SkipStatus if true will skip updating commit status on GitHub and just record exceptions. Useful when crawling through failed runs caused by infrastructure issues.
	SkipStatus bool
}

func (f *Flags) Parse() {
	flag.StringVar(&f.GitHubPayloadPath, "github.payload-path", "", "path to JSON file with GitHub event payload")
	flag.StringVar(&f.GitHubToken, "github.token", "", "GitHub token")
	flag.StringVar(&f.GitHubRunURL, "github.run-url", "", "URL to GitHub actions run")
	flag.StringVar(&f.IssuesRepoOwner, "issues.repo-owner", "sourcegraph", "owner of repo to create issues in")
	flag.StringVar(&f.IssuesRepoName, "issues.repo-name", "sec-pr-audit-trail", "name of repo to create issues in")
	flag.StringVar(&f.ProtectedBranch, "protected-branch", "", "name of branch that if set as the base branch in a PR, will always open an exception")
	flag.StringVar(&f.AdditionalContext, "additional-context", "", "additional information that will be appended to the recorded exception, if any.")
	flag.BoolVar(&f.SkipStatus, "skip-status", false, "skip updating commit status on GitHub and just record exceptions")
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
	switch ref := payload.PullRequest.Base.Ref; ref {
	// This is purely an API call usage optimization, so we don't need to be so specific
	// as to require usage to provide the default branch - we can just rely on a simple
	// allowlist of commonly used default branches.
	case "main", "master", "release":
		log.Printf("performing checks against allow-listed pull request base %q", ref)
	case flags.ProtectedBranch:
		if flags.ProtectedBranch == "" {
			log.Printf("unknown pull request base %q - discarding\n", ref)
			return
		}

		log.Printf("performing checks against protected pull request base %q", ref)
	default:
		log.Printf("unknown pull request base %q - discarding\n", ref)
		return
	}
	if payload.PullRequest.Draft {
		log.Println("skipping event on draft PR")
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
		ProtectedBranch: flags.ProtectedBranch,
	})
	log.Printf("checkPR: %+v\n", result)

	if result.HasTestPlan() && result.Reviewed && !result.ProtectedBranch {
		log.Println("Acceptance checked and PR reviewed, done")
		// Don't create status that likely nobody will check anyway
		return nil
	}

	owner, repo := payload.Repository.GetOwnerAndName()
	if result.Error != nil && !flags.SkipStatus {
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

	issue := generateExceptionIssue(payload, &result, flags.AdditionalContext)

	log.Printf("Ensuring label for repository %q\n", payload.Repository.FullName)
	_, _, err := ghc.Issues.CreateLabel(ctx, flags.IssuesRepoOwner, flags.IssuesRepoName, &github.Label{
		Name: github.String(payload.Repository.FullName),
	})
	if err != nil {
		log.Printf("Ignoring error on CreateLabel: %s\n", err)
	}

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
		ValidateReviews: false, // only validate reviews on post-merge
		ProtectedBranch: flags.ProtectedBranch,
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
	case result.ProtectedBranch:
		prState = "success"
		stateDescription = "No action needed, but an exception will be opened post-merge."
	default:
		prState = "success"
		stateDescription = "No action needed, nice!"
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
