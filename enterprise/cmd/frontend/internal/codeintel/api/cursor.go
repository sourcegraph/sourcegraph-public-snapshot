package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
)

// Cursor holds the complete state necessary to page through a reference result set.
type Cursor struct {
	Phase                  string                  // common
	DumpID                 int                     // common
	Path                   string                  // same-dump/same-dump-monikers/definition-monikers
	Line                   int                     // same-dump/same-dump-monikers
	Character              int                     // same-dump/same-dump-monikers
	Monikers               []lsifstore.MonikerData // same-dump/same-dump-monikers/definition-monikers
	SkipResults            int                     // same-dump/same-dump-monikers/definition-monikers
	Identifier             string                  // same-repo/remote-repo
	Scheme                 string                  // same-repo/remote-repo
	Name                   string                  // same-repo/remote-repo
	Version                string                  // same-repo/remote-repo
	DumpIDs                []int                   // same-repo/remote-repo
	TotalDumpsWhenBatching int                     // same-repo/remote-repo
	SkipDumpsWhenBatching  int                     // same-repo/remote-repo
	SkipDumpsInBatch       int                     // same-repo/remote-repo
	SkipResultsInDump      int                     // same-repo/remote-repo
}

// EncodeCursor returns an encoding of the given cursor suitable for a URL.
func EncodeCursor(cursor Cursor) string {
	rawEncoded, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(rawEncoded)
}

// decodeCursor is the inverse of EncodeCursor.
func decodeCursor(rawEncoded string) (Cursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(rawEncoded)
	if err != nil {
		return Cursor{}, err
	}

	var cursor Cursor
	err = json.Unmarshal([]byte(raw), &cursor)
	return cursor, err
}

// DecodeOrCreateCursor decodes and returns the raw cursor, or creates a new initial page cursor
// if a raw cursor is not supplied.
func DecodeOrCreateCursor(path string, line, character, uploadID int, rawCursor string, dbStore DBStore, lsifStore LSIFStore) (Cursor, error) {
	if rawCursor != "" {
		cursor, err := decodeCursor(rawCursor)
		if err != nil {
			return Cursor{}, err
		}

		return cursor, nil
	}

	dump, exists, err := dbStore.GetDumpByID(context.Background(), uploadID)
	if err != nil {
		return Cursor{}, errors.Wrap(err, "store.GetDumpByID")
	}
	if !exists {
		return Cursor{}, ErrMissingDump
	}

	pathInBundle := strings.TrimPrefix(path, dump.Root)
	rangeMonikers, err := lsifStore.MonikersByPosition(context.Background(), dump.ID, pathInBundle, line, character)
	if err != nil {
		return Cursor{}, errors.Wrap(err, "bundleClient.MonikersByPosition")
	}

	var flattened []lsifstore.MonikerData
	for _, monikers := range rangeMonikers {
		flattened = append(flattened, monikers...)
	}

	return Cursor{
		Phase:       "same-dump",
		DumpID:      dump.ID,
		Path:        pathInBundle,
		Line:        line,
		Character:   character,
		Monikers:    flattened,
		SkipResults: 0,
	}, nil
}
