package githubapp

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	authcheck "github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

const authPrefix = auth.AuthURLPrefix + "/githubapp"

func Middleware(db database.DB) *auth.Middleware {
	return &auth.Middleware{
		API: func(next http.Handler) http.Handler {
			return newMiddleware(db, authPrefix, true, next)
		},
		App: func(next http.Handler) http.Handler {
			return newMiddleware(db, authPrefix, false, next)
		},
	}
}

func newMiddleware(ossDB database.DB, authPrefix string, isAPIHandler bool, next http.Handler) http.Handler {
	db := edb.NewEnterpriseDB(ossDB)
	handler := newServeMux(db, authPrefix)
	traceFamily := "githubapp"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This span should be manually finished before delegating to the next handler or
		// redirecting.
		span, _ := trace.New(r.Context(), traceFamily, "Middleware.Handle")
		span.SetAttributes(attribute.Bool("isAPIHandler", isAPIHandler))
		span.Finish()
		if strings.HasPrefix(r.URL.Path, authPrefix+"/") {
			handler.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func newServeMux(db edb.EnterpriseDB, prefix string) http.Handler {
	r := mux.NewRouter()

	r.HandleFunc(prefix+"/setup/{slug}", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		vars := mux.Vars(req)
		slug, ok := vars["slug"]
		if !ok {
			http.Error(w, "Bad request, slug path param must be present", http.StatusBadRequest)
			return
		}

		installationID, err := strconv.ParseInt(req.URL.Query().Get("installation_id"), 10, 64)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		ctx := req.Context()
		if err := authcheck.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
			status := http.StatusForbidden
			if err == authcheck.ErrNotAuthenticated {
				status = http.StatusUnauthorized
			}
			http.Error(w, "Bad request, user must be a site admin", status)
		}

		baseURL := strings.TrimSuffix(req.Referer(), "/")
		action := req.URL.Query().Get("setup_action")

		if action == "install" {
			app, err := db.GitHubApps().GetBySlug(ctx, slug, baseURL)
			if err != nil {
				http.Error(w, fmt.Sprintf("Unexpected error while fetching github app data: %s", err.Error()), http.StatusInternalServerError)
				return
			}

			// TODO: redirect to github app setup flow
			http.Redirect(w, req, fmt.Sprintf("/site-admin/external-services/new?id=%d&installation_id=%d", app.ID, installationID), http.StatusFound)
			// return
		} else {
			http.Error(w, fmt.Sprintf("Bad request; unsupported setup action: %s", action), http.StatusBadRequest)
			// return
		}
	}))

	return r
}
