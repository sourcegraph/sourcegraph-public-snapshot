package main

import (
	"context"
	"log"
	"net/url"

	"cuelang.org/go/cue/errors"
)

type CodeHost interface {
	ListRepos(ctx context.Context, path string) ([]url.URL, error)
	CreateRepo(ctx context.Context, path string) (url.URL, error)
}

func main() {
	cfg, err := loadConfig("dev/scaletesting/codehostcopy/config.cue")
	if err != nil {
		var cueErr errors.Error
		if errors.As(err, &cueErr) {
			log.Print(errors.Details(err, nil))
		}
		log.Fatal(err)
	}

	println(cfg.From.Kind)

	// ctx := context.Background()
	// tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
	// 	&oauth2.Token{AccessToken: cfg.githubToken},
	// ))
	//
	// gh, err := github.NewEnterpriseClient(cfg.githubURL, cfg.githubURL, tc)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// repos := []*github.Repository{}
	// if cfg.githubOrg != "" {
	// 	opts := github.RepositoryListByOrgOptions{
	// 		ListOptions: github.ListOptions{},
	// 	}
	// 	// If we don't have an org, we want to get the user repositories instead.
	// 	for {
	// 		rs, resp, err := gh.Repositories.ListByOrg(ctx, cfg.githubOrg, &opts)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		repos = append(repos, rs...)
	//
	// 		if resp.NextPage == 0 {
	// 			break
	// 		}
	// 		opts.ListOptions.Page = resp.NextPage
	// 	}
	// } else {
	// 	for {
	// 		rs, _, err := gh.Repositories.ListAll(ctx, nil)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		repos = append(repos, rs...)
	// 	}
	// }
	//
	// println(len(repos))
}
