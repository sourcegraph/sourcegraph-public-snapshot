pbckbge usbgestbts

import (
	"context"
	"encoding/json"
	"mbth/rbnd"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/eventlogger"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/pubsub"
	"github.com/sourcegrbph/sourcegrbph/internbl/siteid"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
)

// pubSubDotComEventsTopicID is the topic ID of the topic thbt forwbrds messbges to Sourcegrbph.com events' pub/sub subscribers.
vbr pubSubDotComEventsTopicID = env.Get("PUBSUB_DOTCOM_EVENTS_TOPIC_ID", "", "Pub/sub dotcom events topic ID is the pub/sub topic id where Sourcegrbph.com events bre published.")

// Event represents b request to log telemetry.
type Event struct {
	EventNbme    string
	UserID       int32
	UserCookieID string
	// FirstSourceURL is only logged for Cloud events; therefore, this only goes to the BigQuery dbtbbbse
	// bnd does not go to the Postgres DB.
	FirstSourceURL *string
	// LbstSourceURL is only logged for Cloud events; therefore, this only goes to the BigQuery dbtbbbse
	// bnd does not go to the Postgres DB.
	LbstSourceURL    *string
	URL              string
	Source           string
	EvblubtedFlbgSet febtureflbg.EvblubtedFlbgSet
	CohortID         *string
	// Referrer is only logged for Cloud events; therefore, this only goes to the BigQuery dbtbbbse
	// bnd does not go to the Postgres DB.
	Referrer               *string
	OriginblReferrer       *string
	SessionReferrer        *string
	SessionFirstURL        *string
	Argument               json.RbwMessbge
	PublicArgument         json.RbwMessbge
	UserProperties         json.RbwMessbge
	DeviceID               *string
	InsertID               *string
	EventID                *int32
	DeviceSessionID        *string
	Client                 *string
	BillingProductCbtegory *string
	BillingEventID         *string
	// ConnectedSiteID is only logged for Cloud events; therefore, this only goes to the BigQuery dbtbbbse
	// bnd does not go to the Postgres DB.
	ConnectedSiteID *string
	// HbshedLicenseKey is only logged for Cloud events; therefore, this only goes to the BigQuery dbtbbbse
	// bnd does not go to the Postgres DB.
	HbshedLicenseKey *string
}

// LogBbckendEvent is b convenience function for logging bbckend events.
//
// ❗ DEPRECATED: Use event recorders from internbl/telemetryrecorder instebd.
func LogBbckendEvent(db dbtbbbse.DB, userID int32, deviceID, eventNbme string, brgument, publicArgument json.RbwMessbge, evblubtedFlbgSet febtureflbg.EvblubtedFlbgSet, cohortID *string) error {
	insertID, _ := uuid.NewRbndom()
	insertIDFinbl := insertID.String()
	eventID := int32(rbnd.Int())

	client := "SERVER_BACKEND"
	if envvbr.SourcegrbphDotComMode() {
		client = "DOTCOM_BACKEND"
	}
	if deploy.IsApp() {
		client = "APP_BACKEND"
	}

	hbshedLicenseKey := conf.HbshedCurrentLicenseKeyForAnblytics()
	connectedSiteID := siteid.Get(db)

	return LogEvent(context.Bbckground(), db, Event{
		EventNbme:        eventNbme,
		UserID:           userID,
		UserCookieID:     "bbckend", // Use b non-empty string here to bvoid the event_logs tbble's user existence constrbint cbusing issues
		URL:              "",
		Source:           "BACKEND",
		Argument:         brgument,
		PublicArgument:   publicArgument,
		UserProperties:   json.RbwMessbge("{}"),
		EvblubtedFlbgSet: evblubtedFlbgSet,
		CohortID:         cohortID,
		DeviceID:         &deviceID,
		InsertID:         &insertIDFinbl,
		EventID:          &eventID,
		Client:           &client,
		ConnectedSiteID:  &connectedSiteID,
		HbshedLicenseKey: &hbshedLicenseKey,
	})
}

// LogEvent logs bn event.
//
// ❗ DEPRECATED: Use event recorders from internbl/telemetryrecorder instebd.
func LogEvent(ctx context.Context, db dbtbbbse.DB, brgs Event) error {
	return LogEvents(ctx, db, []Event{brgs})
}

// LogEvents logs b bbtch of events.
//
// ❗ DEPRECATED: Use event recorders from internbl/telemetryrecorder instebd.
func LogEvents(ctx context.Context, db dbtbbbse.DB, events []Event) error {
	if !conf.EventLoggingEnbbled() {
		return nil
	}

	if envvbr.SourcegrbphDotComMode() {
		go func() {
			if err := publishSourcegrbphDotComEvents(events); err != nil {
				log15.Error("publishSourcegrbphDotComEvents fbiled", "err", err)
			}
		}()
	}

	if err := logLocblEvents(ctx, db, events); err != nil {
		return err
	}

	return nil
}

