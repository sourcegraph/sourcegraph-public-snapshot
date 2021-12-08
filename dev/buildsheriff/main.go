package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/google/go-github/v31/github"
	"golang.org/x/oauth2"
)

func main() {
	var (
		ctx            = context.Background()
		buildkiteToken string
		githubToken    string
		pipeline       string
		branch         string
		threshold      int
	)

	flag.StringVar(&buildkiteToken, "buildkite.token", "", "mandatory buildkite token")
	flag.StringVar(&githubToken, "github.token", "", "mandatory github token")
	flag.StringVar(&pipeline, "pipeline", "sourcegraph", "name of the pipeline to inspect")
	flag.StringVar(&branch, "branch", "main", "name of the branch to inspect")
	flag.IntVar(&threshold, "threshold", 3, "failures required to trigger an incident")

	config, err := buildkite.NewTokenConfig(buildkiteToken, false)
	if err != nil {
		panic(err)
	}
	// Buildkite client
	bkc := buildkite.NewClient(config.Client())

	// GitHub client
	ghc := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)))

	// Newest is returned first https://buildkite.com/docs/apis/rest-api/builds#list-builds-for-a-pipeline
	builds, _, err := bkc.Builds.ListByPipeline("sourcegraph", pipeline, &buildkite.BuildsListOptions{
		Branch: "main",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Scan for first build with a meaningful state
	var firstFailedBuild int
	for i, b := range builds {
		if b.State != nil {
			if *b.State == "passed" {
				fmt.Printf("most recent finished build %d passed\n", *b.Number)
				if err := unlockBranch(ctx, ghc, branch); err != nil {
					log.Fatal(err)
				}
				return
			}
			if *b.State == "failed" {
				fmt.Printf("most recent finished build %d failed\n", *b.Number)
				firstFailedBuild = i
				break
			}

			// Otherwise, keep looking for builds
		}
	}

	// if failed, check if failures are consecutive
	failureAuthorsEmails, exceeded := checkConsecutiveFailures(builds[firstFailedBuild:], threshold)
	if !exceeded {
		if err := unlockBranch(ctx, ghc, branch); err != nil {
			log.Fatal(err)
		}
		return
	}

	fmt.Println("threshold exceeded, this is a big deal!")
	if err := lockBranch(ctx, ghc, branch, failureAuthorsEmails); err != nil {
		log.Fatal(err)
	}
}

func checkConsecutiveFailures(builds []buildkite.Build, threshold int) (authorsEmails []string, thresholdExceeded bool) {
	var consecutiveFailures int
	for _, b := range builds {
		if b.State == nil || *b.State == "" {
			continue
		}

		// if a build in the sequence passed, no big deal
		if *b.State == "passed" {
			fmt.Printf("build %d passed\n", *b.Number)
			return authorsEmails, false
		}

		if *b.State == "failed" {
			consecutiveFailures += 1
			authorsEmails = append(authorsEmails, b.Author.Email)
			fmt.Printf("build %d is %dth consecutive failure\n", *b.Number, consecutiveFailures)
			if consecutiveFailures > threshold {
				break
			}
		}
	}

	// If we get this far we've found a sufficient sequence of failed builds
	return authorsEmails, true
}

func lockBranch(ctx context.Context, ghc *github.Client, branch string, failureAuthorsEmails []string) error {
	users, _, err := ghc.Search.Users(ctx, strings.Join(failureAuthorsEmails, " OR "), &github.SearchOptions{})
	if err != nil {
		return err
	}

	var failureAuthorsUsers []string
	for _, u := range users.Users {
		// Make sure this user is in the Sourcegraph org
		membership, _, err := ghc.Organizations.GetOrgMembership(ctx, *u.Login, "sourcegraph")
		if err != nil {
			return err
		}
		if membership == nil || *membership.State != "active" {
			continue // we don't want this user
		}

		failureAuthorsUsers = append(failureAuthorsUsers, *u.Login)
	}

	restrictions := &github.BranchRestrictionsRequest{
		Users: failureAuthorsUsers,
		Teams: []string{"dev-experience"},
	}
	fmt.Printf("restricting push access to %q to %+v", branch, restrictions)
	_, _, err = ghc.Repositories.UpdateBranchProtection(ctx, "sourcegraph", "sourcegraph", branch, &github.ProtectionRequest{
		Restrictions: restrictions,
	})
	if err != nil {
		return err
	}

	return nil
}

func unlockBranch(ctx context.Context, ghc *github.Client, branch string) error {
	_, _, err := ghc.Repositories.UpdateBranchProtection(ctx, "sourcegraph", "sourcegraph", branch, &github.ProtectionRequest{
		Restrictions: &github.BranchRestrictionsRequest{
			Users: []string{},
			Teams: []string{},
		},
	})
	return err
}
