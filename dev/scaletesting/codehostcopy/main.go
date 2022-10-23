package main

import (
	"context"
	"log"
	"net/url"

	"cuelang.org/go/cue/errors"
)

type CodeHostSource interface {
	ListRepos(ctx context.Context) ([]*Repo, error)
}

type CodeHostDestination interface {
	CreateRepo(ctx context.Context, name string) (*url.URL, error)
}

type Repo struct {
	url  string
	name string
}

func doBitbucket(ctx context.Context, cfg *Config) {
	bt := NewBitbucketCodeHost(ctx, &cfg.From)
	_, err := bt.ListRepos(ctx)
	if err != nil {
		log.Fatal(err)
	}

	repo, err := bt.CreateRepo(ctx, "willdor::test")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v", repo.String())
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
	gl, err := NewGitLabCodeHost(ctx, &cfg.Destination)
	if err != nil {
		log.Fatal(err)
	}
	//
	runner := NewRunner(gh, gl)
	if err := runner.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
