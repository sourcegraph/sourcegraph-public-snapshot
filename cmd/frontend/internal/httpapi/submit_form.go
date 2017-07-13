package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"fmt"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/hubspot/hubspotutil"
)

func serveSubmitForm(w http.ResponseWriter, r *http.Request) error {
	var form map[string]string
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		return err
	}

	hubspotclient, err := hubspotutil.Client()
	if err != nil {
		return err
	}

	formName, ok := form["hubSpotFormName"]
	if !ok {
		return errors.New("httpapi.serveSubmitForm: must provide a HubSpot form name")
	}

	formID := hubspotutil.FormNameToHubSpotID[formName]
	if formID == "" {
		return fmt.Errorf("httpapi.serveSubmitForm: '%s' is not a valid form", formName)
	}

	// Use the registered GitHub user's email address and user ID, if available
	// If not, try the email address provided in the submitted form (i.e., signupEmail)
	actor := actor.FromContext(r.Context())
	if len(actor.Email) > 0 {
		form["email"] = actor.Email
		form["user_id"] = actor.Login
	} else if signupEmail := form["signupEmail"]; signupEmail != "" {
		form["email"] = signupEmail
	} else {
		return errors.New("httpapi.serveSubmitForm: must provide an email address")
	}

	// Perform other form-specific data preparation, including conversion
	// of parameters to snake case (per HubSpot requirements), setting default
	// values for required parameters, etc.
	formData := hubspotutil.PrepareFormData(formName, form)

	if err := hubspotclient.SubmitForm(formID, formData); err != nil {
		return err
	}

	// Return the email address of the submitter
	return writeJSON(w, &sourcegraph.SubmitFormResponse{
		EmailAddress: form["email"],
	})
}
