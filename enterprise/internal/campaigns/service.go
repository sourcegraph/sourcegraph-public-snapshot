package campaigns

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// NewService returns a Service.
func NewService(store *Store, cf *httpcli.Factory) *Service {
	return NewServiceWithClock(store, cf, store.Clock())
}

// NewServiceWithClock returns a Service the given clock used
// to generate timestamps.
func NewServiceWithClock(store *Store, cf *httpcli.Factory, clock func() time.Time) *Service {
	svc := &Service{store: store, cf: cf, sourcer: repos.NewSourcer(cf), clock: clock}

	return svc
}

type Service struct {
	store *Store
	cf    *httpcli.Factory

	sourcer repos.Sourcer

	clock func() time.Time
}

type CreateCampaignSpecOpts struct {
	RawSpec string

	NamespaceUserID int32
	NamespaceOrgID  int32

	ChangesetSpecRandIDs []string
}

// CreateCampaignSpec creates the CampaignSpec.
func (s *Service) CreateCampaignSpec(ctx context.Context, opts CreateCampaignSpecOpts) (spec *campaigns.CampaignSpec, err error) {
	actor := actor.FromContext(ctx)
	tr, ctx := trace.New(ctx, "Service.CreateCampaignSpec", fmt.Sprintf("Actor %s", actor))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	spec, err = campaigns.NewCampaignSpecFromRaw(opts.RawSpec)
	if err != nil {
		return nil, err
	}

	// Check whether the current user has access to either one of the namespaces.
	err = checkNamespaceAccess(ctx, opts.NamespaceUserID, opts.NamespaceOrgID)
	if err != nil {
		return nil, err
	}
	spec.NamespaceOrgID = opts.NamespaceOrgID
	spec.NamespaceUserID = opts.NamespaceUserID
	spec.UserID = actor.UID

	if len(opts.ChangesetSpecRandIDs) == 0 {
		return spec, s.store.CreateCampaignSpec(ctx, spec)
	}

	listOpts := ListChangesetSpecsOpts{Limit: -1, RandIDs: opts.ChangesetSpecRandIDs}
	cs, _, err := s.store.ListChangesetSpecs(ctx, listOpts)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: db.Repos.GetRepoIDsSet uses the authzFilter under the hood and
	// filters out repositories that the user doesn't have access to.
	accessibleReposByID, err := db.Repos.GetReposSetByIDs(ctx, cs.RepoIDs()...)
	if err != nil {
		return nil, err
	}

	byRandID := make(map[string]*campaigns.ChangesetSpec, len(cs))
	for _, changesetSpec := range cs {
		// 🚨 SECURITY: We return an error if the user doesn't have access to one
		// of the repositories associated with a ChangesetSpec.
		if _, ok := accessibleReposByID[changesetSpec.RepoID]; !ok {
			return nil, &db.RepoNotFoundErr{ID: changesetSpec.RepoID}
		}
		byRandID[changesetSpec.RandID] = changesetSpec
	}

	// Check if a changesetSpec was not found
	for _, randID := range opts.ChangesetSpecRandIDs {
		if _, ok := byRandID[randID]; !ok {
			return nil, &changesetSpecNotFoundErr{RandID: randID}
		}
	}

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.CreateCampaignSpec(ctx, spec); err != nil {
		return nil, err
	}

	for _, changesetSpec := range cs {
		changesetSpec.CampaignSpecID = spec.ID

		if err := tx.UpdateChangesetSpec(ctx, changesetSpec); err != nil {
			return nil, err
		}
	}

	return spec, nil
}

// CreateChangesetSpec validates the given raw spec input and creates the ChangesetSpec.
func (s *Service) CreateChangesetSpec(ctx context.Context, rawSpec string, userID int32) (spec *campaigns.ChangesetSpec, err error) {
	tr, ctx := trace.New(ctx, "Service.CreateChangesetSpec", fmt.Sprintf("User %d", userID))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	spec, err = campaigns.NewChangesetSpecFromRaw(rawSpec)
	if err != nil {
		return nil, err
	}
	spec.UserID = userID
	spec.RepoID, err = graphqlbackend.UnmarshalRepositoryID(spec.Spec.BaseRepository)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: We use db.Repos.Get to check whether the user has access to
	// the repository or not.
	if _, err = db.Repos.Get(ctx, spec.RepoID); err != nil {
		return nil, err
	}

	return spec, s.store.CreateChangesetSpec(ctx, spec)
}

