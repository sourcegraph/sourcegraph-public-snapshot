package sources

import (
	"context"
	"fmt"
	"net/url"
	"sort"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/schema"
)

// ErrMissingCredentials is returned by BatchesSource.WithAuthenticatorForUser,
// if the user that applied the last batch change/changeset spec doesn't have
// UserCredentials for the given repository and is not a site-admin (so no
// fallback to the global credentials is possible).
var ErrMissingCredentials = errors.New("no credential found to authenticate BatchesSource")

// ErrNoPushCredentials is returned by BatchesSource.GitserverPushConfig if the
// authenticator cannot be used by git to authenticate a `git push`.
type ErrNoPushCredentials struct{ CredentialsType string }

func (e ErrNoPushCredentials) Error() string {
	return "invalid authenticator provided to push commits"
}

// ErrNoSSHCredential is returned by BatchesSource.GitserverPushConfig, if the
// clone URL of the repository uses the ssh:// scheme, but the authenticator
// doesn't support SSH pushes.
var ErrNoSSHCredential = errors.New("authenticator doesn't support SSH")

type SourcerStore interface {
	DB() dbutil.DB
	GetSiteCredential(ctx context.Context, opts store.GetSiteCredentialOpts) (*btypes.SiteCredential, error)
	GetExternalServiceIDs(ctx context.Context, opts store.GetExternalServiceIDsOpts) ([]int64, error)
	Repos() *database.RepoStore
	ExternalServices() *database.ExternalServiceStore
	UserCredentials() *database.UserCredentialsStore
}

// Sourcer exposes methods to get a BatchesSource based on a changeset, repo or
// external service.
type Sourcer interface {
	ForChangeset(ctx context.Context, ch *btypes.Changeset) (*BatchesSource, error)
	ForRepo(ctx context.Context, repo *types.Repo) (*BatchesSource, error)
	ForExternalService(ctx context.Context, opts store.GetExternalServiceIDsOpts) (*BatchesSource, error)
}

type sourcer struct {
	cf    *httpcli.Factory
	store SourcerStore
}

type fakeSourcer struct {
	err    error
	source *BatchesSource
}

// BatchesSource wraps repos.ChangesetSource and repos.UserSource, which are both
// required to be a valid source used with Batch Changes. It exposes methods to
// be authenticated with a user or site credential.
type BatchesSource struct {
	ChangesetSource
	repos.UserSource

	au    auth.Authenticator
	store SourcerStore
}

// NewSourcer returns a new Sourcer to be used in Batch Changes.
func NewSourcer(cf *httpcli.Factory, store SourcerStore) Sourcer {
	return &sourcer{
		cf,
		store,
	}
}

// NewFakeSourcer returns a new Sourcer to be used in Batch Changes.
func NewFakeSourcer(err error, source *BatchesSource) Sourcer {
	return &fakeSourcer{
		err,
		source,
	}
}

func (s *fakeSourcer) ForChangeset(ctx context.Context, ch *btypes.Changeset) (*BatchesSource, error) {
	return s.source, s.err
}

func (s *fakeSourcer) ForRepo(ctx context.Context, repo *types.Repo) (*BatchesSource, error) {
	return s.source, s.err
}

func (s *fakeSourcer) ForExternalService(ctx context.Context, opts store.GetExternalServiceIDsOpts) (*BatchesSource, error) {
	return s.source, s.err
}

// ForChangeset returns a BatchesSource for the given changeset. The changeset.RepoID
// is used to find the matching code host.
func (s *sourcer) ForChangeset(ctx context.Context, ch *btypes.Changeset) (*BatchesSource, error) {
	repo, err := s.store.Repos().Get(ctx, ch.RepoID)
	if err != nil {
		return nil, errors.Wrap(err, "loading changeset repo")
	}
	return s.ForRepo(ctx, repo)
}

// ForRepo returns a BatchesSource for the given repo.
func (s *sourcer) ForRepo(ctx context.Context, repo *types.Repo) (*BatchesSource, error) {
	// Consider all available external services for this repo.
	return s.loadBatchesSource(ctx, repo.ExternalServiceIDs())
}

// ForExternalService returns a BatchesSource based on the provided external service opts.
func (s *sourcer) ForExternalService(ctx context.Context, opts store.GetExternalServiceIDsOpts) (*BatchesSource, error) {
	extSvcIDs, err := s.store.GetExternalServiceIDs(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, "loading external service IDs")
	}
	return s.loadBatchesSource(ctx, extSvcIDs)
}

// FromRepoSource returns a BatchesSource for a given repos.Source.
// func (s *Sourcer) FromRepoSource(src repos.Source) (*BatchesSource, error) {
// 	return batchesSourceFromRepoSource(src, s.store)
// }

