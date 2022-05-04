package ldfeatureflag

import (
	"context"
	"fmt"
	"time"

	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
	ld "gopkg.in/launchdarkly/go-server-sdk.v5"
	"gopkg.in/launchdarkly/go-server-sdk.v5/interfaces/flagstate"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

const sdkKey = "sdk-2d5b0e1b-b407-4af8-a608-77a4b612f563"
const isDotCom = true

var ldClient, err = ld.MakeCustomClient(sdkKey, ld.Config{Offline: !isDotCom}, 5*time.Second)

// List all feature flags enabled for the user
func AllEnabledFeatureFlags(ctx context.Context, db database.DB) []string {
	flagsMap := ldClient.AllFlagsState(buildVisitorFromContext(ctx, db), flagstate.OptionDetailsOnlyForTrackedFlags()).ToValuesMap()

	var enabledFlags []string

	for flag, value := range flagsMap {
		if value.BoolValue() == true {
			enabledFlags = append(enabledFlags, flag)
		}

	}

	return enabledFlags
}

type FeatureFlag struct {
	Name         string
	DefaultValue bool
}

// Check if the feature flag is enabled for the user
func (f *FeatureFlag) IsEnabledFor(ctx context.Context, db database.DB) bool {
	enabled, err := ldClient.BoolVariation(f.Name, buildVisitorFromContext(ctx, db), f.DefaultValue)

	if err != nil {
		return f.DefaultValue
	}

	return enabled
}

type Experiment struct {
	Name string
}

type Variant struct {
	Value string
}

// Find variant for the experiment for the user
func (e *Experiment) VariantFor(ctx context.Context, db database.DB) *Variant {
	variant, err := ldClient.StringVariation(e.Name, buildVisitorFromContext(ctx, db), "")

	if err != nil {
		return nil
	}

	return &Variant{Value: variant}
}

// List all the already active experiments for the user
func AllParticipatingExperiments(ctx context.Context) {

}

func buildVisitorFromContext(ctx context.Context, db database.DB) lduser.User {
	a := actor.FromContext(ctx)

	if a.IsInternal() {
		return lduser.NewUserBuilder("internal_sourcegraph_service").Build()
	}

	if a.IsAuthenticated() {
		user, err := a.User(ctx, db.Users())
		if err != nil {
			//raise error
		}

		orgs, err := db.Orgs().GetByUserID(ctx, user.ID)
		if err != nil {
			//raise error
		}

		org_ids := make([]ldvalue.Value, 0, len(orgs))

		for _, org := range orgs {
			org_ids = append(org_ids, ldvalue.String(fmt.Sprintf("%v", org.ID)))
		}

		return lduser.NewUserBuilder(fmt.Sprintf("uid_%v", user.ID)).
			Name(user.DisplayName).
			Custom("username", ldvalue.String(user.Username)).
			Custom("site_admin", ldvalue.Bool(user.SiteAdmin)).
			Custom("org_ids", ldvalue.ArrayOf(org_ids...)).
			Build()
		// TODO add source graph team mate bool prop

	}

	return lduser.NewUserBuilder(fmt.Sprintf("aid_%s", a.AnonymousUID)).Build()
}
