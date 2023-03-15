package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

const syncExternalServiceTimeout = 15 * time.Second

type ExternalServicesService interface {
	SyncExternalService(context.Context, *types.ExternalService, time.Duration) error
	ExcludeRepoFromExternalServices(context.Context, []int64, api.RepoID) error
}

type externalServices struct {
	logger            log.Logger
	db                database.DB
	repoupdaterClient *repoupdater.Client
}

func NewExternalServices(logger log.Logger, db database.DB, repoupdaterClient *repoupdater.Client) ExternalServicesService {
	return &externalServices{
		logger:            logger.Scoped("ExternalServices", "service related to external service functionality"),
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
func (e *externalServices) ExcludeRepoFromExternalServices(ctx context.Context, externalServiceIDs []int64, repoID api.RepoID) error {
	// 🚨 SECURITY: check whether user is site-admin
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, e.db); err != nil {
		return err
	}

	logger := e.logger.Scoped("ExcludeRepoFromExternalServices", "excluding a repo from external service config").With(log.Int32("repoID", int32(repoID)))
	for _, extSvcID := range externalServiceIDs {
		logger = logger.With(log.Int64("externalServiceID", extSvcID))
	}

	externalServices, err := e.updateExternalServiceToExcludeRepo(ctx, logger, externalServiceIDs, repoID)
	if err != nil {
		return err
	}
	// Error during triggering a sync is omitted, because this should not prevent
	// from excluding the repo. The repo stays excluded and the sync will come
	// eventually.
	for _, externalService := range externalServices {
		err = e.SyncExternalService(ctx, externalService, syncExternalServiceTimeout)
		if err != nil {
			logger.Warn("Failed to trigger external service sync after adding a repo exclusion.")
		}
	}
	return nil
}

func (e *externalServices) updateExternalServiceToExcludeRepo(
	ctx context.Context,
	logger log.Logger,
	externalServiceIDs []int64,
	repoID api.RepoID,
) (externalServices []*types.ExternalService, err error) {
	err = e.db.WithTransact(ctx, func(tx database.DB) error {
		extSvcStore := tx.ExternalServices()
		externalServices, err = extSvcStore.List(ctx, database.ExternalServicesListOptions{IDs: externalServiceIDs})
		if err != nil {
			return err
		}

		for _, externalService := range externalServices {
			// If external service doesn't support repo exclusion, then return.
			if !externalService.SupportsRepoExclusion() {
				logger.Warn("external service does not support repo exclusion")
				return errors.New("external service does not support repo exclusion")
			}
		}

		repository, err := tx.Repos().Get(ctx, repoID)
		if err != nil {
			return err
		}

		for _, externalService := range externalServices {
			updatedConfig, err := addRepoToExclude(ctx, logger, externalService, repository)
			if err != nil {
				return err
			}
			if err = extSvcStore.Update(ctx, conf.Get().AuthProviders, externalService.ID, &database.ExternalServiceUpdate{Config: &updatedConfig}); err != nil {
				return err
			}
		}

		return nil
	})
	return externalServices, err
}

func addRepoToExclude(ctx context.Context, logger log.Logger, externalService *types.ExternalService, repository *types.Repo) (string, error) {
	config, err := externalService.Configuration(ctx)
	if err != nil {
		return "", err
	}

	// We need to use a name different from `types.Repo.Name` in order for repo to be
	// excluded.
	excludableName := ExcludableRepoName(repository, logger)
	if excludableName == "" {
		return "", errors.New("repository lacks metadata to compose excludable name")
	}

	switch c := config.(type) {
	case *schema.AWSCodeCommitConnection:
		exclusion := &schema.ExcludedAWSCodeCommitRepo{Name: excludableName}
		if !schemaContainsExclusion(c.Exclude, exclusion) {
			c.Exclude = append(c.Exclude, &schema.ExcludedAWSCodeCommitRepo{Name: excludableName})
		}
	case *schema.BitbucketCloudConnection:
		exclusion := &schema.ExcludedBitbucketCloudRepo{Name: excludableName}
		if !schemaContainsExclusion(c.Exclude, exclusion) {
			c.Exclude = append(c.Exclude, &schema.ExcludedBitbucketCloudRepo{Name: excludableName})
		}
	case *schema.BitbucketServerConnection:
		exclusion := &schema.ExcludedBitbucketServerRepo{Name: excludableName}
		if !schemaContainsExclusion(c.Exclude, exclusion) {
			c.Exclude = append(c.Exclude, &schema.ExcludedBitbucketServerRepo{Name: excludableName})
		}
	case *schema.GitHubConnection:
		exclusion := &schema.ExcludedGitHubRepo{Name: excludableName}
		if !schemaContainsExclusion(c.Exclude, exclusion) {
			c.Exclude = append(c.Exclude, &schema.ExcludedGitHubRepo{Name: excludableName})
		}
	case *schema.GitLabConnection:
		exclusion := &schema.ExcludedGitLabProject{Name: excludableName}
		if !schemaContainsExclusion(c.Exclude, exclusion) {
			c.Exclude = append(c.Exclude, &schema.ExcludedGitLabProject{Name: excludableName})
		}
	case *schema.GitoliteConnection:
		exclusion := &schema.ExcludedGitoliteRepo{Name: excludableName}
		if !schemaContainsExclusion(c.Exclude, exclusion) {
			c.Exclude = append(c.Exclude, &schema.ExcludedGitoliteRepo{Name: excludableName})
		}
	}

	strConfig, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(strConfig), nil
}

// ExcludableRepoName returns repo name which should be specified in code host
// config `exclude` section in order to be excluded from syncing.
func ExcludableRepoName(repository *types.Repo, logger log.Logger) (name string) {
	typ, _ := extsvc.ParseServiceType(repository.ExternalRepo.ServiceType)
	switch typ {
	case extsvc.TypeAWSCodeCommit:
		if repo, ok := repository.Metadata.(*awscodecommit.Repository); ok {
			name = repo.Name
		} else {
			logger.Error("invalid repo metadata schema", log.String("extSvcType", extsvc.TypeAWSCodeCommit))
		}
	case extsvc.TypeBitbucketCloud:
		if repo, ok := repository.Metadata.(*bitbucketcloud.Repo); ok {
			name = repo.FullName
		} else {
			logger.Error("invalid repo metadata schema", log.String("extSvcType", extsvc.TypeBitbucketCloud))
		}
	case extsvc.TypeBitbucketServer:
		if repo, ok := repository.Metadata.(*bitbucketserver.Repo); ok {
			if repo.Project == nil {
				return
			}
			name = fmt.Sprintf("%s/%s", repo.Project.Name, repo.Name)
		} else {
			logger.Error("invalid repo metadata schema", log.String("extSvcType", extsvc.TypeBitbucketServer))
		}
	case extsvc.TypeGitHub:
		if repo, ok := repository.Metadata.(*github.Repository); ok {
			name = repo.NameWithOwner
		} else {
			logger.Error("invalid repo metadata schema", log.String("extSvcType", extsvc.TypeGitHub))
		}
	case extsvc.TypeGitLab:
		if project, ok := repository.Metadata.(*gitlab.Project); ok {
			name = project.PathWithNamespace
		} else {
			logger.Error("invalid repo metadata schema", log.String("extSvcType", extsvc.TypeGitLab))
		}
	case extsvc.TypeGitolite:
		if repo, ok := repository.Metadata.(*gitolite.Repo); ok {
			name = repo.Name
		} else {
			logger.Error("invalid repo metadata schema", log.String("extSvcType", extsvc.TypeGitolite))
		}
	}
	return
}

func schemaContainsExclusion[T comparable](exclusions []*T, newExclusion *T) bool {
	for _, exclusion := range exclusions {
		if *exclusion == *newExclusion {
			return true
		}
	}
	return false
}
