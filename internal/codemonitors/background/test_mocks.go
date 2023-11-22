package background

import (
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var externalURLMock, _ = url.Parse("https://www.sourcegraph.com")

var diffResultMock = result.CommitMatch{
	Commit: gitdomain.Commit{
		ID:      api.CommitID("7815187511872asbasdfgasd"),
		Parents: []api.CommitID{},
	},
	Repo: types.MinimalRepo{
		Name: api.RepoName("github.com/test/test"),
	},
	DiffPreview: &result.MatchedString{
		Content: "file1.go file2.go\n@@ -97,5 +97,5 @@ func Test() {\n leading context\n+matched added\n-matched removed\n trailing context\n",
		MatchedRanges: result.Ranges{{
			Start: result.Location{Line: 3, Offset: 66, Column: 1},
			End:   result.Location{Line: 3, Offset: 73, Column: 8},
		}, {
			Start: result.Location{Line: 4, Offset: 91, Column: 1},
			End:   result.Location{Line: 4, Offset: 98, Column: 8},
		}},
	}}

var diffDisplayResultMock = toDisplayResult(&diffResultMock, externalURLMock)

var commitResultMock = result.CommitMatch{
	Commit: gitdomain.Commit{
		ID:      api.CommitID("7815187511872asbasdfgasd"),
		Parents: []api.CommitID{},
	},
	Repo: types.MinimalRepo{
		Name: api.RepoName("github.com/test/test"),
	},
	MessagePreview: &result.MatchedString{
		Content: "summary line\n\nvery\nlong\nmessage\nbody\nwith\nmore\nthan\nten\nlines\nthat\nwill\nbe\ntruncated\n",
		MatchedRanges: result.Ranges{{
			Start: result.Location{Line: 2, Offset: 15, Column: 0},
			End:   result.Location{Line: 2, Offset: 19, Column: 4},
		}},
	},
}

var commitDisplayResultMock = toDisplayResult(&commitResultMock, externalURLMock)

var longCommitResultMock = result.CommitMatch{
	Commit: gitdomain.Commit{
		ID:      api.CommitID("9cb4a43a052f8178566"),
		Parents: []api.CommitID{},
	},
	Repo: types.MinimalRepo{
		Name: api.RepoName("github.com/test/test"),
	},
	MessagePreview: &result.MatchedString{
		Content: "summary line\nnext line\n longlonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglong\n longlonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglong\n longlonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglong\n longlonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglong\n longlonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglong",
		MatchedRanges: result.Ranges{{
			Start: result.Location{Line: 2, Offset: 15, Column: 0},
			End:   result.Location{Line: 2, Offset: 19, Column: 4},
		}},
	},
}