func (s *sourcer) loadBatchesSource(ctx context.Context, externalServiceIDs []int64) (*BatchesSource, error) {
	extSvc, err := loadExternalService(ctx, s.store.ExternalServices(), database.ExternalServicesListOptions{
		IDs: externalServiceIDs,
	})
	if err != nil {
		return nil, errors.Wrap(err, "loading external service")
	}
	css, err := buildChangesetSource(s.store, extSvc)
	if err != nil {
		return nil, errors.Wrap(err, "building changeset source")
	}
	// TODO: This should be the default, once we don't use external service tokens anymore.
	// This ensures that we never use a changeset source without an authenticator.
	// cred, err := loadSiteCredential(ctx, s.store, repo)
	// if err != nil {
	// 	return nil, err
	// }
	// if cred != nil {
	// 	return s.WithAuthenticator(css, cred)
	// }
	return css, nil
}

func (s *BatchesSource) GitserverPushConfig(repo *types.Repo) (*protocol.PushConfig, error) {
	extSvcType := repo.ExternalRepo.ServiceType
	cloneURL, err := extractCloneURL(repo)
	if err != nil {
		return nil, err
	}
	u, err := vcs.ParseURL(cloneURL)
	if err != nil {
		return nil, errors.Wrap(err, "parsing repository clone URL")
	}

	if s.au == nil {
		// This is OK: we'll just send no key and gitserver will use
		// the keys installed locally for SSH and the token from the
		// clone URL for https.
		// This path is only triggered when `loadAuthenticator` returns
		// nil, which is only the case for site-admins currently.
		// We want to revisit this once we start disabling usage of global
		// credentials altogether in RFC312.
		return &protocol.PushConfig{RemoteURL: u.String()}, nil
	}

	// If the repo is cloned using SSH, we need to pass along a private key and passphrase.
	if u.Scheme == "ssh" {
		sshA, ok := s.au.(auth.AuthenticatorWithSSH)
		if !ok {
			return nil, ErrNoSSHCredential
		}
		privateKey, passphrase := sshA.SSHPrivateKey()
		return &protocol.PushConfig{
			RemoteURL:  cloneURL,
			PrivateKey: privateKey,
			Passphrase: passphrase,
		}, nil
	}

	switch av := s.au.(type) {
	case *auth.OAuthBearerTokenWithSSH:
		if err := setOAuthTokenAuth(u, extSvcType, av.Token); err != nil {
			return nil, err
		}
	case *auth.OAuthBearerToken:
		if err := setOAuthTokenAuth(u, extSvcType, av.Token); err != nil {
			return nil, err
		}

	case *auth.BasicAuthWithSSH:
		if err := setBasicAuth(u, extSvcType, av.Username, av.Password); err != nil {
			return nil, err
		}
	case *auth.BasicAuth:
		if err := setBasicAuth(u, extSvcType, av.Username, av.Password); err != nil {
			return nil, err
		}
	default:
		return nil, ErrNoPushCredentials{CredentialsType: fmt.Sprintf("%T", s.au)}
	}

	return &protocol.PushConfig{RemoteURL: u.String()}, nil
}

// DraftChangesetSource returns a repos.DraftChangesetSource, if the underlying
// source supports it. Returns an error if not.
func (s *BatchesSource) DraftChangesetSource() (DraftChangesetSource, error) {
	draftCss, ok := s.ChangesetSource.(DraftChangesetSource)
	if !ok {
		return nil, errors.New("changeset source doesn't implement DraftChangesetSource")
	}
	return draftCss, nil
}

func (s *BatchesSource) WithAuthenticatorForActor(ctx context.Context, repo *types.Repo) (*BatchesSource, error) {
	act := actor.FromContext(ctx)
	if !act.IsAuthenticated() {
		return nil, errors.New("cannot get authenticator from actor: no user in context")
	}
	return s.WithAuthenticatorForUser(ctx, act.UID, repo)
}

func (s *BatchesSource) WithAuthenticatorForUser(ctx context.Context, userID int32, repo *types.Repo) (*BatchesSource, error) {
	cred, err := loadUserCredential(ctx, s.store, userID, repo)
	if err != nil {
		return nil, errors.Wrap(err, "loading user credential")
	}
	if cred != nil {
		return s.WithAuthenticator(cred)
	}

	cred, err = loadSiteCredential(ctx, s.store, repo)
	if err != nil {
		return nil, errors.Wrap(err, "loading site credential")
	}
	if cred != nil {
		return s.WithAuthenticator(cred)
	}
	// For now, default to the internal authenticator of the source.
	// This is either a site-credential or the external service token.

	// If neither exist, we need to check if the user is an admin: if they are,
	// then we can use the nil return from loadUserCredential() to fall
	// back to the global credentials used for the code host. If
	// not, then we need to error out.
	// Once we tackle https://github.com/sourcegraph/sourcegraph/issues/16814,
	// this code path should be removed.
	user, err := database.Users(s.store.DB()).GetByID(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load user")
	}
	if user.SiteAdmin {
		return s, nil
	}

	// Otherwise, we can't authenticate the given ChangesetSource, so we need to bail out.
	return nil, ErrMissingCredentials
}

