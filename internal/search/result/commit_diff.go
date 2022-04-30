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
	Commit gitdomain.Commit
	Repo   types.MinimalRepo
	*DiffFile
}

func (cd *CommitDiffMatch) RepoName() types.MinimalRepo {
	return cd.Repo
}

// Return a file path associated with this diff. If the file was created, it
// returns the created file path. IF the file was deleted, it returns the
// deleted file path. If modified, returns the modified path.
func (cm *CommitDiffMatch) Path() string {
	if cm.OrigName == "/dev/null" {
		return cm.NewName
	}
	if cm.NewName == "/dev/null" {
		return cm.OrigName
	}
	return cm.OrigName
}

// Key implements Match interface's Key() method
func (cm *CommitDiffMatch) Key() Key {
	var nonEmptyPath string
	pathStatus := Modified
	if cm.OrigName == "/dev/null" {
		nonEmptyPath = cm.NewName
		pathStatus = Added
	}
	if cm.NewName == "/dev/null" {
		nonEmptyPath = cm.OrigName
		pathStatus = Deleted
	}
	return Key{
		TypeRank:   rankDiffMatch,
		Repo:       cm.Repo.Name,
		AuthorDate: cm.Commit.Author.Date,
		Commit:     cm.Commit.ID,
		Path:       nonEmptyPath,
		PathStatus: pathStatus,
	}
}

func (cm *CommitDiffMatch) ResultCount() int {
	return 0 // TODO
}

func (cm *CommitDiffMatch) Limit(int) int {
	return 0 // TODO
}

func (cm *CommitDiffMatch) Select(filter.SelectPath) Match {
	return nil // TODO
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

var headerRegex = regexp.MustCompile(`@@ -(\d+),(\d+) \+(\d+),(\d+) @@ (.*)`)

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
	oldStart, newStart int
	oldCount, newCount int
	header             string
	Lines              []string
}
