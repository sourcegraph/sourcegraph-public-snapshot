package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	internalrepos "github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type ExternalServicesService interface {
	ValidateConnection(ctx context.Context, svc *types.ExternalService) error
	ListNamespaces(ctx context.Context, externalServiceID *int64, kind string, config string) ([]*types.ExternalServiceNamespace, error)
	DiscoverRepos(ctx context.Context, externalServiceID *int64, kind string, config string, first int32, query string, excludeRepos []string) ([]*types.ExternalServiceRepository, error)
	ExcludeRepoFromExternalServices(context.Context, []int64, api.RepoID) error
}

type externalServices struct {
	logger log.Logger
	db     database.DB

	mockSourcer internalrepos.Sourcer
}

func NewExternalServices(logger log.Logger, db database.DB) ExternalServicesService {
	return &externalServices{
		logger: logger.Scoped("ExternalServices"),
		db:     db,
	}
}

func NewMockExternalServices(logger log.Logger, db database.DB, mockSourcer internalrepos.Sourcer) ExternalServicesService {
	return &externalServices{
		logger:      logger.Scoped("ExternalServices"),
		db:          db,
		mockSourcer: mockSourcer,
	}
}

const validateConnectionTimeout = 15 * time.Second

func (e *externalServices) ValidateConnection(ctx context.Context, svc *types.ExternalService) error {
	ctx, cancel := context.WithTimeout(ctx, validateConnectionTimeout)
	defer cancel()

	genericSourcer := newGenericSourcer(log.Scoped("externalservice.validateconnection"), e.db)
	genericSrc, err := genericSourcer(ctx, svc)
	if err != nil {
		if ctx.Err() != nil && ctx.Err() == context.DeadlineExceeded {
			return errors.Newf("failed to validate external service connection within %s", validateConnectionTimeout)
		}
		return err
	}

	return externalServiceValidate(ctx, genericSrc)
}

