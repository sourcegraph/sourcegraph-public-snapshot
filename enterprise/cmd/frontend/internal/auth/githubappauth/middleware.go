package githubapp

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/types"
	authcheck "github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

const cacheTTLSeconds = 60 * 60 // 1 hour

func newMiddleware(ossDB database.DB, authPrefix string, isAPIHandler bool, next http.Handler) http.Handler {
	db := edb.NewEnterpriseDB(ossDB)
	ghAppState := rcache.NewWithTTL("github_app_state", cacheTTLSeconds)
	handler := newServeMux(db, authPrefix, ghAppState)
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

// checkSiteAdmin checks if the current user is a site admin and sets http error if not
func checkSiteAdmin(db edb.EnterpriseDB, w http.ResponseWriter, req *http.Request) error {
	err := authcheck.CheckCurrentUserIsSiteAdmin(req.Context(), db)
	if err == nil {
		return nil
	}
	status := http.StatusForbidden
	if err == authcheck.ErrNotAuthenticated {
		status = http.StatusUnauthorized
	}
	http.Error(w, "Bad request, user must be a site admin", status)
	return err
}

// randomState returns a random sha256 hash that can be used as a state parameter
func randomState(n int) (string, error) {
	data := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil)), nil
}

type GitHubAppResponse struct {
	AppID         int               `json:"id"`
	Slug          string            `json:"slug"`
	Name          string            `json:"name"`
	HtmlURL       string            `json:"html_url"`
	ClientID      string            `json:"client_id"`
	ClientSecret  string            `json:"client_secret"`
	PEM           string            `json:"pem"`
	WebhookSecret string            `json:"webhook_secret"`
	Permissions   map[string]string `json:"permissions"`
	Events        []string          `json:"events"`
}

func newServeMux(db edb.EnterpriseDB, prefix string, cache *rcache.Cache) http.Handler {
	r := mux.NewRouter()

	r.HandleFunc(prefix+"/state", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// 🚨 SECURITY: only site admins can create github apps
		if err := checkSiteAdmin(db, w, req); err != nil {
			return
		}

		s, err := randomState(128)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error when creating redirect URL: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		cache.Set(s, []byte{1})

		_, _ = w.Write([]byte(s))
	})

	r.HandleFunc(prefix+"/redirect", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		query := req.URL.Query()
		state := query.Get("state")
		code := query.Get("code")
		if state == "" || code == "" {
			http.Error(w, "Bad request, code and state query params must be present", http.StatusBadRequest)
			return
		}

		_, ok := cache.Get(state)
		if !ok {
			http.Error(w, "Bad request, state query param does not match", http.StatusBadRequest)
			return
		}
		cache.Delete(state)

		u, err := getAPIUrl(req, code)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error when creating github API url: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		app, err := createGitHubApp(u)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error while converting github app: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		id, err := db.GitHubApps().Create(req.Context(), app)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error while storing github app in DB; %s", err.Error()), http.StatusInternalServerError)
			return
		}

		state, err = randomState(128)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error when creating state param: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		cache.Set(state, []byte(strconv.Itoa(id)))

		redirectURL, err := url.JoinPath(app.AppURL, "installations/new")
		if err != nil {
			// if there is an error, try to redirect to app url, which should show Install button as well
			redirectURL = app.AppURL
		}
		http.Redirect(w, req, redirectURL+fmt.Sprintf("?state=%s", state), http.StatusSeeOther)
	})

	r.HandleFunc(prefix+"/setup", func(w http.ResponseWriter, req *http.Request) {
		// 🚨 SECURITY: only site admins can setup github apps
		if err := checkSiteAdmin(db, w, req); err != nil {
			return
		}

		query := req.URL.Query()
		state := query.Get("state")
		instID := query.Get("installation_id")
		if state == "" || instID == "" {
			http.Error(w, "Bad request, installation_id and state query params must be present", http.StatusBadRequest)
			return
		}
		idBytes, ok := cache.Get(state)
		if !ok {
			http.Error(w, "Bad request, state query param does not match", http.StatusBadRequest)
			return
		}
		cache.Delete(state)

		id, err := strconv.Atoi(string(idBytes))
		if err != nil {
			http.Error(w, "Bad request, cannot parse appID", http.StatusBadRequest)
		}

		installationID, err := strconv.Atoi(instID)
		if err != nil {
			http.Error(w, "Bad request, cannot parse installation_id", http.StatusBadRequest)
			return
		}

		action := query.Get("setup_action")
		if action == "install" {
			ctx := req.Context()
			app, err := db.GitHubApps().GetByID(ctx, id)
			if err != nil {
				http.Error(w, fmt.Sprintf("Unexpected error while fetching github app data: %s", err.Error()), http.StatusInternalServerError)
				return
			}

			err = db.GitHubApps().Install(ctx, id, installationID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Unexpected error while installing github app: %s", err.Error()), http.StatusInternalServerError)
				return
			}

			// TODO: redirect to github app configuration page once it's ready
			http.Redirect(w, req, fmt.Sprintf("/site-admin/external-services/new-gh-app?id=%d&installation_id=%d", app.ID, installationID), http.StatusFound)
			// return
		} else {
			http.Error(w, fmt.Sprintf("Bad request; unsupported setup action: %s", action), http.StatusBadRequest)
			// return
		}
	})

	return r
}

func getAPIUrl(req *http.Request, code string) (string, error) {
	referer, err := url.Parse(req.Referer())
	if err != nil {
		return "", err
	}
	api := referer.Scheme + "://api." + referer.Host
	u, err := url.JoinPath(api, "/app-manifests", code, "conversions")
	if err != nil {
		return "", err
	}
	return u, nil
}

func createGitHubApp(conversionURL string) (*types.GitHubApp, error) {
	r, err := http.NewRequest("POST", conversionURL, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 201 {
		return nil, errors.Newf("expected 201 statusCode, got: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	fmt.Printf("Body: %s", body)
	if err != nil {
		return nil, err
	}

	var response GitHubAppResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	htmlURL, err := url.Parse(response.HtmlURL)
	if err != nil {
		return nil, err
	}

	return &types.GitHubApp{
		AppID:        response.AppID,
		Name:         response.Name,
		Slug:         response.Slug,
		ClientID:     response.ClientID,
		ClientSecret: response.ClientSecret,
		PrivateKey:   response.PEM,
		BaseURL:      htmlURL.Scheme + "://" + htmlURL.Host,
		AppURL:       htmlURL.String(),
		// logo: https://github.com/identicons/app/app/milan-test-app-manifest
	}, nil
}
