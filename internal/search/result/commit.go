package result

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/xeonx/timeago"
)

type CommitMatch struct {
	Commit         git.Commit
	Repo           types.RepoName
	Refs           []string
	SourceRefs     []string
	MessagePreview *HighlightedString
	DiffPreview    *HighlightedString
	Body           HighlightedString
}

// ResultCount for CommitSearchResult returns the number of highlights if there
// are highlights and 1 otherwise. We implemented this method because we want to
// return a more meaningful result count for streaming while maintaining backward
// compatibility for our GraphQL API. The GraphQL API calls ResultCount on the
// resolver, while streaming calls ResultCount on CommitSearchResult.
func (r *CommitMatch) ResultCount() int {
	if n := len(r.Body.Highlights); n > 0 {
		return n
	}
	// Queries such as type:commit after:"1 week ago" don't have highlights. We count
	// those results as 1.
	return 1
}

func (r *CommitMatch) RepoName() types.RepoName {
	return r.Repo
}

func (r *CommitMatch) Limit(limit int) int {
	if len(r.Body.Highlights) == 0 {
		return limit - 1 // just counting the commit
	} else if len(r.Body.Highlights) > limit {
		r.Body.Highlights = r.Body.Highlights[:limit]
		return 0
	}
	return limit - len(r.Body.Highlights)
}

func (r *CommitMatch) Select(path filter.SelectPath) Match {
	switch path.Type {
	case filter.Repository:
		return &RepoMatch{
			Name: r.Repo.Name,
			ID:   r.Repo.ID,
		}
	case filter.Commit:
		if len(path.Fields) > 0 && path.Fields[0] == "diff" {
			if r.DiffPreview == nil {
				return nil // Not a diff result.
			}
			if len(path.Fields) == 1 {
				return r
			}
			if len(path.Fields) == 2 {
				return selectCommitDiffKind(r, path.Fields[1])
			}
			return nil
		}
		return r
	}
	return nil
}

// Key implements Match interface's Key() method
func (r *CommitMatch) Key() Key {
	typeRank := rankCommitMatch
	if r.DiffPreview != nil {
		typeRank = rankDiffMatch
	}
	return Key{
		TypeRank: typeRank,
		Repo:     r.Repo.Name,
		Commit:   r.Commit.ID,
	}
}

func (r *CommitMatch) Label() string {
	message := r.Commit.Message.Subject()
	author := r.Commit.Author.Name
	repoName := displayRepoName(string(r.Repo.Name))
	repoURL := (&RepoMatch{Name: r.Repo.Name, ID: r.Repo.ID}).URL().String()
	commitURL := r.URL().String()

	return fmt.Sprintf("[%s](%s) › [%s](%s): [%s](%s)", repoName, repoURL, author, commitURL, message, commitURL)
}

func (r *CommitMatch) Detail() string {
	commitHash := r.Commit.ID.Short()
	timeagoConfig := timeago.NoMax(timeago.English)
	return fmt.Sprintf("[`%v` %v](%v)", commitHash, timeagoConfig.Format(r.Commit.Author.Date), r.URL())
}

func (r *CommitMatch) URL() *url.URL {
	u := (&RepoMatch{Name: r.Repo.Name, ID: r.Repo.ID}).URL()
	u.Path = u.Path + "/-/commit/" + string(r.Commit.ID)
	return u
}

func displayRepoName(repoPath string) string {
	parts := strings.Split(repoPath, "/")
	if len(parts) >= 3 && strings.Contains(parts[0], ".") {
		parts = parts[1:] // remove hostname from repo path (reduce visual noise)
	}
	return strings.Join(parts, "/")
}

// selectModifiedLines extracts the highlight ranges that correspond to lines
// that have a `+` or `-` prefix (corresponding to additions resp. removals).
func selectModifiedLines(lines []string, highlights []HighlightedRange, prefix string, offset int32) []HighlightedRange {
	if len(lines) == 0 {
		return highlights
	}
	include := make([]HighlightedRange, 0, len(highlights))
	for _, h := range highlights {
		if h.Line-offset < 0 {
			// Skip negative line numbers. See: https://github.com/sourcegraph/sourcegraph/issues/20286.
			continue
		}
		if strings.HasPrefix(lines[h.Line-offset], prefix) {
			include = append(include, h)
		}
	}
	return include
}

// modifiedLinesExist checks whether any `line` in lines starts with `prefix`.
func modifiedLinesExist(lines []string, prefix string) bool {
	for _, l := range lines {
		if strings.HasPrefix(l, prefix) {
			return true
		}
	}
	return false
}

// selectCommitDiffKind returns a commit match `c` if it contains `added` (resp.
// `removed`) lines set by `field. It ensures that highlight information only
// applies to the modified lines selected by `field`. If there are no matches
// (i.e., no highlight information) coresponding to modified lines, it is
// removed from the result set (returns nil).
func selectCommitDiffKind(c *CommitMatch, field string) Match {
	diff := c.DiffPreview
	if diff == nil {
		return nil // Not a diff result.
	}
	var prefix string
	if field == "added" {
		prefix = "+"
	}
	if field == "removed" {
		prefix = "-"
	}
	if len(diff.Highlights) == 0 {
		// No highlights, implying no pattern was specified. Filter by
		// whether there exists lines corresponding to additions or
		// removals. Inspect c.Body, which is the diff markdown in the
		// format ```diff <...>``` and which doesn't contain a unified
		// diff header with +++ or --- in diff.Value, which would would
		// otherwise confuse this check.
		if modifiedLinesExist(strings.Split(c.Body.Value, "\n"), prefix) {
			return c
		}
		return nil
	}
	// We have two data structures storing highlight information for diff
	// results. We must keep these in sync. Additionally the diff highlights
	// line number is offset by 1.
	bodyHighlights := selectModifiedLines(strings.Split(c.Body.Value, "\n"), c.Body.Highlights, prefix, 0)
	diffHighlights := selectModifiedLines(strings.Split(diff.Value, "\n"), diff.Highlights, prefix, 1)
	if len(bodyHighlights) > 0 {
		// Only rely on bodyHighlights since the header in diff.Value
		// will create bogus highlights due to `+++` or `---`.
		c.Body.Highlights = bodyHighlights
		c.DiffPreview.Highlights = diffHighlights
		return c
	}
	return nil // No matching lines.
}

func (r *CommitMatch) searchResultMarker() {}
