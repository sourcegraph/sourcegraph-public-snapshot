package service

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/global"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	streamapi "github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

// New returns a Service.
func New(store *store.Store) *Service {
	return NewWithClock(store, store.Clock())
}

// NewWithClock returns a Service the given clock used
// to generate timestamps.
func NewWithClock(store *store.Store, clock func() time.Time) *Service {
	svc := &Service{store: store, sourcer: sources.NewSourcer(httpcli.ExternalClientFactory), clock: clock}

	return svc
}

type Service struct {
	store *store.Store

	sourcer sources.Sourcer

	clock func() time.Time
}

// WithStore returns a copy of the Service with its store attribute set to the
// given Store.
func (s *Service) WithStore(store *store.Store) *Service {
	return &Service{store: store, sourcer: s.sourcer, clock: s.clock}
}

type CreateBatchSpecOpts struct {
	RawSpec string `json:"raw_spec"`

	NamespaceUserID int32 `json:"namespace_user_id"`
	NamespaceOrgID  int32 `json:"namespace_org_id"`

	ChangesetSpecRandIDs []string `json:"changeset_spec_rand_ids"`
}

// CreateBatchSpec creates the BatchSpec.
func (s *Service) CreateBatchSpec(ctx context.Context, opts CreateBatchSpecOpts) (spec *btypes.BatchSpec, err error) {
	actor := actor.FromContext(ctx)
	tr, ctx := trace.New(ctx, "Service.CreateBatchSpec", fmt.Sprintf("Actor %s", actor))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	spec, err = btypes.NewBatchSpecFromRaw(opts.RawSpec)
	if err != nil {
		return nil, err
	}

	// Check whether the current user has access to either one of the namespaces.
	err = s.CheckNamespaceAccess(ctx, opts.NamespaceUserID, opts.NamespaceOrgID)
	if err != nil {
		return nil, err
	}
	spec.NamespaceOrgID = opts.NamespaceOrgID
	spec.NamespaceUserID = opts.NamespaceUserID
	spec.UserID = actor.UID

	if len(opts.ChangesetSpecRandIDs) == 0 {
		return spec, s.store.CreateBatchSpec(ctx, spec)
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

	byRandID := make(map[string]*btypes.ChangesetSpec, len(cs))
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

	if err := tx.CreateBatchSpec(ctx, spec); err != nil {
		return nil, err
	}

	for _, changesetSpec := range cs {
		changesetSpec.BatchSpecID = spec.ID

		if err := tx.UpdateChangesetSpec(ctx, changesetSpec); err != nil {
			return nil, err
		}
	}

	return spec, nil
}

// CreateChangesetSpec validates the given raw spec input and creates the ChangesetSpec.
func (s *Service) CreateChangesetSpec(ctx context.Context, rawSpec string, userID int32) (spec *btypes.ChangesetSpec, err error) {
	tr, ctx := trace.New(ctx, "Service.CreateChangesetSpec", fmt.Sprintf("User %d", userID))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	spec, err = btypes.NewChangesetSpecFromRaw(rawSpec)
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

// changesetSpecNotFoundErr is returned by CreateBatchSpec if a
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

// GetBatchChangeMatchingBatchSpec returns the batch change that the BatchSpec
// applies to, if that BatchChange already exists.
// If it doesn't exist yet, both return values are nil.
// It accepts a *store.Store so that it can be used inside a transaction.
func (s *Service) GetBatchChangeMatchingBatchSpec(ctx context.Context, spec *btypes.BatchSpec) (*btypes.BatchChange, error) {
	opts := store.GetBatchChangeOpts{
		Name:            spec.Spec.Name,
		NamespaceUserID: spec.NamespaceUserID,
		NamespaceOrgID:  spec.NamespaceOrgID,
	}

	batchChange, err := s.store.GetBatchChange(ctx, opts)
	if err != nil {
		if err != store.ErrNoResults {
			return nil, err
		}
		err = nil
	}
	return batchChange, err
}

// GetNewestBatchSpec returns the newest batch spec that matches the given
// spec's namespace and name and is owned by the given user, or nil if none is found.
func (s *Service) GetNewestBatchSpec(ctx context.Context, tx *store.Store, spec *btypes.BatchSpec, userID int32) (*btypes.BatchSpec, error) {
	opts := store.GetNewestBatchSpecOpts{
		UserID:          userID,
		NamespaceUserID: spec.NamespaceUserID,
		NamespaceOrgID:  spec.NamespaceOrgID,
		Name:            spec.Spec.Name,
	}

	newest, err := tx.GetNewestBatchSpec(ctx, opts)
	if err != nil {
		if err != store.ErrNoResults {
			return nil, err
		}
		return nil, nil
	}

	return newest, nil
}

type MoveBatchChangeOpts struct {
	BatchChangeID int64

	NewName string

	NewNamespaceUserID int32
	NewNamespaceOrgID  int32
}

func (o MoveBatchChangeOpts) String() string {
	return fmt.Sprintf(
		"BatchChangeID %d, NewName %q, NewNamespaceUserID %d, NewNamespaceOrgID %d",
		o.BatchChangeID,
		o.NewName,
		o.NewNamespaceUserID,
		o.NewNamespaceOrgID,
	)
}

// MoveBatchChange moves the batch change from one namespace to another and/or renames
// the batch change.
func (s *Service) MoveBatchChange(ctx context.Context, opts MoveBatchChangeOpts) (batchChange *btypes.BatchChange, err error) {
	tr, ctx := trace.New(ctx, "Service.MoveBatchChange", opts.String())
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	batchChange, err = tx.GetBatchChange(ctx, store.GetBatchChangeOpts{ID: opts.BatchChangeID})
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only the Author of the batch change can move it.
	if err := backend.CheckSiteAdminOrSameUser(ctx, s.store.DB(), batchChange.InitialApplierID); err != nil {
		return nil, err
	}
	// Check if current user has access to target namespace if set.
	if opts.NewNamespaceOrgID != 0 || opts.NewNamespaceUserID != 0 {
		err = s.CheckNamespaceAccess(ctx, opts.NewNamespaceUserID, opts.NewNamespaceOrgID)
		if err != nil {
			return nil, err
		}
	}

	if opts.NewNamespaceOrgID != 0 {
		batchChange.NamespaceOrgID = opts.NewNamespaceOrgID
		batchChange.NamespaceUserID = 0
	} else if opts.NewNamespaceUserID != 0 {
		batchChange.NamespaceUserID = opts.NewNamespaceUserID
		batchChange.NamespaceOrgID = 0
	}

	if opts.NewName != "" {
		batchChange.Name = opts.NewName
	}

	return batchChange, tx.UpdateBatchChange(ctx, batchChange)
}

// CloseBatchChange closes the BatchChange with the given ID if it has not been closed yet.
func (s *Service) CloseBatchChange(ctx context.Context, id int64, closeChangesets bool) (batchChange *btypes.BatchChange, err error) {
	traceTitle := fmt.Sprintf("batchChange: %d, closeChangesets: %t", id, closeChangesets)
	tr, ctx := trace.New(ctx, "service.CloseBatchChange", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	batchChange, err = s.store.GetBatchChange(ctx, store.GetBatchChangeOpts{ID: id})
	if err != nil {
		return nil, errors.Wrap(err, "getting batch change")
	}

	if batchChange.Closed() {
		return batchChange, nil
	}

	if err := backend.CheckSiteAdminOrSameUser(ctx, s.store.DB(), batchChange.InitialApplierID); err != nil {
		return nil, err
	}

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	batchChange.ClosedAt = s.clock()
	if err := tx.UpdateBatchChange(ctx, batchChange); err != nil {
		return nil, err
	}

	if !closeChangesets {
		return batchChange, nil
	}

	// At this point we don't know which changesets have ExternalStateOpen,
	// since some might still be being processed in the background by the
	// reconciler.
	// So enqueue all, except the ones that are completed and closed/merged,
	// for closing. If after being processed they're not open, it'll be a noop.
	if err := tx.EnqueueChangesetsToClose(ctx, batchChange.ID); err != nil {
		return nil, err
	}

	return batchChange, nil
}

// DeleteBatchChange deletes the BatchChange with the given ID if it hasn't been
// deleted yet.
func (s *Service) DeleteBatchChange(ctx context.Context, id int64) (err error) {
	traceTitle := fmt.Sprintf("BatchChange: %d", id)
	tr, ctx := trace.New(ctx, "service.BatchChange", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	batchChange, err := s.store.GetBatchChange(ctx, store.GetBatchChangeOpts{ID: id})
	if err != nil {
		return err
	}

	if err := backend.CheckSiteAdminOrSameUser(ctx, s.store.DB(), batchChange.InitialApplierID); err != nil {
		return err
	}

	return s.store.DeleteBatchChange(ctx, id)
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

	batchChanges, _, err := s.store.ListBatchChanges(ctx, store.ListBatchChangesOpts{ChangesetID: id})
	if err != nil {
		return err
	}

	// Check whether the user has admin rights for one of the batches.
	var (
		authErr        error
		hasAdminRights bool
	)

	for _, c := range batchChanges {
		err := backend.CheckSiteAdminOrSameUser(ctx, s.store.DB(), c.InitialApplierID)
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
// enqueues it by calling ResetReconcilerState.
func (s *Service) ReenqueueChangeset(ctx context.Context, id int64) (changeset *btypes.Changeset, repo *types.Repo, err error) {
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

	attachedBatchChanges, _, err := s.store.ListBatchChanges(ctx, store.ListBatchChangesOpts{ChangesetID: id})
	if err != nil {
		return nil, nil, err
	}

	// Check whether the user has admin rights for one of the batches.
	var (
		authErr        error
		hasAdminRights bool
	)

	for _, c := range attachedBatchChanges {
		err := backend.CheckSiteAdminOrSameUser(ctx, s.store.DB(), c.InitialApplierID)
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

	if err := s.store.EnqueueChangeset(ctx, changeset, global.DefaultReconcilerEnqueueState(), btypes.ReconcilerStateFailed); err != nil {
		return nil, nil, err
	}

	return changeset, repo, nil
}

// CheckNamespaceAccess checks whether the current user in the ctx has access
// to either the user ID or the org ID as a namespace.
// If the userID is non-zero that will be checked. Otherwise the org ID will be
// checked.
// If the current user is an admin, true will be returned.
// Otherwise it checks whether the current user _is_ the namespace user or has
// access to the namespace org.
// If both values are zero, an error is returned.
func (s *Service) CheckNamespaceAccess(ctx context.Context, namespaceUserID, namespaceOrgID int32) error {
	if namespaceOrgID != 0 {
		return backend.CheckOrgAccessOrSiteAdmin(ctx, s.store.DB(), namespaceOrgID)
	} else if namespaceUserID != 0 {
		return backend.CheckSiteAdminOrSameUser(ctx, s.store.DB(), namespaceUserID)
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
	css, err := s.sourcer.ForExternalService(ctx, s.store, store.GetExternalServiceIDsOpts{
		ExternalServiceType: externalServiceType,
		ExternalServiceID:   externalServiceID,
	})
	if err != nil {
		return "", err
	}
	css, err = css.WithAuthenticator(&auth.OAuthBearerToken{Token: token})
	if err != nil {
		return "", err
	}

	usernameSource, ok := css.(usernameSource)
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

var _ usernameSource = &sources.BitbucketServerSource{}

// ValidateAuthenticator creates a ChangesetSource, configures it with the given
// authenticator and validates it can correctly access the remote server.
func (s *Service) ValidateAuthenticator(ctx context.Context, externalServiceID, externalServiceType string, a auth.Authenticator) error {
	if Mocks.ValidateAuthenticator != nil {
		return Mocks.ValidateAuthenticator(ctx, externalServiceID, externalServiceType, a)
	}

	css, err := s.sourcer.ForExternalService(ctx, s.store, store.GetExternalServiceIDsOpts{
		ExternalServiceType: externalServiceType,
		ExternalServiceID:   externalServiceID,
	})
	if err != nil {
		return err
	}
	css, err = css.WithAuthenticator(a)
	if err != nil {
		return err
	}

	if err := css.ValidateAuthenticator(ctx); err != nil {
		return err
	}
	return nil
}

// ErrChangesetsForJobNotFound can be returned by (*Service).CreateChangesetJobs
// if the number of changesets returned from the database doesn't match the
// number if IDs passed in. That can happen if some of the changesets are not
// published.
var ErrChangesetsForJobNotFound = errors.New("some changesets could not be found")

// CreateChangesetJobs creates one changeset job for each given Changeset in the
// given BatchChange, checking whether the actor in the context has permission to
// trigger a job, and enqueues it.
func (s *Service) CreateChangesetJobs(ctx context.Context, batchChangeID int64, ids []int64, jobType btypes.ChangesetJobType, payload interface{}, listOpts store.ListChangesetsOpts) (bulkGroupID string, err error) {
	traceTitle := fmt.Sprintf("batchChangeID: %d, len(changesets): %d", batchChangeID, len(ids))
	tr, ctx := trace.New(ctx, "service.CreateChangesetJobs", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// Load the BatchChange to check for write permissions.
	batchChange, err := s.store.GetBatchChange(ctx, store.GetBatchChangeOpts{ID: batchChangeID})
	if err != nil {
		return bulkGroupID, errors.Wrap(err, "loading batch change")
	}

	// 🚨 SECURITY: Only the author of the batch change can create jobs.
	if err := backend.CheckSiteAdminOrSameUser(ctx, s.store.DB(), batchChange.InitialApplierID); err != nil {
		return bulkGroupID, err
	}

	// Construct list options.
	opts := listOpts
	opts.IDs = ids
	opts.BatchChangeID = batchChangeID
	// We only want to allow changesets the user has access to.
	opts.EnforceAuthz = true
	cs, _, err := s.store.ListChangesets(ctx, opts)
	if err != nil {
		return bulkGroupID, errors.Wrap(err, "listing changesets")
	}

	if len(cs) != len(ids) {
		return bulkGroupID, ErrChangesetsForJobNotFound
	}

	bulkGroupID, err = store.RandomID()
	if err != nil {
		return bulkGroupID, errors.Wrap(err, "creating bulkGroupID failed")
	}

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return bulkGroupID, errors.Wrap(err, "starting transaction")
	}
	defer func() { err = tx.Done(err) }()

	userID := actor.FromContext(ctx).UID
	changesetJobs := make([]*btypes.ChangesetJob, 0, len(cs))
	for _, changeset := range cs {
		changesetJobs = append(changesetJobs, &btypes.ChangesetJob{
			BulkGroup:     bulkGroupID,
			ChangesetID:   changeset.ID,
			BatchChangeID: batchChangeID,
			UserID:        userID,
			State:         btypes.ChangesetJobStateQueued,
			JobType:       jobType,
			Payload:       payload,
		})
	}

	// Bulk-insert all changeset jobs into the database.
	if err := tx.CreateChangesetJob(ctx, changesetJobs...); err != nil {
		return bulkGroupID, errors.Wrap(err, "creating changeset jobs")
	}

	return bulkGroupID, nil
}

type RepoRevision struct {
	Repo   *types.Repo
	Branch string
	Commit api.CommitID
}

func (r *RepoRevision) HasBranch() bool {
	return r.Branch != ""
}

type ResolveRepositoriesForBatchSpecOpts struct {
	AllowIgnored     bool
	AllowUnsupported bool
}

func (s *Service) ResolveRepositoriesForBatchSpec(ctx context.Context, batchSpec *batcheslib.BatchSpec, opts ResolveRepositoriesForBatchSpecOpts) (_ []*RepoRevision, err error) {
	traceTitle := fmt.Sprintf("len(On): %d", len(batchSpec.On))
	tr, ctx := trace.New(ctx, "service.ResolveRepositoriesForBatchSpec", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	seen := map[api.RepoID]*RepoRevision{}
	unsupported := UnsupportedRepoSet{}
	ignored := IgnoredRepoSet{}

	// TODO: this could be trivially parallelised in the future.
	for _, on := range batchSpec.On {
		repos, err := s.resolveRepositoriesOn(ctx, &on)
		if err != nil {
			return nil, errors.Wrapf(err, "resolving %q", on.String())
		}

		for _, repo := range repos {
			// Skip repos where no branch exists.
			if !repo.HasBranch() {
				continue
			}

			if other, ok := seen[repo.Repo.ID]; !ok {
				seen[repo.Repo.ID] = repo

				switch st := repo.Repo.ExternalRepo.ServiceType; st {
				case extsvc.TypeGitHub, extsvc.TypeGitLab, extsvc.TypeBitbucketServer:
				default:
					if !opts.AllowUnsupported {
						unsupported.Append(repo.Repo)
					}
				}
			} else {
				// If we've already seen this repository, we overwrite the
				// Commit/Branch fields with the latest value we have
				other.Commit = repo.Commit
				other.Branch = repo.Branch
			}
		}
	}

	final := make([]*RepoRevision, 0, len(seen))
	// TODO: Limit concurrency.
	var wg sync.WaitGroup
	var errs *multierror.Error
	for _, repo := range seen {
		repo := repo
		wg.Add(1)
		go func(repo *RepoRevision) {
			defer wg.Done()
			ignore, err := s.hasBatchIgnoreFile(ctx, repo)
			if err != nil {
				errs = multierror.Append(errs, err)
				return
			}
			if !opts.AllowIgnored && ignore {
				ignored.Append(repo.Repo)
			}

			if !unsupported.Includes(repo.Repo) && !ignored.Includes(repo.Repo) {
				final = append(final, repo)
			}
		}(repo)
	}
	wg.Wait()
	if err := errs.ErrorOrNil(); err != nil {
		return nil, err
	}

	if unsupported.HasUnsupported() {
		return final, unsupported
	}

	if ignored.HasIgnored() {
		return final, ignored
	}

	return final, nil
}

var ErrMalformedOnQueryOrRepository = errors.New("malformed 'on' field; missing either a repository name or a query")

func (s *Service) resolveRepositoriesOn(ctx context.Context, on *batcheslib.OnQueryOrRepository) (_ []*RepoRevision, err error) {
	traceTitle := fmt.Sprintf("On: %+v", on)
	tr, ctx := trace.New(ctx, "service.resolveRepositoriesOn", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if on.RepositoriesMatchingQuery != "" {
		return s.resolveRepositorySearch(ctx, on.RepositoriesMatchingQuery)
	} else if on.Repository != "" && on.Branch != "" {
		repo, err := s.resolveRepositoryNameAndBranch(ctx, on.Repository, on.Branch)
		if err != nil {
			return nil, err
		}
		return []*RepoRevision{repo}, nil
	} else if on.Repository != "" {
		repo, err := s.resolveRepositoryName(ctx, on.Repository)
		if err != nil {
			return nil, err
		}
		return []*RepoRevision{repo}, nil
	}

	// This shouldn't happen on any batch spec that has passed validation, but,
	// alas, software.
	return nil, ErrMalformedOnQueryOrRepository
}

func (s *Service) resolveRepositoryName(ctx context.Context, name string) (_ *RepoRevision, err error) {
	traceTitle := fmt.Sprintf("Name: %q", name)
	tr, ctx := trace.New(ctx, "service.resolveRepositoryName", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	repo, err := s.store.Repos().GetByName(ctx, api.RepoName(name))
	if err != nil {
		return nil, err
	}

	return s.repoToRepoRevision(ctx, repo)
}

func (s *Service) repoToRepoRevision(ctx context.Context, repo *types.Repo) (_ *RepoRevision, err error) {
	traceTitle := fmt.Sprintf("Repo: %q", repo.Name)
	tr, ctx := trace.New(ctx, "service.resolveRepositoriesOn", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	repoRev := &RepoRevision{
		Repo: repo,
	}

	// TODO: Fill default branch.
	refBytes, _, exitCode, err := git.ExecSafe(ctx, repo.Name, []string{"symbolic-ref", "HEAD"})
	repoRev.Branch = string(bytes.TrimSpace(refBytes))
	if err == nil && exitCode == 0 {
		// Check that our repo is not empty
		repoRev.Commit, err = git.ResolveRevision(ctx, repo.Name, "HEAD", git.ResolveRevisionOptions{NoEnsureRevision: true})
	}
	// TODO: Handle repoCloneInProgressErr
	return repoRev, err
}

func (s *Service) resolveRepositoryNameAndBranch(ctx context.Context, name, branch string) (_ *RepoRevision, err error) {
	traceTitle := fmt.Sprintf("Name: %q Branch: %q", name, branch)
	tr, ctx := trace.New(ctx, "service.resolveRepositoryNameAndBranch", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	repo, err := s.resolveRepositoryName(ctx, name)
	if err != nil {
		return repo, err
	}

	commit, err := git.ResolveRevision(ctx, repo.Repo.Name, branch, git.ResolveRevisionOptions{
		NoEnsureRevision: true,
	})
	if err != nil && errors.HasType(err, &gitserver.RevisionNotFoundError{}) {
		return repo, fmt.Errorf("no branch matching %q found for repository %s", branch, name)
	}

	repo.Branch = branch
	repo.Commit = commit

	return repo, err
}

func (s *Service) resolveRepositorySearch(ctx context.Context, query string) (_ []*RepoRevision, err error) {
	traceTitle := fmt.Sprintf("Query: %q", query)
	tr, ctx := trace.New(ctx, "service.resolveRepositorySearch", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	query = setDefaultQueryCount(query)
	query = setDefaultQuerySelect(query)

	repoIDs := []api.RepoID{}
	s.runSearch(ctx, query, func(matches []streamhttp.EventMatch) {
		for _, match := range matches {
			switch m := match.(type) {
			case *streamhttp.EventRepoMatch:
				repoIDs = append(repoIDs, api.RepoID(m.RepositoryID))
			case *streamhttp.EventContentMatch:
				repoIDs = append(repoIDs, api.RepoID(m.RepositoryID))
			}
		}
	})

	accessibleRepos, err := s.store.Repos().List(ctx, database.ReposListOptions{IDs: repoIDs})
	if err != nil {
		return nil, err
	}
	revs := make([]*RepoRevision, 0, len(accessibleRepos))
	for _, repo := range accessibleRepos {
		rev, err := s.repoToRepoRevision(ctx, repo)
		if err != nil {
			{
				return nil, err
			}
		}
		revs = append(revs, rev)
	}

	return revs, nil
}

func (s *Service) runSearch(ctx context.Context, query string, onMatches func(matches []streamhttp.EventMatch)) (err error) {
	// TODO: Duh, why do I need to add .internal here.
	req, err := streamhttp.NewRequest(api.InternalClient.URL+"/.internal", query)
	if err != nil {
		return err
	}
	req.WithContext(ctx)
	// TODO: Document why it's okay to not pass along the ctx.User here.
	req.Header.Set("User-Agent", "Batch Changes repository resolver")

	resp, err := httpcli.InternalClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := streamhttp.FrontendStreamDecoder{
		OnMatches: func(matches []streamhttp.EventMatch) {
			onMatches(matches)
		},
		OnError: func(ee *streamhttp.EventError) {
			err = errors.New(ee.Message)
		},
		OnProgress: func(p *streamapi.Progress) {
			// TODO: Evaluate skipped for values we care about.
		},
	}
	return dec.ReadAll(resp.Body)
}

func (s *Service) hasBatchIgnoreFile(ctx context.Context, r *RepoRevision) (_ bool, err error) {
	traceTitle := fmt.Sprintf("Repo: %q Revision: %q", r.Repo.Name, r.Branch)
	tr, ctx := trace.New(ctx, "service.hasBatchIgnoreFile", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	path := ".batchignore"
	stat, err := git.Stat(ctx, r.Repo.Name, r.Commit, path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if !stat.Mode().IsRegular() {
		return false, errors.Errorf("not a blob: %q", path)
	}
	return true, nil
}

var defaultQueryCountRegex = regexp.MustCompile(`\bcount:(\d+|all)\b`)

const hardCodedCount = " count:all"

func setDefaultQueryCount(query string) string {
	if defaultQueryCountRegex.MatchString(query) {
		return query
	}

	return query + hardCodedCount
}

var selectRegex = regexp.MustCompile(`\bselect:(.+)\b`)

const hardCodedSelectRepo = " select:repo"

func setDefaultQuerySelect(query string) string {
	if selectRegex.MatchString(query) {
		return query
	}

	return query + hardCodedSelectRepo
}

// TODO(mrnugget): Merge these two types (give them an "errorfmt" function,
// rename "Has*" methods to "NotEmpty" or something)

// UnsupportedRepoSet provides a set to manage repositories that are on
// unsupported code hosts. This type implements error to allow it to be
// returned directly as an error value if needed.
type UnsupportedRepoSet map[*types.Repo]struct{}

func (e UnsupportedRepoSet) Includes(r *types.Repo) bool {
	_, ok := e[r]
	return ok
}

func (e UnsupportedRepoSet) Error() string {
	repos := []string{}
	typeSet := map[string]struct{}{}
	for repo := range e {
		repos = append(repos, string(repo.Name))
		typeSet[repo.ExternalRepo.ServiceType] = struct{}{}
	}

	types := []string{}
	for t := range typeSet {
		types = append(types, t)
	}

	return fmt.Sprintf(
		"found repositories on unsupported code hosts: %s\nrepositories:\n\t%s",
		strings.Join(types, ", "),
		strings.Join(repos, "\n\t"),
	)
}

func (e UnsupportedRepoSet) Append(repo *types.Repo) {
	e[repo] = struct{}{}
}

func (e UnsupportedRepoSet) HasUnsupported() bool {
	return len(e) > 0
}

// IgnoredRepoSet provides a set to manage repositories that are on
// unsupported code hosts. This type implements error to allow it to be
// returned directly as an error value if needed.
type IgnoredRepoSet map[*types.Repo]struct{}

func (e IgnoredRepoSet) Includes(r *types.Repo) bool {
	_, ok := e[r]
	return ok
}

func (e IgnoredRepoSet) Error() string {
	repos := []string{}
	for repo := range e {
		repos = append(repos, string(repo.Name))
	}

	return fmt.Sprintf(
		"found repositories containing .batchignore files:\n\t%s",
		strings.Join(repos, "\n\t"),
	)
}

func (e IgnoredRepoSet) Append(repo *types.Repo) {
	e[repo] = struct{}{}
}

func (e IgnoredRepoSet) HasIgnored() bool {
	return len(e) > 0
}
