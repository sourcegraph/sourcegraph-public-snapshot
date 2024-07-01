package types

import (
	"context"
	"time"

	gogithub "github.com/google/go-github/v55/github"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

// GitHubApp represents a GitHub App.
type GitHubApp struct {
	ID            int
	AppID         int
	Name          string
	Domain        types.GitHubAppDomain
	Slug          string
	BaseURL       string
	AppURL        string
	ClientID      string
	ClientSecret  string
	WebhookSecret string
	WebhookID     *int
	PrivateKey    string
	EncryptionKey string
	Logo          string
	Kind          GitHubAppKind
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type GitHubAppKind string

const (
	CommitSigningGitHubAppKind  GitHubAppKind = "COMMIT_SIGNING"
	RepoSyncGitHubAppKind       GitHubAppKind = "REPO_SYNC"
	UserCredentialGitHubAppKind GitHubAppKind = "USER_CREDENTIAL"
	SiteCredentialGitHubAppKind GitHubAppKind = "SITE_CREDENTIAL"
)

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
