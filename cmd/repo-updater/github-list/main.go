package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/dnaeon/go-vcr/recorder"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/internal/externalservice/github"
)

const githubDotComAPI = "https://api.github.com"

func main() {
	var (
		token = flag.String("token", "", "REQUIRED: Your github token")
		api   = flag.String("url", githubDotComAPI, "base URL of GitHub API")
	)
	flag.Parse()

	if *token == "" {
		log.Fatal("-token is required")
	}
	u, err := url.Parse(*api)
	if err != nil {
		log.Fatal(u)
	}

	name := "githublist-" + time.Now().Format("20060102150405")
	r, err := recorder.NewAsMode(name, recorder.ModeRecording, rtFunc(func(r *http.Request) (*http.Response, error) {
		// We remove the auth here to avoid recording it
		defer r.Header.Set("Authorization", "bearer XXX")
		return http.DefaultTransport.RoundTrip(r)
	}))
	if err != nil {
		log.Fatal(err)
	}
	defer fmt.Printf("\n# Recorded to %s.yaml\n", name)
	defer r.Stop()

	client := github.NewClient(u, *token, r)
	ctx := context.Background()

	var endCursor *string // GraphQL pagination cursor
	for {
		var repos []*github.Repository
		var err error
		repos, endCursor, _, err = client.ListViewerRepositories(ctx, 100, endCursor)
		if err != nil {
			log.Printf("error: %v", err)
			log.Printf("error: %+v", err)
			break
		}
		for _, r := range repos {
			fmt.Println(r.NameWithOwner)
		}
		if endCursor == nil {
			break
		}
	}
}

type rtFunc func(*http.Request) (*http.Response, error)

func (rt rtFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt(r)
}
