package worker

import (
	"context"
	"net/url"

	gogithub "github.com/google/go-github/v41/github"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/auth"
	ghtypes "github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewGitHubInstallationWorker returns a goroutine.Handler that will backfill GitHub App
// installation information from the GitHub API into the database.
func NewGitHubInstallationWorker(db database.EnterpriseDB, logger log.Logger) goroutine.Handler {
	return &githubAppInstallationWorker{
		db:     db,
		logger: logger,
	}
}

type githubAppInstallationWorker struct {
	db     database.EnterpriseDB
	logger log.Logger
}

func (g *githubAppInstallationWorker) Handle(ctx context.Context) error {
	store := g.db.GitHubApps()
	apps, err := store.List(ctx, nil)
	if err != nil {
		g.logger.Error("fetching github apps", log.Error(err))
		return errors.Wrap(err, "fetching github apps")
	}

	var errs errors.MultiError
	for _, app := range apps {
		g.logger.Info("github app installation job", log.String("appName", app.Name))

		client, err := newGithubClient(app, g.logger)
		if err != nil {
			g.logger.Error("creating github client", log.Error(err), log.Int("id", app.ID))
			errs = errors.Append(errs, err)
			continue
		}

		remoteInstallations, err := client.GetAppInstallations(ctx)
		if err != nil {
			g.logger.Error("fetching app installations from GitHub", log.Error(err), log.Int("id", app.ID))
			errs = errors.Append(errs, err)
			continue
		}

		var remoteInstallsMap = make(map[int]struct{}, len(remoteInstallations))
		for _, in := range remoteInstallations {
			remoteInstallsMap[int(*in.ID)] = struct{}{}
		}

		dbInstallations, err := store.GetInstallations(ctx, app.ID)
		if err != nil {
			g.logger.Error("fetching app installations from database", log.Error(err), log.Int("id", app.ID))
			errs = errors.Append(errs, err)
			continue
		}

		var dbInstallsMap = make(map[int]struct{}, len(dbInstallations))
		for _, in := range dbInstallations {
			dbInstallsMap[in.InstallationID] = struct{}{}
		}

		var toBeAdded []int
		var toBeDeleted []int

		for id := range dbInstallsMap {
			_, exists := remoteInstallsMap[id]
			if !exists {
				// if the installation id exists in the database but doesn't exist on GitHub, we add it to the
				// slice of installations to be deleted.
				toBeDeleted = append(toBeDeleted, id)
			}
		}

		for id := range remoteInstallsMap {
			_, exists := dbInstallsMap[id]
			if !exists {
				// if the installation exists on GitHub but we don't have it in our database, we add it to the
				// slice of installations to be added.
				toBeAdded = append(toBeAdded, id)
			}
		}

		if len(toBeAdded) > 0 {
			err = store.BulkInstall(ctx, app.ID, toBeAdded)
			if err != nil {
				g.logger.Error("failed to save new installations", log.Error(err), log.Int("id", app.ID))
				errs = errors.Append(errs, err)
				continue
			}
		}

		if len(toBeDeleted) > 0 {
			err = store.BulkRemoveInstallations(ctx, app.ID, toBeDeleted)
			if err != nil {
				g.logger.Error("failed to revoke invalid installations", log.Error(err), log.Int("id", app.ID))
				errs = errors.Append(errs, err)
				continue
			}
		}
	}

	return errs
}

type GitHubAppClient interface {
	GetAppInstallations(context.Context) ([]*gogithub.Installation, error)
}

var MockGitHubClient func(app *ghtypes.GitHubApp, logger log.Logger) (GitHubAppClient, error)

func newGithubClient(app *ghtypes.GitHubApp, logger log.Logger) (GitHubAppClient, error) {
	if MockGitHubClient != nil {
		return MockGitHubClient(app, logger)
	}
	auther, err := auth.NewGitHubAppAuthenticator(app.AppID, []byte(app.PrivateKey))
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(app.BaseURL)
	if err != nil {
		return nil, err
	}

	apiURL, _ := github.APIRoot(baseURL)
	return github.NewV3Client(logger, "", apiURL, auther, nil), nil
}