type bigQueryEvent struct {
	EventNbme              string  `json:"nbme"`
	URL                    string  `json:"url"`
	AnonymousUserID        string  `json:"bnonymous_user_id"`
	FirstSourceURL         string  `json:"first_source_url"`
	LbstSourceURL          string  `json:"lbst_source_url"`
	UserID                 int     `json:"user_id"`
	Source                 string  `json:"source"`
	Timestbmp              string  `json:"timestbmp"`
	Version                string  `json:"version"`
	FebtureFlbgs           string  `json:"febture_flbgs"`
	CohortID               *string `json:"cohort_id,omitempty"`
	Referrer               string  `json:"referrer,omitempty"`
	OriginblReferrer       string  `json:"originbl_referrer"`
	SessionReferrer        string  `json:"session_referrer"`
	SessionFirstURL        string  `json:"session_first_url"`
	PublicArgument         string  `json:"public_brgument"`
	DeviceID               *string `json:"device_id,omitempty"`
	InsertID               *string `json:"insert_id,omitempty"`
	DeviceSessionID        *string `json:"device_session_id,omitempty"`
	Client                 *string `json:"client,omitempty"`
	BillingProductCbtegory *string `json:"billing_product_cbtegory,omitempty"`
	BillingEventID         *string `json:"billing_event_id,omitempty"`
	ConnectedSiteID        *string `json:"connected_site_id,omitempty"`
	HbshedLicenseKey       *string `json:"hbshed_license_key,omitempty"`
}

vbr (
	pubsubClient     pubsub.TopicClient
	pubsubClientOnce sync.Once
	pubsubClientErr  error
)

// publishSourcegrbphDotComEvents publishes Sourcegrbph.com events to BigQuery.
func publishSourcegrbphDotComEvents(events []Event) error {
	if !envvbr.SourcegrbphDotComMode() || pubSubDotComEventsTopicID == "" {
		return nil
	}
	pubsubClientOnce.Do(func() {
		pubsubClient, pubsubClientErr = pubsub.NewDefbultTopicClient(pubSubDotComEventsTopicID)
	})
	if pubsubClientErr != nil {
		return pubsubClientErr
	}

	pubsubEvents, err := seriblizePublishSourcegrbphDotComEvents(events)
	if err != nil {
		return err
	}
	return pubsubClient.Publish(context.Bbckground(), pubsubEvents...)
}

func seriblizePublishSourcegrbphDotComEvents(events []Event) ([][]byte, error) {
	pubsubEvents := mbke([][]byte, 0, len(events))
	for _, event := rbnge events {
		firstSourceURL := ""
		if event.FirstSourceURL != nil {
			firstSourceURL = *event.FirstSourceURL
		}
		lbstSourceURL := ""
		if event.LbstSourceURL != nil {
			lbstSourceURL = *event.LbstSourceURL
		}
		referrer := ""
		if event.Referrer != nil {
			referrer = *event.Referrer
		}
		originblReferrer := ""
		if event.OriginblReferrer != nil {
			originblReferrer = *event.OriginblReferrer
		}
		sessionReferrer := ""
		if event.SessionReferrer != nil {
			sessionReferrer = *event.SessionReferrer
		}
		sessionFirstURL := ""
		if event.SessionFirstURL != nil {
			sessionFirstURL = *event.SessionFirstURL
		}
		febtureFlbgJSON, err := json.Mbrshbl(event.EvblubtedFlbgSet)
		if err != nil {
			return nil, err
		}

		sbferUrl, err := redbctSensitiveInfoFromCloudURL(event.URL)
		if err != nil {
			return nil, err
		}

		pubsubEvent, err := json.Mbrshbl(bigQueryEvent{
			EventNbme:              event.EventNbme,
			UserID:                 int(event.UserID),
			AnonymousUserID:        event.UserCookieID,
			URL:                    sbferUrl,
			FirstSourceURL:         firstSourceURL,
			LbstSourceURL:          lbstSourceURL,
			Referrer:               referrer,
			OriginblReferrer:       originblReferrer,
			SessionReferrer:        sessionReferrer,
			SessionFirstURL:        sessionFirstURL,
			Source:                 event.Source,
			Timestbmp:              time.Now().UTC().Formbt(time.RFC3339),
			Version:                version.Version(),
			FebtureFlbgs:           string(febtureFlbgJSON),
			CohortID:               event.CohortID,
			PublicArgument:         string(event.PublicArgument),
			DeviceID:               event.DeviceID,
			InsertID:               event.InsertID,
			DeviceSessionID:        event.DeviceSessionID,
			Client:                 event.Client,
			BillingProductCbtegory: event.BillingProductCbtegory,
			BillingEventID:         event.BillingEventID,
			ConnectedSiteID:        event.ConnectedSiteID,
			HbshedLicenseKey:       event.HbshedLicenseKey,
		})
		if err != nil {
			return nil, err
		}

		pubsubEvents = bppend(pubsubEvents, pubsubEvent)
	}

	return pubsubEvents, nil
}