// changesetSpecNotFoundErr is returned by CreateCampaignSpec if a
// ChangesetSpec with the given RandID doesn't exist.
// It fulfills the interface required by errcode.IsNotFound.
type changesetSpecNotFoundErr struct {
	RandID string
}

func (e *changesetSpecNotFoundErr) Error() string {
	if e.RandID != "" {
		return fmt.Sprintf("changesetSpec not found: id=%s", e.RandID)
	}
	return "changesetSpec not found"
}

func (e *changesetSpecNotFoundErr) NotFound() bool { return true }

// GetCampaignMatchingCampaignSpec returns the Campaign that the CampaignSpec
// applies to, if that Campaign already exists.
// If it doesn't exist yet, both return values are nil.
// It accepts a *Store so that it can be used inside a transaction.
func (s *Service) GetCampaignMatchingCampaignSpec(ctx context.Context, tx *Store, spec *campaigns.CampaignSpec) (*campaigns.Campaign, error) {
	opts := GetCampaignOpts{
		Name:            spec.Spec.Name,
		NamespaceUserID: spec.NamespaceUserID,
		NamespaceOrgID:  spec.NamespaceOrgID,
	}

	campaign, err := tx.GetCampaign(ctx, opts)
	if err != nil {
		if err != ErrNoResults {
			return nil, err
		}
		err = nil
	}
	return campaign, err
}

type MoveCampaignOpts struct {
	CampaignID int64

	NewName string

	NewNamespaceUserID int32
	NewNamespaceOrgID  int32
}

func (o MoveCampaignOpts) String() string {
	return fmt.Sprintf(
		"CampaignID %d, NewName %q, NewNamespaceUserID %d, NewNamespaceOrgID %d",
		o.CampaignID,
		o.NewName,
		o.NewNamespaceUserID,
		o.NewNamespaceOrgID,
	)
}

// MoveCampaign moves the campaign from one namespace to another and/or renames
// the campaign.
func (s *Service) MoveCampaign(ctx context.Context, opts MoveCampaignOpts) (campaign *campaigns.Campaign, err error) {
	tr, ctx := trace.New(ctx, "Service.MoveCampaign", opts.String())
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	campaign, err = tx.GetCampaign(ctx, GetCampaignOpts{ID: opts.CampaignID})
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only the Author of the campaign can move it.
	if err := backend.CheckSiteAdminOrSameUser(ctx, campaign.InitialApplierID); err != nil {
		return nil, err
	}
	// Check if current user has access to target namespace if set.
	if opts.NewNamespaceOrgID != 0 || opts.NewNamespaceUserID != 0 {
		err = checkNamespaceAccess(ctx, opts.NewNamespaceUserID, opts.NewNamespaceOrgID)
		if err != nil {
			return nil, err
		}
	}

	if opts.NewNamespaceOrgID != 0 {
		campaign.NamespaceOrgID = opts.NewNamespaceOrgID
		campaign.NamespaceUserID = 0
	} else if opts.NewNamespaceUserID != 0 {
		campaign.NamespaceUserID = opts.NewNamespaceUserID
		campaign.NamespaceOrgID = 0
	}

	if opts.NewName != "" {
		campaign.Name = opts.NewName
	}

	return campaign, tx.UpdateCampaign(ctx, campaign)
}

// ErrEnsureCampaignFailed is returned by ApplyCampaign when a ensureCampaignID
// is provided but a campaign with the name specified the campaignSpec exists
// in the given namespace but has a different ID.
var ErrEnsureCampaignFailed = errors.New("a campaign in the given namespace and with the given name exists but does not match the given ID")

// ErrCloseProcessingCampaign is returned by CloseCampaign if the Campaign has
// been published at the time of closing but its Changesets are still being
// processed by the reconciler.
var ErrCloseProcessingCampaign = errors.New("cannot close a campaign while changesets are being processed")

