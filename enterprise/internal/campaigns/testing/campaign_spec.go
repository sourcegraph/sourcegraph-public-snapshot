package testing

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

type CreateCampaignSpecer interface {
	CreateCampaignSpec(ctx context.Context, campaignSpec *campaigns.CampaignSpec) error
}

func CreateCampaignSpec(t *testing.T, ctx context.Context, store CreateCampaignSpecer, name string, userID int32) *campaigns.CampaignSpec {
	t.Helper()

	s := &campaigns.CampaignSpec{
		UserID:          userID,
		NamespaceUserID: userID,
		Spec: campaigns.CampaignSpecFields{
			Name:        name,
			Description: "the description",
			ChangesetTemplate: campaigns.ChangesetTemplate{
				Branch: "branch-name",
			},
		},
	}

	if err := store.CreateCampaignSpec(ctx, s); err != nil {
		t.Fatal(err)
	}

	return s
}
