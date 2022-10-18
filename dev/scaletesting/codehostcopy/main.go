package main

import (
	"context"
	"log"
	"net/url"

	"cuelang.org/go/cue/errors"
)

type CodeHost interface {
	ListRepos(ctx context.Context) ([]*url.URL, error)
	// CreateRepo(ctx context.Context, path string) (*url.URL, error)
}

func main() {
	ctx := context.Background()
	cfg, err := loadConfig("dev/scaletesting/codehostcopy/config.cue")
	if err != nil {
		var cueErr errors.Error
		if errors.As(err, &cueErr) {
			log.Print(errors.Details(err, nil))
		}
		log.Fatal(err)
	}

	gh, err := NewGithubCodeHost(ctx, &cfg.From)
	if err != nil {
		log.Fatal(err)
	}

	res, err := gh.ListRepos(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, u := range res {
		println(u.String())
	}
}