func externalServiceValidate(ctx context.Context, src internalrepos.Source) error {
	if v, ok := src.(internalrepos.UserSource); ok {
		return v.ValidateAuthenticator(ctx)
	}

	ctx, cancel := context.WithCancel(ctx)
	results := make(chan internalrepos.SourceResult)

	defer func() {
		cancel()

		// We need to drain the rest of the results to not leak a blocked goroutine.
		for range results {
		}
	}()

	go func() {
		src.ListRepos(ctx, results)
		close(results)
	}()

	select {
	case res := <-results:
		// As soon as we get the first result back, we've got what we need to validate the external service.
		return res.Err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (e *externalServices) ListNamespaces(ctx context.Context, externalServiceID *int64, kind string, config string) ([]*types.ExternalServiceNamespace, error) {
	var externalSvc *types.ExternalService
	if externalServiceID != nil {
		var err error
		externalSvc, err = e.db.ExternalServices().GetByID(ctx, *externalServiceID)
		if err != nil {
			return nil, err
		}
	} else {
		externalSvc = &types.ExternalService{
			Kind:   kind,
			Config: extsvc.NewUnencryptedConfig(config),
		}
	}

	var (
		genericSrc internalrepos.Source
		err        error
	)
	if e.mockSourcer != nil {
		genericSrc, err = e.mockSourcer(ctx, externalSvc)
		if err != nil {
			return nil, err
		}
	} else {
		genericSourcer := newGenericSourcer(log.Scoped("externalservice.namespacediscovery"), e.db)
		genericSrc, err = genericSourcer(ctx, externalSvc)
		if err != nil {
			return nil, err
		}
	}

	if err := genericSrc.CheckConnection(ctx); err != nil {
		return nil, err
	}

	discoverableSrc, ok := genericSrc.(internalrepos.DiscoverableSource)
	if !ok {
		return nil, errors.New(internalrepos.UnimplementedDiscoverySource)
	}

	results := make(chan internalrepos.SourceNamespaceResult)
	go func() {
		discoverableSrc.ListNamespaces(ctx, results)
		close(results)
	}()

	var sourceErrs error
	namespaces := make([]*types.ExternalServiceNamespace, 0)

	for res := range results {
		if res.Err != nil {
			sourceErrs = errors.Append(sourceErrs, &internalrepos.SourceError{Err: res.Err, ExtSvc: externalSvc})
			continue
		}
		namespaces = append(namespaces, &types.ExternalServiceNamespace{
			ID:         res.Namespace.ID,
			Name:       res.Namespace.Name,
			ExternalID: res.Namespace.ExternalID,
		})
	}

	if sourceErrs != nil {
		return nil, err
	}

	return namespaces, nil
}

func (e *externalServices) DiscoverRepos(ctx context.Context, externalServiceID *int64, kind string, config string, first int32, query string, excludeRepos []string) ([]*types.ExternalServiceRepository, error) {
	var externalSvc *types.ExternalService
	if externalServiceID != nil {
		var err error
		externalSvc, err = e.db.ExternalServices().GetByID(ctx, *externalServiceID)
		if err != nil {
			return nil, err
		}
	} else {
		externalSvc = &types.ExternalService{
			Kind:   kind,
			Config: extsvc.NewUnencryptedConfig(config),
		}
	}

	var (
		genericSrc internalrepos.Source
		err        error
	)
	if e.mockSourcer != nil {
		genericSrc, err = e.mockSourcer(ctx, externalSvc)
		if err != nil {
			return nil, err
		}
	} else {
		genericSourcer := newGenericSourcer(log.Scoped("externalservice.repodiscovery"), e.db)
		genericSrc, err = genericSourcer(ctx, externalSvc)
		if err != nil {
			return nil, err
		}
	}

	if err = genericSrc.CheckConnection(ctx); err != nil {
		return nil, err
	}

	discoverableSrc, ok := genericSrc.(internalrepos.DiscoverableSource)
	if !ok {
		return nil, errors.New(internalrepos.UnimplementedDiscoverySource)
	}

	results := make(chan internalrepos.SourceResult)

	if first > 100 {
		first = 100
	}

	go func() {
		discoverableSrc.SearchRepositories(ctx, query, int(first), excludeRepos, results)
		close(results)
	}()

	var sourceErrs error
	repositories := make([]*types.ExternalServiceRepository, 0)

	for res := range results {
		if res.Err != nil {
			sourceErrs = errors.Append(sourceErrs, &internalrepos.SourceError{Err: res.Err, ExtSvc: externalSvc})
			continue
		}
		repositories = append(repositories, &types.ExternalServiceRepository{
			ID:         res.Repo.ID,
			Name:       res.Repo.Name,
			ExternalID: res.Repo.ExternalRepo.ID,
		})
	}

	if sourceErrs != nil {
		return nil, sourceErrs
	}

	return repositories, nil
}

// ExcludeRepoFromExternalServices excludes given repo from given external service config.
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

	logger := e.logger.Scoped("ExcludeRepoFromExternalServices").With(log.Int32("repoID", int32(repoID)))
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
	s := internalrepos.NewStore(logger, e.db)
	for _, externalService := range externalServices {
		err = s.EnqueueSingleSyncJob(ctx, externalService.ID)
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
			name = fmt.Sprintf("%s/%s", repo.Project.Key, repo.Name)
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

func newGenericSourcer(logger log.Logger, db database.DB) internalrepos.Sourcer {
	// We use the generic sourcer that doesn't have observability attached to it here because the way externalServiceValidate is set up,
	// using the regular sourcer will cause a large dump of errors to be logged when it exits ListRepos prematurely.
	sourcerLogger := logger.Scoped("repos.Sourcer")
	db = database.NewDBWith(sourcerLogger.Scoped("db"), db)
	dependenciesService := dependencies.NewService(observation.NewContext(logger), db)
	cf := httpcli.NewExternalClientFactory(httpcli.NewLoggingMiddleware(sourcerLogger))
	return internalrepos.NewSourcer(sourcerLogger, db, cf, internalrepos.WithDependenciesService(dependenciesService))
}
