package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	scenario "github.com/sourcegraph/sourcegraph/dev/codehost_scenario"
	"github.com/sourcegraph/sourcegraph/dev/codehost_scenario/config"
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
	scenario, err := scenario.NewGithubScenario(context.Background(), t, *cfg)
	if err != nil {
		t.Fatalf("error creating scenario: %v\n", err)
	}
	org := scenario.CreateOrg("tst-org")
	user := scenario.CreateUser("tst-user")
	admin := scenario.GetAdmin()

	//ctx := context.Background()

	org.AllowPrivateForks()
	team := org.CreateTeam("team-1")
	team.AddUser(user)
	adminTeam := org.CreateTeam("team-admin")
	adminTeam.AddUser(admin)

	publicRepo := org.CreateRepoFork("sgtest/go-diff")
	publicRepo.AddTeam(team)
	privateRepo := org.CreateRepoFork("sgtest/private")
	privateRepo.AddTeam(adminTeam)

	fmt.Println(scenario.Plan())

	scenario.Verbose()
	if err := scenario.Apply(context.Background()); err != nil {
		t.Fatalf("error applying scenario: %v", err)
	}
}

func TestMain(m *testing.M) {
	flag.BoolVar(&runGithub, "run.github", false, "Run example github scenario setup")
	flag.Parse()
	os.Exit(m.Run())
}