// CloseCampaign closes the Campaign with the given ID if it has not been closed yet.
func (s *Service) CloseCampaign(ctx context.Context, id int64, closeChangesets, closeAsync bool) (campaign *campaigns.Campaign, err error) {
	traceTitle := fmt.Sprintf("campaign: %d, closeChangesets: %t", id, closeChangesets)
	tr, ctx := trace.New(ctx, "service.CloseCampaign", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	transaction := func() (err error) {
		tx, err := s.store.Transact(ctx)
		if err != nil {
			return err
		}
		defer func() { err = tx.Done(err) }()

		campaign, err = tx.GetCampaign(ctx, GetCampaignOpts{ID: id})
		if err != nil {
			return errors.Wrap(err, "getting campaign")
		}

		if campaign.Closed() {
			return nil
		}

		if err := backend.CheckSiteAdminOrSameUser(ctx, campaign.InitialApplierID); err != nil {
			return err
		}

		if closeChangesets {
			processingState := campaigns.ReconcilerStateProcessing
			countOpts := CountChangesetsOpts{
				CampaignID:      campaign.ID,
				ReconcilerState: &processingState,
			}
			processingCount, err := tx.CountChangesets(ctx, countOpts)
			if err != nil {
				return errors.Wrap(err, "checking for processing changesets")
			}
			if processingCount != 0 {
				err = ErrCloseProcessingCampaign
				return err
			}
		}

		campaign.ClosedAt = time.Now().UTC()

		return tx.UpdateCampaign(ctx, campaign)
	}

	err = transaction()
	if err != nil {
		return nil, err
	}

	if closeChangesets {
		user := actor.FromContext(ctx)
		actorCtx := contextWithActor(context.Background(), user.UID)
		ctx := trace.ContextWithTrace(actorCtx, tr)

		closer := func() {
			open := campaigns.ChangesetExternalStateOpen
			published := campaigns.ChangesetPublicationStatePublished
			cs, _, err := s.store.ListChangesets(ctx, ListChangesetsOpts{
				OwnedByCampaignID: campaign.ID,
				ExternalState:     &open,
				PublicationState:  &published,
				Limit:             -1,
			})
			if err != nil {
				log15.Error("ListChangesets", "err", err)
				return
			}

			// Close only the changesets that are open
			err = s.CloseOpenChangesets(ctx, cs)
			if err != nil {
				log15.Error("CloseCampaignChangesets", "err", err)
			}
		}
		if closeAsync {
			go closer()
		} else {
			closer()
		}
	}

	return campaign, nil
}

// DeleteCampaign deletes the Campaign with the given ID if it hasn't been
// deleted yet.
func (s *Service) DeleteCampaign(ctx context.Context, id int64) (err error) {
	traceTitle := fmt.Sprintf("campaign: %d", id)
	tr, ctx := trace.New(ctx, "service.DeleteCampaign", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	campaign, err := s.store.GetCampaign(ctx, GetCampaignOpts{ID: id})
	if err != nil {
		return err
	}

	if err := backend.CheckSiteAdminOrSameUser(ctx, campaign.InitialApplierID); err != nil {
		return err
	}

	return s.store.DeleteCampaign(ctx, id)
}

// mockCloseOpenChangesets is used to test CloseOpenChangesets closing
// the correct changesets with the correct context.
// This is a temporary mock that should be removed once we move closing of
// changesets into the background.
var mockCloseChangesets func(context.Context, campaigns.Changesets)

// CloseOpenChangesets closes the given Changesets on their respective codehosts and syncs them.
func (s *Service) CloseOpenChangesets(ctx context.Context, cs campaigns.Changesets) (err error) {
	if mockCloseChangesets != nil {
		mockCloseChangesets(ctx, cs)
		return nil
	}

	cs = cs.Filter(func(c *campaigns.Changeset) bool {
		return c.ExternalState == campaigns.ChangesetExternalStateOpen
	})

	if len(cs) == 0 {
		return nil
	}

	// 🚨 SECURITY: db.Repos.GetRepoIDsSet uses the authzFilter under the hood and
	// filters out repositories that the user doesn't have access to.
	accessibleReposByID, err := db.Repos.GetReposSetByIDs(ctx, cs.RepoIDs()...)
	if err != nil {
		return err
	}

	reposStore := repos.NewDBStore(s.store.DB(), sql.TxOptions{})
	bySource, err := groupChangesetsBySource(ctx, reposStore, s.sourcer, cs...)
	if err != nil {
		return err
	}

	errs := &multierror.Error{}
	for _, group := range bySource {
		for _, c := range group.Changesets {
			if _, ok := accessibleReposByID[c.RepoID]; !ok {
				continue
			}

			if err := group.CloseChangeset(ctx, c); err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}

	if len(errs.Errors) != 0 {
		return errs
	}

	// Here we need to sync the just-closed changesets (even though
	// CloseChangesets updates the given Changesets too), because closing a
	// Changeset often produces a ChangesetEvent on the codehost and if we were
	// to close the Changesets and not update the events (which is what
	// syncChangesetsWithSources does) our burndown chart will be outdated
	// until the next run of campaigns.Syncer.
	return syncChangesetsWithSources(ctx, s.store, bySource)
}

// EnqueueChangesetSync loads the given changeset from the database, checks
// whether the actor in the context has permission to enqueue a sync and then
// enqueues a sync by calling the repoupdater client.
func (s *Service) EnqueueChangesetSync(ctx context.Context, id int64) (err error) {
	traceTitle := fmt.Sprintf("changeset: %d", id)
	tr, ctx := trace.New(ctx, "service.EnqueueChangesetSync", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// Check for existence of changeset so we don't swallow that error.
	changeset, err := s.store.GetChangeset(ctx, GetChangesetOpts{ID: id})
	if err != nil {
		return err
	}

	// 🚨 SECURITY: We use db.Repos.Get to check whether the user has access to
	// the repository or not.
	if _, err = db.Repos.Get(ctx, changeset.RepoID); err != nil {
		return err
	}

	campaigns, _, err := s.store.ListCampaigns(ctx, ListCampaignsOpts{ChangesetID: id})
	if err != nil {
		return err
	}

	// Check whether the user has admin rights for one of the campaigns.
	var (
		authErr        error
		hasAdminRights bool
	)

	for _, c := range campaigns {
		err := backend.CheckSiteAdminOrSameUser(ctx, c.InitialApplierID)
		if err != nil {
			authErr = err
		} else {
			hasAdminRights = true
			break
		}
	}

	if !hasAdminRights {
		return authErr
	}

	if err := repoupdater.DefaultClient.EnqueueChangesetSync(ctx, []int64{id}); err != nil {
		return err
	}

	return nil
}

// ErrCampaignNameBlank is returned by CreateCampaign or UpdateCampaign if the
// specified Campaign name is blank.
var ErrCampaignNameBlank = errors.New("Campaign title cannot be blank")

// checkNamespaceAccess checks whether the current user in the ctx has access
// to either the user ID or the org ID as a namespace.
// If the userID is non-zero that will be checked. Otherwise the org ID will be
// checked.
// If the current user is an admin, true will be returned.
// Otherwise it checks whether the current user _is_ the namespace user or has
// access to the namespace org.
// If both values are zero, an error is returned.
func checkNamespaceAccess(ctx context.Context, namespaceUserID, namespaceOrgID int32) error {
	if namespaceOrgID != 0 {
		return backend.CheckOrgAccess(ctx, namespaceOrgID)
	} else if namespaceUserID != 0 {
		return backend.CheckSiteAdminOrSameUser(ctx, namespaceUserID)
	} else {
		return ErrNoNamespace
	}
}

// ErrNoNamespace is returned by checkNamespaceAccess if no valid namespace ID is given.
var ErrNoNamespace = errors.New("no namespace given")

// checkRepoSupported checks whether the given repository is supported by campaigns
// and if not it returns an error.
func checkRepoSupported(repo *types.Repo) error {
	if campaigns.IsRepoSupported(&repo.ExternalRepo) {
		return nil
	}

	return errors.Errorf(
		"External service type %s of repository %q is currently not supported for use with campaigns",
		repo.ExternalRepo.ServiceType,
		repo.Name,
	)
}
