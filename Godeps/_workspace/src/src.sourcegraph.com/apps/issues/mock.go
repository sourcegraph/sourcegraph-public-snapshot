package issues

import (
	"time"

	"src.sourcegraph.com/apps/issues/issues"
)

type mockState struct {
	state
}

func (s mockState) Issue() issues.Issue {
	return issues.Issue{
		ID:      123,
		State:   issues.OpenState,
		Title:   "Mock issue.",
		Comment: s.Comment(),
		Replies: 0,
	}
}

func (s mockState) Comment() issues.Comment {
	return issues.Comment{
		User:      mockUser,
		CreatedAt: time.Unix(1443244474, 0).UTC(),
		Body:      "I've resolved this in [`4387efb`](https://www.example.com). Please re-open or leave a comment if there's still room for improvement here.",
	}
}

func (s mockState) Event() issues.Event {
	return issues.Event{
		Actor:     mockUser,
		CreatedAt: time.Now(),
		Type:      issues.Closed,
	}
}

func (s mockState) Reference() issues.Issue {
	return issues.Issue{
		Comment: issues.Comment{
			User:      mockUser,
			CreatedAt: time.Unix(1443244474, 0).UTC(),
		},
		Reference: &issues.Reference{
			Path:      "api/server/server.go",
			Repo:      issues.RepoSpec{URI: "docker/docker"},
			CommitID:  "7a19164c179601898e748f1b45d0c82b949a6433",
			StartLine: 100,
			EndLine:   101,
			Contents: `<span>	</span><span class="kwd">return</span><span> </span><a href="http://localhost:3000/github.com/golang/go/.GoPackage/builtin/.def/nil" class="ref" rel="nofollow"><span class="kwd">nil</span></a>
<span class="pun">}</span>`,
		},
	}
}

var mockUser = issues.User{
	Login:     "shurcooL",
	AvatarURL: "https://avatars.githubusercontent.com/u/1924134?v=3&s=96",
	HTMLURL:   "https://github.com/shurcooL",
}
