package campaigns

// This file contains methods that exist solely to migrate campaigns and
// changesets lingering from before specs were added in Sourcegraph 3.19 into
// the new world.
//
// It should be removed in or after Sourcegraph 3.21.

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

func (svc *Service) MigratePreSpecCampaigns(ctx context.Context) (err error) {
	log15.Info("migrating campaigns created before 3.19")

	// Since this is a backend migration, we're yolo-ing the authentication checks.
	ctx = backend.WithAuthzBypass(ctx)

	// We want to do all this in a transaction, so let's get that set up. We
	// also need a service instance that uses the transactional store.
	store, err := svc.store.Transact(ctx)
	if err != nil {
		err = errors.Wrap(err, "beginning transaction")
		return
	}

	// Basically: if err is not nil, we'll rollback at the end of the function,
	// otherwise we commit.
	defer func() { store.Done(err) }()

	// We also need a service instance that uses the transaction we just began.
	txSvc := NewServiceWithClock(store, svc.cf, svc.clock)

	// We're going to need the old campaigns in a few places, so let's just get
	// them ready up front.
	cmpgns, err := store.listPreSpecCampaigns(ctx)
	if err != nil {
		err = errors.Wrap(err, "listing pre-spec campaigns")
		return
	}

	// It's easiest if we track a few things while we iterate over the
	// changesets: most notably, the changeset spec process needs to know what
	// user to pretend to be and needs to know which campaigns will need the
	// changeset specs once they're created.
	campaignsByID := map[int64]*campaigns.Campaign{}
	campaignChangesetSpecs := map[int64][]*campaigns.ChangesetSpec{}
	usersByCampaign := map[int64]int32{}
	for _, c := range cmpgns {
		campaignsByID[c.ID] = c
		usersByCampaign[c.ID] = c.InitialApplierID
	}

	// First, we have to create changeset specs to track the changesets that
	// were previously created or tracked.
	cs, err := store.listPreSpecChangesets(ctx)
	if err != nil {
		err = errors.Wrap(err, "listing pre-spec changesets")
		return
	}

	for _, c := range cs {
		// We also need a user to own the changeset spec, which we'll
		// arbitrarily choose to be the creator of the first campaign the old
		// changeset was attached to.
		if len(c.CampaignIDs) == 0 {
			err = errors.Errorf("changeset %d was not attached to any campaigns", c.ID)
			return
		}

		// Check if we have a campaign. If a previous migration run was partly
		// successful, the changeset may no longer have a pre-spec campaign to
		// attach to. That's OK, we'll just skip it.
		user, ok := usersByCampaign[c.CampaignIDs[0]]
		if !ok {
			log15.Debug("skipping changeset without any remaining pre-spec campaigns", "changeset", *c)
			err = nil
			continue
		}

		// Fake up just enough changeset spec JSON to track the changeset.
		desc := campaigns.ChangesetSpecDescription{
			BaseRepository: graphqlbackend.MarshalRepositoryID(c.RepoID),
			ExternalID:     c.ExternalID,
		}

		var raw []byte
		raw, err = json.Marshal(&desc)
		if err != nil {
			err = errors.Wrap(err, "marshalling changeset spec")
			return
		}

		// Now we're ready to create a changeset spec with the JSON we faked
		// up.
		var spec *campaigns.ChangesetSpec
		spec, err = txSvc.CreateChangesetSpec(contextWithActor(ctx, user), string(raw), user)
		if err != nil {
			err = errors.Wrapf(err, "creating changeset spec for changeset %d", c.ID)
			return
		}

		// We'll need to know the changeset spec when we create the campaign
		// specs, so let's keep track of it now while we have all the bits we
		// need.
		for _, cid := range c.CampaignIDs {
			campaignChangesetSpecs[cid] = append(campaignChangesetSpecs[cid], spec)
		}
	}

	// Now for the campaigns: we have the changeset specs we need, so it's time
	// to create the campaign specs we can use to recreate the campaigns.
	for _, c := range cmpgns {
		// As with the changesets, we now need to fake some campaign spec JSON.
		// Unlike changesets, we can't use CampaignSpecFields, as the new
		// importChangesets field isn't yet represented on it. That's OK, we
		// can define some types here.
		type importChangeset struct {
			Repository  string   `json:"repository"`
			ExternalIDs []string `json:"externalIDs"`
		}
		var in struct {
			Name             string            `json:"name"`
			Description      string            `json:"description"`
			ImportChangesets []importChangeset `json:"importChangesets"`
		}

		// We'll start building the campaign spec structure.
		in.Name = campaignSlug(c.Name)

		// Since the name conversion is lossy, we'll reproduce the name in the
		// description as well.
		in.Description = fmt.Sprintf("%s\n\n%s", c.Name, c.Description)

		// We need to iterate over the changeset specs for this campaign both
		// to build the ImportChangesets slice and to get the random changeset
		// IDs that we need when we call CreateCampaignSpec.
		randIDs := []string{}
		repos := map[string][]string{}
		for _, spec := range campaignChangesetSpecs[c.ID] {
			repo := string(graphqlbackend.MarshalRepositoryID(spec.RepoID))
			repos[repo] = append(repos[repo], spec.Spec.ExternalID)
			randIDs = append(randIDs, spec.RandID)
		}
		for repo, ids := range repos {
			in.ImportChangesets = append(in.ImportChangesets, importChangeset{
				Repository:  repo,
				ExternalIDs: ids,
			})
		}

		// Let's get that JSON.
		var raw []byte
		raw, err = json.Marshal(&in)
		if err != nil {
			err = errors.Wrapf(err, "marshalling campaign spec for campaign %d", c.ID)
			return
		}

		// The campaign service rightly checks that we have permission to do
		// whatever it is that we're doing, as well as using the user to
		// populate the user fields on the campaign record. Since this is a
		// backend service, we'll bypass the authentication checks, but we also
		// need to make sure the right actor is attached to the context.
		campaignCtx := contextWithActor(ctx, c.InitialApplierID)

		// Now we can create the campaign spec.
		var spec *campaigns.CampaignSpec
		spec, err = txSvc.CreateCampaignSpec(campaignCtx, CreateCampaignSpecOpts{
			RawSpec:              string(raw),
			NamespaceUserID:      c.InitialApplierID,
			ChangesetSpecRandIDs: randIDs,
		})
		if err != nil {
			err = errors.Wrapf(err, "creating campaign spec for campaign %d\n%s\n", c.ID, string(raw))
			return
		}

		// Let's roll right on into applying the campaign spec so we get a real
		// campaign.
		var nc *campaigns.Campaign
		nc, err = txSvc.ApplyCampaign(campaignCtx, ApplyCampaignOpts{
			CampaignSpecRandID:   spec.RandID,
			FailIfCampaignExists: true,
		})
		if err != nil {
			// Extremely evil hackery: because ApplyCampaign does a synchronous
			// sync (say that five times fast), if the sync fails due to code
			// host issues that are outside of our control (service issues,
			// missing pull requests, et cetera) we probably want to continue
			// and try again next time repo-updater restarts, rather than
			// reporting an error.
			//
			// The fact that we're basing this behaviour on knowing what to
			// look for in the error message is, to state the obvious,
			// horrific.
			if strings.Contains(err.Error(), "syncing changeset failed") {
				log15.Info("sync error when applying campaign, continuing", "err", err, "campaign", *c)
				err = nil
				continue
			}
			err = errors.Wrapf(err, "applying campaign for campaign %d", c.ID)
			return
		}

		// If you liked the above evil, here's some more: we're going to change
		// the campaign ID. This is safe because of how the migration works: we
		// copied the table exactly but didn't recreate campaigns or its ID
		// sequence, which means we know the sequence is already past the ID
		// that was on the old campaign.
		err = store.updateCampaignID(ctx, nc, c.ID)
		if err != nil {
			err = errors.Wrapf(err, "resetting campaign ID %d", c.ID)
			return
		}

		// Finally, we can delete the pre-spec campaign record, which will save
		// us from doing all this again.
		err = store.deletePreSpecCampaign(ctx, c.ID)
		if err != nil {
			err = errors.Wrapf(err, "deleting pre-spec campaign %d", c.ID)
			return
		}

		log15.Info("campaign converted", "old", *c, "new", *nc)
	}

	return
}

var campaignSlugCleanRegex = regexp.MustCompile(`[^\w.-]`)
var campaignSlugDashRegex = regexp.MustCompile(`-{2,}`)

func campaignSlug(name string) string {
	clean := campaignSlugCleanRegex.ReplaceAllLiteralString(name, "-")
	return campaignSlugDashRegex.ReplaceAllLiteralString(clean, "-")
}

func contextWithActor(ctx context.Context, uid int32) context.Context {
	return actor.WithActor(ctx, actor.FromUser(uid))
}
