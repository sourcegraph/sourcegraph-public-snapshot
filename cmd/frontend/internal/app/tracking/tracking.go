package tracking

import (
	"encoding/json"
	"log"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-github/github"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/tracking/slack"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/gcstracker"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/hubspot"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/hubspot/hubspotutil"
)

// Limits to the number of errors we can recieve from external GitHub
// requests before giving up on fetching more data
//
// If GitHub returns an error on a fetch, we want to continue to try
// with the next item (unless it was a rate limit error, which is
// hanleded separately). Only after a sufficient number of unexplained
// errors do we want to give up
var maxOrgMemberErrors = 1

// TrackUser handles user data logging during auth flows
//
// Specifically, updating user data properties in HubSpot
func TrackUser(a *actor.Actor, event string) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic in tracking.TrackUser: %s", err)
		}
	}()

	// If the user is in a dev environment, don't do any data pulls from GitHub, or any tracking
	if env.Version == "dev" {
		return
	}

	// Generate a single set of user parameters for HubSpot and Slack exports
	contactParams := &hubspot.ContactProperties{
		UserID: a.Email,
		UID:    a.UID,
	}

	// Update or create user contact information in HubSpot
	hsResponse, err := trackHubSpotContact(a.Email, event, contactParams)
	if err != nil {
		log15.Warn("trackHubSpotContact: failed to create or update HubSpot contact on auth", "source", "HubSpot", "error", err)
	}

	// Finally, post signup notification to Slack
	if event == "SignupCompleted" {
		err = slack.NotifyOnSignup(a, contactParams, hsResponse)
		if err != nil {
			log15.Error("Error sending new signup details to Slack", "error", err)
			return
		}
	}
}

func trackHubSpotContact(email string, eventLabel string, params *hubspot.ContactProperties) (*hubspot.ContactResponse, error) {
	if email == "" {
		return nil, errors.New("User must have a valid email address.")
	}

	c, err := hubspotutil.Client()
	if err != nil {
		return nil, errors.Wrap(err, "hubspotutil.Client")
	}

	if eventLabel == "SignupCompleted" {
		today := time.Now().Truncate(24 * time.Hour)
		// Convert to milliseconds
		params.RegisteredAt = today.UTC().Unix() * 1000
	}

	// Create or update the contact
	resp, err := c.CreateOrUpdateContact(email, params)
	if err != nil {
		return nil, err
	}

	// Log the event if relevant (in this case, for "SignupCompleted" events)
	if _, ok := hubspotutil.EventNameToHubSpotID[eventLabel]; ok {
		err := c.LogEvent(email, hubspotutil.EventNameToHubSpotID[eventLabel], map[string]string{})
		if err != nil {
			return nil, errors.Wrap(err, "LogEvent")
		}
	}

	// Parse response
	hsResponse := &hubspot.ContactResponse{}
	err = json.Unmarshal(resp, hsResponse)
	if err != nil {
		return nil, err
	}

	return hsResponse, nil
}

// TrackGitHubWebhook handles webhooks received from GitHub.com regarding Sourcegraph
// installations on GitHub orgs
func TrackGitHubWebhook(eventType string, event interface{}) error {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic in tracking.TrackGitHubWebhook: %s", err)
		}
	}()

	switch event := event.(type) {
	case *github.InstallationEvent:
		return trackInstallationEvent(event)
	case *github.InstallationRepositoriesEvent:
		return trackInstallationRepositoriesEvent(event)
	}

	log15.Warn("Unhandled GitHub webhook received", "type", eventType)
	return nil
}

func trackInstallationEvent(event *github.InstallationEvent) error {
	if event.Sender.Login == nil {
		return errors.New("Installing user must have a Login")
	}
	var senderEmail string
	if event.Sender.Email != nil {
		senderEmail = *event.Sender.Email
	}

	gcsClient, err := gcstracker.NewFromUserProperties(*event.Sender.Login, senderEmail, "")
	if err != nil {
		return errors.Wrap(err, "Error creating a new GCS client")
	}
	tos := gcsClient.NewTrackedObjects("GitHubAppInstallationCompleted")
	tos.AddGitHubInstallationEvent(event)
	err = gcsClient.Write(tos)
	if err != nil {
		return errors.Wrap(err, "Error writing to GCS")
	}

	err = slack.NotifyOnAppInstall(*event.Sender.Login, gitHubLink(*event.Sender.Login), lookerUserLink(*event.Sender.Login), event.Installation.Account, gitHubLink(*event.Installation.Account.Login))
	if err != nil {
		return errors.Wrap(err, "Error sending new app install details to Slack")
	}
	return nil
}

func trackInstallationRepositoriesEvent(event *github.InstallationRepositoriesEvent) error {
	if event.Sender.Login == nil {
		return errors.New("Installing user must have a Login")
	}
	var senderEmail string
	if event.Sender.Email != nil {
		senderEmail = *event.Sender.Email
	}

	gcsClient, err := gcstracker.NewFromUserProperties(*event.Sender.Login, senderEmail, "")
	if err != nil {
		return errors.Wrap(err, "Error creating a new GCS client")
	}

	tos := gcsClient.NewTrackedObjects("GitHubAppRepositoriesUpdated")
	tos.AddGitHubRepositoriesEvent(event)
	err = gcsClient.Write(tos)
	if err != nil {
		return errors.Wrap(err, "Error writing to GCS")
	}

	return nil
}

func gitHubLink(login string) string {
	return "https://github.com/" + login
}

func lookerUserLink(login string) string {
	return "https://sourcegraph.looker.com/dashboards/9?User%20ID=" + login
}
