package issues

import (
	"time"

	"src.sourcegraph.com/apps/issues/issues"
)

func (s state) Comment() issues.Comment {
	return issues.Comment{
		User:      is.CurrentUser(),
		CreatedAt: time.Unix(1443244474, 0).UTC(),
		Body:      "I've resolved this in 4387efb. Please re-open or leave a comment if there's still room for improvement here.",
	}
}

func (s state) Event() issues.Event {
	return issues.Event{
		Actor:     is.CurrentUser(),
		CreatedAt: time.Now(),
		Type:      issues.Closed,
	}
}
