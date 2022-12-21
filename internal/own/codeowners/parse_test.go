package codeowners_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"

	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/proto"
)

func TestParseGithubExample(t *testing.T) {
	got, err := codeowners.Parse(
		`# This is a comment.
		# Each line is a file pattern followed by one or more owners.

		# These owners will be the default owners for everything in
		# the repo. Unless a later match takes precedence,
		# @global-owner1 and @global-owner2 will be requested for
		# review when someone opens a pull request.
		*       @global-owner1 @global-owner2

		# Order is important; the last matching pattern takes the most
		# precedence. When someone opens a pull request that only
		# modifies JS files, only @js-owner and not the global
		# owner(s) will be requested for a review.
		*.js    @js-owner #This is an inline comment.

		# You can also use email addresses if you prefer. They'll be
		# used to look up users just like we do for commit author
		# emails.
		*.go docs@example.com

		# Teams can be specified as code owners as well. Teams should
		# be identified in the format @org/team-name. Teams must have
		# explicit write access to the repository. In this example,
		# the octocats team in the octo-org organization owns all .txt files.
		*.txt @octo-org/octocats

		# In this example, @doctocat owns any files in the build/logs
		# directory at the root of the repository and any of its
		# subdirectories.
		/build/logs/ @doctocat

		# The docs/* pattern will match files like
		# docs/getting-started.md but not further nested files like
		# docs/build-app/troubleshooting.md.
		docs/*  docs@example.com

		# In this example, @octocat owns any file in an apps directory
		# anywhere in your repository.
		apps/ @octocat

		# In this example, @doctocat owns any file in the /docs
		# directory in the root of your repository and any of its
		# subdirectories.
		/docs/ @doctocat

		# In this example, any change inside the /scripts directory
		# will require approval from @doctocat or @octocat.
		/scripts/ @doctocat @octocat

		# In this example, @octocat owns any file in the /apps
		# directory in the root of your repository except for the /apps/github
		# subdirectory, as its owners are left empty.
		/apps/ @octocat
		/apps/github`)
	require.NoError(t, err)
	want := []*codeownerspb.Rule{
		{
			Pattern: "*",
			Owner: []*codeownerspb.Owner{
				{Handle: "global-owner1"},
				{Handle: "global-owner2"},
			},
		},
		{
			Pattern: "*.js",
			Owner: []*codeownerspb.Owner{
				{Handle: "js-owner"},
			},
		},
		{
			Pattern: "*.go",
			Owner: []*codeownerspb.Owner{
				{Email: "docs@example.com"},
			},
		},
		{
			Pattern: "*.txt",
			Owner: []*codeownerspb.Owner{
				{Handle: "octo-org/octocats"},
			},
		},
		{
			Pattern: "/build/logs/",
			Owner: []*codeownerspb.Owner{
				{Handle: "doctocat"},
			},
		},
		{
			Pattern: "docs/*",
			Owner: []*codeownerspb.Owner{
				{Email: "docs@example.com"},
			},
		},
		{
			Pattern: "apps/",
			Owner: []*codeownerspb.Owner{
				{Handle: "octocat"},
			},
		},
		{
			Pattern: "/docs/",
			Owner: []*codeownerspb.Owner{
				{Handle: "doctocat"},
			},
		},
		{
			Pattern: "/scripts/",
			Owner: []*codeownerspb.Owner{
				{Handle: "doctocat"},
				{Handle: "octocat"},
			},
		},
		{
			Pattern: "/apps/",
			Owner: []*codeownerspb.Owner{
				{Handle: "octocat"},
			},
		},
		{
			Pattern: "/apps/github",
			Owner:   nil,
		},
	}
	assert.Equal(t, &codeownerspb.File{Rule: want}, got)
}

func TestParseAtHandle(t *testing.T) {
	got, err := codeowners.Parse("README.md @readme-team")
	require.NoError(t, err)
	want := []*codeownerspb.Rule{{
		Pattern: "README.md",
		Owner: []*codeownerspb.Owner{
			{Handle: "readme-team"},
		},
	}}
	assert.Equal(t, &codeownerspb.File{Rule: want}, got)
}

func TestParseAtHandleSupportsNesting(t *testing.T) {
	got, err := codeowners.Parse("README.md @readme-team/readme-subteam")
	require.NoError(t, err)
	want := []*codeownerspb.Rule{{
		Pattern: "README.md",
		Owner: []*codeownerspb.Owner{
			{Handle: "readme-team/readme-subteam"},
		},
	}}
	assert.Equal(t, &codeownerspb.File{Rule: want}, got)
}

func TestParseEmailHandle(t *testing.T) {
	got, err := codeowners.Parse("README.md me@example.com")
	require.NoError(t, err)
	want := []*codeownerspb.Rule{{
		Pattern: "README.md",
		Owner: []*codeownerspb.Owner{
			{Email: "me@example.com"},
		},
	}}
	assert.Equal(t, &codeownerspb.File{Rule: want}, got)
}

func TestParseTwoHandles(t *testing.T) {
	got, err := codeowners.Parse("README.md @readme-team me@example.com")
	require.NoError(t, err)
	want := []*codeownerspb.Rule{{
		Pattern: "README.md",
		Owner: []*codeownerspb.Owner{
			{Handle: "readme-team"},
			{Email: "me@example.com"},
		},
	}}
	assert.Equal(t, &codeownerspb.File{Rule: want}, got)
}
