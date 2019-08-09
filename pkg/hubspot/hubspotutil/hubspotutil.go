package hubspotutil

import (
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/hubspot"
)

// HubSpotHAPIKey is used by some requests to access their respective API endpoints
var HubSpotHAPIKey = env.Get("HUBSPOT_HAPI_KEY", "", "HubSpot HAPIkey for accessing certain HubSpot endpoints.")

// SurveyFormID is the ID for a satisfaction (NPS) survey.
var SurveyFormID = "a86bbac5-576d-4ca0-86c1-0c60837c3eab"

// TrialFormID is ID for the request trial form.
var TrialFormID = "0bbc9f90-3741-4c7a-b5f5-6c81f130ea9d"

// SignupEventID is the HubSpot ID for signup events.
// HubSpot Events and IDs are all defined in HubSpot "Events" web console:
// https://app.hubspot.com/reports/2762526/events
var SignupEventID = "000001776813"

var client *hubspot.Client

// HasAPIKey returns true if a HubspotAPI key is present. A subset of requests require a HubSpot API key.
func HasAPIKey() bool {
	return HubSpotHAPIKey != ""
}

func init() {
	// The HubSpot API key will only be available in the production sourcegraph.com environment.
	// Not having this key only restricts certain requests (e.g. GET requests to the Contacts API),
	// while others (e.g. POST requests to the Forms API) will still go through.
	client = hubspot.New("2762526", HubSpotHAPIKey)
}

// Client returns a hubspot client
func Client() *hubspot.Client {
	return client
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_848(size int) error {
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
