package gitserver

import (
	"context"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestParseShortLog(t *testing.T) {
	tests := []struct {
		name    string
		input   string // in the format of `git shortlog -sne`
		want    []*gitdomain.PersonCount
		wantErr error
	}{
		{
			name: "basic",
			input: `
  1125	Jane Doe <jane@sourcegraph.com>
   390	Bot Of Doom <bot@doombot.com>
`,
			want: []*gitdomain.PersonCount{
				{
					Name:  "Jane Doe",
					Email: "jane@sourcegraph.com",
					Count: 1125,
				},
				{
					Name:  "Bot Of Doom",
					Email: "bot@doombot.com",
					Count: 390,
				},
			},
		},
		{
			name: "commonly malformed (email address as name)",
			input: `  1125	jane@sourcegraph.com <jane@sourcegraph.com>
   390	Bot Of Doom <bot@doombot.com>
`,
			want: []*gitdomain.PersonCount{
				{
					Name:  "jane@sourcegraph.com",
					Email: "jane@sourcegraph.com",
					Count: 1125,
				},
				{
					Name:  "Bot Of Doom",
					Email: "bot@doombot.com",
					Count: 390,
				},
			},
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got, gotErr := parseShortLog([]byte(tst.input))
			if (gotErr == nil) != (tst.wantErr == nil) {
				t.Fatalf("gotErr %+v wantErr %+v", gotErr, tst.wantErr)
			}
			if !reflect.DeepEqual(got, tst.want) {
				t.Logf("got %q", got)
				t.Fatalf("want %q", tst.want)
			}
		})
	}
}

func TestDiff(t *testing.T) {
	ctx := context.Background()
	db := database.NewMockDB()

	t.Run("invalid bases", func(t *testing.T) {
		for _, input := range []string{
			"",
			"-foo",
			".foo",
		} {
			t.Run("invalid base: "+input, func(t *testing.T) {
				i, err := NewClient(db).Diff(ctx, DiffOptions{Base: input})
				if i != nil {
					t.Errorf("unexpected non-nil iterator: %+v", i)
				}
				if err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("rangeSpec calculation", func(t *testing.T) {
		for _, tc := range []struct {
			opts DiffOptions
			want string
		}{
			{opts: DiffOptions{Base: "foo", Head: "bar"}, want: "foo...bar"},
		} {
			t.Run("rangeSpec: "+tc.want, func(t *testing.T) {
				c := NewClient(db)
				Mocks.ExecReader = func(args []string) (reader io.ReadCloser, err error) {
					// The range spec is the sixth argument.
					if args[5] != tc.want {
						t.Errorf("unexpected rangeSpec: have: %s; want: %s", args[5], tc.want)
					}
					return nil, nil
				}
				t.Cleanup(ResetMocks)
				_, _ = c.Diff(ctx, tc.opts)
			})
		}
	})

	t.Run("ExecReader error", func(t *testing.T) {
		c := NewClient(db)
		Mocks.ExecReader = func(args []string) (reader io.ReadCloser, err error) {
			return nil, errors.New("ExecReader error")
		}
		t.Cleanup(ResetMocks)

		i, err := c.Diff(ctx, DiffOptions{Base: "foo", Head: "bar"})
		if i != nil {
			t.Errorf("unexpected non-nil iterator: %+v", i)
		}
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("success", func(t *testing.T) {
		const testDiffFiles = 3
		const testDiff = `diff --git INSTALL.md INSTALL.md
index e5af166..d44c3fc 100644
--- INSTALL.md
+++ INSTALL.md
@@ -3,10 +3,10 @@
 Line 1
 Line 2
 Line 3
-Line 4
+This is cool: Line 4
 Line 5
 Line 6
-Line 7
-Line 8
+Another Line 7
+Foobar Line 8
 Line 9
 Line 10
diff --git JOKES.md JOKES.md
index ea80abf..1b86505 100644
--- JOKES.md
+++ JOKES.md
@@ -4,10 +4,10 @@ Joke #1
 Joke #2
 Joke #3
 Joke #4
-Joke #5
+This is not funny: Joke #5
 Joke #6
-Joke #7
+This one is good: Joke #7
 Joke #8
-Joke #9
+Waffle: Joke #9
 Joke #10
 Joke #11
diff --git README.md README.md
index 9bd8209..d2acfa9 100644
--- README.md
+++ README.md
@@ -1,12 +1,13 @@
 # README

-Line 1
+Foobar Line 1
 Line 2
 Line 3
 Line 4
 Line 5
-Line 6
+Barfoo Line 6
 Line 7
 Line 8
 Line 9
 Line 10
+Another line
`

		var testDiffFileNames = []string{
			"INSTALL.md",
			"JOKES.md",
			"README.md",
		}

		c := NewClient(db)
		Mocks.ExecReader = func(args []string) (reader io.ReadCloser, err error) {
			return io.NopCloser(strings.NewReader(testDiff)), nil
		}
		t.Cleanup(ResetMocks)

		i, err := c.Diff(ctx, DiffOptions{Base: "foo", Head: "bar"})
		if i == nil {
			t.Error("unexpected nil iterator")
		}
		if err != nil {
			t.Errorf("unexpected non-nil error: %+v", err)
		}
		defer i.Close()

		count := 0
		for {
			diff, err := i.Next()
			if err == io.EOF {
				break
			} else if err != nil {
				t.Errorf("unexpected iteration error: %+v", err)
			}

			if diff.OrigName != testDiffFileNames[count] {
				t.Errorf("unexpected diff file name: have: %s; want: %s", diff.OrigName, testDiffFileNames[count])
			}
			count++
		}
		if count != testDiffFiles {
			t.Errorf("unexpected diff count: have %d; want %d", count, testDiffFiles)
		}
	})
}
