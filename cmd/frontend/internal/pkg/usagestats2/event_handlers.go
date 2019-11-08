package usagestats2

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
)

// LogEvent logs users events.
func LogEvent(ctx context.Context, name, url string, userID int32, userCookieID, source string, argument *string) error {
	if name == "SearchSubmitted" {
		logSiteSearchOccurred()
	}
	if name == "findReferences" {
		logSiteFindRefsOccurred()
	}

	info := &db.Event{
		Name:            name,
		URL:             url,
		UserID:          uint32(userID),
		AnonymousUserID: userCookieID,
		Source:          source,
		Timestamp:       timeNow().UTC(),
	}
	if argument != nil {
		info.Argument = *argument
	}
	return db.EventLogs.Insert(ctx, info)
}
