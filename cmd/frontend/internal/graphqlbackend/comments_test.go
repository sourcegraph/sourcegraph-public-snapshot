package graphqlbackend

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func TestComments_appendUniqueEmailsFromMentions(t *testing.T) {
	tests := []struct {
		Input  string
		Output []string
	}{
		{Input: "@alice", Output: []string{"alice"}},
		{Input: "@alice-w", Output: []string{"alice-w"}},
		{Input: "@alice_w", Output: []string{"alice"}},
		{Input: "@org", Output: []string{"org"}},
		{Input: "hello @alice", Output: []string{"alice"}},
		{Input: "hello @alice look at this", Output: []string{"alice"}},
		{Input: "hello @alice. hello @bob.", Output: []string{"alice", "bob"}},
		{Input: "@alice. @bob?", Output: []string{"alice", "bob"}},
		{Input: "@alice@bob", Output: []string{"alice"}},
		{Input: "@alice@bob@nick", Output: []string{"alice"}},
		{Input: "hello.@alice.@bob!@nick?", Output: []string{"alice", "bob", "nick"}},

		{Input: "@", Output: nil},
		{Input: "@@", Output: nil},
		{Input: "@", Output: nil},
		{Input: "@_al", Output: nil},
		{Input: "@-@", Output: nil},
		{Input: "hello@alice", Output: nil},
		{Input: "renfred@sourcegraph.com", Output: nil},
	}

	for _, test := range tests {
		out := usernamesFromMentions(test.Input)
		if !reflect.DeepEqual(out, test.Output) {
			t.Errorf("expected %s for input \"%s\", got: %v", test.Output, test.Input, out)
		}
	}
}

func TestComments_emailsToNotify(t *testing.T) {
	testOrg := sourcegraph.Org{
		ID:   42,
		Name: "sgtest",
	}

	// Mock users
	nick := sourcegraph.User{
		Username: "nick",
		AuthID:   "auth0|1",
		Email:    "nick@sourcegraph.com",
	}
	renfred := sourcegraph.User{
		Username: "renfred",
		AuthID:   "auth0|2",
		Email:    "renfred@sourcegraph.com",
	}
	john := sourcegraph.User{
		Username: "john",
		AuthID:   "auth0|3",
		Email:    "john@sourcegraph.com",
	}
	sqs := sourcegraph.User{
		Username: "sqs",
		AuthID:   "auth0|4",
		Email:    "sqs@sourcegraph.com",
	}
	kingy := sourcegraph.User{
		Username: "kingy",
		AuthID:   "auth0|5",
		Email:    "kingy@sourcegraph.com",
	}
	testUsers := []sourcegraph.User{nick, renfred, sqs, john, kingy}

	// Mock Users.ListByOrg
	store.Mocks.Users.ListByOrg = func(ctx context.Context, orgID int32, auth0IDs, usernames []string) ([]*sourcegraph.User, error) {
		if orgID != testOrg.ID {
			return nil, fmt.Errorf(`expected to be called with testOrg ID "%d", got "%d"`, testOrg.ID, orgID)
		}
		var users []*sourcegraph.User
		for _, id := range auth0IDs {
			u, ok := map[string]sourcegraph.User{
				nick.AuthID:    nick,
				renfred.AuthID: renfred,
				sqs.AuthID:     sqs,
				john.AuthID:    john,
				kingy.AuthID:   kingy,
			}[id]
			if ok {
				users = append(users, &u)
			}
		}
		for _, n := range usernames {
			u, ok := map[string]sourcegraph.User{
				nick.Username:    nick,
				renfred.Username: renfred,
				sqs.Username:     sqs,
				john.Username:    john,
				kingy.Username:   kingy,
			}[n]
			if ok {
				users = append(users, &u)
			}
		}
		return users, nil
	}

	// Mock allOrgEmails
	mockAllEmailsForOrg = func(ctx context.Context, orgID int32, excludeByUserID []string) ([]string, error) {
		exclude := map[string]struct{}{}
		for _, id := range excludeByUserID {
			exclude[id] = struct{}{}
		}
		var emails []string
		for _, u := range testUsers {
			if _, ok := exclude[u.AuthID]; !ok {
				emails = append(emails, u.Email)
			}
		}
		return emails, nil
	}

	// Mock comments
	one := &sourcegraph.Comment{
		Contents:     "Yo @renfred",
		AuthorUserID: nick.AuthID,
		AuthorEmail:  nick.Email,
	}
	two := &sourcegraph.Comment{
		Contents:     "Did you see this comment?",
		AuthorUserID: nick.AuthID,
		AuthorEmail:  nick.Email,
	}
	three := &sourcegraph.Comment{
		Contents:     "Going to mention myself to test notifications @nick",
		AuthorUserID: nick.AuthID,
		AuthorEmail:  nick.Email,
	}
	four := &sourcegraph.Comment{
		Contents:     "Dude, I am on vacation. Ask @sqs or @John",
		AuthorUserID: renfred.AuthID,
		AuthorEmail:  renfred.Email,
	}
	five := &sourcegraph.Comment{
		Contents:     "Stop bothering Renfred!",
		AuthorUserID: sqs.AuthID,
		AuthorEmail:  sqs.Email,
	}
	six := &sourcegraph.Comment{
		Contents:     "Maybe @linus could take a look?",
		AuthorUserID: nick.AuthID,
		AuthorEmail:  nick.Email,
	}
	seven := &sourcegraph.Comment{
		Contents:     "Feels like yelling into @the-void",
		AuthorUserID: nick.AuthID,
		AuthorEmail:  nick.Email,
	}
	eight := &sourcegraph.Comment{
		Contents:     "Nevermind. Just going to ask the whole @org",
		AuthorUserID: nick.AuthID,
		AuthorEmail:  nick.Email,
	}

	tests := []struct {
		previousComments []*sourcegraph.Comment
		newComment       *sourcegraph.Comment
		author           sourcegraph.User
		expected         []string
	}{
		{
			[]*sourcegraph.Comment{},
			one,
			nick,
			[]string{renfred.Email},
		},
		{
			[]*sourcegraph.Comment{one},
			two,
			nick,
			[]string{renfred.Email},
		},
		{
			[]*sourcegraph.Comment{one, two},
			three,
			nick,
			[]string{renfred.Email, nick.Email},
		},
		{
			[]*sourcegraph.Comment{one, two, three},
			four,
			renfred,
			[]string{nick.Email, sqs.Email, john.Email},
		},
		{
			[]*sourcegraph.Comment{one, two, three, four},
			five,
			sqs,
			[]string{nick.Email, renfred.Email, john.Email},
		},
		{
			[]*sourcegraph.Comment{one, two, three, four, five},
			six,
			nick,
			[]string{renfred.Email, sqs.Email, john.Email},
		},
		{
			[]*sourcegraph.Comment{one, two, three, four, five, six},
			seven,
			nick,
			[]string{renfred.Email, sqs.Email, john.Email},
		},
		{
			[]*sourcegraph.Comment{one, two, three, four, five, six, seven, eight},
			eight,
			nick,
			[]string{renfred.Email, sqs.Email, john.Email, kingy.Email},
		},
	}

	for _, test := range tests {
		actual, err := emailsToNotify(context.Background(), append(test.previousComments, test.newComment), test.author, testOrg)
		if err != nil {
			t.Errorf("emailsToNotify error: %s", err)
		}
		if !reflect.DeepEqual(actual, test.expected) {
			t.Fatalf("emailsToNotify(%+v, %+v) expected %#v; got %#v", test.previousComments, test.newComment, test.expected, actual)
		}
	}
}

