package compute

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/grafana/regexp"
	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func Test_output(t *testing.T) {
	test := func(input string, cmd *Output) string {
		content, err := output(context.Background(), logtest.Scoped(t), input, cmd.SearchPattern, cmd.OutputPattern, cmd.Separator)
		if err != nil {
			return err.Error()
		}
		return content
	}

	autogold.Expect("(1)~(2)~(3)~").
		Equal(t, test("a 1 b 2 c 3", &Output{
			SearchPattern: &Regexp{Value: regexp.MustCompile(`(\d)`)},
			OutputPattern: "($1)",
			Separator:     "~",
		}))

	// If we are not on CI skip the test if comby is not installed.
	if os.Getenv("CI") == "" && !comby.Exists() {
		t.Skip("comby is not installed on the PATH. Try running 'bash <(curl -sL get.comby.dev)'.")
	}

	autogold.Expect(`train(regional, intercity)
train(commuter, lightrail)`).
		Equal(t, test("Im a train. train(intercity, regional). choo choo. train(lightrail, commuter)", &Output{
			SearchPattern: &Comby{Value: `train(:[x], :[y])`},
			OutputPattern: "train(:[y], :[x])",
		}))
}

func fileMatch(chunks ...string) result.Match {
	return fileMatchWithPath("my/awesome/path.ml", chunks...)
}

func fileMatchWithPath(path string, chunks ...string) result.Match {
	matches := make([]result.ChunkMatch, 0, len(chunks))
	for _, content := range chunks {
		matches = append(matches, result.ChunkMatch{
			Content:      content,
			ContentStart: result.Location{Offset: 0, Line: 1, Column: 0},
			Ranges: result.Ranges{{
				Start: result.Location{Offset: 0, Line: 1, Column: 0},
				End:   result.Location{Offset: len(content), Line: 1, Column: len(content)},
			}},
		})
	}

	return &result.FileMatch{
		File: result.File{
			Repo: types.MinimalRepo{Name: "my/awesome/repo"},
			Path: path,
		},
		ChunkMatches: matches,
	}
}

func commitMatch(content string) result.Match {
	return &result.CommitMatch{
		Commit: gitdomain.Commit{
			Author:    gitdomain.Signature{Name: "bob"},
			Committer: &gitdomain.Signature{},
			Message:   gitdomain.Message(content),
		},
	}
}

func commitDiffMatch(path string, content string) result.Match {
	return &result.CommitDiffMatch{
		Commit: gitdomain.Commit{
			Author:    gitdomain.Signature{Name: "bob"},
			Committer: &gitdomain.Signature{},
			Message:   gitdomain.Message(content),
		},
		DiffFile: &result.DiffFile{
			NewName: path,
		},
	}
}

func TestRun(t *testing.T) {
	test := func(q string, m result.Match) string {
		computeQuery, _ := Parse(q)
		commandResult, err := computeQuery.Command.Run(context.Background(), gitserver.NewMockClient(), m)
		if err != nil {
			return err.Error()
		}

		switch r := commandResult.(type) {
		case *Text:
			return r.Value
		case *TextExtra:
			commandResult, _ := json.Marshal(r)
			return string(commandResult)
		}
		return "Error, unrecognized result type returned"
	}

	autogold.Expect("(1)\n(2)\n(3)\n").
		Equal(t, test(`content:output((\d) -> ($1))`, fileMatch("a 1 b 2 c 3")))

	autogold.Expect("my/awesome/repo").
		Equal(t, test(`lang:ocaml content:output((\d) -> $repo) select:repo`, fileMatch("a 1 b 2 c 3")))

	autogold.Expect("my/awesome/path.ml content is my/awesome/path.ml with extension: ml\n").
		Equal(t, test(`content:output(awesome/.+\.(\w+) -> $path content is $content with extension: $1) type:path`, fileMatch("a 1 b 2 c 3")))

	autogold.Expect("bob: (1)\nbob: (2)\nbob: (3)\n").
		Equal(t, test(`content:output((\d) -> $author: ($1))`, commitMatch("a 1 b 2 c 3")))

	autogold.Expect("test\nstring\n").
		Equal(t, test(`content:output((\b\w+\b) -> $1)`, fileMatch("test", "string")))

	// If we are not on CI skip the test if comby is not installed.
	if os.Getenv("CI") == "" && !comby.Exists() {
		t.Skip("comby is not installed on the PATH. Try running 'bash <(curl -sL get.comby.dev)'.")
	}

	autogold.Expect(">bar<").
		Equal(t, test(`content:output.structural(foo(:[arg]) -> >:[arg]<)`, fileMatch("foo(bar)")))

	autogold.Expect("OCaml\n").
		Equal(t, test(`content:output((.|\n)* -> $lang)`, fileMatch("anything")))
	autogold.Expect("Magik\n").
		Equal(t, test(`content:output((.|\n)* -> $lang)`, fileMatchWithPath("foo/bar/lang.magik", "anything")))
	autogold.Expect("Magik\n").
		Equal(t, test(`content:output((.|\n)* -> $lang)`, commitDiffMatch("foo/bar/lang.magik", "anything")))
	autogold.Expect("C#\n").
		Equal(t, test(`content:output((.|\n)* -> $lang)`, commitDiffMatch("foo/bar/lang.cs", "anything")))
	autogold.Expect(`{"value":"OCaml\n","kind":"output","repositoryID":0,"repository":"my/awesome/repo"}`).
		Equal(t, test(`content:output.extra((.|\n)* -> $lang)`, fileMatch("anything")))

		// We may have empty string for language if there is not known language for file path
	autogold.Expect("\n").
		Equal(t, test(`content:output((.|\n)* -> $lang)`, commitDiffMatch("foo/bar/lang.blahblahblah", "anything")))
}
