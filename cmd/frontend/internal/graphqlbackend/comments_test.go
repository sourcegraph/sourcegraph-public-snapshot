package graphqlbackend

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/txemail"
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
	testOrg := types.Org{
		ID:   42,
		Name: "sgtest",
	}

	// Mock users
	nick := types.User{
		ID:       1,
		Username: "nick",
	}
	renfred := types.User{
		ID:       2,
		Username: "renfred",
	}
	john := types.User{
		ID:       3,
		Username: "john",
	}
	sqs := types.User{
		ID:       4,
		Username: "sqs",
	}
	kingy := types.User{
		ID:       5,
		Username: "kingy",
	}
	testUsers := []types.User{nick, renfred, sqs, john, kingy}

	db.Mocks.Users.GetByID = func(ctx context.Context, id int32) (*types.User, error) {
		for _, u := range testUsers {
			if u.ID == id {
				return &u, nil
			}
		}
		return nil, fmt.Errorf("user with ID %d not found", id)
	}

	db.Mocks.UserEmails.GetEmail = func(ctx context.Context, id int32) (string, bool, error) {
		for _, u := range testUsers {
			if u.ID == id {
				return u.Username + "@sourcegraph.com", true, nil
			}
		}
		return "", false, fmt.Errorf("user with ID %d not found", id)
	}

	// Mock Users.ListByOrg
	db.Mocks.Users.ListByOrg = func(ctx context.Context, orgID int32, userIDs []int32, usernames []string) ([]*types.User, error) {
		if orgID != testOrg.ID {
			return nil, fmt.Errorf(`expected to be called with testOrg ID "%d", got "%d"`, testOrg.ID, orgID)
		}
		var users []*types.User
		for _, id := range userIDs {
			u, ok := map[int32]types.User{
				nick.ID:    nick,
				renfred.ID: renfred,
				sqs.ID:     sqs,
				john.ID:    john,
				kingy.ID:   kingy,
			}[id]
			if ok {
				users = append(users, &u)
			}
		}
		for _, n := range usernames {
			u, ok := map[string]types.User{
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
	mockAllEmailsForOrg = func(ctx context.Context, orgID int32, excludeByUserID []int32) ([]string, error) {
		exclude := map[int32]struct{}{}
		for _, id := range excludeByUserID {
			exclude[id] = struct{}{}
		}
		var emails []string
		for _, u := range testUsers {
			if _, ok := exclude[u.ID]; !ok {
				emails = append(emails, u.Username+"@sourcegraph.com")
			}
		}
		return emails, nil
	}

	// Mock comments
	one := &types.Comment{
		Contents:     "Yo @renfred",
		AuthorUserID: nick.ID,
	}
	two := &types.Comment{
		Contents:     "Did you see this comment?",
		AuthorUserID: nick.ID,
	}
	three := &types.Comment{
		Contents:     "Going to mention myself to test notifications @nick",
		AuthorUserID: nick.ID,
	}
	four := &types.Comment{
		Contents:     "Dude, I am on vacation. Ask @sqs or @John",
		AuthorUserID: renfred.ID,
	}
	five := &types.Comment{
		Contents:     "Stop bothering Renfred!",
		AuthorUserID: sqs.ID,
	}
	six := &types.Comment{
		Contents:     "Maybe @linus could take a look?",
		AuthorUserID: nick.ID,
	}
	seven := &types.Comment{
		Contents:     "Feels like yelling into @the-void",
		AuthorUserID: nick.ID,
	}
	eight := &types.Comment{
		Contents:     "Nevermind. Just going to ask the whole @org",
		AuthorUserID: nick.ID,
	}

	tests := []struct {
		previousComments []*types.Comment
		newComment       *types.Comment
		author           types.User
		expected         []string
	}{
		{
			[]*types.Comment{},
			one,
			nick,
			[]string{renfred.Username + "@sourcegraph.com"},
		},
		{
			[]*types.Comment{one},
			two,
			nick,
			[]string{renfred.Username + "@sourcegraph.com"},
		},
		{
			[]*types.Comment{one, two},
			three,
			nick,
			[]string{renfred.Username + "@sourcegraph.com", nick.Username + "@sourcegraph.com"},
		},
		{
			[]*types.Comment{one, two, three},
			four,
			renfred,
			[]string{nick.Username + "@sourcegraph.com", sqs.Username + "@sourcegraph.com", john.Username + "@sourcegraph.com"},
		},
		{
			[]*types.Comment{one, two, three, four},
			five,
			sqs,
			[]string{nick.Username + "@sourcegraph.com", renfred.Username + "@sourcegraph.com", john.Username + "@sourcegraph.com"},
		},
		{
			[]*types.Comment{one, two, three, four, five},
			six,
			nick,
			[]string{renfred.Username + "@sourcegraph.com", sqs.Username + "@sourcegraph.com", john.Username + "@sourcegraph.com"},
		},
		{
			[]*types.Comment{one, two, three, four, five, six},
			seven,
			nick,
			[]string{renfred.Username + "@sourcegraph.com", sqs.Username + "@sourcegraph.com", john.Username + "@sourcegraph.com"},
		},
		{
			[]*types.Comment{one, two, three, four, five, six, seven, eight},
			eight,
			nick,
			[]string{renfred.Username + "@sourcegraph.com", sqs.Username + "@sourcegraph.com", john.Username + "@sourcegraph.com", kingy.Username + "@sourcegraph.com"},
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
	ctx := context.Background()

	repo := types.OrgRepo{
		ID:                1,
		CanonicalRemoteID: "github.com/foo/bar",
	}
	thread := types.Thread{
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
	wantComment := types.Comment{
		ThreadID:     1,
		Contents:     "Hello",
		AuthorUserID: 1,
	}

	db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
		return &types.User{ID: 1, DisplayName: "Alice"}, nil
	}
	db.Mocks.OrgMembers.MockGetByOrgIDAndUserID_Return(t, &types.OrgMembership{}, nil)
	mockEmailsToNotify = func(ctx context.Context, comments []*types.Comment, author types.User, org types.Org) ([]string, error) {
		return []string{"a@example.com"}, nil
	}
	defer func() { mockEmailsToNotify = nil }()
	db.Mocks.UserEmails.GetEmail = func(ctx context.Context, id int32) (string, bool, error) { return "b@example.com", true, nil }
	db.Mocks.Orgs.MockGetByID_Return(t, &types.Org{}, nil)
	db.Mocks.Comments.GetAllForThread = func(context.Context, int32) ([]*types.Comment, error) { return nil, nil }
	db.Mocks.OrgRepos.MockGetByID_Return(t, &repo, nil)
	db.Mocks.Threads.MockGet_Return(t, &thread, nil)
	db.Mocks.Settings.GetLatest = func(context.Context, api.ConfigurationSubject) (*api.Settings, error) {
		return &api.Settings{}, nil
	}
	called, calledWith := db.Mocks.Comments.MockCreate(t)
	txemail.MockSend = func(context.Context, txemail.Message) error { return nil }

	r := &schemaResolver{}
	_, err := r.AddCommentToThread(ctx, &struct {
		ThreadID threadID
		Contents string
	}{
		ThreadID: threadID{int32Value: thread.ID},
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

	db.Mocks.Threads.MockGet_Return(t, &types.Thread{OrgRepoID: 1}, nil)
	db.Mocks.OrgRepos.MockGetByID_Return(t, nil, &errcode.Mock{Message: "repo not found", IsNotFound: true})
	called, calledWith := db.Mocks.Comments.MockCreate(t)

	r := &schemaResolver{}
	comment, err := r.AddCommentToThread(ctx, &struct {
		ThreadID threadID
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
		In  api.RepoURI
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

func TestSendNewCommentEmails(t *testing.T) {
	var mockSent []txemail.Message
	txemail.MockSend = func(ctx context.Context, message txemail.Message) error {
		mockSent = append(mockSent, message)
		return nil
	}
	defer func() { txemail.MockSend = nil }()

	url, _ := url.Parse("http://example.com")

	sendNewCommentEmails(
		context.Background(),
		types.OrgRepo{CanonicalRemoteID: "r"},
		types.Comment{Contents: "foo'bar<b>baz</b>**qux**"},
		types.Thread{
			ID:               123,
			RepoRevisionPath: "f",
			Branch:           strptr("b"),
			StartLine:        10,
			EndLine:          11,
			Lines: &types.ThreadLines{
				HTMLBefore: "h0",
				HTML:       "h1",
				HTMLAfter:  "h2",
				TextBefore: "t0",
				Text:       "t",
				TextAfter:  "t1",
			},
		},
		[]*types.Comment{},
		[]string{"a@a.com"},
		types.User{},
		url,
	)

	if want := ([]txemail.Message{
		{
			FromName: "",
			To:       []string{"a@a.com"},
			Template: newCommentEmailTemplates,
			Data: struct {
				threadEmailTemplateCommonData
				Location     string
				ContextLines string
				Contents     string
			}{
				threadEmailTemplateCommonData: threadEmailTemplateCommonData{
					Reply:    false,
					RepoName: "r",
					Branch:   "@b",
					Title:    "foo'bar<b>baz</b>**qux**",
					Number:   123,
					URL:      url.String(),
				},
				Location:     "r/f:L10",
				ContextLines: "t0\nt",
				Contents:     "foo'bar<b>baz</b>**qux**",
			},
		},
	}); !reflect.DeepEqual(mockSent, want) {
		t.Errorf("got  %+v\n\nwant %+v", mockSent, want)
	}
	if len(mockSent) == 0 {
		t.Fatal()
	}

	// Check rendered message.
	rendered, err := txemail.Render(mockSent[0])
	if err != nil {
		t.Fatal(err)
	}
	if want := `
<pre style="color:#555">t0
t</pre>


<p>foo&#39;bar<b>baz</b><strong>qux</strong></p>


<p>View discussion on Sourcegraph: <a href="http://example.com">r/f:L10</a></p>`; rendered.HTMLBody != want {
		t.Errorf("got  %q\nwant %q", rendered.HTMLBody, want)
	}
	if want := `r/f:L10


------------------------------------------------------------------------------
t0
t
------------------------------------------------------------------------------


foo'barbaz**qux**

View discussion on Sourcegraph: http://example.com`; rendered.Body != want {
		t.Errorf("got  %q\nwant %q", rendered.Body, want)
	}
}

func strptr(s string) *string { return &s }
