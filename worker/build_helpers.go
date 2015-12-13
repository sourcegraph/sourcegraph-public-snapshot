package worker

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

func getBuild(repoURI string, id uint64) (*sourcegraph.Build, *sourcegraph.Repo, error) {
	cl := cli.Client()

	build, err := cl.Builds.Get(cli.Ctx, &sourcegraph.BuildSpec{
		ID:   id,
		Repo: sourcegraph.RepoSpec{URI: repoURI},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("getting build: %s", err)
	}

	repo, err := cl.Repos.Get(cli.Ctx, &sourcegraph.RepoSpec{URI: build.Repo})
	if err != nil {
		return nil, nil, fmt.Errorf("getting repository %q: %s", build.Repo, err)
	}

	if repo.HTTPCloneURL != "" {
		checkHTTPCloneURL(repo.HTTPCloneURL)
	}
	if repo.SSHCloneURL != "" {
		checkSSHCloneURL(string(repo.SSHCloneURL))
	}
	checkCommitID(build.CommitID)

	return build, repo, nil
}

func checkHTTPCloneURL(urlStr string) {
	u, err := url.Parse(urlStr)
	if err != nil {
		log.Fatal(err)
	}
	if u.Scheme == "" || strings.ToLower(u.Scheme) == "file" || u.User != nil {
		log.Fatalf("attempt to use insecure URL %q", urlStr)
	}
}

func checkSSHCloneURL(urlStr string) {
	u, err := url.Parse(urlStr)
	if err != nil {
		log.Fatal(err)
	}
	if u.Scheme != "ssh" {
		log.Fatalf("attempt to use insecure/malformed ssh URL %q", urlStr)
	}
}

func checkCommitID(id string) {
	if id == "" {
		log.Fatal("empty commit ID")
	}
	if !isLowerAlnum(id) {
		log.Fatalf("commit ID is not lowercase alphanumeric: %q", id)
	}
}

func isLowerAlnum(s string) bool {
	return strings.IndexFunc(s, func(c rune) bool {
		return !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9'))
	}) == -1
}
