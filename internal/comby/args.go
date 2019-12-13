package comby

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/inconshreveable/log15.v2"
)

func (args Args) String() string {
	s := []string{
		args.MatchTemplate,
		args.RewriteTemplate,
		fmt.Sprintf("-f (%d file patterns)", len(args.FilePatterns)),
		"-json-lines",
	}
	if args.MatchOnly {
		s = append(s, "-match-only")
	} else {
		s = append(s, "-json-only-diff")
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
	default:
		log15.Error("unrecognized input type: %T", i)
		panic("unreachable")
	}

	return strings.Join(s, " ")
}
