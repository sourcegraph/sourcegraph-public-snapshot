package logging

import (
	"fmt"

	"github.com/OpenPeeDeeP/depguard/v2"
	"golang.org/x/tools/go/analysis"

	"github.com/sourcegraph/sourcegraph/dev/linters/nolint"
)

// This analyzer is modeled after the one in
// https://sourcegraph.sourcegraph.com/github.com/sourcegraph/sourcegraph@f6ae87add606c65876b87d378929fcb80c3bb493/-/blob/dev/linters/depguard/depguard.go
// These could potentially be combined into one analyzer.
var Analyzer *analysis.Analyzer = createAnalyzer()

const useLogInsteadMessage = `use "github.com/sourcegraph/log" instead`

// Deny is a map which contains all the deprecated logging packages
// The key of the map is the package name that is not allowed - globs can be used as keys.
// The value of the key is the reason to give as to why the logger is not allowed.
var Deny = map[string]string{
	"log$":                              useLogInsteadMessage,
	"github.com/inconshreveable/log15$": useLogInsteadMessage,
	"go.uber.org/zap":                   useLogInsteadMessage,
}

func createAnalyzer() *analysis.Analyzer {
	settings := &depguard.LinterSettings{
		"deprecated loggers": &depguard.List{
			Deny: Deny,
			Files: []string{
				// Let everything in dev use whatever they want
				"!**/dev/**/*.go",
				// We allow one usage of a direct zap import here
				"!**/internal/observation/fields.go",
				// Inits old loggers
				"!**/internal/logging/main.go",
				// Dependencies require direct usage of zap
				"!**/cmd/frontend/internal/app/otlpadapter/register.go",
				// Legacy and special case handling of panics in background routines
				"!**/lib/background/goroutine.go",
				// Legacy loghandlers for log15
				"!**/cmd/frontend/internal/cli/loghandlers/loghandlers.go",
				// goreman does not have a consistent context to easily setup logging
				"!**/cmd/server/internal/goreman/proc.go",
				// These are temporary
				"!**/internal/honey/honey.go",
				"!**/cmd/server/shared/conf_parse.go",
				"!**/cmd/server/shared/copy.go",
				"!**/cmd/server/shared/shared.go",
				"!**/internal/rcache/rcache.go",
				"!**/internal/database/dbconn/open.go",
				"!**/internal/profiler/profiler.go",
				"!**/internal/auth/providers/providers.go",
				"!**/internal/extsvc/gitolite/repos.go",
				"!**/cmd/frontend/hubspot/hubspotutil/hubspotutil.go",
				"!**/cmd/worker/internal/executorqueue/aws_reporter.go",
				"!**/cmd/worker/internal/executorqueue/gcp_reporter.go",
				"!**/cmd/worker/internal/executorqueue/queue_allocation.go",
				"!**/internal/conf/computed.go",
				"!**/internal/conf/conf.go",
				"!**/internal/conf/service_watcher.go",
				"!**/internal/diskcache/cache.go",
				"!**/internal/httpserver/server.go",
				"!**/internal/uploadstore/gcs_client.go",
				"!**/cmd/frontend/globals/globals.go",
				"!**/cmd/symbols/internal/database/janitor/cache_evicter.go",
				"!**/internal/batches/types/scheduler/config/config.go",
				"!**/internal/cmd/ghe-feeder/progress.go",
				"!**/internal/cmd/ghe-feeder/pump.go",
				"!**/internal/cmd/ghe-feeder/worker.go",
				"!**/internal/conf/reposource/custom.go",
				"!**/internal/extsvc/bitbucketserver/client.go",
				"!**/internal/licensing/licensing.go",
				"!**/internal/search/backend/index_options.go",
				"!**/cmd/frontend/internal/handlerutil/handler.go",
				"!**/cmd/frontend/internal/licensing/enforcement/users.go",
				"!**/cmd/frontend/webhooks/github_webhooks.go",
				"!**/cmd/frontend/webhooks/middleware.go",
				"!**/cmd/searcher/internal/search/zipcache.go",
				"!**/cmd/symbols/parser/parser.go",
				"!**/cmd/symbols/squirrel/http_handlers.go",
				"!**/cmd/worker/internal/webhooks/handler.go",
				"!**/internal/batches/types/changeset.go",
				"!**/internal/batches/types/changeset_event.go",
				"!**/internal/rockskip/search.go",
				"!**/internal/rockskip/server.go",
				"!**/internal/rockskip/status.go",
				"!**/internal/search/searchcontexts/search_contexts.go",
				"!**/internal/session/session.go",
				"!**/internal/siteid/siteid.go",
				"!**/cmd/frontend/internal/handlerutil/handler.go",
				"!**/cmd/frontend/internal/app/errorutil/handlers.go",
				"!**/cmd/symbols/internal/database/store/store.go",
				"!**/internal/batches/scheduler/scheduler.go",
				"!**/internal/batches/scheduler/ticker.go",
				"!**/internal/batches/sources/bitbucketserver.go",
				"!**/internal/batches/state/changeset_history.go",
				"!**/internal/usagestats/event_handlers.go",
				"!**/cmd/frontend/graphqlbackend/git_tree_entry.go",
				"!**/cmd/frontend/graphqlbackend/repository_comparison.go",
				"!**/cmd/frontend/graphqlbackend/search_results.go",
				"!**/cmd/frontend/graphqlbackend/site_admin.go",
				"!**/cmd/frontend/graphqlbackend/site_alerts.go",
				"!**/cmd/frontend/graphqlbackend/site_reload.go",
				"!**/cmd/frontend/graphqlbackend/survey_response.go",
				"!**/cmd/frontend/graphqlbackend/user_usage_stats.go",
				"!**/cmd/frontend/internal/app/debugproxies/scanner.go",
				"!**/cmd/frontend/internal/auth/httpheader/middleware.go",
				"!**/cmd/frontend/internal/auth/oauth/middleware.go",
				"!**/cmd/frontend/internal/auth/oauth/provider.go",
				"!**/cmd/frontend/internal/auth/openidconnect/middleware.go",
				"!**/cmd/frontend/internal/auth/saml/config.go",
				"!**/cmd/frontend/internal/auth/saml/config.go",
				"!**/cmd/frontend/internal/auth/saml/middleware.go",
				"!**/cmd/frontend/internal/batches/webhooks/webhooks.go",
				"!**/cmd/frontend/internal/bg/check_redis_cache_eviction_policy.go",
				"!**/cmd/frontend/internal/bg/delete_old_cache_data_in_redis.go",
				"!**/internal/batches/reconciler/executor.go",
				"!**/cmd/frontend/internal/app/ui/landing.go",
				"!**/cmd/frontend/internal/auth/githuboauth/provider.go",
				"!**/cmd/frontend/internal/auth/githuboauth/session.go",
				"!**/cmd/frontend/internal/compute/resolvers/resolvers.go",
				"!**/cmd/frontend/internal/httpapi/httpapi.go",
				"!**/cmd/frontend/internal/app/debug.go",
				"!**/cmd/frontend/internal/app/opensearch.go",
				"!**/cmd/frontend/internal/app/ping.go",
				"!**/cmd/frontend/internal/app/usage_stats.go",
				"!**/cmd/frontend/internal/cli/middleware/goimportpath.go",
				"!**/cmd/frontend/internal/cli/middleware/trace.go",
				"!**/cmd/frontend/internal/cli/serve_cmd.go",
				"!**/cmd/frontend/internal/cli/sysreq.go",
			},
		},
	}
	analyzer, err := depguard.NewAnalyzer(settings)
	if err != nil {
		panic(fmt.Sprintf("failed to create deprecated logging analyzer: %v", err))
	}

	return nolint.Wrap(analyzer)
}
