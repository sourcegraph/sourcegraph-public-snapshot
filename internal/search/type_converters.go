package search

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// FromMatchOptions controls the behaviour of FromMatch.
type FromMatchOptions struct {
	// MaxContentLineLength will truncate lines in ChunkMatch.Content if they
	// are greater than MaxContentLineLength. If truncation happens
	// ChunkMatch.ContentTruncated is set to true.
	//
	// If MaxContentLineLength <= 0 the feature is disabled.
	MaxContentLineLength int

	// ChunkMatches is true if you want to create ChunkMatches instead of
	// LineMatches.
	ChunkMatches bool
}

func FromMatch(match result.Match, repoCache map[api.RepoID]*types.SearchedRepo, opts FromMatchOptions) http.EventMatch {
	switch v := match.(type) {
	case *result.FileMatch:
		return fromFileMatch(v, repoCache, opts)
	case *result.RepoMatch:
		return fromRepository(v, repoCache)
	case *result.CommitMatch:
		return fromCommit(v, repoCache)
	case *result.OwnerMatch:
		return fromOwner(v)
	default:
		panic(fmt.Sprintf("unknown match type %T", v))
	}
}

func fromFileMatch(fm *result.FileMatch, repoCache map[api.RepoID]*types.SearchedRepo, opts FromMatchOptions) http.EventMatch {
	if len(fm.Symbols) > 0 {
		return fromSymbolMatch(fm, repoCache)
	} else if fm.ChunkMatches.MatchCount() > 0 {
		return fromContentMatch(fm, repoCache, opts)
	}
	return fromPathMatch(fm, repoCache)
}

func fromPathMatch(fm *result.FileMatch, repoCache map[api.RepoID]*types.SearchedRepo) *http.EventPathMatch {
	pathEvent := &http.EventPathMatch{
		Type:         http.PathMatchType,
		Path:         fm.Path,
		PathMatches:  fromRanges(fm.PathMatches),
		Repository:   string(fm.Repo.Name),
		RepositoryID: int32(fm.Repo.ID),
		Commit:       string(fm.CommitID),
		Language:     fm.MostLikelyLanguage(),
	}

	if r, ok := repoCache[fm.Repo.ID]; ok {
		pathEvent.RepoStars = r.Stars
		pathEvent.RepoLastFetched = r.LastFetched
	}

	if fm.InputRev != nil {
		pathEvent.Branches = []string{*fm.InputRev}
	}

	if fm.Debug != nil {
		pathEvent.Debug = *fm.Debug
	}

	return pathEvent
}

func fromChunkMatches(cms result.ChunkMatches, opts FromMatchOptions) []http.ChunkMatch {
	res := make([]http.ChunkMatch, 0, len(cms))
	for _, cm := range cms {
		res = append(res, fromChunkMatch(cm, opts))
	}
	return res
}

func fromChunkMatch(cm result.ChunkMatch, opts FromMatchOptions) http.ChunkMatch {
	content, truncated := truncateLines(cm.Content, opts.MaxContentLineLength)
	return http.ChunkMatch{
		Content:          content,
		ContentStart:     fromLocation(cm.ContentStart),
		Ranges:           fromRanges(cm.Ranges),
		ContentTruncated: truncated,
	}
}

func fromLocation(l result.Location) http.Location {
	return http.Location{
		Offset: l.Offset,
		Line:   l.Line,
		Column: l.Column,
	}
}

func fromRanges(rs result.Ranges) []http.Range {
	res := make([]http.Range, 0, len(rs))
	for _, r := range rs {
		res = append(res, http.Range{
			Start: fromLocation(r.Start),
			End:   fromLocation(r.End),
		})
	}
	return res
}

func fromContentMatch(fm *result.FileMatch, repoCache map[api.RepoID]*types.SearchedRepo, opts FromMatchOptions) *http.EventContentMatch {

	var (
		eventLineMatches  []http.EventLineMatch
		eventChunkMatches []http.ChunkMatch
	)

	if opts.ChunkMatches {
		eventChunkMatches = fromChunkMatches(fm.ChunkMatches, opts)
	} else {
		lineMatches := fm.ChunkMatches.AsLineMatches()
		eventLineMatches = make([]http.EventLineMatch, 0, len(lineMatches))
		for _, lm := range lineMatches {
			eventLineMatches = append(eventLineMatches, http.EventLineMatch{
				Line:             lm.Preview,
				LineNumber:       lm.LineNumber,
				OffsetAndLengths: lm.OffsetAndLengths,
			})
		}
	}

	contentEvent := &http.EventContentMatch{
		Type:         http.ContentMatchType,
		Path:         fm.Path,
		PathMatches:  fromRanges(fm.PathMatches),
		RepositoryID: int32(fm.Repo.ID),
		Repository:   string(fm.Repo.Name),
		Commit:       string(fm.CommitID),
		LineMatches:  eventLineMatches,
		ChunkMatches: eventChunkMatches,
		Language:     fm.MostLikelyLanguage(),
	}

	if fm.InputRev != nil {
		contentEvent.Branches = []string{*fm.InputRev}
	}

	if r, ok := repoCache[fm.Repo.ID]; ok {
		contentEvent.RepoStars = r.Stars
		contentEvent.RepoLastFetched = r.LastFetched
	}

	if fm.Debug != nil {
		contentEvent.Debug = *fm.Debug
	}

	return contentEvent
}

