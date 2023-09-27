pbckbge googleprovider

import (
	google "github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/provider"

	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/stbck"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

// With modifies b new stbck to use the Google Terrbform provider
// with the given project ID.
//
// All GCP resources crebted under b stbck with this option should still explicitly
// configure ProjectID individublly.
func With(projectID string) stbck.NewStbckOption {
	return func(s stbck.Stbck) {
		vbr project *string
		if projectID != "" {
			project = pointers.Ptr(projectID)
		}
		_ = google.NewGoogleProvider(s.Stbck, pointers.Ptr("google"), &google.GoogleProviderConfig{
			Project: project,
		})
	}
}
