package sources

import (
	"context"
	"fmt"
	"net/url"
	"sort"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// ErrMissingCredentials is returned by WithAuthenticatorForUser,
// if the user that applied the last batch change/changeset spec doesn't have
// UserCredentials for the given repository and is not a site-admin (so no
// fallback to the global credentials is possible).
var ErrMissingCredentials = errors.New("no credential found that can authenticate to the code host")

// ErrNoPushCredentials is returned by gitserverPushConfig if the
// authenticator cannot be used by git to authenticate a `git push`.
type ErrNoPushCredentials struct{ CredentialsType string }

func (e ErrNoPushCredentials) Error() string {
	return "invalid authenticator provided to push commits"
}

// ErrNoSSHCredential is returned by gitserverPushConfig, if the
// clone URL of the repository uses the ssh:// scheme, but the authenticator
// doesn't support SSH pushes.
var ErrNoSSHCredential = errors.New("authenticator doesn't support SSH")

type SourcerStore interface {
	GetBatchChange(ctx context.Context, opts store.GetBatchChangeOpts) (*btypes.BatchChange, error)
	GetSiteCredential(ctx context.Context, opts store.GetSiteCredentialOpts) (*btypes.SiteCredential, error)
	GetExternalServiceIDs(ctx context.Context, opts store.GetExternalServiceIDsOpts) ([]int64, error)
	Repos() database.RepoStore
	ExternalServices() database.ExternalServiceStore
	UserCredentials() database.UserCredentialsStore
}

// Sourcer exposes methods to get a ChangesetSource based on a changeset, repo or
// external service.
type Sourcer interface {
	// ForChangeset returns a ChangesetSource for the given changeset. The changeset.RepoID
	// is used to find the matching code host.
	// It authenticates the given ChangesetSource with a
	// credential appropriate to sync or reconcile the given changeset. If the
	// changeset was created by a batch change, then authentication will be based on
	// the first available option of:
	//
	// 1. The last applying user's credentials.
	// 2. Any available site credential matching the changesets repo.
	//
	// If the changeset was not created by a batch change, then a site credential
	// will be used.
	ForChangeset(ctx context.Context, tx SourcerStore, ch *btypes.Changeset) (ChangesetSource, error)
	// ForUser returns a ChangesetSource for changesets on the given repo.
	// It will be authenticated with the given authenticator.
	ForUser(ctx context.Context, tx SourcerStore, uid int32, repo *types.Repo) (ChangesetSource, error)
	// ForExternalService returns a ChangesetSource based on the provided external service opts.
	// It will be authenticated with the given authenticator.
	ForExternalService(ctx context.Context, tx SourcerStore, au auth.Authenticator, opts store.GetExternalServiceIDsOpts) (ChangesetSource, error)
}

// NewSourcer returns a new Sourcer to be used in Batch Changes.
func NewSourcer(cf *httpcli.Factory) Sourcer {
	return newSourcer(cf, loadBatchesSource)
}

type changesetSourceFactory func(ctx context.Context, tx SourcerStore, cf *httpcli.Factory, externalServiceIDs []int64) (ChangesetSource, error)

type sourcer struct {
	cf        *httpcli.Factory
	newSource changesetSourceFactory
}

func newSourcer(cf *httpcli.Factory, csf changesetSourceFactory) Sourcer {
	return &sourcer{
		cf:        cf,
		newSource: csf,
	}
}

func (s *sourcer) ForChangeset(ctx context.Context, tx SourcerStore, ch *btypes.Changeset) (ChangesetSource, error) {
	repo, err := tx.Repos().Get(ctx, ch.RepoID)
	if err != nil {
		return nil, errors.Wrap(err, "loading changeset repo")
	}
	// Consider all available external services for this repo.
	css, err := s.newSource(ctx, tx, s.cf, repo.ExternalServiceIDs())
	if err != nil {
		return nil, err
	}
	if ch.OwnedByBatchChangeID != 0 {
		batchChange, err := loadBatchChange(ctx, tx, ch.OwnedByBatchChangeID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load owning batch change")
		}

		return withAuthenticatorForUser(ctx, tx, css, batchChange.LastApplierID, repo)
	}

	return withSiteAuthenticator(ctx, tx, css, repo)
}

func (s *sourcer) ForUser(ctx context.Context, tx SourcerStore, uid int32, repo *types.Repo) (ChangesetSource, error) {
	// Consider all available external services for this repo.
	css, err := s.newSource(ctx, tx, s.cf, repo.ExternalServiceIDs())
	if err != nil {
		return nil, err
	}
	return withAuthenticatorForUser(ctx, tx, css, uid, repo)
}

func (s *sourcer) ForExternalService(ctx context.Context, tx SourcerStore, au auth.Authenticator, opts store.GetExternalServiceIDsOpts) (ChangesetSource, error) {
	// Empty authenticators are not allowed.
	if au == nil {
		return nil, ErrMissingCredentials
	}

	extSvcIDs, err := tx.GetExternalServiceIDs(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, "loading external service IDs")
	}
	css, err := s.newSource(ctx, tx, s.cf, extSvcIDs)
	if err != nil {
		return nil, err
	}
	return css.WithAuthenticator(au)
}

