package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/service/servegit"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type appResolver struct {
	logger    log.Logger
	db        database.DB
	gitClient gitserver.Client
}

var _ graphqlbackend.AppResolver = &appResolver{}

func NewAppResolver(logger log.Logger, db database.DB, gitClient gitserver.Client) *appResolver {
	return &appResolver{
		logger:    logger,
		db:        db,
		gitClient: gitClient,
	}
}

func (r *appResolver) checkLocalDirectoryAccess(ctx context.Context) error {
	return auth.CheckCurrentUserIsSiteAdmin(ctx, r.db)
}

func (r *appResolver) LocalDirectories(ctx context.Context, args *graphqlbackend.LocalDirectoryArgs) (graphqlbackend.LocalDirectoryResolver, error) {
	// 🚨 SECURITY: Only site admins on app may use API which accesses local filesystem.
	if err := r.checkLocalDirectoryAccess(ctx); err != nil {
		return nil, err
	}

	// Make sure all paths are absolute
	absPaths := make([]string, 0, len(args.Paths))
	for _, path := range args.Paths {
		if path == "" {
			return nil, errors.New("Path must be non-empty string")
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		absPaths = append(absPaths, absPath)
	}

	return &localDirectoryResolver{paths: absPaths}, nil
}

func (r *appResolver) SetupNewAppRepositoriesForEmbedding(ctx context.Context, args graphqlbackend.SetupNewAppRepositoriesForEmbeddingArgs) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins may schedule embedding jobs.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	// Create a global policy to embed all the repos
	err := createGlobalEmbeddingsPolicy(ctx)
	if err != nil {
		r.logger.Error("unable to create a global indexing policy", log.Error(err))
	}

	repoEmbeddingsStore := repo.NewRepoEmbeddingJobsStore(r.db)
	jobContext, cancel := context.WithDeadline(ctx, time.Now().Add(60*time.Second))
	defer cancel()
	p := pool.New().WithMaxGoroutines(10).WithContext(jobContext)
	for _, repo := range args.RepoNames {
		repoName := api.RepoName(repo)
		p.Go(func(ctx context.Context) error {
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-jobContext.Done():
					return errors.New("time limit exceeded unable to schedule repo")
				case <-ticker.C:
					r.logger.Debug("Checking repo")
					branch, _, err := r.gitClient.GetDefaultBranch(ctx, repoName, true)
					if err == nil && branch != "" {
						if err := embeddings.ScheduleRepositories(
							ctx,
							[]api.RepoName{repoName},
							false,
							r.db,
							repoEmbeddingsStore,
							r.gitClient,
						); err == nil {
							r.logger.Debug("Repo scheduled")
							return nil
						}
					}
					r.logger.Debug("Repo not cloned")
				}
			}
		})
	}
	err = p.Wait()
	if err != nil {
		r.logger.Warn("error scheduling repos for embedding", log.Error(err))
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *appResolver) EmbeddingsSetupProgress(ctx context.Context, args graphqlbackend.EmbeddingSetupProgressArgs) (graphqlbackend.EmbeddingsSetupProgressResolver, error) {
	// 🚨 SECURITY: Only site admins may schedule embedding jobs.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	return &embeddingsSetupProgressResolver{repos: args.RepoNames, db: r.db}, nil
}

