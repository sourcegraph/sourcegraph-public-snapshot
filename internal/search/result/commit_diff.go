package result

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/types"
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

// Key implements Match interface's Key() method
func (cm *CommitDiffMatch) Key() Key {
	var nonEmptyPath string
	if cm.OrigName == "/dev/null" {
		nonEmptyPath = cm.NewName
	}
	if cm.NewName == "/dev/null" {
		nonEmptyPath = cm.OrigName
	}
	return Key{
		TypeRank:   rankDiffMatch,
		Repo:       cm.Repo.Name,
		AuthorDate: cm.Commit.Author.Date,
		Commit:     cm.Commit.ID,
		Path:       nonEmptyPath,
		PathStatus: cm.PathStatus,
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
				cm.Preview = filteredMatch
				return cm
			}
			return nil
		}
	}
	return nil
}

func (cm *CommitDiffMatch) searchResultMarker() {}

func ParseDiffString(diff string) (res []DiffFile, err error) {
	const (
		INIT = iota
		IN_DIFF
		IN_HUNK
	)

	state := INIT
	var currentDiff DiffFile
	var currentHunk Hunk
	for _, line := range strings.Split(diff, "\n") {
		if len(line) == 0 {
			continue
		}
		switch state {
		case INIT:
			currentDiff.OrigName, currentDiff.NewName, err = splitDiffFiles(line)
			if currentDiff.OrigName == "/dev/null" {
				currentDiff.PathStatus = Added
			} else if currentDiff.NewName == "/dev/null" {
				currentDiff.PathStatus = Deleted
			} else {
				currentDiff.PathStatus = Modified
			}
			state = IN_DIFF
		case IN_DIFF:
			currentHunk.oldStart, currentHunk.oldCount, currentHunk.newStart, currentHunk.newCount, currentHunk.header, err = parseHunkHeader(line)
			state = IN_HUNK
		case IN_HUNK:
			switch line[0] {
			case '-', '+', ' ':
				currentHunk.Lines = append(currentHunk.Lines, line)
			case '@':
				currentDiff.Hunks = append(currentDiff.Hunks, currentHunk)
				currentHunk = Hunk{}
				currentHunk.oldStart, currentHunk.oldCount, currentHunk.newStart, currentHunk.newCount, currentHunk.header, err = parseHunkHeader(line)
				state = IN_HUNK
			default:
				res = append(res, currentDiff)
				currentDiff.OrigName, currentDiff.NewName, err = splitDiffFiles(line)
				if currentDiff.OrigName == "/dev/null" {
					currentDiff.PathStatus = Added
				} else if currentDiff.NewName == "/dev/null" {
					currentDiff.PathStatus = Deleted
				} else {
					currentDiff.PathStatus = Modified
				}
				state = IN_DIFF
			}
		}
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

var errInvalidDiff = errors.New("invalid diff format")

func splitDiffFiles(fileLine string) (oldFile, newFile string, err error) {
	split := strings.Fields(fileLine)
	if len(split) != 2 {
		return "", "", errInvalidDiff
	}
	return split[0], split[1], nil
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
	PathStatus        PathStatus
	Hunks             []Hunk
}

type Hunk struct {
	oldStart, newStart int
	oldCount, newCount int
	header             string
	Lines              []string
}

type PathStatus int

const (
	Modified PathStatus = iota
	Added
	Deleted
)
