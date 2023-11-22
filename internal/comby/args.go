package comby

import (
	"fmt"
	"strconv"
	"strings"
)

func (args Args) String() string {
	s := []string{
		args.MatchTemplate,
		args.RewriteTemplate,
		"-json-lines",
	}

	if len(args.FilePatterns) > 0 {
		s = append(s, fmt.Sprintf("-f (%d file patterns)", len(args.FilePatterns)))
	}

	switch args.ResultKind {
	case MatchOnly:
		s = append(s, "-match-only")
	case Diff:
		s = append(s, "-json-only-diff")
	case Replacement:
		// Output contains replacement data in rewritten_source of JSON.
	}

	if args.NumWorkers == 0 {
		s = append(s, "-sequential")
	} else {
		s = append(s, "-jobs", strconv.Itoa(args.NumWorkers))
	}

	if args.Matcher != "" {
		s = append(s, "-matcher", args.Matcher)
	}

	switch i := args.Input.(type) {
	case ZipPath:
		s = append(s, "-zip", string(i))
	case DirPath:
		s = append(s, "-directory", string(i))
	case FileContent:
		s = append(s, fmt.Sprintf("<stdin content, length %d>", len(string(i))))
	case Tar:
		s = append(s, "-tar", "-chunk-matches", "0")
	default:
		s = append(s, fmt.Sprintf("~comby mccombyface is sad and can't handle type %T~", i))
	}

	return strings.Join(s, " ")
}