// WithSiteAuthenticator uses the site credential of the code host of the passed-in repo.
// If no credential is found, the original source is returned and uses the external service
// config.
func (s *BatchesSource) WithSiteAuthenticator(ctx context.Context, repo *types.Repo) (*BatchesSource, error) {
	cred, err := loadSiteCredential(ctx, s.store, repo)
	if err != nil {
		return nil, errors.Wrap(err, "loading site credential")
	}
	if cred != nil {
		return s.WithAuthenticator(cred)
	}
	return s, nil
}

func (s *BatchesSource) WithAuthenticator(au auth.Authenticator) (*BatchesSource, error) {
	return authenticateSource(s, au)
}

// loadExternalService looks up all external services that are connected to the given repo.
// The first external service to have a token configured will be returned then.
// If no external service matching the above criteria is found, an error is returned.
func loadExternalService(ctx context.Context, s *database.ExternalServiceStore, opts database.ExternalServicesListOptions) (*types.ExternalService, error) {
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
		cfg, err := e.Configuration()
		if err != nil {
			return nil, err
		}

		switch cfg := cfg.(type) {
		case *schema.GitHubConnection:
			if cfg.Token != "" {
				return e, nil
			}
		case *schema.BitbucketServerConnection:
			if cfg.Token != "" {
				return e, nil
			}
		case *schema.GitLabConnection:
			if cfg.Token != "" {
				return e, nil
			}
		}
	}

	return nil, errors.New("no external services found")
}

// buildChangesetSource get an authenticated ChangesetSource for the given repo
// to load the changeset state from.
func buildChangesetSource(store SourcerStore, externalService *types.ExternalService) (*BatchesSource, error) {
	var source ChangesetSource
	switch externalService.Kind {
	case extsvc.KindGitHub:
		source = &GithubSource{}
	case extsvc.KindGitLab:
		source = &GitLabSource{}
	case extsvc.KindBitbucketServer:
		source = &BitbucketServerSource{}
	default:
		return nil, fmt.Errorf("unsupported external service type %q", extsvc.KindToType(externalService.Kind))
	}
	return batchesSourceFromRepoSource(source, store)
}

func batchesSourceFromRepoSource(src ChangesetSource, store SourcerStore) (*BatchesSource, error) {
	us, ok := src.(repos.UserSource)
	if !ok {
		return nil, fmt.Errorf("cannot create UserSource from external service")
	}
	return &BatchesSource{
		ChangesetSource: src,
		UserSource:      us,
		store:           store,
	}, nil
}

func authenticateSource(src *BatchesSource, au auth.Authenticator) (*BatchesSource, error) {
	repoSource, err := src.UserSource.WithAuthenticator(au)
	if err != nil {
		return nil, err
	}
	clone, err := batchesSourceFromRepoSource(repoSource, src.store)
	if err != nil {
		return nil, err
	}
	clone.au = au
	return clone, nil
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
		return cred.Credential, nil
	}
	return nil, nil
}

// loadSiteCredential attempts to find a site credential for the given repo.
// When no credential is found, nil is returned.
func loadSiteCredential(ctx context.Context, s SourcerStore, repo *types.Repo) (auth.Authenticator, error) {
	cred, err := s.GetSiteCredential(ctx, store.GetSiteCredentialOpts{
		ExternalServiceType: repo.ExternalRepo.ServiceType,
		ExternalServiceID:   repo.ExternalRepo.ServiceID,
	})
	if err != nil && err != store.ErrNoResults {
		return nil, err
	}
	if cred != nil {
		return cred.Credential, nil
	}
	return nil, nil
}

func setOAuthTokenAuth(u *url.URL, extsvcType, token string) error {
	switch extsvcType {
	case extsvc.TypeGitHub:
		u.User = url.User(token)

	case extsvc.TypeGitLab:
		u.User = url.UserPassword("git", token)

	case extsvc.TypeBitbucketServer:
		return errors.New("require username/token to push commits to BitbucketServer")
	}
	return nil
}

func setBasicAuth(u *url.URL, extSvcType, username, password string) error {
	switch extSvcType {
	case extsvc.TypeGitHub, extsvc.TypeGitLab:
		return errors.New("need token to push commits to " + extSvcType)

	case extsvc.TypeBitbucketServer:
		u.User = url.UserPassword(username, password)
	}
	return nil
}

// extractCloneURL returns a remote URL from the repo, preferring HTTPS over SSH.
func extractCloneURL(repo *types.Repo) (string, error) {
	if len(repo.Sources) == 0 {
		return "", errors.New("no clone URL found for repo")
	}
	cloneURLs := make([]*url.URL, 0, len(repo.Sources))
	for _, source := range repo.Sources {
		parsedURL, err := vcs.ParseURL(source.CloneURL)
		if err != nil {
			return "", err
		}
		cloneURLs = append(cloneURLs, parsedURL)
	}
	sort.SliceStable(cloneURLs, func(i, j int) bool {
		return cloneURLs[i].Scheme != "ssh"
	})
	cloneURL := cloneURLs[0]
	// TODO: Do this once we don't want to use existing credentials anymore.
	// // Remove any existing credentials from the clone URL.
	// parsedU.User = nil
	return cloneURL.String(), nil
}
