package conf

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

func TestPropertyActions(t *testing.T) {
	// Set up a fake schema that covers the required scenarios.
	schema := configPropertyActionSchema{
		"experimentalFeatures::automation": {FrontendReloadRequired: true},
		"externalURL":                      {FrontendReloadRequired: true, ServerRestartRequired: true},
		"email.address":                    {},
		"licenseKey":                       {ServerRestartRequired: true},
	}

	// Set up an empty configuration to use as the "before" in our tests.
	empty, err := ParseConfig(conftypes.RawUnified{
		Site: `{}`,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("empty configurations", func(t *testing.T) {
		have := needActionToApply(empty, empty, schema)
		want := PostConfigWriteActions{FrontendReloadRequired: false, ServerRestartRequired: false}
		if diff := cmp.Diff(have, want); diff != "" {
			t.Fatalf("unexpected actions: %s", diff)
		}
	})

	t.Run("no action required", func(t *testing.T) {
		config, err := ParseConfig(conftypes.RawUnified{
			Site: `{"email.address":"foo@bar.quux"}`,
		})
		if err != nil {
			t.Fatal(err)
		}

		have := needActionToApply(empty, config, schema)
		want := PostConfigWriteActions{FrontendReloadRequired: false, ServerRestartRequired: false}
		if diff := cmp.Diff(have, want); diff != "" {
			t.Fatalf("unexpected actions: %s", diff)
		}
	})

	t.Run("frontend reload required", func(t *testing.T) {
		config, err := ParseConfig(conftypes.RawUnified{
			Site: `{"email.address":"foo@bar.quux", "experimentalFeatures": {"automation": "enabled"}}`,
		})
		if err != nil {
			t.Fatal(err)
		}

		have := needActionToApply(empty, config, schema)
		want := PostConfigWriteActions{FrontendReloadRequired: true, ServerRestartRequired: false}
		if diff := cmp.Diff(have, want); diff != "" {
			t.Fatalf("unexpected actions: %s", diff)
		}
	})

	t.Run("server restart required", func(t *testing.T) {
		config, err := ParseConfig(conftypes.RawUnified{
			Site: `{"email.address":"foo@bar.quux", "licenseKey": "foo"}`,
		})
		if err != nil {
			t.Fatal(err)
		}

		have := needActionToApply(empty, config, schema)
		want := PostConfigWriteActions{FrontendReloadRequired: false, ServerRestartRequired: true}
		if diff := cmp.Diff(have, want); diff != "" {
			t.Fatalf("unexpected actions: %s", diff)
		}
	})

	t.Run("frontend reload and server restart required", func(t *testing.T) {
		config, err := ParseConfig(conftypes.RawUnified{
			Site: `{"email.address":"foo@bar.quux", "experimentalFeatures": {"automation": "enabled"}, "licenseKey": "foo"}`,
		})
		if err != nil {
			t.Fatal(err)
		}

		have := needActionToApply(empty, config, schema)
		want := PostConfigWriteActions{FrontendReloadRequired: true, ServerRestartRequired: true}
		if diff := cmp.Diff(have, want); diff != "" {
			t.Fatalf("unexpected actions: %s", diff)
		}
	})
}
