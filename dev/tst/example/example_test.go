package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/dev/tst"
	"github.com/sourcegraph/sourcegraph/dev/tst/config"
)

var tstCfg *config.Config

func TestRepo(t *testing.T) {
	cfg, err := config.FromFile("config.json")
	if err != nil {
		fmt.Printf("error loading scenario config: %v\n", err)
	}

	builder, err := tst.NewGitHubScenario(context.Background(), cfg, t)
	if err != nil {
		fmt.Printf("failed to create scenario: %v", err)
	}

	s := builder.Verbose().Org("tst-org").
		Users(tst.Admin, tst.User1).
		Teams(tst.Team("public-team", tst.Admin), tst.Team("private-team", tst.User1)).
		Repos(tst.PublicRepo("sgtest/go-diff", "public-team", true), tst.PrivateRepo("sgtest/private", "private-team", true))

	t.Log(s)

	ctx := context.Background()
	scenario, teardown, err := s.Setup(ctx)
	if err != nil {
		fmt.Printf("error during scenario setup: %v\n", err)
		teardown(ctx)
		os.Exit(1)
	}
	defer teardown(ctx)

	if _, err := scenario.GetOrg(); err != nil {
		t.Errorf("scenario org invalid - did the builder not populate it? %v", err)
	}
	if _, err := scenario.GetClient(); err != nil {
		t.Errorf("scenario client invalid - did the builder not populate it? %v", err)
	}

	if users := scenario.Users(); len(users) != 2 {
		t.Errorf("scenario is invalid - expected 2 users got %d", len(users))
	}
	if teams := scenario.Teams(); len(teams) != 2 {
		t.Errorf("scenario is invalid - expected 2 teams got %d", len(teams))
	}
	if repos := scenario.Repos(); len(repos) != 2 {
		t.Errorf("scenario is invalid - expected 2 repos got %d", len(repos))
	}
}
