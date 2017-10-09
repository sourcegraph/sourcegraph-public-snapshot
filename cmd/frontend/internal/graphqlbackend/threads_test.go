package graphqlbackend

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func TestThreads_Create(t *testing.T) {
	ctx := context.Background()

	wantRepo := sourcegraph.OrgRepo{
		RemoteURI: "test",
	}
	store.Mocks.OrgMembers.MockGetByOrgIDAndUserID_Return(t, &sourcegraph.OrgMember{}, nil)
	store.Mocks.OrgRepos.MockGetByRemoteURI_Return(t, nil, store.ErrRepoNotFound)
	repoCreateCalled, repoCreateCalledWith := store.Mocks.OrgRepos.MockCreate_Return(t, &sourcegraph.OrgRepo{
		ID:        1,
		RemoteURI: wantRepo.RemoteURI,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil)

	store.Mocks.Orgs.MockGetByID_Return(t, &sourcegraph.Org{}, nil)
	threadCreateCalled, _ := store.Mocks.Threads.MockCreate_Return(t, &sourcegraph.Thread{
		ID:        1,
		OrgRepoID: wantRepo.ID,
		File:      "foo.go",
		Revision:  "abcd",
	}, nil)
	commentCreateCalled, _ := store.Mocks.Comments.MockCreate(t)

	r := &schemaResolver{}
	_, err := r.CreateThread(ctx, &struct {
		OrgID          int32
		RemoteURI      string
		File           string
		Revision       string
		Branch         *string
		StartLine      int32
		EndLine        int32
		StartCharacter int32
		EndCharacter   int32
		RangeLength    int32
		Contents       string
	}{
		RemoteURI: wantRepo.RemoteURI,
		File:      "foo.go",
		Revision:  "abcd",
		Contents:  "Hello",
	})

	if err != nil {
		t.Error(err)
	}
	if !*repoCreateCalled || !reflect.DeepEqual(wantRepo, *repoCreateCalledWith) {
		t.Errorf("want LocalRepos.Create call to be %v not %v", wantRepo, repoCreateCalledWith)
	}
	if !*threadCreateCalled {
		t.Error("expected Threads.Create to be called")
	}
	if !*commentCreateCalled {
		t.Error("expected Comments.Create to be called")
	}
}

func TestThreads_Update(t *testing.T) {
	wantRepo := sourcegraph.OrgRepo{
		RemoteURI: "test",
	}

	store.Mocks.Threads.MockGet_Return(t, &sourcegraph.Thread{OrgRepoID: 1}, nil)
	store.Mocks.OrgRepos.MockGetByID_Return(t, &wantRepo, nil)
	called := store.Mocks.Threads.MockUpdate_Return(t, &sourcegraph.Thread{OrgRepoID: 1}, nil)
	store.Mocks.OrgMembers.MockGetByOrgIDAndUserID_Return(t, nil, nil)
	store.Mocks.Orgs.MockGetByID_Return(t, nil, nil)

	r := &schemaResolver{}
	archived := true
	_, err := r.UpdateThread(context.Background(), &struct {
		ThreadID int32
		Archived *bool
	}{
		ThreadID: 1,
		Archived: &archived,
	})

	if err != nil {
		t.Error(err)
	}
	if !*called {
		t.Error("expected Threads.Update to be called")
	}
}

func TestTitleFromContents(t *testing.T) {
	tests := []struct {
		In  string
		Out string
	}{
		{In: "Hello", Out: "Hello"},
		{In: "Hello.", Out: "Hello."},
		{In: "Hello...", Out: "Hello..."},
		{In: "Hello?", Out: "Hello?"},
		{In: "Hello???", Out: "Hello???"},
		{In: "Hello?? Are you there?", Out: "Hello??"},
		{In: "Hello!", Out: "Hello!"},
		{In: "Hello!!!", Out: "Hello!!!"},
		{In: "Hello!?", Out: "Hello!?"},
		{In: "Hello there!", Out: "Hello there!"},
		{In: "Check this out. Weird code huh?", Out: "Check this out."},
		{In: "Hello world\n", Out: "Hello world"},
		{In: "Hello world\nAnd all who inhabit it.", Out: "Hello world"},
		{In: "Hello title\n\nSome contents?", Out: "Hello title"},
		{In: "I have a question about this.\nWhat's going on here", Out: "I have a question about this."},
		{In: "Hello title\n\nSome contents?", Out: "Hello title"},
		{In: "What does foo.bar do?", Out: "What does foo.bar do?"},
		{In: "It should be 1 != 2\nFYI 1 != 1 is wrong.", Out: "It should be 1 != 2"},
		{In: "This\nis\na\nweird\ncomment. With two sentences.", Out: "This"},
		{In: strings.Repeat("a", 141), Out: strings.Repeat("a", 137) + "..."},
	}

	for _, test := range tests {
		out := titleFromContents(test.In)
		if out != test.Out {
			t.Errorf("\n   input: \"%s\"\nexpected: \"%s\"\n     got: \"%s\"", test.In, test.Out, out)
		}
		// Adding trailing whitespace should not change the title
		outTrailingSpace := titleFromContents(test.In + " ")
		if outTrailingSpace != test.Out {
			t.Errorf("\n   input: \"%s\"\nexpected: \"%s\"\n     got: \"%s\"", test.In, test.Out, outTrailingSpace)
		}
		// Adding trailing newline should not change the title
		outTrailingNewline := titleFromContents(test.In + "\n")
		if outTrailingNewline != test.Out {
			t.Errorf("\n   input: \"%s\"\nexpected: \"%s\"\n     got: \"%s\"", test.In, test.Out, outTrailingNewline)
		}
	}
}
