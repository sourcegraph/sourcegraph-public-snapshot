package backend

import (
	"context"
	json "encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

const syncExternalServiceTimeout = 15 * time.Second

type ExternalServicesService interface {
	SyncExternalService(context.Context, *types.ExternalService, time.Duration) error
	ExcludeRepoFromExternalService(context.Context, int64, api.RepoID) error
}

type externalServices struct {
	logger            log.Logger
	db                database.DB
	repoupdaterClient *repoupdater.Client
}

func NewExternalServices(logger log.Logger, db database.DB, repoupdaterClient *repoupdater.Client) ExternalServicesService {
	return &externalServices{
		logger:            logger,
		db:                db,
		repoupdaterClient: repoupdaterClient,
	}
}

// SyncExternalService will eagerly trigger a repo-updater sync. It accepts a
// timeout as an argument which is recommended to be 5 seconds unless the caller
// has special requirements for it to be larger or smaller.
func (e *externalServices) SyncExternalService(ctx context.Context, svc *types.ExternalService, timeout time.Duration) (err error) {
	logger := e.logger.Scoped("SyncExternalService", "handles triggering of repo-updater syncing for a particular external service")
	// Set a timeout to validate external service sync. It usually fails in
	// under 5s if there is a problem.
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	defer func() {
		// err is either nil or contains an actual error from the API call. And we return it
		// nonetheless.
		err = errors.Wrapf(err, "error in SyncExternalService for service %q with ID %d", svc.Kind, svc.ID)

		// If context error is anything but a deadline exceeded error, we do not want to propagate
		// it. But we definitely want to log the error as a warning.
		if ctx.Err() != nil && ctx.Err() != context.DeadlineExceeded {
			logger.Warn("context error discarded", log.Error(ctx.Err()))
			err = nil
		}
	}()

	_, err = e.repoupdaterClient.SyncExternalService(ctx, svc.ID)
	return err
}

// ExcludeRepoFromExternalService excludes given repo from given external service config.
//
// Function is pretty beefy, what it does is:
// - finds an external service by ID and checks if it supports repo exclusion
// - adds repo to `exclude` config parameter and updates an external service
// - triggers external service sync
func (e *externalServices) ExcludeRepoFromExternalService(ctx context.Context, externalServiceID int64, repoID api.RepoID) (err error) {
	tx, err := e.db.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	externalServices := tx.ExternalServices()
	externalService, err := externalServices.GetByID(ctx, externalServiceID)
	if err != nil {
		return err
	}

	logger := e.logger.Scoped("ExcludeRepoFromExternalService", "excluding a repo from external service config").With(
		log.Int64("externalServiceID", externalServiceID),
		log.Int32("repoID", int32(repoID)),
	)

	// If external service doesn't support repo exclusion, then return.
	if !externalService.SupportsRepoExclusion() {
		logger.Warn("external service does not support repo exclusion")
		return nil
	}

	repository, err := tx.Repos().Get(ctx, repoID)
	if err != nil {
		return err
	}

	updatedConfig, err := addRepoToExclude(ctx, externalService, repository)
	if err != nil {
		return err
	}

	err = externalServices.Update(ctx, conf.Get().AuthProviders, externalServiceID, &database.ExternalServiceUpdate{Config: &updatedConfig})
	if err != nil {
		return err
	}

	// Error during triggering a sync is omitted, because this should not prevent
	// from excluding the repo. The repo stays excluded and the sync will come
	// eventually.
	err = e.SyncExternalService(ctx, externalService, syncExternalServiceTimeout)
	if err != nil {
		logger.Warn("Failed to trigger external service sync after adding a repo exclusion.")
	}
	return nil
}

func addRepoToExclude(ctx context.Context, externalService *types.ExternalService, repository *types.Repo) (string, error) {
	config, err := externalService.Configuration(ctx)
	if err != nil {
		return "", err
	}

	// we need to use a `org/repo` repo name format in `exclude` section of code host
	// config, hence trimming the host: `github.com/sourcegraph/sourcegraph` becomes
	// `sourcegraph/sourcegraph`.
	repoName := trimHostFromRepoName(string(repository.Name))

	switch c := config.(type) {
	case *schema.AWSCodeCommitConnection:
		exclusion := &schema.ExcludedAWSCodeCommitRepo{Id: strconv.FormatInt(int64(repository.ID), 10), Name: repoName}
		if !schemaContainsExclusion(c.Exclude, exclusion) {
			c.Exclude = append(c.Exclude, &schema.ExcludedAWSCodeCommitRepo{Id: strconv.FormatInt(int64(repository.ID), 10), Name: repoName})
		}
	case *schema.BitbucketCloudConnection:
		exclusion := &schema.ExcludedBitbucketCloudRepo{Name: repoName}
		if !schemaContainsExclusion(c.Exclude, exclusion) {
			c.Exclude = append(c.Exclude, &schema.ExcludedBitbucketCloudRepo{Name: repoName})
		}
	case *schema.BitbucketServerConnection:
		exclusion := &schema.ExcludedBitbucketServerRepo{Id: int(repository.ID), Name: repoName}
		if !schemaContainsExclusion(c.Exclude, exclusion) {
			c.Exclude = append(c.Exclude, &schema.ExcludedBitbucketServerRepo{Id: int(repository.ID), Name: repoName})
		}
	case *schema.GitHubConnection:
		exclusion := &schema.ExcludedGitHubRepo{Id: strconv.FormatInt(int64(repository.ID), 10), Name: repoName}
		if !schemaContainsExclusion(c.Exclude, exclusion) {
			c.Exclude = append(c.Exclude, &schema.ExcludedGitHubRepo{Id: strconv.FormatInt(int64(repository.ID), 10), Name: repoName})
		}
	case *schema.GitLabConnection:
		exclusion := &schema.ExcludedGitLabProject{Name: repoName}
		if !schemaContainsExclusion(c.Exclude, exclusion) {
			c.Exclude = append(c.Exclude, &schema.ExcludedGitLabProject{Name: repoName})
		}
	case *schema.GitoliteConnection:
		exclusion := &schema.ExcludedGitoliteRepo{Name: repoName}
		if !schemaContainsExclusion(c.Exclude, exclusion) {
			c.Exclude = append(c.Exclude, &schema.ExcludedGitoliteRepo{Name: repoName})
		}
	}

	strConfig, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(strConfig), nil
}

// trimHostFromRepoName removes host from full repo name. It does it in the same
// way as it is done before rendering repo name on the UI. See RepoLink
// component.
func trimHostFromRepoName(name string) string {
	parts := strings.Split(name, "/")
	if len(parts) >= 3 && strings.Contains(parts[0], ".") {
		return strings.Join(parts[1:], "/")
	}
	return strings.Join(parts, "/")
}

func schemaContainsExclusion[T comparable](exclusions []*T, newExclusion *T) bool {
	for _, exclusion := range exclusions {
		if *exclusion == *newExclusion {
			return true
		}
	}
	return false
}
