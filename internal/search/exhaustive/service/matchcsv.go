package service

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type matchCSVWriter struct {
	w         CSVWriter
	headerTyp string
	host      *url.URL
}

func newMatchCSVWriter(w CSVWriter) (*matchCSVWriter, error) {
	externalURL := conf.Get().ExternalURL
	u, err := url.Parse(externalURL)
	if err != nil {
		return nil, err
	}
	return &matchCSVWriter{w: w, host: u}, nil
}

func (w *matchCSVWriter) Write(match result.Match) error {
	// TODO compare to logic used by the webapp to convert
	// results into csv. See
	// client/web/src/search/results/export/searchResultsExport.ts

	switch m := match.(type) {
	case *result.FileMatch:
		return w.writeFileMatch(m)
	default:
		return errors.Errorf("match type %T not yet supported", match)
	}
}

func (w *matchCSVWriter) writeFileMatch(fm *result.FileMatch) error {
	// Differences to "Export CSV" in webapp. We have removed columns since it
	// is easier to add columns than to remove them.
	//
	// Spaces :: We remove spaces from all column names. This is to avoid
	// needing to quote them. This makes processing of the output more
	// pleasant in tools like shell pipelines, sqlite's csv mode, etc.
	//
	// Match type :: Excluded since we only have one type for now. When we add
	// other types we may want to include them in different ways.
	//
	// Repository export URL :: We don't like it. It is verbose and is just
	// repo + rev fields. Unsure why someone would want to click on it.
	//
	// File URL :: We like this, but since we leave out actual ranges we
	// instead include an example URL to a match.
	//
	// Chunk Matches :: We are unsure who this field is for. It is hard for a
	// human to read and similarly weird for a machine to parse JSON out of a
	// CSV file. Instead we have "First match url" for a human to help
	// validate and "Match count" for calculating aggregate counts.
	//
	// First match url :: This is a new field which is a convenient URL for a
	// human to click on. We only have one URL to prevent blowing up the size
	// of the CSV. We find this field useful for building confidence.
	//
	// Match count :: In general a useful field for humans and machines.
	//
	// While we are EAP, feel free to drastically change this based on
	// feedback. After that adjusting these columns (including order) may
	// break customer workflows.

	if ok, err := w.writeHeader("content"); err != nil {
		return err
	} else if ok {
		if err := w.w.WriteHeader(
			"repository",
			"revision",
			"file_path",
			"match_count",
			"first_match_url",
		); err != nil {
			return err
		}
	}

	firstMatchURL := *w.host
	firstMatchURL.Path = fm.File.URLAtCommit().Path

	if queryParam, ok := firstMatchRawQuery(fm.ChunkMatches); ok {
		firstMatchURL.RawQuery = queryParam
	}

	return w.w.WriteRow(
		// repository
		string(fm.Repo.Name),

		// revision
		string(fm.CommitID),

		// file_path
		fm.Path,

		// match_count
		strconv.Itoa(fm.ChunkMatches.MatchCount()),

		// first_match_url
		firstMatchURL.String(),
	)
}

// firstMatchRawQuery returns the raw query parameter for the location of the
// first match. This is what is appended to the sourcegraph URL when clicking
// on a search result. eg if the match is on line 11 it is "L11". If it is
// multiline to line 13 it will be L11-13.
func firstMatchRawQuery(cms result.ChunkMatches) (string, bool) {
	cm, ok := minChunkMatch(cms)
	if !ok {
		return "", false
	}
	r, ok := minRange(cm.Ranges)
	if !ok {
		return "", false
	}

	// TODO validate how we use r.End. It is documented to be [Start, End) but
	// that would be weird for line numbers.

	// Note: Range.Line is 0-based but our UX is 1-based for line.
	if r.Start.Line != r.End.Line {
		return fmt.Sprintf("L%d-%d", r.Start.Line+1, r.End.Line+1), true
	}
	return fmt.Sprintf("L%d", r.Start.Line+1), true
}

func minChunkMatch(cms result.ChunkMatches) (result.ChunkMatch, bool) {
	if len(cms) == 0 {
		return result.ChunkMatch{}, false
	}
	min := cms[0]
	for _, cm := range cms[1:] {
		if cm.ContentStart.Line < min.ContentStart.Line {
			min = cm
		}
	}
	return min, true
}

func minRange(ranges result.Ranges) (result.Range, bool) {
	if len(ranges) == 0 {
		return result.Range{}, false
	}
	min := ranges[0]
	for _, r := range ranges[1:] {
		if r.Start.Offset < min.Start.Offset || (r.Start.Offset == min.Start.Offset && r.End.Offset < min.End.Offset) {
			min = r
		}
	}
	return min, true
}

func (w *matchCSVWriter) writeHeader(typ string) (bool, error) {
	if w.headerTyp == "" {
		w.headerTyp = typ
		return true, nil
	}
	if w.headerTyp != typ {
		return false, errors.Errorf("cant write result type %q since we have already written %q", typ, w.headerTyp)
	}
	return false, nil
}
