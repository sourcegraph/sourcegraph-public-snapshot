package main

import (
	"context"
	"log"
	"math"

	"github.com/google/go-github/v41/github"

	"github.com/sourcegraph/sourcegraph/lib/group"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func create(ctx context.Context, orgs []*org, cfg config) {
	var err error

	// load or generate users
	var users []*user
	if users, err = store.loadUsers(); err != nil {
		log.Fatalf("Failed to load users from state: %s", err)
	}

	if len(users) == 0 {
		if users, err = store.generateUsers(cfg); err != nil {
			log.Fatalf("Failed to generate users: %s", err)
		}
		writeSuccess(out, "generated user jobs in %s", cfg.resume)
	} else {
		writeSuccess(out, "resuming user jobs from %s", cfg.resume)
	}

	// load or generate teams
	var teams []*team
	if teams, err = store.loadTeams(); err != nil {
		log.Fatalf("Failed to load teams from state: %s", err)
	}

	if len(teams) == 0 {
		if teams, err = store.generateTeams(cfg); err != nil {
			log.Fatalf("Failed to generate teams: %s", err)
		}
		writeSuccess(out, "generated team jobs in %s", cfg.resume)
	} else {
		writeSuccess(out, "resuming team jobs from %s", cfg.resume)
	}

	var repos []*repo
	if repos, err = store.loadRepos(); err != nil {
		log.Fatalf("Failed to load repos from state: %s", err)
	}

	if len(repos) == 0 {
		remoteRepos := getGitHubRepos(ctx)

		if repos, err = store.insertRepos(remoteRepos); err != nil {
			log.Fatalf("Failed to insert repos in state: %s", err)
		}
		writeSuccess(out, "Fetched %d private repos and stored in state", len(remoteRepos))
	} else {
		writeSuccess(out, "resuming repo jobs from %s", cfg.resume)
	}

	bars := []output.ProgressBar{
		{Label: "Creating orgs", Max: float64(cfg.subOrgCount + 1)},
		{Label: "Creating teams", Max: float64(cfg.teamCount)},
		{Label: "Creating users", Max: float64(cfg.userCount)},
		{Label: "Adding users to teams", Max: float64(cfg.userCount)},
		{Label: "Assigning repos", Max: float64(len(repos))},
	}
	if cfg.generateTokens {
		bars = append(bars, output.ProgressBar{Label: "Generating OAuth tokens", Max: float64(cfg.userCount)})
	}
	progress = out.Progress(bars, nil)
	var usersDone int64
	var orgsDone int64
	var teamsDone int64
	var tokensDone int64
	var membershipsDone int64
	var reposDone int64

	g := group.New().WithMaxConcurrency(1000)

	for _, o := range orgs {
		currentOrg := o
		g.Go(func() {
			executeCreateOrg(ctx, currentOrg, cfg.orgAdmin, &orgsDone)
		})
	}
	g.Wait()

	// Default permission is "read"; we need members to not have access by default on the main organisation.
	defaultRepoPermission := "none"
	var res *github.Response
	_, res, err = gh.Organizations.Edit(ctx, "main-org", &github.Organization{DefaultRepoPermission: &defaultRepoPermission})
	if err != nil && res.StatusCode != 409 {
		// 409 means the base repo permissions are currently being updated already due to a previous run
		log.Fatalf("Failed to make main-org private by default: %s", err)
	}

	for _, t := range teams {
		currentTeam := t
		g.Go(func() {
			executeCreateTeam(ctx, currentTeam, &teamsDone)
		})
	}
	g.Wait()

	for _, u := range users {
		currentUser := u
		g.Go(func() {
			executeCreateUser(ctx, currentUser, &usersDone)
		})
	}
	g.Wait()

	membershipsPerTeam := int(math.Ceil(float64(cfg.userCount) / float64(cfg.teamCount)))
	g2 := group.New().WithMaxConcurrency(100)

	for i, t := range teams {
		currentTeam := t
		currentIter := i
		var usersToAssign []*user

		for j := currentIter * membershipsPerTeam; j < ((currentIter + 1) * membershipsPerTeam); j++ {
			usersToAssign = append(usersToAssign, users[j])
		}

		g2.Go(func() {
			executeCreateTeamMembershipsForTeam(ctx, currentTeam, usersToAssign, &membershipsDone)
		})
	}
	g2.Wait()

	mainOrg, orgRepos := categorizeOrgRepos(cfg, repos, orgs)
	executeAssignOrgRepos(ctx, orgRepos, users, &reposDone, g2)
	g2.Wait()

	// 0.5% repos with only users attached
	amountReposWithOnlyUsers := int(math.Ceil(float64(len(repos)) * 0.005))
	reposWithOnlyUsers := orgRepos[mainOrg][:amountReposWithOnlyUsers]
	// slice out the user repos
	orgRepos[mainOrg] = orgRepos[mainOrg][amountReposWithOnlyUsers:]

	teamRepos := categorizeTeamRepos(cfg, orgRepos[mainOrg], teams)
	userRepos := categorizeUserRepos(reposWithOnlyUsers, users)

	executeAssignTeamRepos(ctx, teamRepos, &reposDone, g2)
	g2.Wait()

	executeAssignUserRepos(ctx, userRepos, &reposDone, g2)
	g2.Wait()

	if cfg.generateTokens {
		generateUserOAuthCsv(ctx, users, tokensDone)
	}
}
