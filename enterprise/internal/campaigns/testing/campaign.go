package testing

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

type CreateCampaigner interface {
	CreateCampaign(ctx context.Context, campaign *campaigns.Campaign) error
	Clock() func() time.Time
}

func CreateCampaign(t *testing.T, ctx context.Context, store CreateCampaigner, name string, userID int32, spec int64) *campaigns.Campaign {
	t.Helper()

	c := &campaigns.Campaign{
		InitialApplierID: userID,
		LastApplierID:    userID,
		LastAppliedAt:    store.Clock()(),
		NamespaceUserID:  userID,
		CampaignSpecID:   spec,
		Name:             name,
		Description:      "campaign description",
	}

	if err := store.CreateCampaign(ctx, c); err != nil {
		t.Fatal(err)
	}

	return c
}
