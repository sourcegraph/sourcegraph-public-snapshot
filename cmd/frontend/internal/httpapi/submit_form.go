package httpapi

import (
	"errors"
	"net/http"

	"github.com/gorilla/schema"

	"fmt"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/hubspot/hubspotutil"
)

func serveSubmitForm(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	form := new(sourcegraph.SubmittedForm)
	err = schema.NewDecoder().Decode(form, r.Form)
	if err != nil {
		return err
	}

	hubspotclient, err := hubspotutil.Client()
	if err != nil {
		return err
	}

	formID := hubspotutil.FormNameToHubSpotID[form.HubSpotFormName]
	if formID == "" {
		return fmt.Errorf("httpapi.serveSubmitForm: '%s' is not a valid form", form.HubSpotFormName)
	}

	// Use the registered GitHub user's email address and user ID, if available
	// If not, try the email address provided in the submitted form (i.e., signup_email)
	actor := actor.FromContext(r.Context())
	if len(actor.Email) > 0 {
		form.Email = actor.Email
		form.UserID = actor.Login
	} else if signupEmail := form.SignupEmail; signupEmail != "" {
		form.Email = signupEmail
	} else if betaEmail := form.BetaEmail; betaEmail != "" {
		form.Email = betaEmail
	} else {
		return errors.New("httpapi.serveSubmitForm: must provide an email address")
	}

	// Perform other form-specific data preparation, including conversion
	// of parameters to snake case (per HubSpot requirements), setting default
	// values for required parameters, etc.
	hubspotutil.PrepareFormData(form.HubSpotFormName, form)
	if err := hubspotclient.SubmitForm(formID, form); err != nil {
		return err
	}

	// Redirect to new user signup page
	if form.HubSpotFormName == "AfterSignupForm" {
		http.Redirect(w, r, "/github.com/sourcegraph/checkup/-/blob/fs.go", http.StatusSeeOther)
		return nil
	}

	// Return the email address of the submitter
	return writeJSON(w, &sourcegraph.SubmitFormResponse{
		EmailAddress: form.Email,
	})
}