func fromSymbolMatch(fm *result.FileMatch, repoCache map[api.RepoID]*types.SearchedRepo) *http.EventSymbolMatch {
	symbols := make([]http.Symbol, 0, len(fm.Symbols))
	for _, sym := range fm.Symbols {
		kind := sym.Symbol.LSPKind()
		kindString := "UNKNOWN"
		if kind != 0 {
			kindString = strings.ToUpper(kind.String())
		}

		symbols = append(symbols, http.Symbol{
			URL:           sym.URL().String(),
			Name:          sym.Symbol.Name,
			ContainerName: sym.Symbol.Parent,
			Kind:          kindString,
			Line:          int32(sym.Symbol.Line),
		})
	}

	symbolMatch := &http.EventSymbolMatch{
		Type:         http.SymbolMatchType,
		Path:         fm.Path,
		Repository:   string(fm.Repo.Name),
		RepositoryID: int32(fm.Repo.ID),
		Commit:       string(fm.CommitID),
		Language:     fm.MostLikelyLanguage(),
		Symbols:      symbols,
	}

	if r, ok := repoCache[fm.Repo.ID]; ok {
		symbolMatch.RepoStars = r.Stars
		symbolMatch.RepoLastFetched = r.LastFetched
	}

	if fm.InputRev != nil {
		symbolMatch.Branches = []string{*fm.InputRev}
	}

	return symbolMatch
}

func fromRepository(rm *result.RepoMatch, repoCache map[api.RepoID]*types.SearchedRepo) *http.EventRepoMatch {
	var branches []string
	if rev := rm.Rev; rev != "" {
		branches = []string{rev}
	}

	repoEvent := &http.EventRepoMatch{
		Type:               http.RepoMatchType,
		RepositoryID:       int32(rm.ID),
		Repository:         string(rm.Name),
		RepositoryMatches:  fromRanges(rm.RepoNameMatches),
		Branches:           branches,
		DescriptionMatches: fromRanges(rm.DescriptionMatches),
	}

	if r, ok := repoCache[rm.ID]; ok {
		repoEvent.RepoStars = r.Stars
		repoEvent.RepoLastFetched = r.LastFetched
		repoEvent.Description = r.Description
		repoEvent.Fork = r.Fork
		repoEvent.Archived = r.Archived
		repoEvent.Private = r.Private
		repoEvent.Metadata = r.KeyValuePairs
		repoEvent.Topics = r.Topics
	}

	return repoEvent
}

func fromCommit(commit *result.CommitMatch, repoCache map[api.RepoID]*types.SearchedRepo) *http.EventCommitMatch {
	hls := commit.Body().ToHighlightedString()
	ranges := make([][3]int32, len(hls.Highlights))
	for i, h := range hls.Highlights {
		ranges[i] = [3]int32{h.Line, h.Character, h.Length}
	}

	commitEvent := &http.EventCommitMatch{
		Type:          http.CommitMatchType,
		Label:         commit.Label(),
		URL:           commit.URL().String(),
		Detail:        commit.Detail(),
		Repository:    string(commit.Repo.Name),
		RepositoryID:  int32(commit.Repo.ID),
		OID:           string(commit.Commit.ID),
		Message:       string(commit.Commit.Message),
		AuthorName:    commit.Commit.Author.Name,
		AuthorDate:    commit.Commit.Author.Date,
		CommitterName: commit.Commit.Committer.Name,
		CommitterDate: commit.Commit.Committer.Date,
		Content:       hls.Value,
		Ranges:        ranges,
	}

	if r, ok := repoCache[commit.Repo.ID]; ok {
		commitEvent.RepoStars = r.Stars
		commitEvent.RepoLastFetched = r.LastFetched
	}

	return commitEvent
}

func fromOwner(owner *result.OwnerMatch) http.EventMatch {
	switch v := owner.ResolvedOwner.(type) {
	case *result.OwnerPerson:
		person := &http.EventPersonMatch{
			Type:   http.PersonMatchType,
			Handle: v.Handle,
			Email:  v.Email,
		}
		if v.User != nil {
			person.User = &http.UserMetadata{
				Username:    v.User.Username,
				DisplayName: v.User.DisplayName,
				AvatarURL:   v.User.AvatarURL,
			}
		}
		return person
	case *result.OwnerTeam:
		return &http.EventTeamMatch{
			Type:        http.TeamMatchType,
			Handle:      v.Handle,
			Email:       v.Email,
			Name:        v.Team.Name,
			DisplayName: v.Team.DisplayName,
		}
	default:
		panic(fmt.Sprintf("unknown owner match type %T", v))
	}
}
