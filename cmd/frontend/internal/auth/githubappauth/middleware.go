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
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	authcheck "github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	ghaauth "github.com/sourcegraph/sourcegraph/internal/github_apps/auth"
	ghtypes "github.com/sourcegraph/sourcegraph/internal/github_apps/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
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

func newMiddleware(db database.DB, authPrefix string, isAPIHandler bool, next http.Handler) http.Handler {
	ghAppState := rcache.NewWithTTL("github_app_state", cacheTTLSeconds)
	handler := newServeMux(db, authPrefix, ghAppState)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This span should be manually finished before delegating to the next handler or
		// redirecting.
		span, _ := trace.New(r.Context(), "githubapp")
		span.SetAttributes(attribute.Bool("isAPIHandler", isAPIHandler))
		span.End()
		if strings.HasPrefix(r.URL.Path, authPrefix+"/") {
			handler.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// checkSiteAdmin checks if the current user is a site admin and sets http error if not
func checkSiteAdmin(db database.DB, w http.ResponseWriter, req *http.Request) error {
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

// RandomState returns a random sha256 hash that can be used as a state parameter. It is only
// exported for testing purposes.
func RandomState(n int) (string, error) {
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

type gitHubAppStateDetails struct {
	WebhookUUID string `json:"webhookUUID,omitempty"`
	Domain      string `json:"domain"`
	AppID       int    `json:"app_id,omitempty"`
	BaseURL     string `json:"base_url,omitempty"`
}

func newServeMux(db database.DB, prefix string, cache *rcache.Cache) http.Handler {
	r := mux.NewRouter()

	r.Path(prefix + "/state").Methods("GET").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// 🚨 SECURITY: only site admins can create github apps
		if err := checkSiteAdmin(db, w, req); err != nil {
			http.Error(w, "User must be site admin", http.StatusForbidden)
			return
		}

		s, err := RandomState(128)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error when generating state parameter: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		gqlID := req.URL.Query().Get("id")
		domain := req.URL.Query().Get("domain")
		baseURL := req.URL.Query().Get("baseURL")
		if gqlID == "" {
			// we marshal an empty `gitHubAppStateDetails` struct because we want the values saved in the cache
			// to always conform to the same structure.
			stateDetails, err := json.Marshal(gitHubAppStateDetails{})
			if err != nil {
				http.Error(w, fmt.Sprintf("Unexpected error when marshalling state: %s", err.Error()), http.StatusInternalServerError)
				return
			}
			cache.Set(s, stateDetails)

			_, _ = w.Write([]byte(s))
			return
		}

		id64, err := UnmarshalGitHubAppID(graphql.ID(gqlID))
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error while unmarshalling App ID: %s", err.Error()), http.StatusBadRequest)
			return
		}
		stateDetails, err := json.Marshal(gitHubAppStateDetails{
			AppID:   int(id64),
			Domain:  domain,
			BaseURL: baseURL,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error when marshalling state: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		cache.Set(s, stateDetails)

		_, _ = w.Write([]byte(s))
	})

	r.Path(prefix + "/new-app-state").Methods("GET").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// 🚨 SECURITY: only site admins can create github apps
		if err := checkSiteAdmin(db, w, req); err != nil {
			http.Error(w, "User must be site admin", http.StatusForbidden)
			return
		}

		webhookURN := req.URL.Query().Get("webhookURN")
		appName := req.URL.Query().Get("appName")
		domain := req.URL.Query().Get("domain")
		baseURL := req.URL.Query().Get("baseURL")
		var webhookUUID string
		if webhookURN != "" {
			ws := backend.NewWebhookService(db, keyring.Default())
			hook, err := ws.CreateWebhook(req.Context(), appName, extsvc.KindGitHub, webhookURN, nil)
			if err != nil {
				http.Error(w, fmt.Sprintf("Unexpected error while setting up webhook endpoint: %s", err.Error()), http.StatusInternalServerError)
				return
			}
			webhookUUID = hook.UUID.String()
		}

		s, err := RandomState(128)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error when generating state parameter: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		stateDetails, err := json.Marshal(gitHubAppStateDetails{
			WebhookUUID: webhookUUID,
			Domain:      domain,
			BaseURL:     baseURL,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error when marshalling state: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		cache.Set(s, stateDetails)

		resp := struct {
			State       string `json:"state"`
			WebhookUUID string `json:"webhookUUID,omitempty"`
		}{
			State:       s,
			WebhookUUID: webhookUUID,
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error while writing response: %s", err.Error()), http.StatusInternalServerError)
		}
	})

	r.Path(prefix + "/redirect").Methods("GET").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// 🚨 SECURITY: only site admins can setup github apps
		if err := checkSiteAdmin(db, w, req); err != nil {
			http.Error(w, "User must be site admin", http.StatusForbidden)
			return
		}

		query := req.URL.Query()
		state := query.Get("state")
		code := query.Get("code")
		if state == "" || code == "" {
			http.Error(w, "Bad request, code and state query params must be present", http.StatusBadRequest)
			return
		}

		// Check that the length of state is the expected length
		if len(state) != 64 {
			http.Error(w, "Bad request, state query param is wrong length", http.StatusBadRequest)
			return
		}

		stateValue, ok := cache.Get(state)
		if !ok {
			http.Error(w, "Bad request, state query param does not match", http.StatusBadRequest)
			return
		}

		var stateDetails gitHubAppStateDetails
		err := json.Unmarshal(stateValue, &stateDetails)
		if err != nil {
			http.Error(w, "Bad request, invalid state", http.StatusBadRequest)
			return
		}
		// Wait until we've validated the type of state before deleting it from the cache.
		cache.Delete(state)

		webhookUUID, err := uuid.Parse(stateDetails.WebhookUUID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Bad request, could not parse webhook UUID: %v", err), http.StatusBadRequest)
			return
		}

		baseURL, err := url.Parse(stateDetails.BaseURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Bad request, could not parse baseURL from state: %v, error: %v", stateDetails.BaseURL, err), http.StatusInternalServerError)
			return
		}

		apiURL, _ := github.APIRoot(baseURL)
		u, err := url.JoinPath(apiURL.String(), "/app-manifests", code, "conversions")
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error when building manifest endpoint URL: %v", err), http.StatusInternalServerError)
			return
		}

		domain, err := parseDomain(&stateDetails.Domain)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to parse domain: %v", err), http.StatusBadRequest)
			return
		}

		cli, err := httpcli.NewExternalClientFactory().Client()
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to create external client: %v", err), http.StatusInternalServerError)
			return
		}

		app, err := createGitHubApp(u, *domain, cli)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error while converting github app: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		id, err := db.GitHubApps().Create(req.Context(), app)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error while storing github app in DB: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		webhookDB := db.Webhooks(keyring.Default().WebhookKey)
		hook, err := webhookDB.GetByUUID(req.Context(), webhookUUID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error while fetching webhook: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		hook.Secret = encryption.NewUnencrypted(app.WebhookSecret)
		hook.Name = app.Name
		if _, err := webhookDB.Update(req.Context(), hook); err != nil {
			http.Error(w, fmt.Sprintf("Error while updating webhook secret: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		state, err = RandomState(128)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error when creating state param: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		newStateDetails, err := json.Marshal(gitHubAppStateDetails{
			Domain: stateDetails.Domain,
			AppID:  id,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("unexpected error when marshalling state: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		cache.Set(state, newStateDetails)

		// The installations page often takes a few seconds to become available after the
		// app is first created, so we sleep for a bit to give it time to load. The exact
		// length of time to sleep was determined empirically.
		time.Sleep(3 * time.Second)
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
			http.Error(w, "User must be site admin", http.StatusForbidden)
			return
		}

		query := req.URL.Query()
		state := query.Get("state")
		instID := query.Get("installation_id")
		if state == "" || instID == "" {
			// If neither state or installation ID is set, we redirect to the GitHub Apps page.
			// This can happen when someone installs the App directly from GitHub, instead of
			// following the link from within Sourcegraph.
			http.Redirect(w, req, "/site-admin/github-apps", http.StatusFound)
			return
		}

		// Check that the length of state is the expected length
		if len(state) != 64 {
			http.Error(w, "Bad request, state query param is wrong length", http.StatusBadRequest)
			return
		}

		setupInfo, ok := cache.Get(state)
		if !ok {
			redirectURL := generateRedirectURL(nil, nil, nil, nil, errors.New("Bad request, state query param does not match"))
			http.Redirect(w, req, redirectURL, http.StatusFound)
			return
		}

		var stateDetails gitHubAppStateDetails
		err := json.Unmarshal(setupInfo, &stateDetails)
		if err != nil {
			redirectURL := generateRedirectURL(nil, nil, nil, nil, errors.New("Bad request, invalid state"))
			http.Redirect(w, req, redirectURL, http.StatusFound)
			return
		}
		// Wait until we've validated the type of state before deleting it from the cache.
		cache.Delete(state)

		installationID, err := strconv.Atoi(instID)
		if err != nil {
			redirectURL := generateRedirectURL(&stateDetails.Domain, nil, &stateDetails.AppID, nil, errors.New("Bad request, could not parse installation ID"))
			http.Redirect(w, req, redirectURL, http.StatusFound)
			return
		}

		action := query.Get("setup_action")
		if action == "install" {
			ctx := req.Context()
			app, err := db.GitHubApps().GetByID(ctx, stateDetails.AppID)
			if err != nil {
				redirectURL := generateRedirectURL(&stateDetails.Domain, &installationID, &stateDetails.AppID, nil, errors.Newf("Unexpected error while fetching GitHub App from DB: %s", err.Error()))
				http.Redirect(w, req, redirectURL, http.StatusFound)
				return
			}

			auther, err := ghaauth.NewGitHubAppAuthenticator(app.AppID, []byte(app.PrivateKey))
			if err != nil {
				redirectURL := generateRedirectURL(&stateDetails.Domain, &installationID, &stateDetails.AppID, nil, errors.Newf("Unexpected error while creating GitHubAppAuthenticator: %s", err.Error()))
				http.Redirect(w, req, redirectURL, http.StatusFound)
				return
			}

			baseURL, err := url.Parse(app.BaseURL)
			if err != nil {
				redirectURL := generateRedirectURL(&stateDetails.Domain, &installationID, &stateDetails.AppID, nil, errors.Newf("Unexpected error while parsing App base URL: %s", err.Error()))
				http.Redirect(w, req, redirectURL, http.StatusFound)
				return
			}

			apiURL, _ := github.APIRoot(baseURL)

			logger := log.NoOp()
			client, err := github.NewV3Client(logger, "", apiURL, auther, nil)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// The installation often takes a few seconds to become available after the
			// app is first installed, so we sleep for a bit to give it time to load. The exact
			// length of time to sleep was determined empirically.
			time.Sleep(3 * time.Second)

			remoteInstall, err := client.GetAppInstallation(ctx, int64(installationID))
			if err != nil {
				redirectURL := generateRedirectURL(&stateDetails.Domain, &installationID, &stateDetails.AppID, nil, errors.Newf("Unexpected error while fetching App installation details from GitHub: %s", err.Error()))
				http.Redirect(w, req, redirectURL, http.StatusFound)
				return
			}

			_, err = db.GitHubApps().Install(ctx, ghtypes.GitHubAppInstallation{
				InstallationID:   installationID,
				AppID:            app.ID,
				URL:              remoteInstall.GetHTMLURL(),
				AccountLogin:     remoteInstall.Account.GetLogin(),
				AccountAvatarURL: remoteInstall.Account.GetAvatarURL(),
				AccountURL:       remoteInstall.Account.GetHTMLURL(),
				AccountType:      remoteInstall.Account.GetType(),
			})
			if err != nil {
				redirectURL := generateRedirectURL(&stateDetails.Domain, &installationID, &stateDetails.AppID, &app.Name, errors.Newf("Unexpected error while creating GitHub App installation: %s", err.Error()))
				http.Redirect(w, req, redirectURL, http.StatusFound)
				return
			}

			redirectURL := generateRedirectURL(&stateDetails.Domain, &installationID, &app.ID, &app.Name, nil)
			http.Redirect(w, req, redirectURL, http.StatusFound)
			return
		} else {
			http.Error(w, fmt.Sprintf("Bad request; unsupported setup action: %s", action), http.StatusBadRequest)
			return
		}
	})

	return r
}

func generateRedirectURL(domain *string, installationID, appID *int, appName *string, creationErr error) string {
	// If we got an error but didn't even get far enough to parse a domain for the new
	// GitHub App, we still want to route the user back to somewhere useful, so we send
	// them back to the main site admin GitHub Apps page.
	if domain == nil && creationErr != nil {
		return fmt.Sprintf("/site-admin/github-apps?success=false&error=%s", url.QueryEscape(creationErr.Error()))
	}

	parsedDomain, err := parseDomain(domain)
	if err != nil {
		return fmt.Sprintf("/site-admin/github-apps?success=false&error=%s", url.QueryEscape(fmt.Sprintf("invalid domain: %s", *domain)))
	}

	switch *parsedDomain {
	case types.ReposGitHubAppDomain:
		if creationErr != nil {
			return fmt.Sprintf("/site-admin/github-apps?success=false&error=%s", url.QueryEscape(creationErr.Error()))
		}
		if installationID == nil || appID == nil {
			return fmt.Sprintf("/site-admin/github-apps?success=false&error=%s", url.QueryEscape("missing installation ID or app ID"))
		}

		return fmt.Sprintf("/site-admin/github-apps/%s?installation_id=%d", MarshalGitHubAppID(int64(*appID)), *installationID)
	case types.BatchesGitHubAppDomain:
		if creationErr != nil {
			return fmt.Sprintf("/site-admin/batch-changes?success=false&error=%s", url.QueryEscape(creationErr.Error()))
		}

		// This shouldn't really happen unless we also had an error, but we handle it just
		// in case
		if appName == nil {
			return "/site-admin/batch-changes?success=true"
		}
		return fmt.Sprintf("/site-admin/batch-changes?success=true&app_name=%s", *appName)
	default:
		return fmt.Sprintf("/site-admin/github-apps?success=false&error=%s", url.QueryEscape(fmt.Sprintf("unsupported github apps domain: %v", parsedDomain)))
	}
}

var MockCreateGitHubApp func(conversionURL string, domain types.GitHubAppDomain) (*ghtypes.GitHubApp, error)

func createGitHubApp(conversionURL string, domain types.GitHubAppDomain, httpClient *http.Client) (*ghtypes.GitHubApp, error) {
	if MockCreateGitHubApp != nil {
		return MockCreateGitHubApp(conversionURL, domain)
	}
	r, err := http.NewRequest(http.MethodPost, conversionURL, http.NoBody)
	if err != nil {
		return nil, err
	}

	cf := httpcli.NewExternalClientFactory()
	client, err := cf.Doer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GitHub client")
	}

	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, errors.Newf("expected 201 statusCode, got: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	var response GitHubAppResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	htmlURL, err := url.Parse(response.HtmlURL)
	if err != nil {
		return nil, err
	}

	return &ghtypes.GitHubApp{
		AppID:         response.AppID,
		Name:          response.Name,
		Slug:          response.Slug,
		ClientID:      response.ClientID,
		ClientSecret:  response.ClientSecret,
		WebhookSecret: response.WebhookSecret,
		PrivateKey:    response.PEM,
		BaseURL:       htmlURL.Scheme + "://" + htmlURL.Host,
		AppURL:        htmlURL.String(),
		Domain:        domain,
		Logo:          fmt.Sprintf("%s://%s/identicons/app/app/%s", htmlURL.Scheme, htmlURL.Host, response.Slug),
	}, nil
}
