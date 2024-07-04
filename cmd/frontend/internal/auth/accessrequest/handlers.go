package accessrequest

import (
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/userpasswd"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/types"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/deviceid"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
)

// HandleRequestAccess handles submission of the request access form.
func HandleRequestAccess(logger log.Logger, db database.DB) http.HandlerFunc {
	logger = logger.Scoped("HandleRequestAccess")
	return func(w http.ResponseWriter, r *http.Request) {
		if !conf.IsAccessRequestEnabled() {
			logger.Error("experimental feature accessRequests is disabled, but received request")
			http.Error(w, "experimental feature accessRequests is disabled, but received request", http.StatusForbidden)
			return
		}
		// Check whether builtin signup is enabled.
		builtInAuthProvider, _ := userpasswd.GetProviderConfig()
		if builtInAuthProvider != nil && builtInAuthProvider.AllowSignup {
			logger.Error("signup is enabled, but received access request")
			http.Error(w, "Use sign up instead.", http.StatusConflict)
			return
		}
		handleRequestAccess(logger, db, w, r)
	}
}

type requestAccessData struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	AdditionalInfo string `json:"additionalInfo"`
}

// handleRequestAccess handles submission of the request access form.
func handleRequestAccess(logger log.Logger, db database.DB, w http.ResponseWriter, r *http.Request) {
	var data requestAccessData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "could not decode request body", http.StatusBadRequest)
		return
	}

	if err := userpasswd.CheckEmailFormat(data.Email); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Create the access_request.
	accessRequest := types.AccessRequest{
		Name:           data.Name,
		Email:          data.Email,
		AdditionalInfo: data.AdditionalInfo,
	}
	_, err := db.AccessRequests().Create(r.Context(), &accessRequest)
	if err == nil {
		w.WriteHeader(http.StatusCreated)
		// TODO: Use EventRecorder from internal/telemetryrecorder instead.
		//lint:ignore SA1019 existing usage of deprecated functionality.
		if err = usagestats.LogBackendEvent(db, actor.FromContext(r.Context()).UID, deviceid.FromContext(r.Context()), "CreateAccessRequestSucceeded", nil, nil, featureflag.GetEvaluatedFlagSet(r.Context()), nil); err != nil {
			logger.Warn("Failed to log event CreateAccessRequestSucceeded", log.Error(err))
		}
		return
	}
	logger.Error("Error in access request.", log.String("email", data.Email), log.String("name", data.Name), log.Error(err))
	if database.IsAccessRequestUserWithEmailExists(err) || database.IsAccessRequestWithEmailExists(err) {
		// 🚨 SECURITY: We don't show an error message when the user or access request with the same e-mail address exists
		// as to not leak the existence of a given e-mail address in the database.
		w.WriteHeader(http.StatusCreated)
	} else if errcode.PresentationMessage(err) != "" {
		http.Error(w, errcode.PresentationMessage(err), http.StatusConflict)
	} else {
		// Do not show non-allowed error messages to user, in case they contain sensitive or confusing
		// information.
		http.Error(w, "Request access failed unexpectedly.", http.StatusInternalServerError)
	}

	// TODO: Use EventRecorder from internal/telemetryrecorder instead.
	//lint:ignore SA1019 existing usage of deprecated functionality.
	if err = usagestats.LogBackendEvent(db, actor.FromContext(r.Context()).UID, deviceid.FromContext(r.Context()), "AccessRequestFailed", nil, nil, featureflag.GetEvaluatedFlagSet(r.Context()), nil); err != nil {
		logger.Warn("Failed to log event AccessRequestFailed", log.Error(err))
	}
}
