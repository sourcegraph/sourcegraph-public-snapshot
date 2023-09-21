package service

import (
	"net/url"

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
	if ok, err := w.writeHeader("content"); err != nil {
		return err
	} else if ok {
		if err := w.w.WriteHeader(
			"Match type",
			"Repository",
			"Revision",
			"Repository external URL",
			"File path",
			"File URL",
			"Chunk matches [line [start end]]",
		); err != nil {
			return err
		}
	}

	repoURL := *w.host
	repoURL.Path = "/" + string(fm.Repo.Name) + "@" + string(fm.CommitID)

	fileURL := *w.host
	fileURL.Path = fm.File.URLAtCommit().Path

	return w.w.WriteRow(
		// Match type
		"content",

		// Repository
		string(fm.Repo.Name),

		// Revision
		string(fm.CommitID),

		// Repository external URL
		repoURL.String(),

		// File path
		fm.Path,

		// File URL
		fileURL.String(),

		// Chunk matches
		fm.ChunkMatches.String(),
	)
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