func loadBatchesSource(ctx context.Context, tx SourcerStore, cf *httpcli.Factory, externalServiceIDs []int64) (ChangesetSource, error) {
	extSvc, err := loadExternalService(ctx, tx.ExternalServices(), database.ExternalServicesListOptions{
		IDs: externalServiceIDs,
	})
	if err != nil {
		return nil, errors.Wrap(err, "loading external service")
	}
	css, err := buildChangesetSource(ctx, cf, extSvc)
	if err != nil {
		return nil, errors.Wrap(err, "building changeset source")
	}
	return css, nil
}

// GitserverPushConfig creates a push configuration given a repo and an
// authenticator. This function is only public for testing purposes, and should
// not be used otherwise.
func GitserverPushConfig(repo *types.Repo, au auth.Authenticator) (*protocol.PushConfig, error) {
	// Empty authenticators are not allowed.
	if au == nil {
		return nil, ErrNoPushCredentials{}
	}

	cloneURL, err := extractCloneURL(repo)
	if err != nil {
		return nil, err
	}

	// If the repo is cloned using SSH, we need to pass along a private key and passphrase.
	if cloneURL.IsSSH() {
		sshA, ok := au.(auth.AuthenticatorWithSSH)
		if !ok {
			return nil, ErrNoSSHCredential
		}
		privateKey, passphrase := sshA.SSHPrivateKey()
		return &protocol.PushConfig{
			RemoteURL:  cloneURL.String(),
			PrivateKey: privateKey,
			Passphrase: passphrase,
		}, nil
	}

	extSvcType := repo.ExternalRepo.ServiceType
	switch av := au.(type) {
	case *auth.OAuthBearerTokenWithSSH:
		if err := setOAuthTokenAuth(cloneURL, extSvcType, av.Token); err != nil {
			return nil, err
		}
	case *auth.OAuthBearerToken:
		if err := setOAuthTokenAuth(cloneURL, extSvcType, av.Token); err != nil {
			return nil, err
		}

	case *auth.BasicAuthWithSSH:
		if err := setBasicAuth(cloneURL, extSvcType, av.Username, av.Password); err != nil {
			return nil, err
		}
	case *auth.BasicAuth:
		if err := setBasicAuth(cloneURL, extSvcType, av.Username, av.Password); err != nil {
			return nil, err
		}
	default:
		return nil, ErrNoPushCredentials{CredentialsType: fmt.Sprintf("%T", au)}
	}

	return &protocol.PushConfig{RemoteURL: cloneURL.String()}, nil
}

// ToDraftChangesetSource returns a DraftChangesetSource, if the underlying
// source supports it. Returns an error if not.
func ToDraftChangesetSource(css ChangesetSource) (DraftChangesetSource, error) {
	draftCss, ok := css.(DraftChangesetSource)
	if !ok {
		return nil, errors.New("changeset source doesn't implement DraftChangesetSource")
	}
	return draftCss, nil
}

type getBatchChanger interface {
	GetBatchChange(ctx context.Context, opts store.GetBatchChangeOpts) (*btypes.BatchChange, error)
}

