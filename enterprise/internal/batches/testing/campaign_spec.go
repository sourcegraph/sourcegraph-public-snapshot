package testing

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/batches"
)

type CreateCampaignSpecer interface {
	CreateCampaignSpec(ctx context.Context, campaignSpec *batches.CampaignSpec) error
}

func CreateCampaignSpec(t *testing.T, ctx context.Context, store CreateCampaignSpecer, name string, userID int32) *batches.CampaignSpec {
	t.Helper()

	s := &batches.CampaignSpec{
		UserID:          userID,
		NamespaceUserID: userID,
		Spec: batches.CampaignSpecFields{
			Name:        name,
			Description: "the description",
			ChangesetTemplate: batches.ChangesetTemplate{
				Branch: "branch-name",
			},
		},
	}

	if err := store.CreateCampaignSpec(ctx, s); err != nil {
		t.Fatal(err)
	}

	return s
}
