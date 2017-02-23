package localstore

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
)

func TestPaywallLogic(t *testing.T) {
	username := "badActor12"
	ctx := auth.WithActor(testContext(), &auth.Actor{Login: username})
	p := payments{}

	// Test public repos aren't blocked:
	publicRepo := sourcegraph.Repo{Private: false}
	err := p.CheckPaywallForRepo(ctx, &publicRepo)
	if err != nil {
		t.Errorf("Public repos shouldn't be blocked, but got error %s.", err)
	}

	// Test personal private repos aren't blocked:
	ppRepo := sourcegraph.Repo{Private: true, Owner: username}
	err = p.CheckPaywallForRepo(ctx, &ppRepo)
	if err != nil {
		t.Errorf("Personal private repos shouldn't be blocked, but got error %s.", err)
	}

	// Test on a blocked org:
	err = p.BlockOrg(ctx, "someOrg")
	if err != nil {
		t.Fatal(err)
	}

	orgRepo := sourcegraph.Repo{Private: true, Owner: "someOrg"}
	err = p.CheckPaywallForRepo(ctx, &orgRepo)
	if err != (ErrBlocked{}) {
		t.Errorf("Org private repos should be blocked, but got %s.", err)
	}
}