func loadBatchChange(ctx context.Context, tx getBatchChanger, id int64) (*btypes.BatchChange, error) {
	if id == 0 {
		return nil, errors.New("changeset has no owning batch change")
	}

	batchChange, err := tx.GetBatchChange(ctx, store.GetBatchChangeOpts{ID: id})
	if err != nil && err != store.ErrNoResults {
		return nil, errors.Wrapf(err, "retrieving owning batch change: %d", id)
	} else if batchChange == nil {
		return nil, errors.Errorf("batch change not found: %d", id)
	}

	return batchChange, nil
}

// withAuthenticatorForUser authenticates the given ChangesetSource with a credential
// usable by the given user with userID. User credentials are preferred, with a
// fallback to site credentials. If none of these exist, ErrMissingCredentials
// is returned.
func withAuthenticatorForUser(ctx context.Context, tx SourcerStore, css ChangesetSource, userID int32, repo *types.Repo) (ChangesetSource, error) {
	cred, err := loadUserCredential(ctx, tx, userID, repo)
	if err != nil {
		return nil, errors.Wrap(err, "loading user credential")
	}
	if cred != nil {
		return css.WithAuthenticator(cred)
	}

	// Fall back to site credentials.
	return withSiteAuthenticator(ctx, tx, css, repo)
}

// withSiteAuthenticator uses the site credential of the code host of the passed-in repo.
// If no credential is found, the original source is returned and uses the external service
// config.
func withSiteAuthenticator(ctx context.Context, tx SourcerStore, css ChangesetSource, repo *types.Repo) (ChangesetSource, error) {
	cred, err := loadSiteCredential(ctx, tx, store.GetSiteCredentialOpts{
		ExternalServiceType: repo.ExternalRepo.ServiceType,
		ExternalServiceID:   repo.ExternalRepo.ServiceID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "loading site credential")
	}
	if cred != nil {
		return css.WithAuthenticator(cred)
	}
	return nil, ErrMissingCredentials
}

