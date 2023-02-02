package enterprise

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// Services is a bag of HTTP handlers and factory functions that are registered by the
// enterprise frontend setup hook.
type Services struct {
	// Batch Changes Services
	BatchesGitHubWebhook            webhooks.Registerer
	BatchesGitLabWebhook            webhooks.RegistererHandler
	BatchesBitbucketServerWebhook   webhooks.RegistererHandler
	BatchesBitbucketCloudWebhook    webhooks.RegistererHandler
	BatchesChangesFileGetHandler    http.Handler
	BatchesChangesFileExistsHandler http.Handler
	BatchesChangesFileUploadHandler http.Handler

	// Repo related webhook handlers, currently only handle `push` events.
	ReposGithubWebhook          webhooks.Registerer
	ReposGitLabWebhook          webhooks.Registerer
	ReposBitbucketServerWebhook webhooks.Registerer
	ReposBitbucketCloudWebhook  webhooks.Registerer

	SCIMHandler http.Handler

	// Handler for exporting code insights data.
	CodeInsightsDataExportHandler http.Handler

	PermissionsGitHubWebhook    webhooks.Registerer
	NewCodeIntelUploadHandler   NewCodeIntelUploadHandler
	RankingService              RankingService
	NewExecutorProxyHandler     NewExecutorProxyHandler
	NewGitHubAppSetupHandler    NewGitHubAppSetupHandler
	NewComputeStreamHandler     NewComputeStreamHandler
	AuthzResolver               graphqlbackend.AuthzResolver
	BatchChangesResolver        graphqlbackend.BatchChangesResolver
	CodeIntelResolver           graphqlbackend.CodeIntelResolver
	InsightsResolver            graphqlbackend.InsightsResolver
	CodeMonitorsResolver        graphqlbackend.CodeMonitorsResolver
	LicenseResolver             graphqlbackend.LicenseResolver
	DotcomResolver              graphqlbackend.DotcomRootResolver
	SearchContextsResolver      graphqlbackend.SearchContextsResolver
	NotebooksResolver           graphqlbackend.NotebooksResolver
	ComputeResolver             graphqlbackend.ComputeResolver
	InsightsAggregationResolver graphqlbackend.InsightsAggregationResolver
	WebhooksResolver            graphqlbackend.WebhooksResolver
}

// NewCodeIntelUploadHandler creates a new handler for the LSIF upload endpoint. The
// resulting handler skips auth checks when the internal flag is true.
type NewCodeIntelUploadHandler func(internal bool) http.Handler

// RankingService is a subset of codeintel.ranking.Service methods we use.
type RankingService interface {
	LastUpdatedAt(ctx context.Context, repoIDs []api.RepoID) (map[api.RepoID]time.Time, error)
	GetRepoRank(ctx context.Context, repoName api.RepoName) (_ []float64, err error)
	GetDocumentRanks(ctx context.Context, repoName api.RepoName) (_ map[string][]float64, err error)
}

// NewExecutorProxyHandler creates a new proxy handler for routes accessible to the
// executor services deployed separately from the k8s cluster. This handler is protected
// via a shared username and password.
type NewExecutorProxyHandler func() http.Handler

// NewGitHubAppSetupHandler creates a new handler for the Sourcegraph
// GitHub App setup URL endpoint (Cloud and on-prem).
type NewGitHubAppSetupHandler func() http.Handler

// NewComputeStreamHandler creates a new handler for the Sourcegraph Compute streaming endpoint.
type NewComputeStreamHandler func() http.Handler

// DefaultServices creates a new Services value that has default implementations for all services.
func DefaultServices() Services {
	return Services{
		ReposGithubWebhook:              &emptyWebhookHandler{name: "github sync webhook"},
		ReposGitLabWebhook:              &emptyWebhookHandler{name: "gitlab sync webhook"},
		ReposBitbucketServerWebhook:     &emptyWebhookHandler{name: "bitbucket server sync webhook"},
		ReposBitbucketCloudWebhook:      &emptyWebhookHandler{name: "bitbucket cloud sync webhook"},
		PermissionsGitHubWebhook:        &emptyWebhookHandler{name: "permissions github webhook"},
		BatchesGitHubWebhook:            &emptyWebhookHandler{name: "batches github webhook"},
		BatchesGitLabWebhook:            &emptyWebhookHandler{name: "batches gitlab webhook"},
		BatchesBitbucketServerWebhook:   &emptyWebhookHandler{name: "batches bitbucket server webhook"},
		BatchesBitbucketCloudWebhook:    &emptyWebhookHandler{name: "batches bitbucket cloud webhook"},
		BatchesChangesFileGetHandler:    makeNotFoundHandler("batches file get handler"),
		BatchesChangesFileExistsHandler: makeNotFoundHandler("batches file exists handler"),
		BatchesChangesFileUploadHandler: makeNotFoundHandler("batches file upload handler"),
		SCIMHandler:                     makeNotFoundHandler("SCIM handler"),
		NewCodeIntelUploadHandler:       func(_ bool) http.Handler { return makeNotFoundHandler("code intel upload") },
		RankingService:                  stubRankingService{},
		NewExecutorProxyHandler:         func() http.Handler { return makeNotFoundHandler("executor proxy") },
		NewGitHubAppSetupHandler:        func() http.Handler { return makeNotFoundHandler("Sourcegraph GitHub App setup") },
		NewComputeStreamHandler:         func() http.Handler { return makeNotFoundHandler("compute streaming endpoint") },
		CodeInsightsDataExportHandler:   makeNotFoundHandler("code insights data export handler"),
	}
}

