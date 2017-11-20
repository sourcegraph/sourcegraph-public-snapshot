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
		CanonicalRemoteID: "test",
		CloneURL:          "https://test.com/test",
	}
	store.Mocks.OrgMembers.MockGetByOrgIDAndUserID_Return(t, &sourcegraph.OrgMember{}, nil)
	store.Mocks.Users.MockGetByAuth0ID_Return(t, &sourcegraph.User{}, nil)
	store.Mocks.OrgRepos.MockGetByCanonicalRemoteID_Return(t, nil, store.ErrRepoNotFound)
	repoCreateCalled, repoCreateCalledWith := store.Mocks.OrgRepos.MockCreate_Return(t, &sourcegraph.OrgRepo{
		ID:                1,
		CanonicalRemoteID: wantRepo.CanonicalRemoteID,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}, nil)

	store.Mocks.Orgs.MockGetByID_Return(t, &sourcegraph.Org{}, nil)
	repoRev, lineRev := "abcd", "dcba"
	threadCreateCalled, _ := store.Mocks.Threads.MockCreate_Return(t, &sourcegraph.Thread{
		ID:            1,
		OrgRepoID:     wantRepo.ID,
		File:          "foo.go",
		RepoRevision:  repoRev,
		LinesRevision: lineRev,
	}, nil)
	commentCreateCalled, _ := store.Mocks.Comments.MockCreate(t)

	r := &schemaResolver{}
	_, err := r.CreateThread(ctx, &struct {
		OrgID             orgInt32OrID
		CanonicalRemoteID string
		CloneURL          string
		File              string
		RepoRevision      string
		LinesRevision     string
		Branch            *string
		StartLine         int32
		EndLine           int32
		StartCharacter    int32
		EndCharacter      int32
		RangeLength       int32
		Contents          string
		Lines             *threadLines
	}{
		CanonicalRemoteID: wantRepo.CanonicalRemoteID,
		CloneURL:          wantRepo.CloneURL,
		File:              "foo.go",
		RepoRevision:      repoRev,
		LinesRevision:     lineRev,
		Contents:          "Hello",
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
		CanonicalRemoteID: "test",
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
		out := TitleFromContents(test.In)
		if out != test.Out {
			t.Errorf("\n   input: \"%s\"\nexpected: \"%s\"\n     got: \"%s\"", test.In, test.Out, out)
		}
		// Adding trailing whitespace should not change the title
		outTrailingSpace := TitleFromContents(test.In + " ")
		if outTrailingSpace != test.Out {
			t.Errorf("\n   input: \"%s\"\nexpected: \"%s\"\n     got: \"%s\"", test.In, test.Out, outTrailingSpace)
		}
		// Adding trailing newline should not change the title
		outTrailingNewline := TitleFromContents(test.In + "\n")
		if outTrailingNewline != test.Out {
			t.Errorf("\n   input: \"%s\"\nexpected: \"%s\"\n     got: \"%s\"", test.In, test.Out, outTrailingNewline)
		}
	}
}