// loadExternalService looks up all external services that are connected to the given repo.
// Global external services are preferred over user-owned external services.
// If no external service matching the given criteria is found, an error is returned.
func loadExternalService(ctx context.Context, s database.ExternalServiceStore, opts database.ExternalServicesListOptions) (*types.ExternalService, error) {
	es, err := s.List(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Sort the external services so user owned external service go last.
	// This also retains the initial ORDER BY ID DESC.
	sort.SliceStable(es, func(i, j int) bool {
		return es[i].NamespaceUserID == 0 && es[i].ID > es[j].ID
	})

	for _, e := range es {
		cfg, err := e.Configuration(ctx)
		if err != nil {
			return nil, err
		}

		switch cfg.(type) {
		case *schema.GitHubConnection,
			*schema.BitbucketServerConnection,
			*schema.GitLabConnection,
			*schema.BitbucketCloudConnection:
			return e, nil
		}
	}

	return nil, errors.New("no external services found")
}

// buildChangesetSource builds a ChangesetSource for the given repo to load the
// changeset state from.
func buildChangesetSource(ctx context.Context, cf *httpcli.Factory, externalService *types.ExternalService) (ChangesetSource, error) {
	switch externalService.Kind {
	case extsvc.KindGitHub:
		return NewGithubSource(ctx, externalService, cf)
	case extsvc.KindGitLab:
		return NewGitLabSource(ctx, externalService, cf)
	case extsvc.KindBitbucketServer:
		return NewBitbucketServerSource(ctx, externalService, cf)
	case extsvc.KindBitbucketCloud:
		return NewBitbucketCloudSource(ctx, externalService, cf)
	default:
		return nil, errors.Errorf("unsupported external service type %q", extsvc.KindToType(externalService.Kind))
	}
}

// loadUserCredential attempts to find a user credential for the given repo.
// When no credential is found, nil is returned.
func loadUserCredential(ctx context.Context, s SourcerStore, userID int32, repo *types.Repo) (auth.Authenticator, error) {
	cred, err := s.UserCredentials().GetByScope(ctx, database.UserCredentialScope{
		Domain:              database.UserCredentialDomainBatches,
		UserID:              userID,
		ExternalServiceType: repo.ExternalRepo.ServiceType,
		ExternalServiceID:   repo.ExternalRepo.ServiceID,
	})
	if err != nil && !errcode.IsNotFound(err) {
		return nil, err
	}
	if cred != nil {
		return cred.Authenticator(ctx)
	}
	return nil, nil
}

// loadSiteCredential attempts to find a site credential for the given repo.
// When no credential is found, nil is returned.
func loadSiteCredential(ctx context.Context, s SourcerStore, opts store.GetSiteCredentialOpts) (auth.Authenticator, error) {
	cred, err := s.GetSiteCredential(ctx, opts)
	if err != nil && err != store.ErrNoResults {
		return nil, err
	}
	if cred != nil {
		return cred.Authenticator(ctx)
	}
	return nil, nil
}

// setOAuthTokenAuth sets the user part of the given URL to use the provided OAuth token,
// with the specific quirks per code host.
func setOAuthTokenAuth(u *vcs.URL, extSvcType, token string) error {
	switch extSvcType {
	case extsvc.TypeGitHub:
		u.User = url.User(token)

	case extsvc.TypeGitLab:
		u.User = url.UserPassword("git", token)

	case extsvc.TypeBitbucketServer:
		return errors.New("require username/token to push commits to BitbucketServer")

	default:
		panic(fmt.Sprintf("setOAuthTokenAuth: invalid external service type %q", extSvcType))
	}
	return nil
}

// setBasicAuth sets the user part of the given URL to use the provided username/
// password combination, with the specific quirks per code host.
func setBasicAuth(u *vcs.URL, extSvcType, username, password string) error {
	switch extSvcType {
	case extsvc.TypeGitHub, extsvc.TypeGitLab:
		return errors.New("need token to push commits to " + extSvcType)

	case extsvc.TypeBitbucketServer, extsvc.TypeBitbucketCloud:
		u.User = url.UserPassword(username, password)

	default:
		panic(fmt.Sprintf("setBasicAuth: invalid external service type %q", extSvcType))
	}
	return nil
}

// extractCloneURL returns a remote URL from the repo, preferring HTTPS over SSH.
func extractCloneURL(repo *types.Repo) (*vcs.URL, error) {
	if len(repo.Sources) == 0 {
		return nil, errors.New("no clone URL found for repo")
	}

	cloneURLs := make([]*vcs.URL, 0, len(repo.Sources))
	for _, src := range repo.Sources {
		parsedURL, err := vcs.ParseURL(src.CloneURL)
		if err != nil {
			return nil, err
		}
		cloneURLs = append(cloneURLs, parsedURL)
	}

	sort.SliceStable(cloneURLs, func(i, j int) bool {
		return !cloneURLs[i].IsSSH()
	})

	return cloneURLs[0], nil
}

var ErrChangesetSourceCannotFork = errors.New("forking is enabled, but the changeset source does not support forks")

// GetRemoteRepo returns the remote that should be pushed to for a given
// changeset, changeset source, and target repo. The changeset spec may
// optionally be provided, and is required if the repo will be pushed to.
func GetRemoteRepo(
	ctx context.Context,
	css ChangesetSource,
	targetRepo *types.Repo,
	ch *btypes.Changeset,
	spec *btypes.ChangesetSpec,
) (*types.Repo, error) {
	// If the changeset spec doesn't expect a fork _and_ we're not updating a
	// changeset that was previously created using a fork, then we don't need to
	// even check if the changeset source is forkable, let alone set up the
	// remote repo: we can just return the target repo and be done with it.
	if ch.ExternalForkNamespace == "" && (spec == nil || !spec.IsFork()) {
		return targetRepo, nil
	}

	fss, ok := css.(ForkableChangesetSource)
	if !ok {
		return nil, ErrChangesetSourceCannotFork
	}

	if ch.ExternalForkNamespace != "" {
		// If we're updating an existing changeset, we should push/modify the
		// same fork, even if the user credential would now fork into a
		// different namespace.
		return fss.GetNamespaceFork(ctx, targetRepo, ch.ExternalForkNamespace)
	} else if namespace := spec.GetForkNamespace(); namespace != nil {
		// If the changeset spec requires a specific fork namespace, then we
		// should handle that here.
		return fss.GetNamespaceFork(ctx, targetRepo, *namespace)
	}

	// Otherwise, we're pushing to a user fork.
	return fss.GetUserFork(ctx, targetRepo)
}
