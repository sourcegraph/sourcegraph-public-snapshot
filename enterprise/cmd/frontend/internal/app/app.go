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

	gogithub "github.com/google/go-github/v41/github"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
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
	observationCtx *observation.Context,
	db database.DB,
	codeIntelServices codeintel.Services,
	conf conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
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

	auther, err := github.NewGitHubAppAuthenticator(appID, privateKey)
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
		responseServerError := func(statusCode int, msg string, err error) {
			w.WriteHeader(statusCode)
			_, _ = w.Write([]byte(msg))
			log15.Error(msg, "error", err)
		}

		setupAction := r.URL.Query().Get("setup_action")

		if !isSetupActionValid(setupAction) {
			err := errors.Newf("Invalid setup action %q", setupAction)
			responseServerError(http.StatusBadRequest, err.Error(), err)
			return
		}

		// TODO: Double check that this code still works as intended with the
		// removal org org-based installations on dotcom
		a := actor.FromContext(r.Context())
		if !a.IsAuthenticated() || !envvar.SourcegraphDotComMode() {
			if setupAction == "install" {
				gitHubAppConfig := conf.SiteConfig().GitHubApp
				privateKey, err := base64.StdEncoding.DecodeString(gitHubAppConfig.PrivateKey)
				if err != nil {
					responseServerError(http.StatusBadRequest, "Error while decoding encryption key", err)
					return
				}

				installationID, err := strconv.ParseInt(r.URL.Query().Get("installation_id"), 10, 64)
				if err != nil {
					responseServerError(http.StatusBadRequest, `The "installation_id" is not a valid integer`, err)
					return
				}

				ins, err := client.GetAppInstallation(r.Context(), installationID)
				if err != nil {
					responseServerError(http.StatusInternalServerError, `Failed to get the installation information using the "installation_id"`, err)
					return
				}

				var displayName string
				if ins.Account.Login != nil {
					displayName = fmt.Sprintf("GitHub (%s)", *ins.Account.Login)
				}

				err = createCodeHostConnectionForInstallation(r.Context(), db, installationID, displayName, apiURL.String())
				if err != nil {
					responseServerError(http.StatusInternalServerError, "Failed to create code host connection", err)
					return
				}

				encryptedInstallationID, err := EncryptWithPrivateKey(r.URL.Query().Get("installation_id"), privateKey)
				if err != nil {
					responseServerError(http.StatusBadRequest, "Error while encrypting installation ID", err)
					return
				}

				base64InstallationID := base64.StdEncoding.EncodeToString(encryptedInstallationID)
				http.Redirect(w, r, "/install-github-app-success?installation_id="+url.QueryEscape(base64InstallationID), http.StatusFound)
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
	})
}

func createCodeHostConnectionForInstallation(ctx context.Context, db database.DB, installationID int64, displayName, apiURL string) error {
	if displayName == "" {
		displayName = "GitHub App"
	}

	config := extsvc.NewUnencryptedConfig(fmt.Sprintf(`{"url": "%s", "repos": []} `, apiURL))
	svc := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: displayName,
		Config:      config,
	}

	currentConfig, err := svc.Config.Decrypt(context.Background())
	if err != nil {
		return err
	}
	newConfig, err := jsonc.Edit(currentConfig, strconv.FormatInt(installationID, 10), "githubAppInstallationID")
	if err != nil {
		return err
	}
	newConfig, err = jsonc.Edit(newConfig, false, "pending")
	if err != nil {
		return err
	}
	svc.Config = extsvc.NewUnencryptedConfig(newConfig)

	return db.ExternalServices().Upsert(ctx, svc)
}
