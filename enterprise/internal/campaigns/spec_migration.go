package campaigns

// This file contains methods that exist solely to migrate campaigns and
// changesets lingering from before specs were added in Sourcegraph 3.19 into
// the new world.
//
// It should be removed in or after Sourcegraph 3.21.

import (
	"context"
	"encoding/json"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

func (svc *Service) MigratePreSpecCampaigns(ctx context.Context) (err error) {
	log15.Info("migrating campaigns created before 3.19")

	// We want to do all this in a transaction, so let's get that set up. We
	// also need a service instance that uses the transactional store.
	store, err := svc.store.Transact(ctx)
	if err != nil {
		err = errors.Wrap(err, "beginning transaction")
		return
	}
	defer func() { store.Done(err) }()
	//txSvc := NewServiceWithClock(store, svc.cf, svc.clock)

	// Create changeset specs.

	return
}
