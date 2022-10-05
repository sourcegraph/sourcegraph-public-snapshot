package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
)

type config struct {
	Origin            string
	GitHubToken       string
	GitHubDotComToken string
	RepoCount         int
	Org               string
	GitHubInstanceURL string
}

type originImport struct {
	name string
	imp  *github.Import
}

var cfg config

func init() {
	flag.StringVar(&cfg.Origin, "origin", "wp-plugins", "github.com/[origin] to import repositories from")
	flag.StringVar(&cfg.GitHubToken, "github.token", "", "mandatory GitHub token for the given instance")
	flag.StringVar(&cfg.GitHubDotComToken, "github-dotcom.token", "", "mandatory GitHub.com token, used to fetch the repositories to import")
	flag.IntVar(&cfg.RepoCount, "repo-count", 10000, "number of repositories to import")
	flag.StringVar(&cfg.Org, "github.org", "", "mandatory GitHub organization where the repos will be imported")
	flag.StringVar(&cfg.GitHubInstanceURL, "github.url", "https://ghe.github.org", "URl of the GitHub instance to import the repositories in")
}

func main() {
	flag.Parse()

	if cfg.GitHubToken == "" {
		flag.Usage()
		log.Fatal("need a GitHub token")
	}
	if cfg.GitHubDotComToken == "" {
		flag.Usage()
		log.Fatal("need a GitHub.com token")
	}
	if cfg.Org == "" {
		flag.Usage()
		log.Fatal("need an org name for GitHub")
	}
	if cfg.GitHubInstanceURL == "" {
		flag.Usage()
		log.Fatal("need an instance URL for GitHub")
	}
	if cfg.Origin == "" {
		flag.Usage()
		log.Fatal("need an origin")
	}

	// Create a GitHub client.
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GitHubDotComToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Fetch some repos.
	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{Page: 0},
	}

	var originRepos []*github.Repository
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, cfg.Org, opts)
		if err != nil {
			log.Fatal(err)
		}
		originRepos = append(originRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
	}
	fmt.Printf("Found %d repositories on https://github.com/%s\n", len(originRepos), cfg.Org)

	// Create a client for the target instance
	ts = oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GitHubToken},
	)
	tc = oauth2.NewClient(ctx, ts)
	client = github.NewClient(tc)

	// Create imports.
	for i := 0; i < cfg.RepoCount; i++ {
		repo := originRepos[i]
		_, _, err := client.Migrations.StartImport(ctx, cfg.Org, *repo.Name, &github.Import{
			VCS:    github.String("git"),
			VCSURL: github.String(fmt.Sprintf("https://github.com/google/%s.git", *repo.Name)),
		})
		if err != nil {
			fmt.Println("failed to create import for", repo)
			log.Fatal(err)
		}
	}
	fmt.Printf("Created %d imports", cfg.RepoCount)

	// Poll the import API for the statuses.
	var wg sync.WaitGroup
	wg.Add(cfg.RepoCount)

	remaining := len(originRepos)
	for _, repo := range originRepos {
		go func(repo *github.Repository) {
			ticker := time.NewTicker(30 * time.Second)
			for {
				select {
				case <-ticker.C:
					imp, _, err := client.Migrations.ImportProgress(ctx, cfg.Org, *repo.Name)
					if err != nil {
						fmt.Println("failed to get import status for", repo.Name)
						log.Fatal(err)
					}
					if *imp.Status == "complete" {
						fmt.Println("import completed for", repo.Name)
						remaining--
						fmt.Println(remaining, "repositories to import")
						wg.Done()
						break
					}
				}
			}
		}(repo)
	}

	fmt.Println("done")
}
