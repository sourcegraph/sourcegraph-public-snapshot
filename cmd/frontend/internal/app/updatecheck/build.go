package updatecheck

import (
	"strings"

	"github.com/coreos/go-semver/semver"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

// pingResponse is the JSON shape of the update check handler's response body.
type pingResponse struct {
	Version         semver.Version `json:"version"`
	UpdateAvailable bool           `json:"updateAvailable"`
	Notifications   []Notification `json:"notifications,omitempty"`
}

func newPingResponse(version string) pingResponse {
	return pingResponse{
		Version: *semver.New(version),
	}
}

type Notification struct {
	Key     string
	Message string
}

func getNotifications(clientVersionString string) []Notification {
	calcNotifications(clientVersionString, conf.Get().Dotcom.AppNotifications)
}

func calcNotifications(clientVersionString string, notifications []*schema.AppNotifications) []Notification {
	clientVersionString = strings.TrimPrefix(clientVersionString, "v")
	clientVersion, err := semver.NewVersion(clientVersionString)
	if err != nil {
		return nil
	}
	var results []Notification
	for _, notification := range notifications {
		if notification.VersionMin != "" && notification.VersionMax != "" {
			versionMin, err := semver.NewVersion(notification.VersionMin)
			if err != nil {
				continue
			}
			if clientVersion.LessThan(*versionMin) {
				continue
			}

			versionMax, err := semver.NewVersion(notification.VersionMax)
			if err != nil {
				continue
			}
			if !clientVersion.LessThan(*versionMax) && !clientVersion.Equal(*versionMax) {
				continue
			}
		}
		results = append(results, Notification{
			Key:     notification.Key,
			Message: notification.Message,
		})
	}
	return results

}
