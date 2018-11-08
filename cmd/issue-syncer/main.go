package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func handleHelloWorld(resp http.ResponseWriter, req *http.Request) {
	io.WriteString(resp, "Hello world")
	fmt.Println("hello world")
}

func main() {

	http.HandleFunc("/hello-world", handleHelloWorld)
	http.HandleFunc("/fetch-sourcegraph-issues", fetchIssues)

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}

	addr := net.JoinHostPort(host, "4500")
	log15.Info("issue-syncer: listening", "addr", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// fetchIssues will fetch issues for a specified repository from the GitHub API.
func fetchIssues(resp http.ResponseWriter, req *http.Request) {
	// Create a extsvc/GitHub client, fetch repositories.
	githubConfig := conf.Get().Github[0]
	url, _ := url.Parse(githubConfig.Url)
	apiURL := extsvc.NormalizeBaseURL(url)

	cl := github.NewClient(apiURL, githubConfig.Token, nil, nil)
	ctx := context.Background()
	issues := cl.FetchIssues(ctx, "sourcegraph/sourcegraph")

	fmt.Println("ISSUES: %+v", issues)
	// io.WriteString(resp, "")
}

// fetch comments for a given issue
// func fetchIssueComments() {}

// convert each issue and comment to markdown
// func convertIssueToMarkdown() {}

// convert issue comment to markdown
// func convertIssueCommentToMarkdown() {}

// zipIssues will zip up a set of markdown files containing issues
// func zipIssues() {
// Define a diskcache store, can use diskCache's zip functionality.
// }

// Can we define a different search.Service based on the wueries/requests passed in? We use a different Path, define a new diskcache.Store, and don't need FetchTar:
// https://sourcegraph.sgdev.org/github.com/sourcegraph/sourcegraph@9da467a9fb25b2dee932ab55cd9569b2ac09b01c/-/blob/cmd/searcher/main.go#L55-63
