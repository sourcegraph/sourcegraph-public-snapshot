package graphqlbackend

import (
	"strings"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"

	"context"
	"math/rand"
	"sort"
	"sync"
	"testing"
)

func TestExternalServiceCollaborators_parallelRecentCommitters(t *testing.T) {
	ctx := context.Background()

	var (
		callsMu sync.Mutex
		calls   []*github.RecentCommittersParams
	)
	recentCommittersFunc := func(ctx context.Context, params *github.RecentCommittersParams) (*github.RecentCommittersResults, error) {
		callsMu.Lock()
		calls = append(calls, params)
		callsMu.Unlock()

		var results github.RecentCommittersResults
		results.Nodes = append(results.Nodes, struct {
			Authors struct {
				Nodes []struct {
					Date      string
					Email     string
					Name      string
					User      struct{ Login string }
					AvatarURL string
				}
			}
		}{
			Authors: struct {
				Nodes []struct {
					Date      string
					Email     string
					Name      string
					User      struct{ Login string }
					AvatarURL string
				}
			}{
				Nodes: []struct {
					Date      string
					Email     string
					Name      string
					User      struct{ Login string }
					AvatarURL string
				}{
					{Name: params.Name + "-joe"},
					{Name: params.Name + "-jane"},
					{Name: params.Name + "-janet"},
				},
			},
		})

		return &results, nil
	}

	repos := []string{"gorilla/mux", "golang/go", "sourcegraph/sourcegraph"}
	recentCommitters, err := parallelRecentCommitters(ctx, repos, recentCommittersFunc)
	if err != nil {
		t.Fatal(err)
	}

	sort.Slice(calls, func(i, j int) bool {
		return calls[i].Name < calls[j].Name
	})
	sort.Slice(recentCommitters, func(i, j int) bool {
		return recentCommitters[i].name < recentCommitters[j].name
	})

	autogold.Want("calls", []*github.RecentCommittersParams{
		{
			Name:  "go",
			Owner: "golang",
			First: 100,
		},
		{
			Name:  "mux",
			Owner: "gorilla",
			First: 100,
		},
		{
			Name:  "sourcegraph",
			Owner: "sourcegraph",
			First: 100,
		},
	}).Equal(t, calls)

	autogold.Want("recentCommitters", []*invitableCollaboratorResolver{
		{
			name: "go-jane",
		},
		{name: "go-janet"},
		{name: "go-joe"},
		{name: "mux-jane"},
		{name: "mux-janet"},
		{name: "mux-joe"},
		{name: "sourcegraph-jane"},
		{name: "sourcegraph-janet"},
		{name: "sourcegraph-joe"},
	}).Equal(t, recentCommitters)
}

