package service

import (
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type matchCSVWriter struct {
	w         CSVWriter
	headerTyp string
}

func (w *matchCSVWriter) Write(match result.Match) error {
	// TODO match logic used by the webapp to convert
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
		if err := w.w.WriteHeader("Match type", "Repository", "Repository external URL", "File path", "File URL", "Path matches [path [start end]]", "Chunk matches [line [start end]]"); err != nil {
			return err
		}
	}

	//key := m.Key()
	return w.w.WriteRow(
		// Match type
		"content",

		// Repository
		string(fm.Repo.Name),

		// Repository external URL
		"", // TODO

		// File path
		fm.Path,

		// File URL
		"", // TODO

		// Path matches [path [start end]]
		"",

		// Chunk matches [line [start end]]
		"")
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
