package userpasswd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hubspot"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hubspot/hubspotutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/suspiciousnames"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/cookie"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/deviceid"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type credentials struct {
	Email           string `json:"email"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	AnonymousUserID string `json:"anonymousUserId"`
	FirstSourceURL  string `json:"firstSourceUrl"`
	LastSourceURL   string `json:"lastSourceUrl"`
}

type unlockAccountInfo struct {
	Token string `json:"token"`
}

// HandleSignUp handles submission of the user signup form.
func HandleSignUp(logger log.Logger, db database.DB) http.HandlerFunc {
	logger = logger.Scoped("HandleSignUp", "sign up request handler")
	return func(w http.ResponseWriter, r *http.Request) {
		if handleEnabledCheck(logger, w) {
			return
		}
		if pc, _ := getProviderConfig(); !pc.AllowSignup {
			http.Error(w, "Signup is not enabled (builtin auth provider allowSignup site configuration option)", http.StatusNotFound)
			return
		}
		handleSignUp(logger, db, w, r, false)
	}
}

// HandleSiteInit handles submission of the site initialization form, where the initial site admin user is created.
func HandleSiteInit(logger log.Logger, db database.DB) http.HandlerFunc {
	logger = logger.Scoped("HandleSiteInit", "initial size initialization request handler")
	return func(w http.ResponseWriter, r *http.Request) {
		// This only succeeds if the site is not yet initialized and there are no users yet. It doesn't
		// allow signups after those conditions become true, so we don't need to check the builtin auth
		// provider's allowSignup in site config.
		handleSignUp(logger, db, w, r, true)
	}
}

// checkEmailAbuse performs abuse prevention checks to prevent email abuse, i.e. users using emails
// of other people whom they want to annoy.
func checkEmailAbuse(ctx context.Context, db database.DB, addr string) (abused bool, reason string, err error) {
	email, err := db.UserEmails().GetLatestVerificationSentEmail(ctx, addr)
	if err != nil {
		if errcode.IsNotFound(err) {
			return false, "", nil
		}
		return false, "", err
	}

	// NOTE: We could check if email is already used here but that complicates the logic
	// and the reused problem should be better handled in the user creation.

	if email.NeedsVerificationCoolDown() {
		return true, "too frequent attempt since last verification email sent", nil
	}

	return false, "", nil
}

// doServeSignUp is called to create a new user account. It is called for the normal user signup process (where a
// non-admin user is created) and for the site initialization process (where the initial site admin user account is
// created).
//
// 🚨 SECURITY: Any change to this function could introduce security exploits
// and/or break sign up / initial admin account creation. Be careful.
func handleSignUp(logger log.Logger, db database.DB, w http.ResponseWriter, r *http.Request, failIfNewUserIsNotInitialSiteAdmin bool) {
	if r.Method != "POST" {
		http.Error(w, fmt.Sprintf("unsupported method %s", r.Method), http.StatusBadRequest)
		return
	}
	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "could not decode request body", http.StatusBadRequest)
		return
	}

	const defaultErrorMessage = "Signup failed unexpectedly."

	if err := suspiciousnames.CheckNameAllowedForUserOrOrganization(creds.Username); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := checkEmailFormat(creds.Email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create the user.
	//
	// We don't need to check the builtin auth provider's allowSignup because we assume the caller
	// of doServeSignUp checks it, or else that failIfNewUserIsNotInitialSiteAdmin == true (in which
	// case the only signup allowed is that of the initial site admin).
	newUserData := database.NewUser{
		Email:                 creds.Email,
		Username:              creds.Username,
		Password:              creds.Password,
		FailIfNotInitialUser:  failIfNewUserIsNotInitialSiteAdmin,
		EnforcePasswordLength: true,
		TosAccepted:           true, // Users created via the signup form are considered to have accepted the Terms of Service.
	}
	if failIfNewUserIsNotInitialSiteAdmin {
		// The email of the initial site admin is considered to be verified.
		newUserData.EmailIsVerified = true
	} else {
		code, err := backend.MakeEmailVerificationCode()
		if err != nil {
			logger.Error("Error generating email verification code for new user.", log.String("email", creds.Email), log.String("username", creds.Username), log.Error(err))
			http.Error(w, defaultErrorMessage, http.StatusInternalServerError)
			return
		}
		newUserData.EmailVerificationCode = code
	}

	// Prevent abuse (users adding emails of other people whom they want to annoy) with the
	// following abuse prevention checks.
	if conf.EmailVerificationRequired() && !newUserData.EmailIsVerified {
		abused, reason, err := checkEmailAbuse(r.Context(), db, creds.Email)
		if err != nil {
			logger.Error("Error checking email abuse", log.String("email", creds.Email), log.Error(err))
			http.Error(w, defaultErrorMessage, http.StatusInternalServerError)
			return
		} else if abused {
			logger.Error("Possible email address abuse prevented", log.String("email", creds.Email), log.String("reason", reason))
			http.Error(w, "Email address is possibly being abused, please try again later or use a different email address.", http.StatusTooManyRequests)
			return
		}
	}

	usr, err := db.Users().Create(r.Context(), newUserData)
	if err != nil {
		var (
			message    string
			statusCode int
		)
		switch {
		case database.IsUsernameExists(err):
			message = "Username is already in use. Try a different username."
			statusCode = http.StatusConflict
		case database.IsEmailExists(err):
			message = "Email address is already in use. Try signing into that account instead, or use a different email address."
			statusCode = http.StatusConflict
		case errcode.PresentationMessage(err) != "":
			message = errcode.PresentationMessage(err)
			statusCode = http.StatusConflict
		default:
			// Do not show non-allowed error messages to user, in case they contain sensitive or confusing
			// information.
			message = defaultErrorMessage
			statusCode = http.StatusInternalServerError
		}
		logger.Error("Error in user signup.", log.String("email", creds.Email), log.String("username", creds.Username), log.Error(err))
		http.Error(w, message, statusCode)

		if err = usagestats.LogBackendEvent(db, actor.FromContext(r.Context()).UID, deviceid.FromContext(r.Context()), "SignUpFailed", nil, nil, featureflag.GetEvaluatedFlagSet(r.Context()), nil); err != nil {
			logger.Warn("Failed to log event SignUpFailed", log.Error(err))
		}

		return
	}

	if err = db.Authz().GrantPendingPermissions(r.Context(), &database.GrantPendingPermissionsArgs{
		UserID: usr.ID,
		Perm:   authz.Read,
		Type:   authz.PermRepos,
	}); err != nil {
		logger.Error("Failed to grant user pending permissions", log.Int32("userID", usr.ID), log.Error(err))
	}

	if conf.EmailVerificationRequired() && !newUserData.EmailIsVerified {
		if err := backend.SendUserEmailVerificationEmail(r.Context(), usr.Username, creds.Email, newUserData.EmailVerificationCode); err != nil {
			logger.Error("failed to send email verification (continuing, user's email will be unverified)", log.String("email", creds.Email), log.Error(err))
		} else if err = db.UserEmails().SetLastVerification(r.Context(), usr.ID, creds.Email, newUserData.EmailVerificationCode); err != nil {
			logger.Error("failed to set email last verification sent at (user's email is verified)", log.String("email", creds.Email), log.Error(err))
		}
	}

	// Write the session cookie
	a := &actor.Actor{UID: usr.ID}
	if err := session.SetActor(w, r, a, 0, usr.CreatedAt); err != nil {
		httpLogError(logger, log.LevelError, w, "Could not create new user session", http.StatusInternalServerError, log.Error(err))
	}

	// Track user data
	if r.UserAgent() != "Sourcegraph e2etest-bot" {
		go hubspotutil.SyncUser(creds.Email, hubspotutil.SignupEventID, &hubspot.ContactProperties{AnonymousUserID: creds.AnonymousUserID, FirstSourceURL: creds.FirstSourceURL, LastSourceURL: creds.LastSourceURL, DatabaseID: usr.ID})
	}

	if err = usagestats.LogBackendEvent(db, actor.FromContext(r.Context()).UID, deviceid.FromContext(r.Context()), "SignUpSucceeded", nil, nil, featureflag.GetEvaluatedFlagSet(r.Context()), nil); err != nil {
		logger.Warn("Failed to log event SignUpSucceeded", log.Error(err))
	}
}

func checkEmailFormat(email string) error {
	// Max email length is 320 chars https://datatracker.ietf.org/doc/html/rfc3696#section-3
	if len(email) > 320 {
		return errors.Newf("maximum email length is 320, got %d", len(email))
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return err
	}
	return nil
}

func getByEmailOrUsername(ctx context.Context, db database.DB, emailOrUsername string) (*types.User, error) {
	if strings.Contains(emailOrUsername, "@") {
		return db.Users().GetByVerifiedEmail(ctx, emailOrUsername)
	}
	return db.Users().GetByUsername(ctx, emailOrUsername)
}

// HandleSignIn accepts a POST containing username-password credentials and
// authenticates the current session if the credentials are valid.
//
// The account will be locked out after consecutive failed attempts in a certain
// period of time.
func HandleSignIn(logger log.Logger, db database.DB, store LockoutStore) http.HandlerFunc {
	logger = logger.Scoped("HandleSignin", "sign in request handler")
	return func(w http.ResponseWriter, r *http.Request) {
		if handleEnabledCheck(logger, w) {
			return
		}

		var user types.User

		signInResult := database.SecurityEventNameSignInAttempted
		logSignInEvent(r, db, &user, &signInResult)

		// We have more failure scenarios and ONLY one successful scenario. By default,
		// assume a SignInFailed state so that the deferred logSignInEvent function call
		// will log the correct security event in case of a failure.
		signInResult = database.SecurityEventNameSignInFailed
		defer func() {
			logSignInEvent(r, db, &user, &signInResult)
			checkAccountLockout(store, &user, &signInResult)
		}()

		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("Unsupported method %s", r.Method), http.StatusBadRequest)
			return
		}
		var creds credentials
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "Could not decode request body", http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		// Validate user. Allow login by both email and username (for convenience).
		u, err := getByEmailOrUsername(ctx, db, creds.Email)
		if err != nil {
			httpLogError(logger, log.LevelWarn, w, "Authentication failed", http.StatusUnauthorized, log.Error(err))
			return
		}
		user = *u

		if reason, locked := store.IsLockedOut(user.ID); locked {
			func() {
				if !conf.CanSendEmail() {
					return
				}

				recipient, _, err := db.UserEmails().GetPrimaryEmail(ctx, user.ID)
				if err != nil {
					logger.Error("Error getting primary email address", log.Int32("userID", user.ID), log.Error(err))
					return
				}

				err = store.SendUnlockAccountEmail(ctx, user.ID, recipient)
				if err != nil {
					logger.Error("Error sending unlock account email", log.Int32("userID", user.ID), log.Error(err))
					return
				}
			}()

			httpLogError(logger, log.LevelError, w, fmt.Sprintf("Account has been locked out due to %q", reason), http.StatusUnprocessableEntity)
			return
		}

		// 🚨 SECURITY: check password
		correct, err := db.Users().IsPassword(ctx, user.ID, creds.Password)
		if err != nil {
			httpLogError(logger, log.LevelError, w, "Error checking password", http.StatusInternalServerError, log.Error(err))
			return
		}
		if !correct {
			httpLogError(logger, log.LevelWarn, w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		// Write the session cookie
		actor := actor.Actor{
			UID: user.ID,
		}
		if err := session.SetActor(w, r, &actor, 0, user.CreatedAt); err != nil {
			httpLogError(logger, log.LevelError, w, "Could not create new user session", http.StatusInternalServerError, log.Error(err))
			return
		}

		signInResult = database.SecurityEventNameSignInSucceeded
	}
}

func HandleUnlockAccount(logger log.Logger, _ database.DB, store LockoutStore) http.HandlerFunc {
	logger = logger.Scoped("HandleUnlockAccount", "unlock account request handler")
	return func(w http.ResponseWriter, r *http.Request) {
		if handleEnabledCheck(logger, w) {
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("Unsupported method %s", r.Method), http.StatusBadRequest)
			return
		}

		var unlockAccountInfo unlockAccountInfo
		if err := json.NewDecoder(r.Body).Decode(&unlockAccountInfo); err != nil {
			http.Error(w, "Could not decode request body", http.StatusBadRequest)
			return
		}

		if unlockAccountInfo.Token == "" {
			http.Error(w, "Bad request: missing token", http.StatusBadRequest)
			return
		}

		valid, error := store.VerifyUnlockAccountTokenAndReset(unlockAccountInfo.Token)

		if !valid || error != nil {
			err := "invalid token provided"
			if error != nil {
				err = error.Error()
			}
			httpLogError(logger, log.LevelWarn, w, err, http.StatusUnauthorized)
			return
		}
	}
}

func logSignInEvent(r *http.Request, db database.DB, user *types.User, name *database.SecurityEventName) {
	var anonymousID string
	event := &database.SecurityEvent{
		Name:            *name,
		URL:             r.URL.Path,
		UserID:          uint32(user.ID),
		AnonymousUserID: anonymousID,
		Source:          "BACKEND",
		Timestamp:       time.Now(),
	}

	// Safe to ignore this error
	event.AnonymousUserID, _ = cookie.AnonymousUID(r)
	_ = usagestats.LogBackendEvent(db, user.ID, deviceid.FromContext(r.Context()), string(*name), nil, nil, featureflag.GetEvaluatedFlagSet(r.Context()), nil)
	db.SecurityEventLogs().LogEvent(r.Context(), event)
}

func checkAccountLockout(store LockoutStore, user *types.User, event *database.SecurityEventName) {
	if user.ID <= 0 {
		return
	}

	if *event == database.SecurityEventNameSignInSucceeded {
		store.Reset(user.ID)
	} else if *event == database.SecurityEventNameSignInFailed {
		store.IncreaseFailedAttempt(user.ID)
	}
}

// HandleCheckUsernameTaken checks availability of username for signup form
func HandleCheckUsernameTaken(logger log.Logger, db database.DB) http.HandlerFunc {
	logger = logger.Scoped("HandleCheckUsernameTaken", "checks for username uniqueness")
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		username, err := auth.NormalizeUsername(vars["username"])

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err = db.Namespaces().GetByName(r.Context(), username)
		if err == database.ErrNamespaceNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			httpLogError(logger, log.LevelError, w, "Error checking username uniqueness", http.StatusInternalServerError, log.Error(err))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func httpLogError(logger log.Logger, level log.Level, w http.ResponseWriter, msg string, code int, errArgs ...log.Field) {
	switch level {
	case log.LevelInfo:
		logger.Info(msg, errArgs...)
	case log.LevelWarn:
		logger.Warn(msg, errArgs...)
	case log.LevelDebug:
		logger.Debug(msg, errArgs...)
	case log.LevelError:
		logger.Error(msg, errArgs...)
	default:
		logger.Error(msg, errArgs...)
	}
	http.Error(w, msg, code)
}
