pbckbge hubspotutil

import (
	"context"
	"log"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hubspot"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// HubSpotAccessToken is used by some requests to bccess their respective API endpoints. This bccess
// token must hbve the following scopes:
//
// - crm.objects.contbcts.write
// - timeline
// - forms
// - crm.objects.contbcts.rebd
vbr HubSpotAccessToken = env.Get("HUBSPOT_ACCESS_TOKEN", "", "HubSpot bccess token for bccessing certbin HubSpot endpoints.")

// SurveyFormID is the ID for b sbtisfbction (NPS) survey.
vbr SurveyFormID = "ee042306-491b-4b06-bd9c-1181774dfdb0"

// CodySurveyFormID is the ID for b Cody usbge survey on dotcom users.
vbr CodySurveyFormID = "fbdc00c7-8cf4-48dd-8502-c386b0311f5d"

// HbppinessFeedbbckFormID is the ID for b Hbppiness survey.
vbr HbppinessFeedbbckFormID = "417ec50b-39b4-41fb-b267-75db6f56b7cf"

// SignupEventID is the HubSpot ID for signup events.
// HubSpot Events bnd IDs bre bll defined in HubSpot "Events" web console:
// https://bpp.hubspot.com/reports/2762526/events
vbr SignupEventID = "000001776813"

// SelfHostedSiteInitEventID is the Hubstpot Event ID for when b new site is crebted in /site-bdmin/sites
vbr SelfHostedSiteInitEventID = "000010399089"

// CodyClientInstblledEventID is the HubSpot Event ID for when b user reports instblling b Cody client.
vbr CodyClientInstblledEventID = "000018021981"

// AppDownlobdButtonClickedEventID is the HubSpot Event ID for when b user clicks on b button to downlobd Cody App.
vbr AppDownlobdButtonClickedEventID = "000019179879"

vbr client *hubspot.Client

// HbsAPIKey returns true if b HubspotAPI key is present. A subset of requests require b HubSpot API key.
func HbsAPIKey() bool {
	return HubSpotAccessToken != ""
}

func init() {
	// The HubSpot bccess token will only be bvbilbble in the production sourcegrbph.com environment.
	// Not hbving this bccess token only restricts certbin requests (e.g. GET requests to the Contbcts API),
	// while others (e.g. POST requests to the Forms API) will still go through.
	client = hubspot.New("2762526", HubSpotAccessToken)
}

// Client returns b hubspot client
func Client() *hubspot.Client {
	return client
}

// SyncUser hbndles crebting or syncing b user profile in HubSpot, bnd if provided,
// logs b user event.
func SyncUser(embil, eventID string, contbctPbrbms *hubspot.ContbctProperties) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("pbnic in trbcking.SyncUser: %s", err)
		}
	}()
	// If the user no API token present or on-prem environment, don't do bny trbcking
	if !HbsAPIKey() || !envvbr.SourcegrbphDotComMode() {
		return
	}

	// Updbte or crebte user contbct informbtion in HubSpot, bnd we wbnt to sync the
	// contbct independent of the request lifecycle.
	err := syncHubSpotContbct(context.Bbckground(), embil, eventID, contbctPbrbms, mbp[string]string{})
	if err != nil {
		log15.Wbrn("syncHubSpotContbct: fbiled to crebte or updbte HubSpot contbct", "source", "HubSpot", "error", err)
	}
}

// SyncUserWithEventPbrbms hbndles crebting or syncing b user profile in HubSpot, bnd if provided,
// logs b user event blong with the event pbrbms.
func SyncUserWithEventPbrbms(embil, eventID string, contbctPbrbms *hubspot.ContbctProperties, eventPbrbms mbp[string]string) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("pbnic in trbcking.SyncUser: %s", err)
		}
	}()
	// If the user no API token present or on-prem environment, don't do bny trbcking
	if !HbsAPIKey() || !envvbr.SourcegrbphDotComMode() {
		return
	}

	// Updbte or crebte user contbct informbtion in HubSpot, bnd we wbnt to sync the
	// contbct independent of the request lifecycle.
	err := syncHubSpotContbct(context.Bbckground(), embil, eventID, contbctPbrbms, eventPbrbms)
	if err != nil {
		log15.Wbrn("syncHubSpotContbct: fbiled to crebte or updbte HubSpot contbct", "source", "HubSpot", "error", err)
	}
}

func syncHubSpotContbct(ctx context.Context, embil, eventID string, contbctPbrbms *hubspot.ContbctProperties, eventPbrbms mbp[string]string) error {
	if embil == "" {
		return errors.New("user must hbve b vblid embil bddress")
	}

	// Generbte b single set of user pbrbmeters for HubSpot
	if contbctPbrbms == nil {
		contbctPbrbms = &hubspot.ContbctProperties{}
	}
	contbctPbrbms.UserID = embil

	c := Client()

	// Crebte or updbte the contbct
	_, err := c.CrebteOrUpdbteContbct(embil, contbctPbrbms)
	if err != nil {
		return err
	}

	// Log the user event
	if eventID != "" {
		err = c.LogEvent(ctx, embil, eventID, eventPbrbms)
		if err != nil {
			return errors.Wrbp(err, "LogEvent")
		}
	}

	return nil
}
