package types

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestRoundTripRedactExternalServiceConfig(t *testing.T) {
	// this test simulates the round trip of a user editing external service config via our APIs
	someSecret := "this is a secret, i hope no one steals it"
	cfg := schema.GitHubConnection{
		Token: someSecret,
		Repos: []string{
			"foo",
			"bar",
		},
	}
	buf, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	old := string(buf)
	// first we redact the config as it was received from the DB, then write the redacted form to the user
	svc := ExternalService{
		Kind:   extsvc.KindGitHub,
		Config: old,
	}
	if err := svc.RedactConfig(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// now we have the redacted form, we check that the relevant fields are correctly redacted
	newCfg := schema.GitHubConnection{}
	if err := json.Unmarshal([]byte(svc.Config), &newCfg); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if want, got := RedactedSecret, newCfg.Token; want != got {
		t.Errorf("want: %q, got: %q", want, got)
	}

	// now we simulate a user updating their config, and writing it back to the API containing redacted secrets
	oldSvc := ExternalService{
		Kind:   extsvc.KindGitHub,
		Config: old,
	}
	// the user added a new repo
	newCfg.Repos = append(newCfg.Repos, "baz")
	buf, err = json.Marshal(newCfg)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// this config now contains a redacted token
	newSvc := ExternalService{
		Kind:   extsvc.KindGitHub,
		Config: string(buf),
	}
	// unredact fields in newSvc config
	if err := newSvc.UnredactConfig(&oldSvc); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// the config is now safe to write to the DB, let's unmarshal it again to make sure that no fields are redacted
	// still, and that our updated fields are there
	finalCfg := schema.GitHubConnection{}
	if err := json.Unmarshal([]byte(newSvc.Config), &finalCfg); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// our updated fields are still here
	if diff := cmp.Diff(finalCfg.Repos, newCfg.Repos); diff != "" {
		t.Errorf("unexpected diff: %s", diff)
	}
	// and the secret is no longer redacted
	if want, got := someSecret, finalCfg.Token; want != got {
		t.Errorf("want: %q, got %q", want, got)
	}
}
