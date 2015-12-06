package kv

import (
	"html/template"
	"time"

	"src.sourcegraph.com/apps/notifications/notifications"
	"src.sourcegraph.com/apps/tracker/issues"
)

// repoSpec is an on-disk representation of issues.RepoSpec.
type repoSpec struct {
	URI string
}

func fromRepoSpec(rs issues.RepoSpec) repoSpec {
	return repoSpec{URI: rs.URI}
}

func (rs repoSpec) RepoSpec() issues.RepoSpec {
	return issues.RepoSpec{URI: rs.URI}
}

// octiconID is an on-disk representation of notifications.OcticonID.
type octiconID string

func fromOcticonID(o notifications.OcticonID) octiconID {
	return octiconID(o)
}

func (o octiconID) OcticonID() notifications.OcticonID {
	return notifications.OcticonID(o)
}

// notification is an on-disk representation of notification.
type notification struct {
	RepoSpec  repoSpec
	Title     string
	Icon      octiconID
	UpdatedAt time.Time
	HTMLURL   template.URL
}

func subscribersBucket(appID string, threadID uint64) string {
	return "subscribers" + "-" + appID + "-" + formatUint64(threadID)
}

func notificationKey(repo issues.RepoSpec, appID string, threadID uint64) string {
	return repo.URI + "-" + appID + "-" + formatUint64(threadID)
}
