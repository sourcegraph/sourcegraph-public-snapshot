package userpasswd

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/apptoken"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/session"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const appUsername = "admin"

// appSecret stores the in-memory secret used by Cody App to enable passworldless
// login from the console.
var appSecret secret

// secret is a base64 URL encoded string
type secret struct {
	mu    sync.Mutex
	value string
}

// Value returns the current secret value, or generates one if it has not yet
// been generated. An error can be returned if generation fails.
func (n *secret) Value() (string, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.value != "" {
		return n.value, nil
	}

	value, err := randBase64(32)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate secret from crypto/rand")
	}
	n.value = value

	return n.value, nil
}

// Verify returns true if clientSecret matches the current secret value.
func (n *secret) Verify(clientSecret string) bool {
	// We hold the lock the entire verify period to ensure we do not have
	// any replay attacks.
	n.mu.Lock()
	defer n.mu.Unlock()

	// The secret was never generated.
	if n.value == "" {
		return false
	}

	if subtle.ConstantTimeCompare([]byte(n.value), []byte(clientSecret)) != 1 {
		return false
	}
	return true // success
}

// AppSignInMiddleware will intercept any request containing a secret query
// parameter. If it is the correct secret it will sign in and redirect to
// search. Otherwise it will call the wrapped handler.
func AppSignInMiddleware(db database.DB, handler func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) error {
	// This handler should only be used in App. Extra precaution to enforce
	// that here.
	if !deploy.IsApp() {
		return handler
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		secret := r.URL.Query().Get("s")
		if secret == "" {
			return handler(w, r)
		}

		if !appSecret.Verify(secret) && !env.InsecureDev {
			return errors.New("Authentication failed")
		}

		// Admin should always be UID=0, but just in case we query it.
		user, err := getByEmailOrUsername(r.Context(), db, appUsername)
		if err != nil {
			return errors.Wrap(err, "Failed to find admin account")
		}

		if _, err := session.SetActorFromUser(r.Context(), w, r, user, 0); err != nil {
			return errors.Wrap(err, "Could not create new user session")
		}

		err = apptoken.CreateAppTokenFileIfNotExists(r.Context(), db, user.ID)
		if err != nil {
			fmt.Println("Error creating app token file", errors.Wrap(err, "Could not create app token file"))
		}

		// Success. Redirect to search or to "redirect" param if present.
		redirect := r.URL.Query().Get("redirect")
		u := r.URL
		if redirect != "" {
			redirectUrl, err := url.Parse(redirect)
			if err == nil {
				u.Path = redirectUrl.Path
				u.RawQuery = redirectUrl.RawQuery
			}
		} else {
			u.RawQuery = ""
			u.Path = "/search"
		}
		http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
		return nil
	}
}

// AppSiteInit is called in the case of Cody App to create the initial site admin account.
//
// Returns a sign-in URL which will automatically sign in the user. This URL
// can only be used once.
//
// Returns a nil error if the admin account already exists, or if it was created.
func AppSiteInit(ctx context.Context, logger log.Logger, db database.DB) (string, error) {
	password, err := generatePassword()
	if err != nil {
		return "", errors.Wrap(err, "failed to generate site admin password")
	}

	failIfNewUserIsNotInitialSiteAdmin := true
	err, _, _ = unsafeSignUp(ctx, logger, db, credentials{
		Email:    "app@sourcegraph.com",
		Username: appUsername,
		Password: password,
	}, failIfNewUserIsNotInitialSiteAdmin)
	if err != nil {
		return "", errors.Wrap(err, "failed to create site admin account")
	}

	// We have an account, return a sign in URL.
	return appSignInURL(), nil
}

func generatePassword() (string, error) {
	pw, err := randBase64(64)
	if err != nil {
		return "", err
	}
	if len(pw) > 72 {
		return pw[:72], nil
	}
	return pw, nil
}

func appSignInURL() string {
	externalURL := globals.ExternalURL().String()
	u, err := url.Parse(externalURL)
	if err != nil {
		return externalURL
	}
	secret, err := appSecret.Value()
	if err != nil {
		return externalURL
	}
	u.Path = "/sign-in"
	query := u.Query()
	query.Set("s", secret)
	u.RawQuery = query.Encode()
	return u.String()
}

func randBase64(dataLen int) (string, error) {
	data := make([]byte, dataLen)
	_, err := rand.Read(data)
	if err != nil {
		return "", err
	}
	// Our secret ends up in URLs, so use URLEncoding.
	return base64.URLEncoding.EncodeToString(data), nil
}
