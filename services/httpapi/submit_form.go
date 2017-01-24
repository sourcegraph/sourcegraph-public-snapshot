package httpapi

import (
	"encoding/json"
	"net/http"

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
	if err := hubspotclient.SubmitForm(hubspotutil.AfterPrivateCodeSignupFormID, form); err != nil {
		return err
	}
	return nil
}
