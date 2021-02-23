package service

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// New returns a Service.
func New(store *store.Store) *Service {
	return NewWithClock(store, store.Clock())
}

// NewWithClock returns a Service the given clock used
// to generate timestamps.
func NewWithClock(store *store.Store, clock func() time.Time) *Service {
	svc := &Service{store: store, sourcer: repos.NewSourcer(httpcli.NewExternalHTTPClientFactory()), clock: clock}

	return svc
}

type Service struct {
	store *store.Store

	sourcer repos.Sourcer

	clock func() time.Time
}

// WithStore returns a copy of the Service with its store attribute set to the
// given Store.
func (s *Service) WithStore(store *store.Store) *Service {
	return &Service{store: store, sourcer: s.sourcer, clock: s.clock}
}

type CreateCampaignSpecOpts struct {
	RawSpec string `json:"raw_spec"`

	NamespaceUserID int32 `json:"namespace_user_id"`
	NamespaceOrgID  int32 `json:"namespace_org_id"`

	ChangesetSpecRandIDs []string `json:"changeset_spec_rand_ids"`
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
	err = checkNamespaceAccess(ctx, s.store.DB(), opts.NamespaceUserID, opts.NamespaceOrgID)
	if err != nil {
		return nil, err
	}
	spec.NamespaceOrgID = opts.NamespaceOrgID
	spec.NamespaceUserID = opts.NamespaceUserID
	spec.UserID = actor.UID

	if len(opts.ChangesetSpecRandIDs) == 0 {
		return spec, s.store.CreateCampaignSpec(ctx, spec)
	}

	listOpts := store.ListChangesetSpecsOpts{RandIDs: opts.ChangesetSpecRandIDs}
	cs, _, err := s.store.ListChangesetSpecs(ctx, listOpts)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: database.Repos.GetRepoIDsSet uses the authzFilter under the hood and
	// filters out repositories that the user doesn't have access to.
	accessibleReposByID, err := s.store.Repos().GetReposSetByIDs(ctx, cs.RepoIDs()...)
	if err != nil {
		return nil, err
	}

	byRandID := make(map[string]*campaigns.ChangesetSpec, len(cs))
	for _, changesetSpec := range cs {
		// 🚨 SECURITY: We return an error if the user doesn't have access to one
		// of the repositories associated with a ChangesetSpec.
		if _, ok := accessibleReposByID[changesetSpec.RepoID]; !ok {
			return nil, &database.RepoNotFoundErr{ID: changesetSpec.RepoID}
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

	// 🚨 SECURITY: We use database.Repos.Get to check whether the user has access to
	// the repository or not.
	if _, err = s.store.Repos().Get(ctx, spec.RepoID); err != nil {
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
// It accepts a *store.Store so that it can be used inside a transaction.
func (s *Service) GetCampaignMatchingCampaignSpec(ctx context.Context, spec *campaigns.CampaignSpec) (*campaigns.Campaign, error) {
	opts := store.GetCampaignOpts{
		Name:            spec.Spec.Name,
		NamespaceUserID: spec.NamespaceUserID,
		NamespaceOrgID:  spec.NamespaceOrgID,
	}

	campaign, err := s.store.GetCampaign(ctx, opts)
	if err != nil {
		if err != store.ErrNoResults {
			return nil, err
		}
		err = nil
	}
	return campaign, err
}

// GetNewestCampaignSpec returns the newest campaign spec that matches the given
// spec's namespace and name and is owned by the given user, or nil if none is found.
func (s *Service) GetNewestCampaignSpec(ctx context.Context, tx *store.Store, spec *campaigns.CampaignSpec, userID int32) (*campaigns.CampaignSpec, error) {
	opts := store.GetNewestCampaignSpecOpts{
		UserID:          userID,
		NamespaceUserID: spec.NamespaceUserID,
		NamespaceOrgID:  spec.NamespaceOrgID,
		Name:            spec.Spec.Name,
	}

	newest, err := tx.GetNewestCampaignSpec(ctx, opts)
	if err != nil {
		if err != store.ErrNoResults {
			return nil, err
		}
		return nil, nil
	}

	return newest, nil
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

	campaign, err = tx.GetCampaign(ctx, store.GetCampaignOpts{ID: opts.CampaignID})
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only the Author of the campaign can move it.
	if err := backend.CheckSiteAdminOrSameUser(ctx, campaign.InitialApplierID); err != nil {
		return nil, err
	}
	// Check if current user has access to target namespace if set.
	if opts.NewNamespaceOrgID != 0 || opts.NewNamespaceUserID != 0 {
		err = checkNamespaceAccess(ctx, s.store.DB(), opts.NewNamespaceUserID, opts.NewNamespaceOrgID)
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

// CloseCampaign closes the Campaign with the given ID if it has not been closed yet.
func (s *Service) CloseCampaign(ctx context.Context, id int64, closeChangesets bool) (campaign *campaigns.Campaign, err error) {
	traceTitle := fmt.Sprintf("campaign: %d, closeChangesets: %t", id, closeChangesets)
	tr, ctx := trace.New(ctx, "service.CloseCampaign", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	campaign, err = s.store.GetCampaign(ctx, store.GetCampaignOpts{ID: id})
	if err != nil {
		return nil, errors.Wrap(err, "getting campaign")
	}

	if campaign.Closed() {
		return campaign, nil
	}

	if err := backend.CheckSiteAdminOrSameUser(ctx, campaign.InitialApplierID); err != nil {
		return nil, err
	}

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	campaign.ClosedAt = s.clock()
	if err := tx.UpdateCampaign(ctx, campaign); err != nil {
		return nil, err
	}

	if !closeChangesets {
		return campaign, nil
	}

	// At this point we don't know which changesets have ExternalStateOpen,
	// since some might still be being processed in the background by the
	// reconciler.
	// So enqueue all, except the ones that are completed and closed/merged,
	// for closing. If after being processed they're not open, it'll be a noop.
	if err := tx.EnqueueChangesetsToClose(ctx, campaign.ID); err != nil {
		return nil, err
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

	campaign, err := s.store.GetCampaign(ctx, store.GetCampaignOpts{ID: id})
	if err != nil {
		return err
	}

	if err := backend.CheckSiteAdminOrSameUser(ctx, campaign.InitialApplierID); err != nil {
		return err
	}

	return s.store.DeleteCampaign(ctx, id)
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
	changeset, err := s.store.GetChangeset(ctx, store.GetChangesetOpts{ID: id})
	if err != nil {
		return err
	}

	// 🚨 SECURITY: We use database.Repos.Get to check whether the user has access to
	// the repository or not.
	if _, err = s.store.Repos().Get(ctx, changeset.RepoID); err != nil {
		return err
	}

	campaigns, _, err := s.store.ListCampaigns(ctx, store.ListCampaignsOpts{ChangesetID: id})
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

// ReenqueueChangeset loads the given changeset from the database, checks
// whether the actor in the context has permission to enqueue a reconciler run and then
// enqueues it by calling ResetQueued.
func (s *Service) ReenqueueChangeset(ctx context.Context, id int64) (changeset *campaigns.Changeset, repo *types.Repo, err error) {
	traceTitle := fmt.Sprintf("changeset: %d", id)
	tr, ctx := trace.New(ctx, "service.RenqueueChangeset", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	changeset, err = s.store.GetChangeset(ctx, store.GetChangesetOpts{ID: id})
	if err != nil {
		return nil, nil, err
	}

	// 🚨 SECURITY: We use database.Repos.Get to check whether the user has access to
	// the repository or not.
	repo, err = s.store.Repos().Get(ctx, changeset.RepoID)
	if err != nil {
		return nil, nil, err
	}

	attachedCampaigns, _, err := s.store.ListCampaigns(ctx, store.ListCampaignsOpts{ChangesetID: id})
	if err != nil {
		return nil, nil, err
	}

	// Check whether the user has admin rights for one of the campaigns.
	var (
		authErr        error
		hasAdminRights bool
	)

	for _, c := range attachedCampaigns {
		err := backend.CheckSiteAdminOrSameUser(ctx, c.InitialApplierID)
		if err != nil {
			authErr = err
		} else {
			hasAdminRights = true
			break
		}
	}

	if !hasAdminRights {
		return nil, nil, authErr
	}

	if changeset.ReconcilerState != campaigns.ReconcilerStateFailed {
		return nil, nil, errors.New("cannot re-enqueue changeset not in failed state")
	}

	changeset.ResetQueued()

	if err = s.store.UpdateChangeset(ctx, changeset); err != nil {
		return nil, nil, err
	}

	return changeset, repo, nil
}

// checkNamespaceAccess checks whether the current user in the ctx has access
// to either the user ID or the org ID as a namespace.
// If the userID is non-zero that will be checked. Otherwise the org ID will be
// checked.
// If the current user is an admin, true will be returned.
// Otherwise it checks whether the current user _is_ the namespace user or has
// access to the namespace org.
// If both values are zero, an error is returned.
func checkNamespaceAccess(ctx context.Context, db dbutil.DB, namespaceUserID, namespaceOrgID int32) error {
	if namespaceOrgID != 0 {
		return backend.CheckOrgAccess(ctx, db, namespaceOrgID)
	} else if namespaceUserID != 0 {
		return backend.CheckSiteAdminOrSameUser(ctx, namespaceUserID)
	} else {
		return ErrNoNamespace
	}
}

// ErrNoNamespace is returned by checkNamespaceAccess if no valid namespace ID is given.
var ErrNoNamespace = errors.New("no namespace given")

// FetchUsernameForBitbucketServerToken fetches the username associated with a
// Bitbucket server token.
//
// We need the username in order to use the token as the password in a HTTP
// BasicAuth username/password pair used by gitserver to push commits.
//
// In order to not require from users to type in their BitbucketServer username
// we only ask for a token and then use that token to talk to the
// BitbucketServer API and get their username.
//
// Since Bitbucket sends the username as a header in REST responses, we can
// take it from there and complete the UserCredential.
func (s *Service) FetchUsernameForBitbucketServerToken(ctx context.Context, externalServiceID, externalServiceType, token string) (string, error) {
	extSvcID, err := s.store.GetExternalServiceID(ctx, store.GetExternalServiceIDOpts{
		ExternalServiceID:   externalServiceID,
		ExternalServiceType: externalServiceType,
	})
	if err != nil {
		return "", err
	}

	externalService, err := s.store.ExternalServices().GetByID(ctx, extSvcID)
	if err != nil {
		if errcode.IsNotFound(err) {
			return "", errors.New("no external service found for repo")
		}

		return "", err
	}

	sources, err := s.sourcer(externalService)
	if err != nil {
		return "", err
	}

	userSource, ok := sources[0].(repos.UserSource)
	if !ok {
		return "", errors.New("external service source cannot use other authenticator")
	}

	source, err := userSource.WithAuthenticator(&auth.OAuthBearerToken{Token: token})
	if err != nil {
		return "", err
	}

	usernameSource, ok := source.(usernameSource)
	if !ok {
		return "", errors.New("external service source doesn't implement AuthenticatedUsername")
	}

	return usernameSource.AuthenticatedUsername(ctx)
}

// A usernameSource can fetch the username associated with the credentials used
// by the Source.
// It's only used by FetchUsernameForBitbucketServerToken.
type usernameSource interface {
	// AuthenticatedUsername makes a request to the code host to fetch the
	// username associated with the credentials.
	// If no username could be determined an error is returned.
	AuthenticatedUsername(ctx context.Context) (string, error)
}

var _ usernameSource = &repos.BitbucketServerSource{}
