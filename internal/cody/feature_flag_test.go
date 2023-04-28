package cody

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestIsCodyEnabled(t *testing.T) {
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 0})
	if IsCodyEnabled(ctx) {
		t.Error("Expected IsCodyEnabled to return false for unauthenticated actor")
	}

	ctx = context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
	if !IsCodyEnabled(ctx) {
		t.Error("Expected IsCodyEnabled to return true for authenticated actor")
	}

	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				CodyRestrictUsersFeatureFlag: true,
			},
		},
	})
	t.Cleanup(func() {
		conf.Mock(nil)
	})

	ctx = context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 0})
	if IsCodyEnabled(ctx) {
		t.Error("Expected IsCodyEnabled to return false for unauthenticated user with CodyRestrictUsersFeatureFlag enabled")
	}
	ctx = context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
	if IsCodyEnabled(ctx) {
		t.Error("Expected IsCodyEnabled to return false for authenticated user when CodyRestrictUsersFeatureFlag is set and no feature flag is present for the user")
	}

	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				CodyRestrictUsersFeatureFlag: true,
			},
		},
	})
	ctx = context.Background()
	ctx = featureflag.WithFlags(ctx, featureflag.NewMemoryStore(map[string]bool{"cody-experimental": true}, nil, nil))
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
	if !IsCodyEnabled(ctx) {
		t.Error("Expected IsCodyEnabled to return true when cody-experimental feature flag is enabled")
	}
}
