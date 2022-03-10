package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v41/github"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/oauth2"
)

var ghToken = os.Getenv("GITHUB_TOKEN")

func main() {
	ctx := context.Background()
	ghc := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghToken},
	)))

	sha1 := os.Getenv("CI_PREPROD_COMMIT")
	fmt.Println(sha1)

	lastCommit, err := GetLastPreprodCommit()
	if err != nil {
		panic(err)
	}
	// Force last commit for testing purposes
	lastCommit = "cd0799fa3686c87909ad81570d17469e9840a230"

	pulls, err := GetPullRequestsSinceCommit(ctx, ghc, lastCommit)
	if err != nil {
		panic(err)
	}

	for _, pr := range pulls {
		fmt.Printf("pretending to post a comment in PR#%d\n", pr.GetNumber())
	}
}

func GetLastPreprodCommit() (string, error) {
	resp, err := http.Get("https://preview.sgdev.dev/__version")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	elems := strings.Split(string(body), "_")
	if len(elems) != 3 {
		return "", errors.Errorf("unknown format of /__version response: %q", body)
	}
	return elems[2], nil
}

func GetPullRequestsSinceCommit(ctx context.Context, ghc *github.Client, sha1 string) ([]*github.PullRequest, error) {
	var pullsSinceLastCommit []*github.PullRequest
	lines, err := GitCmd("log", "--format=%H", fmt.Sprintf("HEAD...%s", sha1))
	if err != nil {
		return nil, err
	}
	commits := strings.Split(lines, "\n")
	commits = commits[1 : len(commits)-1]

	for _, sha1 := range commits {
		pulls, _, err := ghc.PullRequests.ListPullRequestsWithCommit(
			ctx,
			"sourcegraph",
			"sourcegraph",
			sha1,
			&github.PullRequestListOptions{
				State: "merged",
			},
		)
		if err != nil {
			return nil, err
		}
		pullsSinceLastCommit = append(pullsSinceLastCommit, pulls...)
	}

	return pullsSinceLastCommit, nil
}

// func markDeployed() {
// 	ctx := context.Background()
// 	ghc := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
// 		&oauth2.Token{AccessToken: ghToken},
// 	)))
//
// 	ref := "ede7d6cf098047b1c79a8e2e9b6eadcb147934ba"
// 	transient := true
// 	environment := "cloud-preprod"
// 	description := "https://github.com/sourcegraph/sourcegraph/pull/32381"
// 	automerge := false
// 	requiredContext := []string{}
// 	deployment, _, err := ghc.Repositories.CreateDeployment(ctx, "sourcegraph", "sourcegraph", &github.DeploymentRequest{
// 		Ref:                  &ref,
// 		TransientEnvironment: &transient,
// 		Environment:          &environment,
// 		Description:          &description,
// 		AutoMerge:            &automerge,
// 		RequiredContexts:     &requiredContext,
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	fmt.Println(deployment.GetID())
//
// 	state := "in_progress"
// 	_, _, err = ghc.Repositories.CreateDeploymentStatus(ctx, "sourcegraph", "sourcegraph", deployment.GetID(), &github.DeploymentStatusRequest{
// 		State: &state,
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
//
// }
//
// func fetchLastPRCommit() {
// 	ctx := context.Background()
// 	ghc := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
// 		&oauth2.Token{AccessToken: ghToken},
// 	)))
//
// 	ref := "072969e28084444bff5c5bc1695ab709e61729b6"
// 	pulls, _, err := ghc.PullRequests.ListPullRequestsWithCommit(ctx, "sourcegraph", "sourcegraph", ref, &github.PullRequestListOptions{
// 		State: "merged",
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	if len(pulls) > 1 {
// 		panic("multiple PRs")
// 	}
//
// 	pr := pulls[0]
// 	fmt.Println(pr.GetTitle())
//
// 	pr, _, err = ghc.PullRequests.Get(ctx, "sourcegraph", "sourcegraph", int(pr.GetNumber()))
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(pr.GetHead().GetSHA())
// }

// Extracted from sg, keeping it here for testing purposes
func GitCmd(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Env = append(os.Environ(),
		// Don't use the system wide git config.
		"GIT_CONFIG_NOSYSTEM=1",
		// And also not any other, because they can mess up output, change defaults, .. which can do unexpected things.
		"GIT_CONFIG=/dev/null")

	return InRoot(cmd)
}

// Extracted from sg, keeping it here for testing purposes
func InRoot(cmd *exec.Cmd) (string, error) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return "", err
	}

	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), errors.Wrapf(err, "'%s' failed: %s", strings.Join(cmd.Args, " "), out)
	}

	return string(out), nil
}
