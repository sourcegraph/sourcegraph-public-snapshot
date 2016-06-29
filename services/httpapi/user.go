package httpapi

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"os"

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
