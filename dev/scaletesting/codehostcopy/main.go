package main

import (
	"context"
	"log"
	"net/url"

	"cuelang.org/go/cue/errors"
)

type CodeHostSource interface {
	ListRepos(ctx context.Context) ([]*url.URL, error)
}

type CodeHostDestination interface {
	CreateRepo(ctx context.Context, name string) (*url.URL, error)
}

type Repo struct {
	url  string
	name string
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

	gl, err := NewGitLabCodeHost(ctx, &cfg.Destination)
	if err != nil {
		log.Fatal(err)
	}

	project, err := gl.CreateRepo(ctx, "fooz")
	if err != nil {
		log.Fatal(err)
	}

	println(project.String())

	// gh, err := NewGithubCodeHost(ctx, &cfg.From)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	//
	// res, err := gh.ListRepos(ctx)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// for _, r := range res {
	// 	println(r.name)
	// }
}
