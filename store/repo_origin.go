package store

import (
	"crypto/rsa"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/context"
)

// RepoOrigin defines the interface for communicating with the
// external origins corresponding to repos we mirror locally.
//
// E.g., consider the various kinds of data that exists when
// Sourcegraph displays a GitHub repo:
//
// * Some of the data is Sourcegraph-specific data (such as srclib
//   build output, and whether the repo is Sourcegraph-enabled) that
//   external services would not even represent in their data
//   model. This data clearly must be stored in Sourcegraph.
//
// * The basic repo metadata data (such as the repo name and clone
//   URL) is generic. While in this case GitHub supplied the data (and
//   hosts the canonical version of the data), it's generally
//   irrelevant where the data came from.
//
// * Some of the data is specific to the external origin (GitHub),
//   such as whether webhooks, commit status, and SSH keys are set
//   up. While Sourcegraph could represent this data locally, the data
//   only makes sense in the context of the repo's external origin.
//
// RepoOrigin stores the 3rd kind (data that's specific to the
// external origin). It does not persist state locally on
// Sourcegraph's side; it generally reads the state on-demand from the
// external origin (e.g., checking whether a webhook is set up) to
// avoid data inconsistencies.
type RepoOrigin interface {
	// BrandName is the popular name of the product/service that this
	// implementation communicates with (e.g., GitHub, GitHub
	// Enterprise, or Bitbucket).
	BrandName() string

	// Host is the hostname of the server that this implementation
	// communicates with (e.g., github.com for GitHub.com or
	// github.example.com for a GitHub Enterprise installation).
	Host() string

	// A RepoOrigin's functionality is determined by the other Go
	// interfaces (named RepoOriginXxx) that it implements.
}

// A RepoOriginWithPushHooks is a repo external origin that supports
// setting and querying push hooks (so that Sourcegraph is notified
// whenever someone pushes to the origin repo).
type RepoOriginWithPushHooks interface {
	// IsPushHookEnabled determines if the push hook is set up.
	IsPushHookEnabled(ctx context.Context, repo string) (bool, error)

	// EnablePushHook enables a push hook.
	EnablePushHook(ctx context.Context, repo string) error

	// TODO(sqs!nodb-ctx): add DeletePushHook method or something, and
	// add arg for the endpoint URL? and make impls of this interface
	// provide a handler, too?
}

// A RepoOriginWithCommitStatuses is a repo external origin that
// supports publishing commit statuses (so that Sourcegraph can
// publish the build statuses to the external origin).
//
// Currently commit statuses are only published for successful builds
// (until Sourcegraph builds become more reliable).
type RepoOriginWithCommitStatuses interface {
	// IsCommitStatusCapable determines if commit statuses may be
	// published (which may require special authorization, such as on
	// GitHub).
	//
	// Note: Even if IsCommitStatusCapable returns true, we must still
	// check that the repo is enabled (in our local RepoConfig) before
	// publishing commit statuses.
	IsCommitStatusCapable(ctx context.Context, repo string) (bool, error)

	// PublishCommitStatus publishes a commit status.
	PublishCommitStatus(ctx context.Context, repo string, status *sourcegraph.RepoStatus) error
}

// A RepoOriginWithAuthorizedSSHKeys is a repo external origin that
// supports accessing repos via SSH private-key authentication.
//
// SSH key access is generally only necessary for private repositories
// and for write operations on public repositories.
type RepoOriginWithAuthorizedSSHKeys interface {
	// IsSSHKeyAuthorized determines if the given public key is
	// authorized for access to the repo.
	IsSSHKeyAuthorized(ctx context.Context, repo string, key ssh.PublicKey) (bool, error)

	// AuthorizeSSHKey authorizes the keypair for access to the repo.
	AuthorizeSSHKey(ctx context.Context, repo string, key ssh.PublicKey) error

	// DeleteSSHKey deauthorizes and removes a previously authorized
	// SSH keypair.
	DeleteSSHKey(ctx context.Context, repo string, key ssh.PublicKey) error
}

// MirroredRepoSSHKeys defines the interface for stores that persist
// and fetch repository SSH keys (e.g., to access private repos on
// some external origin that we mirror).
//
// Unlike RepoOriginConfig values (which are logically stored on the
// external origin), these keys are persisted locally. This is because
// the origin would typically only store the public key when you
// create a keypair (e.g., on the GitHub API), and we need the private
// key. This type is named "MirroredXxx" not "OriginXxx" to emphasize
// the fact that it is persisted on Sourcegraph, not on the external
// repo origin.
type MirroredRepoSSHKeys interface {
	Create(ctx context.Context, repo string, privKey *rsa.PrivateKey) error
	GetPEM(ctx context.Context, repo string) ([]byte, error)
	Delete(ctx context.Context, repo string) error
}
