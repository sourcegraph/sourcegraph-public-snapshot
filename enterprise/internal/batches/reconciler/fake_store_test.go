package reconciler

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/batches"
)

type mockMissingErr struct {
	mockName string
}

func (e mockMissingErr) Error() string {
	return fmt.Sprintf("FakeStore is missing mock for %s", e.mockName)
}

type FakeStore struct {
	GetCampaignMock func(context.Context, store.GetCampaignOpts) (*batches.Campaign, error)
}

func (fs *FakeStore) GetCampaign(ctx context.Context, opts store.GetCampaignOpts) (*batches.Campaign, error) {
	if fs.GetCampaignMock != nil {
		return fs.GetCampaignMock(ctx, opts)
	}
	return nil, mockMissingErr{"GetCampaign"}
}
