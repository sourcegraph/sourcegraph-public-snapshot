package worker

import (
	"context"
	"net/url"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewGitHubInstallationHandler(db database.EnterpriseDB) goroutine.Handler {
	return &githubAppInstallationHandler{
		db: db,
	}
}

type githubAppInstallationHandler struct {
	db database.EnterpriseDB
}

func (g *githubAppInstallationHandler) Handle(ctx context.Context) error {
	logger := log.Scoped("GitHubAppInstallationWorker", "")

	store := g.db.GitHubApps()
	apps, err := store.List(ctx, nil)
	if err != nil {
		logger.Error("fetching github apps", log.Error(err))
		return errors.Wrap(err, "fetching github apps")
	}

	var errs errors.MultiError
	for _, app := range apps {
		logger.Info("github app installation job", log.String("appName", app.Name))
		auther, err := auth.NewGitHubAppAuthenticator(app.AppID, []byte(app.PrivateKey))
		if err != nil {
			logger.Error("fetching installation token", log.Error(err), log.Int("id", app.ID))
			errs = errors.Append(errs, err)
			continue
		}

		baseURL, err := url.Parse(app.BaseURL)
		if err != nil {
			logger.Error("parsing github app base URL", log.Error(err), log.Int("id", app.ID))
			errs = errors.Append(errs, err)
			continue
		}

		apiURL, _ := github.APIRoot(baseURL)
		client := github.NewV3Client(logger, "", apiURL, auther, nil)

		remoteInstallations, err := client.GetAppInstallations(ctx)
		if err != nil {
			logger.Error("fetching app installations from GitHub", log.Error(err), log.Int("id", app.ID))
			errs = errors.Append(errs, err)
			continue
		}

		var remoteInstallsMap = make(map[int]struct{}, len(remoteInstallations))
		for _, in := range remoteInstallations {
			remoteInstallsMap[int(*in.ID)] = struct{}{}
		}

		dbInstallations, err := store.GetInstallations(ctx, app.ID)
		if err != nil {
			logger.Error("fetching app installations from database", log.Error(err), log.Int("id", app.ID))
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
				logger.Error("failed to save new installations", log.Error(err), log.Int("id", app.ID))
				errs = errors.Append(errs, err)
				continue
			}
		}

		if len(toBeDeleted) > 0 {
			err = store.BulkRevoke(ctx, app.ID, toBeDeleted)
			if err != nil {
				logger.Error("failed to revoke invalid installations", log.Error(err), log.Int("id", app.ID))
				errs = errors.Append(errs, err)
				continue
			}
		}
	}

	return errs
}
