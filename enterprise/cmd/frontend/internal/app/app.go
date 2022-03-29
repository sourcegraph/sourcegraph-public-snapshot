package app

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	gogithub "github.com/google/go-github/v41/github"
	"github.com/graph-gophers/graphql-go"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

	dotcomConfig := conf.SiteConfig().Dotcom
	if !repos.IsGitHubAppCloudEnabled(dotcomConfig) {
		enterpriseServices.NewGitHubAppCloudSetupHandler = func() http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("Sourcegraph Cloud GitHub App setup is not enabled"))
			})
		}
		return nil
	}

	privateKey, err := base64.StdEncoding.DecodeString(dotcomConfig.GithubAppCloud.PrivateKey)
	if err != nil {
		return errors.Wrap(err, "decode private key")
	}

	auther, err := auth.NewOAuthBearerTokenWithGitHubApp(dotcomConfig.GithubAppCloud.AppID, privateKey)
	if err != nil {
		return errors.Wrap(err, "new authenticator with GitHub App")
	}

	apiURL, err := url.Parse("https://github.com")
	if err != nil {
		return errors.Wrap(err, "parse github.com")
	}
	client := github.NewV3Client(extsvc.URNGitHubAppCloud, apiURL, auther, nil)

	enterpriseServices.NewGitHubAppCloudSetupHandler = func() http.Handler {
		return newGitHubAppCloudSetupHandler(db, apiURL, client)
	}
	return nil
}

func checkIfOrgCanInstallGitHubApp(ctx context.Context, db database.DB, orgID int32) error {
	enabled, err := db.FeatureFlags().GetOrgFeatureFlag(ctx, orgID, "github-app-cloud")
	if err != nil {
		return err
	} else if !enabled {
		return errors.New("Sourcegraph Cloud GitHub App setup is not enabled for the organization")
	}
	return nil
}

type githubClient interface {
	GetAppInstallation(ctx context.Context, installationID int64) (*gogithub.Installation, error)
}

func isSetupActionValid(setupAction string) bool {
	for _, a := range []string{"install", "request"} {
		if setupAction == a {
			return true
		}
	}

	return false
}

func encryptWithPrivateKey(msg string, privateKey []byte) (string, error) {
	block, _ := pem.Decode(privateKey)

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", errors.Wrap(err, "parse private key")
	}

	hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, &key.PublicKey, []byte(msg), nil)
	if err != nil {
		return "", errors.Wrap(err, "encrypt message")
	}

	return url.QueryEscape(string(ciphertext)), nil
}

func newGitHubAppCloudSetupHandler(db database.DB, apiURL *url.URL, client githubClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !envvar.SourcegraphDotComMode() {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("Sourcegraph Cloud GitHub App setup is only available on sourcegraph.com"))
			return
		}

		setupAction := r.URL.Query().Get("setup_action")

		if !isSetupActionValid(setupAction) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(fmt.Sprintf("Invalid setup action '%s'", setupAction)))
			return
		}

		a := actor.FromContext(r.Context())
		if !a.IsAuthenticated() {
			if setupAction == "install" {
				dotcomConfig := conf.SiteConfig().Dotcom

				privateKey, err := base64.StdEncoding.DecodeString(dotcomConfig.GithubAppCloud.PrivateKey)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte(`Error while encrypting installation ID`))
					return
				}

				installationID := r.URL.Query().Get("installation_id")
				encryptedInstallationID, err := encryptWithPrivateKey(installationID, privateKey)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte(`Error while encrypting installation ID`))
					return
				}
				base64InstallationID := base64.StdEncoding.EncodeToString([]byte(encryptedInstallationID))
				http.Redirect(w, r, "/install-github-app-success?installation_id="+base64InstallationID, http.StatusFound)
				return
			}

			if setupAction == "request" {
				http.Redirect(w, r, "/install-github-app-request", http.StatusFound)
				return
			}
		}

		state := r.URL.Query().Get("state")
		if state == "" && setupAction == "install" {
			http.Redirect(w, r, "/settings", http.StatusFound)
			return
		}

		orgID, err := graphqlbackend.UnmarshalOrgID(graphql.ID(state))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`The "state" is not a valid graphql.ID of an organization`))
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

		err = checkIfOrgCanInstallGitHubApp(r.Context(), db, orgID)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		err = backend.CheckOrgAccess(r.Context(), db, orgID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("the authenticated user does not belong to the organization requested"))
			return
		}

		var svc *types.ExternalService
		displayName := "GitHub"
		now := time.Now()

		if len(svcs) == 0 {
			svc = &types.ExternalService{
				Kind:        extsvc.KindGitHub,
				DisplayName: displayName,
				Config: fmt.Sprintf(`
{
  "url": "%s",
  "repos": []
}
`, apiURL.String()),
				NamespaceOrgID: org.ID,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
		} else if len(svcs) == 1 {
			// We have an existing github service, update it
			svc = svcs[0]
			svc.DisplayName = displayName
		} else {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Multiple code host connections of same kind found"))
			return
		}

		if setupAction == "request" {
			newConfig, err := jsonc.Edit(svc.Config, true, "pending")
			if err != nil {
				responseServerError("Failed to edit config", err)
				return
			}
			svc.Config = newConfig
		} else if setupAction == "install" {
			installationID, err := strconv.ParseInt(r.URL.Query().Get("installation_id"), 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`The "installation_id" is not a valid integer`))
				return
			}

			ins, err := client.GetAppInstallation(r.Context(), installationID)
			if err != nil {
				responseServerError(`Failed to get the installation information using the "installation_id"`, err)
				return
			}

			if ins.Account.Login != nil {
				displayName = fmt.Sprintf("GitHub (%s)", *ins.Account.Login)
			}

			svc.DisplayName = displayName
			newConfig, err := jsonc.Edit(svc.Config, strconv.FormatInt(installationID, 10), "githubAppInstallationID")
			if err != nil {
				responseServerError("Failed to edit config", err)
				return
			}
			newConfig, err = jsonc.Edit(newConfig, false, "pending")
			if err != nil {
				responseServerError("Failed to edit config", err)
			}
			svc.Config = newConfig
			svc.UpdatedAt = now
		}

		err = db.ExternalServices().Upsert(r.Context(), svc)
		if err != nil {
			responseServerError("Failed to upsert code host connection", err)
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
