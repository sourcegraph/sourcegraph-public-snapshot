package types

import (
	"context"
	"time"

	gogithub "github.com/google/go-github/v55/github"

	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GitHubApp represents a GitHub App.
type GitHubApp struct {
	ID              int
	AppID           int
	Name            string
	Domain          types.GitHubAppDomain
	Slug            string
	BaseURL         string
	AppURL          string
	ClientID        string
	ClientSecret    string
	WebhookSecret   string
	WebhookID       *int
	PrivateKey      string
	EncryptionKey   string
	Logo            string
	Kind            GitHubAppKind
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CreatedByUserId int
}

type GitHubAppKind string

const (
	CommitSigningGitHubAppKind  GitHubAppKind = "COMMIT_SIGNING"
	RepoSyncGitHubAppKind       GitHubAppKind = "REPO_SYNC"
	UserCredentialGitHubAppKind GitHubAppKind = "USER_CREDENTIAL"
	SiteCredentialGitHubAppKind GitHubAppKind = "SITE_CREDENTIAL"
)

func (s GitHubAppKind) Valid() bool {
	switch s {
	case CommitSigningGitHubAppKind,
		RepoSyncGitHubAppKind,
		UserCredentialGitHubAppKind,
		SiteCredentialGitHubAppKind:
		return true
	default:
		return false
	}
}

func (s GitHubAppKind) Validate() (GitHubAppKind, error) {
	if !s.Valid() {
		return "", errors.Newf("Not a valid GitHubAppKind: %s", s)
	}
	return s, nil
}

// GitHubAppInstallation represents an installation of a GitHub App.
type GitHubAppInstallation struct {
	ID               int
	AppID            int
	InstallationID   int
	URL              string
	AccountLogin     string
	AccountAvatarURL string
	AccountURL       string
	AccountType      string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type GitHubAppClient interface {
	GetAppInstallations(ctx context.Context, page int) (_ []*gogithub.Installation, hasNextPage bool, _ error)
}
