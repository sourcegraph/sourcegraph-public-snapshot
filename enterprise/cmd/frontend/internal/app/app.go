package app

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/cockroachdb/errors"
	gogithub "github.com/google/go-github/v41/github"
	"github.com/graph-gophers/graphql-go"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// Init initializes the app endpoints.
func Init(
	db database.DB,
	conf conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	if !envvar.SourcegraphDotComMode() {
		enterpriseServices.NewGitHubAppCloudSetupHandler = func() http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("Sourcegraph Cloud GitHub App setup is only available on sourcegraph.com"))
			})
		}
		return nil
	}

	appConfig := conf.SiteConfig().Dotcom.GithubAppCloud
	if appConfig.AppID == "" {
		enterpriseServices.NewGitHubAppCloudSetupHandler = func() http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("Sourcegraph Cloud GitHub App setup is not enabled"))
			})
		}
		return nil
	}

	privateKey, err := base64.StdEncoding.DecodeString(appConfig.PrivateKey)
	if err != nil {
		return errors.Wrap(err, "decode private key")
	}

	auther, err := auth.NewOAuthBearerTokenWithGitHubApp(appConfig.AppID, privateKey)
	if err != nil {
		return errors.Wrap(err, "new authenticator with GitHub App")
	}

	apiURL, err := url.Parse("https://github.com")
	if err != nil {
		return errors.Wrap(err, "parse github.com")
	}
	client := github.NewV3Client(apiURL, auther, nil)

	enterpriseServices.NewGitHubAppCloudSetupHandler = func() http.Handler {
		return newGitHubAppCloudSetupHandler(db, apiURL, client)
	}
	return nil
}

type githubClient interface {
	GetAppInstallation(ctx context.Context, installationID int64) (*gogithub.Installation, error)
}

func newGitHubAppCloudSetupHandler(db database.DB, apiURL *url.URL, client githubClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !envvar.SourcegraphDotComMode() {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("Sourcegraph Cloud GitHub App setup is only available on sourcegraph.com"))
			return
		}

		state := r.URL.Query().Get("state")
		orgID, err := graphqlbackend.UnmarshalOrgID(graphql.ID(state))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`The "state" is not a valid graphql.ID of an organization`))
			return
		}

		installationID, err := strconv.ParseInt(r.URL.Query().Get("installation_id"), 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`The "installation_id" is not a valid integer`))
			return
		}

		err = backend.CheckOrgAccess(r.Context(), db, orgID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("the authenticated user does not belong to the organization requested"))
			return
		}

		responseServerError := func(msg string, err error) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(msg))
			log15.Error(msg, "error", err)
		}

		org, err := db.Orgs().GetByID(r.Context(), orgID)
		if err != nil {
			responseServerError("Failed to get organization", err)
			return
		}

		externalServices := db.ExternalServices()
		svcs, err := externalServices.List(r.Context(),
			database.ExternalServicesListOptions{
				NamespaceOrgID: org.ID,
				Kinds:          []string{extsvc.KindGitHub},
			},
		)
		if err != nil {
			responseServerError("Failed to list organization code host connections", err)
			return
		}

		ins, err := client.GetAppInstallation(r.Context(), installationID)
		if err != nil {
			responseServerError(`Failed to get the installation information using the "installation_id"`, err)
			return
		}

		displayName := "GitHub"
		if ins.Account.Login != nil {
			displayName = fmt.Sprintf("GitHub (%s)", *ins.Account.Login)
		}
		now := time.Now()

		var svc *types.ExternalService
		if len(svcs) == 0 {
			svc = &types.ExternalService{
				Kind:        extsvc.KindGitHub,
				DisplayName: displayName,
				Config: fmt.Sprintf(`
{
  "url": "%s",
  "githubAppInstallationID": "%d",
  "repos": []
}
`, apiURL.String(), installationID),
				NamespaceOrgID: org.ID,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
		} else if len(svcs) == 1 {
			// We have an existing github service, update it
			svc = svcs[0]
			svc.DisplayName = displayName
			newConfig, err := jsonc.Edit(svc.Config, strconv.FormatInt(installationID, 10), "githubAppInstallationID")
			if err != nil {
				responseServerError("Failed to edit config", err)
				return
			}
			svc.Config = newConfig
			svc.UpdatedAt = now
		} else {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Multiple code host connections of same kind found"))
			return
		}

		err = db.ExternalServices().Upsert(r.Context(), svc)
		if err != nil {
			responseServerError("Failed to upsert code host connection", err)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/organizations/%s/settings/code-hosts", org.Name), http.StatusFound)
	})
}
