package userpasswd

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"sync"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	sgactor "github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func appRegenerateNonce() {
	appSignInNonceMu.Lock()
	defer appSignInNonceMu.Unlock()

	data := make([]byte, 32)
	_, err := rand.Read(data)
	if err != nil {
		appSignInNonce = ""
		return
	}
	appSignInNonce = base64.StdEncoding.EncodeToString(data)
}

// Returns the current nonce. If it returns an empty string, generating the nonce failed.
func AppReadCurrentNonce() string {
	appSignInNonceMu.Lock()
	defer appSignInNonceMu.Unlock()
	if appSignInNonce == "" {
		appSignInNonceMu.Unlock()
		appRegenerateNonce() // try to generate it
		appSignInNonceMu.Lock()
	}
	return appSignInNonce
}

var (
	appSignInNonceMu sync.Mutex
	appSignInNonce   string
)

func AppSignIn(w http.ResponseWriter, r *http.Request, db database.DB, userProvidedNonce string) error {
	currentNonce := AppReadCurrentNonce()
	if currentNonce == "" {
		return errors.New("failed to generate nonce")
	}

	if subtle.ConstantTimeCompare([]byte(appSignInNonce), []byte(userProvidedNonce)) != 1 {
		return errors.New("Authentication failed")
	}

	// Admin should always be UID=0, but just in case we query it.
	user, err := getByEmailOrUsername(r.Context(), db, "admin")
	if err != nil {
		return errors.Wrap(err, "Failed to find admin account")
	}

	// Write the session cookie
	actor := sgactor.Actor{
		UID: user.ID,
	}
	if err := session.SetActor(w, r, &actor, 0, user.CreatedAt); err != nil {
		return errors.Wrap(err, "Could not create new user session")
	}

	// Sign in was successful, so regenerate the nonce.
	appRegenerateNonce()
	return nil
}

// AppSiteInit is called in the case of Sourcegraph App to create the initial site admin account.
//
// Returns a nil error if the admin account already exists, or if it was created.
func AppSiteInit(ctx context.Context, logger log.Logger, db database.DB, email, username, password string) error {
	failIfNewUserIsNotInitialSiteAdmin := true
	err, _, _ := unsafeSignUp(ctx, logger, db, credentials{
		Email:    email,
		Username: username,
		Password: password,
	}, failIfNewUserIsNotInitialSiteAdmin)
	return err
}
