package localstore

import (
	"testing"

	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
)

func TestPaywallLogic(t *testing.T) {
	username := "badActor12"
	ctx := auth.WithActor(testContext(), &auth.Actor{Login: username})
	p := payments{}

	// Test public repos aren't blocked:
	publicRepo := sourcegraph.Repo{Private: false}
	err := p.CheckPaywallForRepo(ctx, publicRepo)
	if err != nil {
		t.Errorf("Public repos shouldn't be blocked, but got error %s.", err)
	}

	// Test personal private repos aren't blocked:
	ppRepo := sourcegraph.Repo{Private: true, Owner: username}
	err = p.CheckPaywallForRepo(ctx, ppRepo)
	if err != nil {
		t.Errorf("Personal private repos shouldn't be blocked, but got error %s.", err)
	}

	// Test on a blocked org:
	_, err = appDBH(ctx).Db.Query("INSERT INTO "+orgTableName+" (org_name, plan) VALUES ($1, $2);", "someOrg", string(Blocked))
	if err != nil {
		t.Fatalf("Unexected error: %s", err)
	}

	orgRepo := sourcegraph.Repo{Private: true, Owner: "someOrg"}
	err = p.CheckPaywallForRepo(ctx, orgRepo)
	if err != (ErrBlocked{}) {
		t.Errorf("Org private repos should be blocked, but got %s.", err)
	}
}

func TestExpirationLogic(t *testing.T) {
	username := "badActor12"
	ctx := auth.WithActor(testContext(), &auth.Actor{Login: username})
	p := payments{}

	repo := sourcegraph.Repo{Private: true, Owner: username}

	date, err := p.TrialExpirationDate(ctx, repo)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if date != nil {
		t.Errorf("Should have gotten nil time for repo not on trial, got %v", date)
	}

	// Expired repos are blocked:
	_, err = appDBH(ctx).Db.Query("INSERT INTO "+orgTableName+" (org_name, trial_expiration, plan) VALUES ($1, $2, $3);", "someOrg", time.Now().UTC(), string(None))
	if err != nil {
		t.Fatal(err)
	}

	orgRepo := sourcegraph.Repo{Private: true, Owner: "someOrg"}
	err = p.CheckPaywallForRepo(ctx, orgRepo)
	if err != (ErrTrialExpired{}) {
		t.Errorf("Org private repos should be blocked, but got %s.", err)
	}
}
