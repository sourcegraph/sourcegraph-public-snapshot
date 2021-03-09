package queryrunner

import (
	"encoding/json"
	"fmt"
)

// TODO(slimsag): future: It's really nasty that our GraphQL search API is like this:
//
// 1. GraphQL makes it really annoying for us to extract the individual result values out
//    due to the `... on Type` switches.
// 2. More annoying, the GraphQL API has "results" which contain multiple matches. The web UI
//    reported the aggregate "match" count, not the number of "results" (and misleadingly
//    _calls these_ "results") - since we cannot use the global/aggregate count returned by the
//    API (we need per-repository match counts) we have to calculate this information ourselves
//    in the same way the search backend does (see the matchCount() methods) in order to get the
//    correct value. Yuck.

type result interface {
	matchCount() int
}

func decodeResult(result json.RawMessage) (result, error) {
	typeName := struct {
		TypeName string `json:"__typeName"`
	}{}
	if err := json.Unmarshal(result, &typeName); err != nil {
		return nil, err
	}
	switch typeName.TypeName {
	case "FileMatch":
		var v fileMatch
		if err := json.Unmarshal(result, &v); err != nil {
			return nil, err
		}
		return &v, nil
	case "CommitSearchResult":
		var v commitSearchResult
		if err := json.Unmarshal(result, &v); err != nil {
			return nil, err
		}
		return &v, nil
	case "Repository":
		var v repository
		if err := json.Unmarshal(result, &v); err != nil {
			return nil, err
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("cannot decode search result: unexpected TypeName: %s", string(result))
	}
}

type fileMatch struct {
	Repository struct {
		ID string
	}
	LineMatches []struct {
		OffsetAndLengths [][]int
	}
	Symbols []struct {
		Name string
	}
}

func (r *fileMatch) matchCount() int {
	matches := len(r.Symbols)
	for _, lineMatch := range r.LineMatches {
		matches += len(lineMatch.OffsetAndLengths)
	}
	if matches == 0 {
		matches = 1 // 1 to count "empty" results like type:path results
	}
	return matches
}

type commitSearchResult struct {
	Matches struct {
		Highlights []struct {
			Line int
		}
	}
	Commit struct {
		Repository struct {
			ID string
		}
	}
}

func (r *commitSearchResult) matchCount() int {
	matches := 1
	if len(r.Matches.Highlights) > 0 {
		matches = len(r.Matches.Highlights)
	}
	return matches
}

type repository struct {
	ID string
}

func (r *repository) matchCount() int {
	return 1
}
