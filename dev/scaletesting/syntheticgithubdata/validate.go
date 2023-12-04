package main

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/sourcegraph/conc/pool"
)

// validate calculates statistics regarding orgs, teams, users, and repos on the GitHub instance.
func validate(ctx context.Context) {
	localTeams, err := store.loadTeams()
	if err != nil {
		log.Fatalf("Failed to load teams from state: %s", err)
	}

	localRepos, err := store.loadRepos()
	if err != nil {
		log.Fatalf("Failed to load repos from state: %s", err)
	}

	teamSizes := make(map[int]int)
	for _, t := range localTeams {
		users, _, err := gh.Teams.ListTeamMembersBySlug(ctx, t.Org, t.Name, &github.TeamListTeamMembersOptions{
			Role:        "member",
			ListOptions: github.ListOptions{PerPage: 100},
		})
		if err != nil {
			log.Fatal(err)
		}
		teamSizes[len(users)]++
	}

	remoteTeams := 0
	for k, v := range teamSizes {
		remoteTeams += v
		writeInfo(out, "Found %d teams with %d members", v, k)
	}

	remoteOrgs := getGitHubOrgs(ctx)
	remoteUsers := getGitHubUsers(ctx)

	writeInfo(out, "Total orgs on instance: %d", len(remoteOrgs))
	writeInfo(out, "Total teams on instance: %d", remoteTeams)
	writeInfo(out, "Total users on instance: %d", len(remoteUsers))

	p := pool.New().WithMaxGoroutines(1000)

	var orgRepoCount int64
	var teamRepoCount int64
	var userRepoCount int64

	for i, r := range localRepos {
		cI := i
		cR := r

		p.Go(func() {
			writeInfo(out, "Processing repo %d", cI)
		retryRepoContributors:
			contributors, res, err := gh.Repositories.ListContributors(ctx, cR.Owner, cR.Name, &github.ListContributorsOptions{
				Anon:        "false",
				ListOptions: github.ListOptions{},
			})
			if err != nil {
				log.Fatalf("Failed getting contributors for repo %s/%s: %s", cR.Owner, cR.Name, err)
			}
			if res != nil && (res.StatusCode == 502 || res.StatusCode == 504) {
				time.Sleep(30 * time.Second)
				goto retryRepoContributors
			}
			if len(contributors) != 0 {
				// Permissions assigned on user level
				atomic.AddInt64(&userRepoCount, 1)
				return
			}

		retryRepoTeams:
			teams, res, err := gh.Repositories.ListTeams(ctx, cR.Owner, cR.Name, &github.ListOptions{})
			if err != nil {
				log.Fatalf("Failed getting teams for repo %s/%s: %s", cR.Owner, cR.Name, err)
			}
			if res != nil && (res.StatusCode == 502 || res.StatusCode == 504) {
				time.Sleep(30 * time.Second)
				goto retryRepoTeams
			}
			if len(teams) != 0 {
				// Permissions assigned on user level
				atomic.AddInt64(&teamRepoCount, 1)
				return
			}

			// If we get this far the repo is org-wide
			atomic.AddInt64(&orgRepoCount, 1)
		})
	}
	p.Wait()

	writeInfo(out, "Total org-scoped repos: %d", orgRepoCount)
	writeInfo(out, "Total team-scoped repos: %d", teamRepoCount)
	writeInfo(out, "Total user-scoped repos: %d", userRepoCount)
}