func TestSanitize(t *testing.T) {
	for _, test := range []struct {
		name, input, want string
	}{
		{
			// Output format of VS Code, i.e. we expect this to pass through unscathed.
			name: "good_vscode_copy_with_syntax_highlighting",
			input: `<div><span style="color: #2b8a3e;">// sharedItems provides access to the 'shared_items' table.</span></div>
				<div><span style="color: #2b8a3e;">//</span></div><div><span style="color: #2b8a3e;">// For a detailed overview of the schema, see schema.md.</span></div>
				<div><span style="color: #329af0;">type</span> <span style="color: #4ec9b0;">sharedItems</span> <span style="color: #329af0;">struct</span>{}</div><br>
				<div><span style="color: #329af0;">func</span> (s <span style="color: #d4d4d4;">*</span>sharedItems) <span style="color: #fff3bf;">Create</span>(ctx context.Context, item <span style="color: #d4d4d4;">*</span>sourcegraph.SharedItem) (<span style="color: #329af0;">string</span>, <span style="color: #329af0;">error</span>) {</div>
				<div>&nbsp;&nbsp;&nbsp;&nbsp;<span style="color: #c586c0;">if</span> item.ULID <span style="color: #d4d4d4;">!=</span> <span style="color: #ffa8a8;">""</span> {</div>
				<div>&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;<span style="color: #c586c0;">return</span> <span style="color: #ffa8a8;">""</span>, errors.<span style="color: #fff3bf;">New</span>(<span style="color: #ffa8a8;">"SharedItems.Create: cannot specify ULID when creating shared item"</span>)</div>
				<div>&nbsp;&nbsp;&nbsp;&nbsp;}</div>
				<div></div>`,
			want: `<div><span style="color: #2b8a3e;">// sharedItems provides access to the &#39;shared_items&#39; table.</span></div>
				<div><span style="color: #2b8a3e;">//</span></div><div><span style="color: #2b8a3e;">// For a detailed overview of the schema, see schema.md.</span></div>
				<div><span style="color: #329af0;">type</span> <span style="color: #4ec9b0;">sharedItems</span> <span style="color: #329af0;">struct</span>{}</div><br>
				<div><span style="color: #329af0;">func</span> (s <span style="color: #d4d4d4;">*</span>sharedItems) <span style="color: #fff3bf;">Create</span>(ctx context.Context, item <span style="color: #d4d4d4;">*</span>sourcegraph.SharedItem) (<span style="color: #329af0;">string</span>, <span style="color: #329af0;">error</span>) {</div>
				<div>&nbsp;&nbsp;&nbsp;&nbsp;<span style="color: #c586c0;">if</span> item.ULID <span style="color: #d4d4d4;">!=</span> <span style="color: #ffa8a8;">&#34;&#34;</span> {</div>
				<div>&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;<span style="color: #c586c0;">return</span> <span style="color: #ffa8a8;">&#34;&#34;</span>, errors.<span style="color: #fff3bf;">New</span>(<span style="color: #ffa8a8;">&#34;SharedItems.Create: cannot specify ULID when creating shared item&#34;</span>)</div>
				<div>&nbsp;&nbsp;&nbsp;&nbsp;}</div>
				<div></div>`,
		}, {
			name:  "bad_script_tags",
			input: `Hello <script>alert("world!")</script>`,
			want:  `Hello `,
		}, {
			name:  "bad_div_style",
			input: `<div style="color: #329af0;">Hello</div>`,
			want:  `<div>Hello</div>`,
		}, {
			name:  "good_span_style",
			input: `<span style="color: #329af0;">Hello</span>`,
			want:  `<span style="color: #329af0;">Hello</span>`,
		}, {
			name:  "bad_span_style_3",
			input: `<span style="color: #329;">Hello</span>`,
			want:  `<span>Hello</span>`,
		}, {
			name:  "bad_span_style_suffix",
			input: `<span style="color: #329af0; pwnd: yes;">Hello</span>`,
			want:  `<span>Hello</span>`,
		}, {
			name:  "bad_span_style_prefix",
			input: `<span style="pwnd: yes; color: #329af0;">Hello</span>`,
			want:  `<span>Hello</span>`,
		}, {
			name:  "bad_span_style_olor_not_color",
			input: `<span style="olor: #329af0;">Hello</span>`,
			want:  `<span>Hello</span>`,
		}, {
			name:  "bad_span_style_colo_not_color",
			input: `<span style="colo: #329af0;">Hello</span>`,
			want:  `<span>Hello</span>`,
		}, {
			name:  "bad_span_style_missing_semicolon",
			input: `<span style="color: #329af0">Hello</span>`,
			want:  `<span>Hello</span>`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := sanitize(test.input)

			// Replace "&nbsp;" with "\00a0", since internally our sanitizer
			// does this and it's quite hard to write (and distinguish) the
			// literal unicode codepoint above.
			test.want = strings.Replace(test.want, "&nbsp;", "\u00a0", -1)
			if got != test.want {
				gotLines := strings.Split(got, "\n")
				wantLines := strings.Split(test.want, "\n")
				for i, gotLine := range gotLines {
					if gotLine != wantLines[i] {
						t.Log("(NOT match)")
					} else {
						t.Log("(match)")
					}
					t.Logf(" got line %d: %q\n", i, gotLine)
					t.Logf("want line %d: %q\n\n", i, wantLines[i])
				}
				t.Fatal("got != want")
			}
		})
	}
}
