// Package tracking contains code for tracking users on Sourcegraph.com
package tracking

import (
	"log"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/pkg/hubspot"
	"github.com/sourcegraph/sourcegraph/pkg/hubspot/hubspotutil"
)

// SyncUser handles creating or syncing a user profile in HubSpot, and if provided,
// logs a user event.
func SyncUser(email string, eventID string, contactParams *hubspot.ContactProperties) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic in tracking.SyncUser: %s", err)
		}
	}()
	// If the user no API token present or on-prem environment, don't do any tracking
	if !hubspotutil.HasAPIKey() || !envvar.SourcegraphDotComMode() {
		return
	}

	// Update or create user contact information in HubSpot
	err := syncHubSpotContact(email, eventID, contactParams)
	if err != nil {
		log15.Warn("syncHubSpotContact: failed to create or update HubSpot contact", "source", "HubSpot", "error", err)
	}
}

func syncHubSpotContact(email, eventID string, contactParams *hubspot.ContactProperties) error {
	if email == "" {
		return errors.New("user must have a valid email address")
	}

	// Generate a single set of user parameters for HubSpot
	if contactParams == nil {
		contactParams = &hubspot.ContactProperties{}
	}
	contactParams.UserID = email

	c := hubspotutil.Client()

	// Create or update the contact
	_, err := c.CreateOrUpdateContact(email, contactParams)
	if err != nil {
		return err
	}

	// Log the user event
	if eventID != "" {
		err = c.LogEvent(email, eventID, map[string]string{})
		if err != nil {
			return errors.Wrap(err, "LogEvent")
		}
	}

	return nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_287(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
