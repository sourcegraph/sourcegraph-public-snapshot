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
		g.logger.Info("github app installation job", log.String("appName", app.Name), log.Int("id", app.ID))

		client, err := newGithubClient(app, g.logger)
		if err != nil {
			g.logger.Error("creating github client", log.Error(err), log.String("appName", app.Name), log.Int("id", app.ID))
			errs = errors.Append(errs, err)
			continue
		}

		remoteInstallations, err := client.GetAppInstallations(ctx)
		if err != nil {
			g.logger.Error("fetching app installations from GitHub", log.Error(err), log.String("appName", app.Name), log.Int("id", app.ID))
			errs = errors.Append(errs, err)

			// This likely means the App has been deleted from GitHub, so we should remove
			// all installations of it from our database.
			dbInstallations, err := store.GetInstallations(ctx, app.ID)
			if err != nil {
				g.logger.Error("fetching app installations from database", log.Error(err), log.String("appName", app.Name), log.Int("id", app.ID))
				errs = errors.Append(errs, err)
				continue
			}

			var toBeDeleted []int
			for _, install := range dbInstallations {
				toBeDeleted = append(toBeDeleted, install.InstallationID)
			}
			if len(toBeDeleted) > 0 {
				g.logger.Info("deleting github app installations", log.String("appName", app.Name), log.Ints("installationIDs", toBeDeleted))
				err = store.BulkRemoveInstallations(ctx, app.ID, toBeDeleted)
				if err != nil {
					g.logger.Error("failed to delete invalid installations", log.Error(err), log.String("appName", app.Name), log.Int("id", app.ID))
					errs = errors.Append(errs, err)
					continue
				}
			}

			continue
		}

		var remoteInstallsMap = make(map[int]struct{}, len(remoteInstallations))
		for _, in := range remoteInstallations {
			remoteInstallsMap[int(*in.ID)] = struct{}{}
		}

		dbInstallations, err := store.GetInstallations(ctx, app.ID)
		if err != nil {
			g.logger.Error("fetching app installations from database", log.Error(err), log.String("appName", app.Name), log.Int("id", app.ID))
			errs = errors.Append(errs, err)
			continue
		}

		var dbInstallsMap = make(map[int]struct{}, len(dbInstallations))
		for _, in := range dbInstallations {
			dbInstallsMap[in.InstallationID] = struct{}{}
		}

		var toBeAdded []ghtypes.GitHubAppInstallation

		for _, install := range remoteInstallations {
			if install == nil || install.ID == nil {
				continue
			}
			// We add any installation that exists on GitHub regardless of whether or
			// not it already exists in our database, because we will upsert it to
			// ensure that we have the latest metadata for the installation.
			toBeAdded = append(toBeAdded, ghtypes.GitHubAppInstallation{
				InstallationID:   int(install.GetID()),
				AppID:            app.ID,
				URL:              install.GetHTMLURL(),
				AccountLogin:     install.Account.GetLogin(),
				AccountAvatarURL: install.Account.GetAvatarURL(),
				AccountURL:       install.Account.GetHTMLURL(),
				AccountType:      install.Account.GetType(),
			})
			_, exists := dbInstallsMap[int(install.GetID())]
			// If the installation already existed in the DB, we delete it from the
			// map of database installations so that we can determine later which
			// installations need to be deleted from the database. Any installations
			// that remain in the map after this loop will be deleted.
			if exists {
				delete(dbInstallsMap, int(install.GetID()))
			}
		}

		if len(toBeAdded) > 0 {
			for _, install := range toBeAdded {
				g.logger.Info("upserting github app installation", log.String("appName", app.Name), log.Int("installationID", install.InstallationID))
				_, err = store.Install(ctx, install)
				if err != nil {
					g.logger.Error("failed to save new installation", log.Error(err), log.String("appName", app.Name), log.Int("id", app.ID))
					errs = errors.Append(errs, err)
					continue
				}
			}
		}

		if len(dbInstallsMap) > 0 {
			var toBeDeleted []int
			for id := range dbInstallsMap {
				toBeDeleted = append(toBeDeleted, id)
			}
			g.logger.Info("deleting github app installations", log.String("appName", app.Name), log.Ints("installationIDs", toBeDeleted))
			err = store.BulkRemoveInstallations(ctx, app.ID, toBeDeleted)
			if err != nil {
				g.logger.Error("failed to delete invalid installations", log.Error(err), log.String("appName", app.Name), log.Int("id", app.ID))
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
