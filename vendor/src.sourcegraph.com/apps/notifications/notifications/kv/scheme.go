package kv

import (
	"html/template"
	"time"

	"src.sourcegraph.com/apps/notifications/notifications"
)

// repoSpec is an on-disk representation of notifications.RepoSpec.
type repoSpec struct {
	URI string
}

func fromRepoSpec(rs notifications.RepoSpec) repoSpec {
	return repoSpec{URI: rs.URI}
}

func (rs repoSpec) RepoSpec() notifications.RepoSpec {
	return notifications.RepoSpec{URI: rs.URI}
}

// octiconID is an on-disk representation of notifications.OcticonID.
type octiconID string

func fromOcticonID(o notifications.OcticonID) octiconID {
	return octiconID(o)
}

func (o octiconID) OcticonID() notifications.OcticonID {
	return notifications.OcticonID(o)
}

type rgb struct {
	R, G, B uint8
}

func fromRGB(c notifications.RGB) rgb {
	return rgb{R: c.R, G: c.G, B: c.B}
}

func (c rgb) RGB() notifications.RGB {
	return notifications.RGB{R: c.R, G: c.G, B: c.B}
}

// notification is an on-disk representation of notification.
type notification struct {
	RepoSpec  repoSpec
	Title     string
	Icon      octiconID
	Color     rgb
	UpdatedAt time.Time
	HTMLURL   template.URL
}

func subscribersBucket(appID string, threadID uint64) string {
	return "subscribers" + "-" + appID + "-" + formatUint64(threadID)
}

func notificationKey(repo notifications.RepoSpec, appID string, threadID uint64) string {
	return repo.URI + "-" + appID + "-" + formatUint64(threadID)
}