func TestComments_Create(t *testing.T) {
	t.Skip("broken due to test using the DB")

	ctx := context.Background()

	repo := sourcegraph.OrgRepo{
		ID:                1,
		CanonicalRemoteID: "github.com/foo/bar",
	}
	thread := sourcegraph.Thread{
		ID:               1,
		OrgRepoID:        1,
		RepoRevisionPath: "foo.go",
		RepoRevision:     "1234",
		LinesRevision:    "5678",
		StartLine:        1,
		EndLine:          2,
		StartCharacter:   3,
		EndCharacter:     4,
	}
	wantComment := sourcegraph.Comment{
		ThreadID:    1,
		Contents:    "Hello",
		AuthorName:  "Alice",
		AuthorEmail: "alice@acme.com",
	}

	store.Mocks.OrgRepos.MockGetByID_Return(t, &repo, nil)
	store.Mocks.Threads.MockGet_Return(t, &thread, nil)
	called, calledWith := store.Mocks.Comments.MockCreate(t)

	r := &schemaResolver{}
	_, err := r.AddCommentToThread(ctx, &struct {
		ThreadID int32
		Contents string
	}{
		ThreadID: thread.ID,
		Contents: wantComment.Contents,
	})

	if err != nil {
		t.Error(err)
	}
	if !*called || !reflect.DeepEqual(wantComment, *calledWith) {
		t.Errorf("want Comments.Create call to be %v not %v", wantComment, *calledWith)
	}
}

func TestComments_CreateAccessDenied(t *testing.T) {
	ctx := context.Background()

	store.Mocks.Threads.MockGet_Return(t, &sourcegraph.Thread{OrgRepoID: 1}, nil)
	store.Mocks.OrgRepos.MockGetByID_Return(t, nil, store.ErrRepoNotFound)
	called, calledWith := store.Mocks.Comments.MockCreate(t)

	r := &schemaResolver{}
	comment, err := r.AddCommentToThread(ctx, &struct {
		ThreadID int32
		Contents string
	}{})

	if *called {
		t.Errorf("should not have called Comments.Create (called with %v)", calledWith)
	}
	if comment != nil || err == nil {
		t.Error("did not return error for failed Comments.Create")
	}
}

func TestRepoNameFromRemoteID(t *testing.T) {
	tests := []struct {
		In  string
		Out string
	}{
		{In: "github.com/gorilla/mux", Out: "gorilla/mux"},
		{In: "git.acmeinternal.org/acme/acme", Out: "acme/acme"},
		{In: "company.internal/project", Out: "project"},
	}

	for _, test := range tests {
		out := repoNameFromRemoteID(test.In)
		if out != test.Out {
			t.Errorf("\n   input: \"%s\"\nexpected: \"%s\"\n     got: \"%s\"", test.In, test.Out, out)
		}
	}
}