func TestExternalServiceCollaborators_filterInvitableCollaborators(t *testing.T) {
	collaborators := func(emails ...string) []*invitableCollaboratorResolver {
		var v []*invitableCollaboratorResolver
		for _, email := range emails {
			v = append(v, &invitableCollaboratorResolver{email: email})
		}
		return v
	}
	emails := func(values ...string) []*database.UserEmail {
		var v []*database.UserEmail
		for _, email := range values {
			v = append(v, &database.UserEmail{Email: email})
		}
		return v
	}

	tests := []struct {
		want             autogold.Value
		recentCommitters []*invitableCollaboratorResolver
		authUserEmails   []*database.UserEmail
	}{
		{
			recentCommitters: collaborators(),
			authUserEmails:   emails("stephen@sourcegraph.com"),
			want:             autogold.Want("zero committers", []*invitableCollaboratorResolver{}),
		},
		{
			recentCommitters: collaborators("stephen@sourcegraph.com", "sqs@sourcegraph.com", "stephen@sourcegraph.com", "stephen@sourcegraph.com"),
			authUserEmails:   emails(),
			want: autogold.Want("deduplication", []*invitableCollaboratorResolver{
				{
					email: "stephen@sourcegraph.com",
				},
				{email: "sqs@sourcegraph.com"},
			}),
		},
		{
			recentCommitters: collaborators("stephen@sourcegraph.com", "sqs@sourcegraph.com", "stephen@sourcegraph.com", "beyang@sourcegraph.com", "stephen@sourcegraph.com"),
			authUserEmails:   emails("stephen@sourcegraph.com"),
			want: autogold.Want("not ourself", []*invitableCollaboratorResolver{
				{
					email: "sqs@sourcegraph.com",
				},
				{email: "beyang@sourcegraph.com"},
			}),
		},
		{
			recentCommitters: collaborators("noreply@github.com", "noreply.notifications@github.com", "stephen+noreply@sourcegraph.com", "beyang@sourcegraph.com"),
			authUserEmails:   emails(),
			want: autogold.Want("noreply excluded", []*invitableCollaboratorResolver{{
				email: "beyang@sourcegraph.com",
			}}),
		},
		{
			recentCommitters: append(
				collaborators("sqs+sourcegraph-bot@sourcegraph.com", "renovatebot@gmail.com", "stephen@sourcegraph.com"),
				&invitableCollaboratorResolver{email: "campaigns@sourcegraph.com", name: "Sourcegraph Bot"},
			),
			authUserEmails: emails(),
			want: autogold.Want("bots excluded", []*invitableCollaboratorResolver{{
				email: "stephen@sourcegraph.com",
			}}),
		},
		{
			recentCommitters: collaborators("steveexists@github.com", "rando@randi.com", "kimbo@github.com", "stephenexists@sourcegraph.com"),
			authUserEmails:   emails(),
			want: autogold.Want("existing users excluded", []*invitableCollaboratorResolver{
				{
					email: "rando@randi.com",
				},
				{email: "kimbo@github.com"},
			}),
		},
		{
			recentCommitters: collaborators("steve@github.com", "rando@randi.com", "kimbo@github.com", "stephen@sourcegraph.com", "beyang@sourcegraph.com", "sqs@sourcegraph.com"),
			authUserEmails:   emails(),
			want: autogold.Want("same domain first", []*invitableCollaboratorResolver{
				{
					email: "stephen@sourcegraph.com",
				},
				{email: "beyang@sourcegraph.com"},
				{email: "sqs@sourcegraph.com"},
				{email: "steve@github.com"},
				{email: "kimbo@github.com"},
				{email: "rando@randi.com"},
			}),
		},
		{
			recentCommitters: collaborators("steve@gmail.com", "rando@gmail.com", "kimbo@gmail.com", "george@gmail.com", "stephen@sourcegraph.com", "beyang@sourcegraph.com", "sqs@sourcegraph.com"),
			authUserEmails:   emails(),
			want: autogold.Want("popular personal email domains last", []*invitableCollaboratorResolver{
				{
					email: "sqs@sourcegraph.com",
				},
				{email: "stephen@sourcegraph.com"},
				{email: "beyang@sourcegraph.com"},
				{email: "rando@gmail.com"},
				{email: "kimbo@gmail.com"},
				{email: "george@gmail.com"},
				{email: "steve@gmail.com"},
			}),
		},
	}
	for _, tst := range tests {
		t.Run(tst.want.Name(), func(t *testing.T) {
			userExists := func(usernameOrEmail string) bool {
				return strings.Contains(usernameOrEmail, "exists")
			}
			got := filterInvitableCollaborators(tst.recentCommitters, tst.authUserEmails, userExists, userExists)
			tst.want.Equal(t, got)
		})
	}
}

func TestExternalServiceCollaborators_pickReposToScanForCollaborators(t *testing.T) {
	rand.Seed(0)
	tests := []struct {
		possibleRepos  []string
		maxReposToScan int
		want           autogold.Value
	}{
		{
			possibleRepos:  []string{"o", "b", "f", "d", "e", "u", "a", "h", "l", "s", "u", "b", "m"},
			maxReposToScan: 8,
			want:           autogold.Want("three", []string{"f", "a", "b", "u", "l", "o", "u", "s"}),
		},
		{
			possibleRepos:  []string{"c"},
			maxReposToScan: 3,
			want:           autogold.Want("have one", []string{"c"}),
		},
		{
			possibleRepos:  []string{},
			maxReposToScan: 3,
			want:           autogold.Want("have zero", []string{}),
		},
	}
	for _, tst := range tests {
		t.Run(tst.want.Name(), func(t *testing.T) {
			got := pickReposToScanForCollaborators(tst.possibleRepos, tst.maxReposToScan)
			tst.want.Equal(t, got)
		})
	}
}
