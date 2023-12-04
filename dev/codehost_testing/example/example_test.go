package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	scenario "github.com/sourcegraph/sourcegraph/dev/codehost_testing"
	"github.com/sourcegraph/sourcegraph/dev/codehost_testing/config"
)

var runGithub bool

func TestGithubScenario(t *testing.T) {
	if !runGithub {
		t.Skip("`run.github` flag not provided - skipping github scenario as this is only an example test")
	}
	cfg, err := config.FromFile("config.json")
	if err != nil {
		t.Fatalf("error loading scenario config: %v\n", err)
	}
	scenario, err := scenario.NewGitHubScenario(t, *cfg)
	if err != nil {
		t.Fatalf("error creating scenario: %v\n", err)
	}
	org := scenario.CreateOrg("tst-org")
	user := scenario.CreateUser("tst-user")
	otherUser := scenario.CreateUser("other-user")
	admin := scenario.GetAdmin()

	org.AllowPrivateForks()
	team := org.CreateTeam("team-1")
	team.AddUser(user)
	team.AddUser(otherUser)
	adminTeam := org.CreateTeam("team-admin")
	adminTeam.AddUser(admin)

	publicRepo := org.CreateRepoFork("sgtest/go-diff")
	publicRepo.AddTeam(team)
	privateRepo := org.CreateRepoFork("sgtest/private")
	privateRepo.AddTeam(adminTeam)

	fmt.Println(scenario.Plan())
	ctx := context.Background()
	// Get the Organization WILL FAIL since the scenario has not been applied yet
	_, err = org.Get(ctx)
	if err != nil {
		t.Logf("failed to get github.Organization since it hasn't been applied yet: %v", err)
	}

	scenario.SetVerbose()
	if err := scenario.Apply(ctx); err != nil {
		t.Fatalf("error applying scenario: %v", err)
	}

	// Get the Organization
	ghOrg, err := org.Get(ctx)
	if err != nil {
		t.Fatalf("failed to get github.Organization: %v", err)
	}

	t.Logf("GitHub Organization: %s", ghOrg.GetLogin())
}

func TestMain(m *testing.M) {
	flag.BoolVar(&runGithub, "run.github", false, "Run example github scenario setup")
	flag.Parse()
	os.Exit(m.Run())
}
