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
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	extsvcauth "github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Init initializes the app endpoints.
func Init(
	ctx context.Context,
	db database.DB,
	conf conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
	observationContext *observation.Context,
) error {
	var privateKey []byte
	var err error
	var appID string

	gitHubAppConfig := conf.SiteConfig().GitHubApp
	if !repos.IsGitHubAppEnabled(gitHubAppConfig) {
		enterpriseServices.NewGitHubAppSetupHandler = func() http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("Sourcegraph GitHub App setup is not enabled"))
			})
		}
		return nil
	}
	privateKey, err = base64.StdEncoding.DecodeString(gitHubAppConfig.PrivateKey)
	if err != nil {
		return errors.Wrap(err, "decode private key")
	}
	appID = gitHubAppConfig.AppID

	auther, err := extsvcauth.NewOAuthBearerTokenWithGitHubApp(appID, privateKey)
	if err != nil {
		return errors.Wrap(err, "new authenticator with GitHub App")
	}

	apiURL, err := url.Parse("https://github.com")
	if err != nil {
		return errors.Wrap(err, "parse github.com")
	}
	client := github.NewV3Client(log.Scoped("app.github.v3", "github v3 client for frontend app"), extsvc.URNGitHubApp, apiURL, auther, nil)

	enterpriseServices.NewGitHubAppSetupHandler = func() http.Handler {
		return newGitHubAppSetupHandler(db, apiURL, client)
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

// DescryptWithPrivatekey decrypts a message using a provided private key byte string in PEM format.
func DecryptWithPrivateKey(encodedMsg string, privateKey []byte) (string, error) {
	block, _ := pem.Decode(privateKey)

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", errors.Wrap(err, "parse private key")
	}

	hash := sha256.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, key, []byte(encodedMsg), nil)
	if err != nil {
		return "", errors.Wrap(err, "decrypt message")
	}

	return string(plaintext), nil
}

// EncryptWithPrivatekey encrypts a message using a provided private key byte string in PEM format.
// The public key used for encryption is derived from the provided private key.
func EncryptWithPrivateKey(msg string, privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "parse private key")
	}

	hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, &key.PublicKey, []byte(msg), nil)
	if err != nil {
		return nil, errors.Wrap(err, "encrypt message")
	}

	return ciphertext, nil
}

func newGitHubAppSetupHandler(db database.DB, apiURL *url.URL, client githubClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setupAction := r.URL.Query().Get("setup_action")

		if !isSetupActionValid(setupAction) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(fmt.Sprintf("Invalid setup action '%s'", setupAction)))
			return
		}

		responseServerError := func(msg string, err error) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(msg))
			log15.Error(msg, "error", err)
		}

		a := actor.FromContext(r.Context())
		if !a.IsAuthenticated() || !envvar.SourcegraphDotComMode() {
			if setupAction == "install" {
				var privateKey []byte
				var err error

				gitHubAppConfig := conf.SiteConfig().GitHubApp
				privateKey, err = base64.StdEncoding.DecodeString(gitHubAppConfig.PrivateKey)
				if err != nil {
					log15.Error("Error while decoding privatekey.", "error", err)
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte(`Error while decoding encryption key`))
					return
				}

				var svc *types.ExternalService
				displayName := "GitHub App"
				now := time.Now()

				svc = &types.ExternalService{
					Kind:        extsvc.KindGitHub,
					DisplayName: displayName,
					Config: extsvc.NewUnencryptedConfig(fmt.Sprintf(`
{
  "url": "%s",
  "repos": []
}
`, apiURL.String())),
					CreatedAt: now,
					UpdatedAt: now,
				}

				if setupAction == "request" {
					currentConfig, err := svc.Config.Decrypt(context.Background())
					if err != nil {
						responseServerError("Failed to edit config", err)
						return
					}
					newConfig, err := jsonc.Edit(currentConfig, true, "pending")
					if err != nil {
						responseServerError("Failed to edit config", err)
						return
					}
					svc.Config = extsvc.NewUnencryptedConfig(newConfig)
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
					currentConfig, err := svc.Config.Decrypt(context.Background())
					if err != nil {
						responseServerError("Failed to edit config", err)
						return
					}
					newConfig, err := jsonc.Edit(currentConfig, strconv.FormatInt(installationID, 10), "githubAppInstallationID")
					if err != nil {
						responseServerError("Failed to edit config", err)
						return
					}
					newConfig, err = jsonc.Edit(newConfig, false, "pending")
					if err != nil {
						responseServerError("Failed to edit config", err)
					}
					svc.Config = extsvc.NewUnencryptedConfig(newConfig)
					svc.UpdatedAt = now
					err = db.ExternalServices().Upsert(r.Context(), svc)
					if err != nil {
						responseServerError("Failed to upsert code host connection", err)
						return
					}
					encryptedInstallationID, err := EncryptWithPrivateKey(r.URL.Query().Get("installation_id"), privateKey)
					if err != nil {
						log15.Error("Error while encrypting installation ID.", "error", err)
						w.WriteHeader(http.StatusBadRequest)
						_, _ = w.Write([]byte(`Error while encrypting installation ID`))
						return
					}
					base64InstallationID := base64.StdEncoding.EncodeToString(encryptedInstallationID)
					http.Redirect(w, r, "/install-github-app-success?installation_id="+url.QueryEscape(base64InstallationID), http.StatusFound)
					return
				}
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

		err = auth.CheckOrgAccess(r.Context(), db, orgID)
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
				Config: extsvc.NewUnencryptedConfig(fmt.Sprintf(`
{
  "url": "%s",
  "repos": []
}
`, apiURL.String())),
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
			rawConfig, err := svc.Config.Decrypt(r.Context())
			if err != nil {
				responseServerError("Failed to retrieve config", err)
				return
			}

			rawConfig, err = jsonc.Edit(rawConfig, true, "pending")
			if err != nil {
				responseServerError("Failed to edit config", err)
				return
			}
			svc.Config.Set(rawConfig)
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

			rawConfig, err := svc.Config.Decrypt(r.Context())
			if err != nil {
				responseServerError("Failed to retrieve config", err)
				return
			}

			rawConfig, err = jsonc.Edit(rawConfig, strconv.FormatInt(installationID, 10), "githubAppInstallationID")
			if err != nil {
				responseServerError("Failed to edit config", err)
				return
			}
			rawConfig, err = jsonc.Edit(rawConfig, false, "pending")
			if err != nil {
				responseServerError("Failed to edit config", err)
			}
			svc.Config.Set(rawConfig)
			svc.UpdatedAt = now
		}

		err = db.ExternalServices().Upsert(r.Context(), svc)
		if err != nil {
			responseServerError("Failed to upsert code host connection", err)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/organizations/%s/settings/code-hosts", org.Name), http.StatusFound)
	})
}
