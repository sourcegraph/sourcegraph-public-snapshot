package updatecheck

import (
	"context"
	"strings"

	"github.com/coreos/go-semver/semver"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
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
	if !envvar.SourcegraphDotComMode() {
		return []Notification{}
	}
	return calcNotifications(clientVersionString, conf.Get().Dotcom.AppNotifications)
}

func calcNotifications(clientVersionString string, notifications []*schema.AppNotifications) []Notification {
	clientVersionString = strings.TrimPrefix(clientVersionString, "v")
	clientVersion, err := semver.NewVersion(clientVersionString)
	if err != nil {
		return nil
	}
	var results []Notification
	for _, notification := range notifications {
		if len(strings.Split(notification.Key, "-")) < 4 {
			// TODO(app): this is a poor/approximate check for "YYYY-MM-DD-" prefix that we mandate
			continue
		}
		if notification.VersionMin != "" {
			versionMin, err := semver.NewVersion(notification.VersionMin)
			if err != nil {
				continue
			}
			if clientVersion.LessThan(*versionMin) {
				continue
			}
		}
		if notification.VersionMax != "" {
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

// handleNotifications is called on a Sourcegraph client instance to handle notification messages that
// the client recieved from the server (sourcegraph.com). They get stored in the site config.
func (r pingResponse) handleNotifications() {
	ctx := context.Background()

	server := globals.ConfigurationServerFrontendOnly
	if server == nil {
		// Cannot ever happen, as updatecheck only runs in the frontend, but just in case do nothing.
		return
	}

	// Update the site configuration "app.notifications" field. Note that this also removes notifications
	// if they are no longer present in the sourcegraph.com site configuration.
	var notifications []*schema.Notifications
	for _, v := range r.Notifications {
		notifications = append(notifications, &schema.Notifications{
			Key:     v.Key,
			Message: v.Message,
		})
	}
	updated := conf.Raw()
	var err error
	updated.Site, err = jsonc.Edit(updated.Site, notifications, "notifications")
	if err != nil {
		return // clearly our edit logic would be broken, so do nothing (better than panic in the case of pings.)
	}

	if err := server.Write(ctx, updated, updated.ID, 0); err != nil {
		// error or conflicting edit; do nothing, the next updatecheck will try again.
		return
	}
}