func (r *appResolver) AddLocalRepositories(ctx context.Context, args graphqlbackend.AddLocalRepositoriesArgs) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins may schedule embedding jobs.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	if envvar.ExtsvcConfigFile() != "" && !envvar.ExtsvcConfigAllowEdits() {
		return nil, errors.New("adding external service not allowed when using EXTSVC_CONFIG_FILE")
	}

	var services []*types.ExternalService

	// Inspect paths and append /* if the target is not a git repo, to
	// create a blob pattern that matches all repos inside the path.
	for _, path := range args.Paths {
		if !isGitRepo(path) {
			path = filepath.Join(path, "*")
		}

		serviceConfig, err := json.Marshal(schema.LocalGitExternalService{
			Repos: []*schema.LocalGitRepoPattern{{Pattern: path}},
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal external service configuration")
		}

		services = append(services, &types.ExternalService{
			Kind:        extsvc.VariantLocalGit.AsKind(),
			DisplayName: fmt.Sprintf("Local repositories (%s)", path),
			Config:      extsvc.NewUnencryptedConfig(string(serviceConfig)),
		})
	}

	for _, service := range services {
		err := r.db.ExternalServices().Create(ctx, conf.Get, service)
		if err != nil {
			return nil, err
		}
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *appResolver) LocalExternalServices(ctx context.Context) ([]graphqlbackend.LocalExternalServiceResolver, error) {
	// 🚨 SECURITY: Only site admins on app may use API which accesses local filesystem.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	externalServices, err := backend.NewAppExternalServices(r.db).LocalExternalServices(ctx)
	if err != nil {
		return nil, err
	}

	var localExternalServices []graphqlbackend.LocalExternalServiceResolver
	for _, externalService := range externalServices {
		config, err := extsvc.ParseEncryptableConfig(ctx, externalService.Kind, externalService.Config)
		if err != nil {
			return nil, err
		}
		localExternalServices = append(localExternalServices, localExternalServiceResolver{
			config:  config,
			service: externalService,
			db:      r.db,
		})
	}

	return localExternalServices, nil
}

type localDirectoryResolver struct {
	paths []string
}

func (r *localDirectoryResolver) Paths() []string {
	return r.paths
}

func (r *localDirectoryResolver) Repositories(ctx context.Context) ([]graphqlbackend.LocalRepositoryResolver, error) {
	var allRepos []graphqlbackend.LocalRepositoryResolver

	for _, path := range r.paths {
		repos, err := servegit.Service.Repos(ctx, path)
		if err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			allRepos = append(allRepos, localRepositoryResolver{
				name: repo.Name,
				path: repo.AbsFilePath,
			})
		}
	}

	return allRepos, nil
}

type localRepositoryResolver struct {
	name string
	path string
}

func (r localRepositoryResolver) Name() string {
	return r.name
}

func (r localRepositoryResolver) Path() string {
	return r.path
}

type localExternalServiceResolver struct {
	service *types.ExternalService
	db      database.DB
	config  any
}

func (r localExternalServiceResolver) ID() graphql.ID {
	return graphqlbackend.MarshalExternalServiceID(r.service.ID)
}

func (r localExternalServiceResolver) Path() string {
	switch c := r.config.(type) {
	case *schema.OtherExternalServiceConnection:
		return c.Root
	case *schema.LocalGitExternalService:
		var patterns []string
		for _, repo := range c.Repos {
			patterns = append(patterns, repo.Pattern)
		}
		// This will almost always be only a single path, but the automatically generated
		// local git service from the config file can specify multiple.
		return strings.Join(patterns, ",")
	}

	return ""
}

func (r localExternalServiceResolver) Autogenerated() bool {
	return r.service.ID == servegit.ExtSVCID
}

// Repositories returns the configured repositories as they exist on the filesystem. Due to scheduling delays it can take
// some until repositories are synced from the service to the DB and so we cannot rely on the DB in this case.
func (r localExternalServiceResolver) Repositories(ctx context.Context) ([]graphqlbackend.LocalRepositoryResolver, error) {
	// 🚨 SECURITY: Only site admins on app may use API which accesses local filesystem.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	var allRepos []graphqlbackend.LocalRepositoryResolver

	switch c := r.config.(type) {
	case *schema.OtherExternalServiceConnection:
		absPath, err := filepath.Abs(c.Root)
		if err != nil {
			return nil, err
		}
		repos, err := servegit.Service.Repos(ctx, absPath)
		if err != nil {
			return nil, err
		}

		for _, r := range repos {
			allRepos = append(allRepos, localRepositoryResolver{
				name: r.Name,
				path: r.AbsFilePath,
			})
		}
	case *schema.LocalGitExternalService:
		src, err := repos.NewLocalGitSource(ctx, log.Scoped("localExternalServiceResolver.Repositories"), r.service)
		if err != nil {
			return nil, err
		}
		for _, r := range src.Repos(ctx) {
			allRepos = append(allRepos, localRepositoryResolver{
				name: string(r.Name),
				path: r.Metadata.(*extsvc.LocalGitMetadata).AbsRepoPath,
			})
		}
	}

	return allRepos, nil
}

