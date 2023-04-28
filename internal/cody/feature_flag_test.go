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
	t.Run("Unauthenticated user", func(t *testing.T) {
		ctx := context.Background()
		ctx = actor.WithActor(ctx, &actor.Actor{UID: 0})
		if IsCodyEnabled(ctx) {
			t.Error("Expected IsCodyEnabled to return false for unauthenticated actor")
		}
	})

	t.Run("Authenticated user", func(t *testing.T) {
		ctx := context.Background()
		ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
		if !IsCodyEnabled(ctx) {
			t.Error("Expected IsCodyEnabled to return true for authenticated actor")
		}
	})

	t.Run("CodyRestrictUsersFeatureFlag", func(t *testing.T) {
		t.Run("feature flag disabled", func(t *testing.T) {
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

			ctx := context.Background()
			ctx = actor.WithActor(ctx, &actor.Actor{UID: 0})
			if IsCodyEnabled(ctx) {
				t.Error("Expected IsCodyEnabled to return false for unauthenticated user with CodyRestrictUsersFeatureFlag enabled")
			}
			ctx = context.Background()
			ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
			if IsCodyEnabled(ctx) {
				t.Error("Expected IsCodyEnabled to return false for authenticated user when CodyRestrictUsersFeatureFlag is set and no feature flag is present for the user")
			}
		})
		t.Run("feature flag enabled", func(t *testing.T) {
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

			ctx := context.Background()
			ctx = featureflag.WithFlags(ctx, featureflag.NewMemoryStore(map[string]bool{"cody-experimental": true}, map[string]bool{"cody-experimental": true}, nil))
			ctx = actor.WithActor(ctx, &actor.Actor{UID: 0})
			if IsCodyEnabled(ctx) {
				t.Error("Expected IsCodyEnabled to return false when cody-experimental feature flag is enabled")
			}
			ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
			if !IsCodyEnabled(ctx) {
				t.Error("Expected IsCodyEnabled to return true when cody-experimental feature flag is enabled")
			}
		})
	})
}