// makeNotFoundHandler returns an HTTP handler that respond 404 for all requests.
func makeNotFoundHandler(handlerName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(fmt.Sprintf("%s is only available in enterprise", handlerName)))
	})
}

type emptyWebhookHandler struct {
	name string
}

func (e *emptyWebhookHandler) Register(w *webhooks.Router) {}

func (e *emptyWebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	makeNotFoundHandler(e.name)
}

type ErrBatchChangesDisabledDotcom struct{}

func (e ErrBatchChangesDisabledDotcom) Error() string {
	return "batch changes is not available on Sourcegraph.com; use Sourcegraph Cloud or self-hosted instead"
}

type ErrBatchChangesDisabled struct{}

func (e ErrBatchChangesDisabled) Error() string {
	return "batch changes are disabled. Ask a site admin to set 'batchChanges.enabled' in the site configuration to enable the feature."
}

type ErrBatchChangesDisabledForUser struct{}

func (e ErrBatchChangesDisabledForUser) Error() string {
	return "batch changes are disabled for non-site-admin users. Ask a site admin to unset 'batchChanges.restrictToAdmins' in the site configuration to enable the feature for all users."
}

// BatchChangesEnabledForSite checks if Batch Changes are enabled at the site-level and returns `nil` if they are, or
// else an error indicating why they're disabled
func BatchChangesEnabledForSite() error {
	if !conf.BatchChangesEnabled() {
		return ErrBatchChangesDisabled{}
	}

	// Batch Changes are disabled on sourcegraph.com
	if envvar.SourcegraphDotComMode() {
		return ErrBatchChangesDisabledDotcom{}
	}

	return nil
}

// BatchChangesEnabledForUser checks if Batch Changes are enabled for the current user and returns `nil` if they are,
// or else an error indicating why they're disabled
func BatchChangesEnabledForUser(ctx context.Context, db database.DB) error {
	if err := BatchChangesEnabledForSite(); err != nil {
		return err
	}

	if conf.BatchChangesRestrictedToAdmins() && auth.CheckCurrentUserIsSiteAdmin(ctx, db) != nil {
		return ErrBatchChangesDisabledForUser{}
	}
	return nil
}

// IsCodeInsightsEnabled tells if code insights are enabled or not.
func IsCodeInsightsEnabled() bool {
	if envvar.SourcegraphDotComMode() {
		return false
	}
	if v, _ := strconv.ParseBool(os.Getenv("DISABLE_CODE_INSIGHTS")); v {
		// Code insights can always be disabled. This can be a helpful escape hatch if e.g. there
		// are issues with (or connecting to) the codeinsights-db deployment and it is preventing
		// the Sourcegraph frontend or repo-updater from starting.
		//
		// It is also useful in dev environments if you do not wish to spend resources running Code
		// Insights.
		return false
	}
	if deploy.IsDeployTypeSingleDockerContainer(deploy.Type()) {
		// Code insights is not supported in single-container Docker demo deployments unless
		// explicity allowed, (for example by backend integration tests.)
		if v, _ := strconv.ParseBool(os.Getenv("ALLOW_SINGLE_DOCKER_CODE_INSIGHTS")); v {
			return true
		}
		return false
	}
	return true
}

type stubRankingService struct{}

func (s stubRankingService) LastUpdatedAt(ctx context.Context, repoIDs []api.RepoID) (map[api.RepoID]time.Time, error) {
	return nil, nil
}

func (s stubRankingService) GetRepoRank(ctx context.Context, repoName api.RepoName) (_ []float64, err error) {
	return nil, nil
}

func (s stubRankingService) GetDocumentRanks(ctx context.Context, repoName api.RepoName) (_ map[string][]float64, err error) {
	return nil, nil
}