func globalEmbeddingsPolicyExists(ctx context.Context) (bool, error) {
	const queryPayload = `{
		"operationName": "CodeIntelligenceConfigurationPolicies",
		"variables": {
			"repository": null,
			"query": "",
			"forDataRetention": null,
			"forIndexing": null,
			"forEmbeddings": true,
			"first": 20,
			"after": null,
			"protected": null
		},
		"query": "query CodeIntelligenceConfigurationPolicies($repository: ID, $query: String, $forDataRetention: Boolean, $forIndexing: Boolean, $forEmbeddings: Boolean, $first: Int, $after: String, $protected: Boolean) {codeIntelligenceConfigurationPolicies(repository: $repository query: $query forDataRetention: $forDataRetention forIndexing: $forIndexing forEmbeddings: $forEmbeddings first: $first after: $after protected: $protected) { totalCount }}"
	}`

	url, err := gqlURL("CodeIntelligenceConfigurationPolicies")
	if err != nil {
		return false, err
	}
	cli := httpcli.InternalDoer
	payload := strings.NewReader(queryPayload)

	// Send GraphQL request to sourcegraph.com to check if email is verified
	req, err := http.NewRequestWithContext(ctx, "POST", url, payload)
	if err != nil {
		return false, err
	}

	resp, err := cli.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, errors.Newf("request failed with status: %n", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, errors.Wrap(err, "ReadBody")
	}

	var v struct {
		Data struct {
			CodeIntelligenceConfigurationPolicies struct{ TotalCount int }
		}
		Errors []any
	}

	if err := json.Unmarshal(respBody, &v); err != nil {
		return false, errors.Wrap(err, "Decode")
	}

	if len(v.Errors) > 0 {
		return false, errors.Errorf("graphql: errors: %v", v.Errors)
	}
	return v.Data.CodeIntelligenceConfigurationPolicies.TotalCount > 0, nil
}

func createGlobalEmbeddingsPolicy(ctx context.Context) error {
	alreadyExists, _ := globalEmbeddingsPolicyExists(ctx)
	// ignoring error creating multiple policies is not problematic
	if alreadyExists {
		return nil
	}

	const globalEmbeddingsPolicyPayload = `{
		"operationName": "CreateCodeIntelligenceConfigurationPolicy",
		"variables": {
		  "name": "Global",
		  "repositoryPatterns": null,
		  "type": "GIT_COMMIT",
		  "pattern": "HEAD",
		  "retentionEnabled": false,
		  "retentionDurationHours": null,
		  "retainIntermediateCommits": false,
		  "indexingEnabled": false,
		  "indexCommitMaxAgeHours": null,
		  "indexIntermediateCommits": false,
		  "embeddingsEnabled": true
		},
		"query": "mutation CreateCodeIntelligenceConfigurationPolicy($repositoryId: ID, $repositoryPatterns: [String!], $name: String!, $type: GitObjectType!, $pattern: String!, $retentionEnabled: Boolean!, $retentionDurationHours: Int, $retainIntermediateCommits: Boolean!, $indexingEnabled: Boolean!, $indexCommitMaxAgeHours: Int, $indexIntermediateCommits: Boolean!, $embeddingsEnabled: Boolean!) {  createCodeIntelligenceConfigurationPolicy(    repository: $repositoryId    repositoryPatterns: $repositoryPatterns    name: $name    type: $type    pattern: $pattern    retentionEnabled: $retentionEnabled    retentionDurationHours: $retentionDurationHours    retainIntermediateCommits: $retainIntermediateCommits    indexingEnabled: $indexingEnabled    indexCommitMaxAgeHours: $indexCommitMaxAgeHours    indexIntermediateCommits: $indexIntermediateCommits    embeddingsEnabled: $embeddingsEnabled  ) {    id    __typename  }}"
	  }`

	url, err := gqlURL("CreateCodeIntelligenceConfigurationPolicy")
	if err != nil {
		return err
	}
	cli := httpcli.InternalDoer
	payload := strings.NewReader(globalEmbeddingsPolicyPayload)

	// Send GraphQL request to sourcegraph.com to check if email is verified
	req, err := http.NewRequestWithContext(ctx, "POST", url, payload)
	if err != nil {
		return err
	}

	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Newf("request failed with status: %n", resp.StatusCode)
	}

	return nil
}

// gqlURL returns the frontend's internal GraphQL API URL, with the given ?queryName parameter
// which is used to keep track of the source and type of GraphQL queries.
func gqlURL(queryName string) (string, error) {
	u, err := url.Parse(internalapi.Client.URL)
	if err != nil {
		return "", err
	}
	u.Path = "/.internal/graphql"
	u.RawQuery = queryName
	return u.String(), nil
}

// Check if git thinks the given path is a proper git checkout
func isGitRepo(path string) bool {
	// Executing git rev-parse in the root of a worktree returns an error if the
	// path is not a git repo.
	c := exec.Command("git", "-C", path, "rev-parse")
	err := c.Run()
	return err == nil
}
