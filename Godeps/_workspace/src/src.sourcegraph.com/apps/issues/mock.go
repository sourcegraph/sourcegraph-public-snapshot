package issues

import (
	"time"

	"src.sourcegraph.com/apps/issues/issues"
)

func (s state) Comment() issues.Comment {
	return issues.Comment{
		User:      mockUser(),
		CreatedAt: time.Unix(1443244474, 0).UTC(),
		Body:      "I've resolved this in 4387efb. Please re-open or leave a comment if there's still room for improvement here.",
	}
}

func (s state) Event() issues.Event {
	return issues.Event{
		Actor:     mockUser(),
		CreatedAt: time.Now(),
		Type:      issues.Closed,
	}
}

func mockUser() issues.User {
	return issues.User{
		Login:     "shurcooL",
		AvatarURL: "https://avatars.githubusercontent.com/u/1924134?v=3&s=96",
		HTMLURL:   "https://github.com/shurcooL",
	}
}
