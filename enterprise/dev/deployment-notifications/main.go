package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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

	changedFiles, err := getChangedFiles()
	if err != nil {
		panic(err)
	}

	deployedApps := guessDeployedApps(changedFiles)
	fmt.Println(deployedApps)

	return

	sha1 := os.Getenv("CI_PREPROD_COMMIT")
	fmt.Println(sha1)

	lastCommit, err := getLastPreprodCommit()
	if err != nil {
		panic(err)
	}
	// Force last commit for testing purposes
	lastCommit = "cd0799fa3686c87909ad81570d17469e9840a230"

	pulls, err := getPullRequestsSinceCommit(ctx, ghc, lastCommit)
	if err != nil {
		panic(err)
	}

	for _, pr := range pulls {
		fmt.Printf("pretending to post a comment in PR#%d\n", pr.GetNumber())
	}
}

func getChangedFiles() ([]string, error) {
	diffCommand := []string{"diff", "--name-only", "HEAD~5"}
	if output, err := exec.Command("git", diffCommand...).Output(); err != nil {
		return nil, err
	} else {
		return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
	}
}

func guessDeployedApps(changedFiles []string) []string {
	var deployedApps []string
	for _, file := range changedFiles {
		if filepath.Ext(file) != ".yaml" {
			continue
		}
		base := filepath.Base(file)
		components := strings.Split(base, ".")
		if len(components) < 3 {
			continue
		}
		// gitserver.[Deployment|StatefulSet|DaemonSet].yaml
		kind := components[1]

		if kind == "Deployment" || kind == "StatefulSet" || kind == "DaemonSet" {
			deployedApps = append(deployedApps, components[0])
		}
	}
	return deployedApps
}

func getLastPreprodCommit() (string, error) {
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

func getPullRequestsSinceCommit(ctx context.Context, ghc *github.Client, sha1 string) ([]*github.PullRequest, error) {
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
