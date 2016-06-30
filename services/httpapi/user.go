package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
)

var mailchimpKey string = os.Getenv("MAILCHIMP_KEY")

func serveUser(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	userSpec, err := routevar.ParseUserSpec(mux.Vars(r)["User"])
	if err != nil {
		return err
	}

	user, err := cl.Users.Get(ctx, &userSpec)
	if err != nil {
		return err
	}
	return writeJSON(w, user)
}

func serveUserEmails(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	userSpec, err := routevar.ParseUserSpec(mux.Vars(r)["User"])
	if err != nil {
		return err
	}

	emails, err := cl.Users.ListEmails(ctx, &userSpec)
	if err != nil {
		return err
	}
	return writeJSON(w, emails)
}

type betaSubscription struct {
	Email     string
	FirstName string
	LastName  string
	Languages []string
	Editors   []string
	Message   string
}

type mailchimpPayload struct {
	Status       string                 `json:"status"`
	EmailAddress string                 `json:"email_address"`
	MergeFields  map[string]interface{} `json:"merge_fields"`
}

// mailchimpArray returns the given list of strings as a plaintext-string
// usable within Mailchimp as an 'array' of sorts. Basically:
//
//  mailchimpArray("a", "b", "c") == "a,b,c,"
//
// The trailing comma is significant because it allows matching a singular
// element within mailchimp using the "contains" operator. For example:
//
//  mailchimpArray("Visual Studio Code", "Sublime") == "Visual Studio Code,Sublime,"
//
// To match users who have selected Visual Studio (not VSCode), we would use "contains"
// and "Visual Studio,". The trailing comma is important because otherwise the
// query would also match all people who have picked "Visual Studio Code", i.e.
// without a comma the "contains" operation is a substring search.
//
// To match users who signed up with the old single-value registration form,
// just use the following configuration:
//
//  "Subscribers match ___ of the following conditions": "all"
//  "contains": "Visual Studio,"
//  "is": "Visual Studio"
//
func mailchimpArray(values []string) string {
	return strings.Join(values, ",") + ","
}

func serveBetaSubscription(w http.ResponseWriter, r *http.Request) error {
	var sub betaSubscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		return err
	}

	data, err := json.Marshal(&mailchimpPayload{
		Status:       "subscribed",
		EmailAddress: sub.Email,
		MergeFields: map[string]interface{}{
			"FNAME":    sub.FirstName,
			"LNAME":    sub.LastName,
			"LANGUAGE": mailchimpArray(sub.Languages),
			"EDITOR":   mailchimpArray(sub.Editors),
			"MESSAGE":  sub.Message,
		},
	})
	if err != nil {
		return err
	}

	if mailchimpKey == "" {
		return errors.New("mailchimp authorization key only available on production environments")
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://us8.api.mailchimp.com/3.0/lists/dd6c4706a1/members", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("user", mailchimpKey)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return writeJSON(w, string(body))
}

func serveEmailSubscription(w http.ResponseWriter, r *http.Request) error {
	newBody, newErr := ioutil.ReadAll(r.Body)
	if newErr != nil {
		return newErr
	}

	if mailchimpKey == "" {
		return errors.New("mailchimp authorization key only available on production environments")
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://us8.api.mailchimp.com/3.0/lists/dd6c4706a1/members", bytes.NewReader(newBody))
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("user", mailchimpKey)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return writeJSON(w, string(body))

}
