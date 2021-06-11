package git

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestDiff(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid bases", func(t *testing.T) {
		for _, input := range []string{
			"",
			"-foo",
			".foo",
		} {
			t.Run("invalid base: "+input, func(t *testing.T) {
				i, err := Diff(ctx, DiffOptions{Base: input})
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
				Mocks.ExecReader = func(args []string) (io.ReadCloser, error) {
					// The range spec is the sixth argument.
					if args[5] != tc.want {
						t.Errorf("unexpected rangeSpec: have: %s; want: %s", args[5], tc.want)
					}
					return nil, nil
				}
				_, _ = Diff(ctx, tc.opts)
			})
		}
	})

	t.Run("ExecReader error", func(t *testing.T) {
		Mocks.ExecReader = func(args []string) (io.ReadCloser, error) {
			return nil, errors.New("ExecReader error")
		}
		defer ResetMocks()

		i, err := Diff(ctx, DiffOptions{Base: "foo", Head: "bar"})
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

		Mocks.ExecReader = func(args []string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(testDiff)), nil
		}
		defer ResetMocks()

		i, err := Diff(ctx, DiffOptions{Base: "foo", Head: "bar"})
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

func TestDiffFileIterator(t *testing.T) {
	t.Run("Close", func(t *testing.T) {
		c := new(closer)
		i := &DiffFileIterator{rdr: c}
		i.Close()
		if *c != true {
			t.Errorf("iterator did not close the underlying reader: have: %v; want: %v", *c, true)
		}
	})
}

type closer bool

func (c *closer) Read(p []byte) (int, error) {
	return 0, errors.New("testing only; this should never be invoked")
}

func (c *closer) Close() error {
	*c = true
	return nil
}