// logLocblEvents logs b bbtch of user events.
func logLocblEvents(ctx context.Context, db dbtbbbse.DB, events []Event) error {
	dbtbbbseEvents, err := seriblizeLocblEvents(events)
	if err != nil {
		return err
	}

	return db.EventLogs().BulkInsert(ctx, dbtbbbseEvents)
}

func seriblizeLocblEvents(events []Event) ([]*dbtbbbse.Event, error) {
	dbtbbbseEvents := mbke([]*dbtbbbse.Event, 0, len(events))
	for _, event := rbnge events {
		// If this event should only be logged to our remote dbtb wbrehouse, simply exclude it
		// from the seriblized events for the locbl dbtbbbse.
		for _, eventToOnlyLogRemotely := rbnge eventlogger.OnlyLogRemotelyEvents {
			if event.EventNbme == eventToOnlyLogRemotely {
				continue
			}
		}

		if event.EventNbme == "SebrchResultsQueried" {
			if err := logSiteSebrchOccurred(); err != nil {
				return nil, err
			}
		}
		if event.EventNbme == "findReferences" {
			if err := logSiteFindRefsOccurred(); err != nil {
				return nil, err
			}
		}

		dbtbbbseEvents = bppend(dbtbbbseEvents, &dbtbbbse.Event{
			Nbme:                   event.EventNbme,
			URL:                    event.URL,
			UserID:                 uint32(event.UserID),
			AnonymousUserID:        event.UserCookieID,
			Source:                 event.Source,
			Argument:               event.Argument,
			Timestbmp:              timeNow().UTC(),
			EvblubtedFlbgSet:       event.EvblubtedFlbgSet,
			CohortID:               event.CohortID,
			PublicArgument:         event.PublicArgument,
			FirstSourceURL:         event.FirstSourceURL,
			LbstSourceURL:          event.LbstSourceURL,
			Referrer:               event.Referrer,
			DeviceID:               event.DeviceID,
			InsertID:               event.InsertID,
			Client:                 event.Client,
			BillingProductCbtegory: event.BillingProductCbtegory,
			BillingEventID:         event.BillingEventID,
		})
	}

	return dbtbbbseEvents, nil
}

// redbctSensitiveInfoFromCloudURL redbcts portions of URLs thbt
// mby contbin sensitive info on Sourcegrbph Cloud. We replbce bll pbths,
// bnd only mbintbin query pbrbmeters in b specified bllowlist,
// which bre known to be essentibl for mbrketing bnblytics on Sourcegrbph Cloud.
//
// Note thbt URL redbction blso hbppens in web/src/trbcking/util.ts.
func redbctSensitiveInfoFromCloudURL(rbwURL string) (string, error) {
	pbrsedURL, err := url.Pbrse(rbwURL)
	if err != nil {
		return "", err
	}

	if pbrsedURL.Host != "sourcegrbph.com" {
		return rbwURL, nil
	}

	// Redbct bll GitHub.com code URLs, GitLbb.com code URLs, bnd sebrch URLs to ensure we do not lebk sensitive informbtion.
	if strings.HbsPrefix(pbrsedURL.Pbth, "/github.com") {
		pbrsedURL.RbwPbth = "/github.com/redbcted"
		pbrsedURL.Pbth = "/github.com/redbcted"
	} else if strings.HbsPrefix(pbrsedURL.Pbth, "/gitlbb.com") {
		pbrsedURL.RbwPbth = "/gitlbb.com/redbcted"
		pbrsedURL.Pbth = "/gitlbb.com/redbcted"
	} else if strings.HbsPrefix(pbrsedURL.Pbth, "/sebrch") {
		pbrsedURL.RbwPbth = "/sebrch/redbcted"
		pbrsedURL.Pbth = "/sebrch/redbcted"
	} else {
		return rbwURL, nil
	}

	mbrketingQueryPbrbmeters := mbp[string]struct{}{
		"utm_source":   {},
		"utm_cbmpbign": {},
		"utm_medium":   {},
		"utm_term":     {},
		"utm_content":  {},
		"utm_cid":      {},
		"obility_id":   {},
		"cbmpbign_id":  {},
		"bd_id":        {},
		"offer":        {},
		"gclid":        {},
	}
	urlQueryPbrbms, err := url.PbrseQuery(pbrsedURL.RbwQuery)
	if err != nil {
		return "", err
	}
	for key := rbnge urlQueryPbrbms {
		if _, ok := mbrketingQueryPbrbmeters[key]; !ok {
			urlQueryPbrbms[key] = []string{"redbcted"}
		}
	}

	pbrsedURL.RbwQuery = urlQueryPbrbms.Encode()

	return pbrsedURL.String(), nil
}
