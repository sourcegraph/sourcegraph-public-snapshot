package result

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type CommitDiffMatch struct {
	Commit  gitdomain.Commit
	Repo    types.MinimalRepo
	Preview *MatchedString
	*DiffFile
}

func (cd *CommitDiffMatch) RepoName() types.MinimalRepo {
	return cd.Repo
}

// Path returns a nonempty path associated with a diff. This value is the usual
// path when the associated file is modified. When it is created or removed, it
// returns the path of the associated file being created or removed.
func (cm *CommitDiffMatch) Path() string {
	nonEmptyPath := cm.NewName
	if cm.NewName == "/dev/null" {
		nonEmptyPath = cm.OrigName
	}
	return nonEmptyPath
}

func (cm *CommitDiffMatch) PathStatus() PathStatus {
	if cm.OrigName == "/dev/null" {
		return Added
	}

	if cm.NewName == "/dev/null" {
		return Deleted
	}

	return Modified
}

// Key implements Match interface's Key() method
func (cm *CommitDiffMatch) Key() Key {
	return Key{
		TypeRank:   rankDiffMatch,
		Repo:       cm.Repo.Name,
		AuthorDate: cm.Commit.Author.Date,
		Commit:     cm.Commit.ID,
		Path:       cm.Path(),
	}
}

func (cm *CommitDiffMatch) ResultCount() int {
	matchCount := len(cm.Preview.MatchedRanges)
	if matchCount > 0 {
		return matchCount
	}
	// Queries such as type:diff after:"1 week ago" don't have highlights. We count
	// those results as 1.
	return 1
}

func (cm *CommitDiffMatch) Limit(limit int) int {
	limitMatchedString := func(ms *MatchedString) int {
		if len(ms.MatchedRanges) == 0 {
			return limit - 1
		} else if len(ms.MatchedRanges) > limit {
			ms.MatchedRanges = ms.MatchedRanges[:limit]
			return 0
		}
		return limit - len(ms.MatchedRanges)
	}

	return limitMatchedString(cm.Preview)
}

func (cm *CommitDiffMatch) Select(path filter.SelectPath) Match {
	switch path.Root() {
	case filter.Repository:
		return &RepoMatch{
			Name: cm.Repo.Name,
			ID:   cm.Repo.ID,
		}
	case filter.Commit:
		fields := path[1:]
		if len(fields) > 0 && fields[0] == "diff" {
			if len(fields) == 1 {
				return cm
			}
			if len(fields) == 2 {
				filteredMatch := selectCommitDiffKind(cm.Preview, fields[1])
				if filteredMatch == nil {
					// no result after selecting, propagate no result.
					return nil
				}
				cm.Preview = filteredMatch
				return cm
			}
			return nil
		}
		return cm
	}
	return nil
}

func (cm *CommitDiffMatch) searchResultMarker() {}

// FormatDiffFiles inverts ParseDiffString
func FormatDiffFiles(res []DiffFile) string {
	var buf strings.Builder
	for _, diffFile := range res {
		buf.WriteString(escaper.Replace(diffFile.OrigName))
		buf.WriteByte(' ')
		buf.WriteString(escaper.Replace(diffFile.NewName))
		buf.WriteByte('\n')
		for _, hunk := range diffFile.Hunks {
			fmt.Fprintf(&buf, "@@ -%d,%d +%d,%d @@", hunk.OldStart, hunk.OldCount, hunk.NewStart, hunk.NewCount)
			if hunk.Header != "" {
				// Only add a space before the header if the header is non-empty
				fmt.Fprintf(&buf, " %s", hunk.Header)
			}
			buf.WriteByte('\n')
			for _, line := range hunk.Lines {
				buf.WriteString(line)
				buf.WriteByte('\n')
			}
		}
	}
	return buf.String()
}

var escaper = strings.NewReplacer(" ", `\ `)
var unescaper = strings.NewReplacer(`\ `, " ")

func ParseDiffString(diff string) (res []DiffFile, err error) {
	const (
		INIT = iota
		IN_DIFF
		IN_HUNK
	)

	state := INIT
	var currentDiff DiffFile
	finishDiff := func() {
		res = append(res, currentDiff)
		currentDiff = DiffFile{}
	}

	var currentHunk Hunk
	finishHunk := func() {
		currentDiff.Hunks = append(currentDiff.Hunks, currentHunk)
		currentHunk = Hunk{}
	}

	for _, line := range strings.Split(diff, "\n") {
		if len(line) == 0 {
			continue
		}
		switch state {
		case INIT:
			currentDiff.OrigName, currentDiff.NewName, err = splitDiffFiles(line)
			state = IN_DIFF
		case IN_DIFF:
			currentHunk.OldStart, currentHunk.OldCount, currentHunk.NewStart, currentHunk.NewCount, currentHunk.Header, err = parseHunkHeader(line)
			state = IN_HUNK
		case IN_HUNK:
			switch line[0] {
			case '-', '+', ' ':
				currentHunk.Lines = append(currentHunk.Lines, line)
			case '@':
				finishHunk()
				currentHunk.OldStart, currentHunk.OldCount, currentHunk.NewStart, currentHunk.NewCount, currentHunk.Header, err = parseHunkHeader(line)
				state = IN_HUNK
			default:
				finishHunk()
				finishDiff()
				currentDiff.OrigName, currentDiff.NewName, err = splitDiffFiles(line)
				state = IN_DIFF
			}
		}
		if err != nil {
			return nil, err
		}
	}
	finishHunk()
	finishDiff()

	return res, nil
}

var errInvalidDiff = errors.New("invalid diff format")
var splitRegex = lazyregexp.New(`(.*[^\\]) (.*)`)

func splitDiffFiles(fileLine string) (oldFile, newFile string, err error) {
	match := splitRegex.FindStringSubmatch(fileLine)
	if len(match) == 0 {
		return "", "", errInvalidDiff
	}
	return unescaper.Replace(match[1]), unescaper.Replace(match[2]), nil
}

var headerRegex = regexp.MustCompile(`@@ -(\d+),(\d+) \+(\d+),(\d+) @@\ ?(.*)`)

func parseHunkHeader(headerLine string) (oldStart, oldCount, newStart, newCount int, header string, err error) {
	groups := headerRegex.FindStringSubmatch(headerLine)
	if groups == nil {
		return 0, 0, 0, 0, "", errInvalidDiff
	}
	oldStart, err = strconv.Atoi(groups[1])
	if err != nil {
		return 0, 0, 0, 0, "", err
	}
	oldCount, err = strconv.Atoi(groups[2])
	if err != nil {
		return 0, 0, 0, 0, "", err
	}
	newStart, err = strconv.Atoi(groups[3])
	if err != nil {
		return 0, 0, 0, 0, "", err
	}
	newCount, err = strconv.Atoi(groups[4])
	if err != nil {
		return 0, 0, 0, 0, "", err
	}
	return oldStart, oldCount, newStart, newCount, groups[5], nil
}

type DiffFile struct {
	OrigName, NewName string
	Hunks             []Hunk
}

type Hunk struct {
	OldStart, NewStart int
	OldCount, NewCount int
	Header             string
	Lines              []string
}

type PathStatus int

const (
	Modified PathStatus = iota
	Added
	Deleted
)
