package accessrequest

import (
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/log"

	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/userpasswd"
	sgactor "github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/deviceid"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
)

// HandleRequestAccess handles submission of the request access form.
func HandleRequestAccess(logger log.Logger, db database.DB) http.HandlerFunc {
	logger = logger.Scoped("HandleRequestAccess", "request access request handler")
	return func(w http.ResponseWriter, r *http.Request) {
		if handleEnabledCheck(logger, w) {
			return
		}
		builtInAuthProvider, _ := userpasswd.GetProviderConfig();
		if builtInAuthProvider != nil  && builtInAuthProvider.AllowSignup {
			http.Error(w, "Use sign up instead.", http.StatusConflict)
			return
		}
		handleRequestAccess(logger, db, w, r)
	}
}

func handleEnabledCheck(logger log.Logger, w http.ResponseWriter) (handled bool) {
	experimentalFeatures := conf.Get().ExperimentalFeatures
	if experimentalFeatures != nil && experimentalFeatures.AccessRequests != nil && !experimentalFeatures.AccessRequests.Enabled {
		logger.Error("Request access is not enabled.")
		http.Error(w, "Request access is not enabled.", http.StatusConflict)
		return true
	}

	// test whether
	return false
}

type requestAccessData struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	AdditionalInfo string `json:"additionalInfo"`
}

// handleRequestAccess handles requests to /request-access.
func handleRequestAccess(logger log.Logger, db database.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, fmt.Sprintf("unsupported method %s", r.Method), http.StatusBadRequest)
		return
	}
	var data requestAccessData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "could not decode request body", http.StatusBadRequest)
		return
	}

	const defaultErrorMessage = "Request access failed unexpectedly."

	if err := userpasswd.CheckEmailFormat(data.Email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create the access_request.
	accessRequestCreateData := database.AccessRequestCreate{
		Name:           data.Name,
		Email:          data.Email,
		AdditionalInfo: data.AdditionalInfo,
	}

	_, err := db.AccessRequests().Create(r.Context(), accessRequestCreateData)
	if err != nil {
		var (
			message    string
			statusCode int
		)
		switch {
		case database.IsAccessRequestUserWithEmailExists(err):
			// TODO: clarify how to handle this error, since this will be shown to unauthenticated user
			message = "A user with this email already exists."
			statusCode = http.StatusConflict
		case database.IsAccessRequestWithEmailExists(err):
			// TODO: clarify how to handle this error, since this will be shown to unauthenticated user
			message = "A access request was already created previously."
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
		logger.Error("Error in access request.", log.String("email", data.Email), log.String("name", data.Name), log.Error(err))
		http.Error(w, message, statusCode)

		if err = usagestats.LogBackendEvent(db, sgactor.FromContext(r.Context()).UID, deviceid.FromContext(r.Context()), "AccessRequestFailed", nil, nil, featureflag.GetEvaluatedFlagSet(r.Context()), nil); err != nil {
			logger.Warn("Failed to log event AccessRequestFailed", log.Error(err))
		}

		return
	}

	w.WriteHeader(http.StatusCreated)
	if err = usagestats.LogBackendEvent(db, sgactor.FromContext(r.Context()).UID, deviceid.FromContext(r.Context()), "CreateAccessRequestSucceeded", nil, nil, featureflag.GetEvaluatedFlagSet(r.Context()), nil); err != nil {
		logger.Warn("Failed to log event CreateAccessRequestSucceeded", log.Error(err))
	}
}
