package v4

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/serialization"
	jsonserializer "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/serialization/json"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/sqlite/batch"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/sqlite/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/sqlite/util"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/types"
)

type MigrationStep func(context.Context, *store.Store, string, string, serialization.Serializer) error

// Migrate v4: Modify the storage of definition and references. Prior to this version, both
// tables stored scheme, identifier, and location fields as normalized rows. This version
// modifies the tables to store arrays of encoded locations keyed by (scheme, identifier)
// pairs. This makes the storage more uniform with documents and result chunks, and tends
// to save a good amount of space on disk due to the reduce number of tuples.
func Migrate(ctx context.Context, s *store.Store, _ serialization.Serializer) error {
	steps := []MigrationStep{
		createTempTable,
		populateTable,
		swapTables,
	}

	// NOTE: We need to serialize with the JSON serializer, NOT the current serializer. This is
	// because future migrations assume that v4 was written with the most current serializer at
	// that time. Using the current serializer will cause future migrations to fail to read the
	// encoded data.
	serializer := jsonserializer.New()

	for _, tableName := range []string{"definitions", "references"} {
		for _, step := range steps {
			if err := step(ctx, s, tableName, fmt.Sprintf("t_%s", tableName), serializer); err != nil {
				return err
			}
		}
	}

	return nil
}

// createTempTable creates a new, empty table that contains the new format of the definition and reference data.
func createTempTable(ctx context.Context, s *store.Store, tableName, tempTableName string, serializer serialization.Serializer) error {
	return s.Exec(ctx, sqlf.Sprintf(`CREATE TABLE "`+tempTableName+`" ("scheme" text NOT NULL, "identifier" text NOT NULL, "data" blob NOT NULL)`))
}

// populateTable pulls data from the old definition or reference table and inserts the data into the temporary table.
func populateTable(ctx context.Context, s *store.Store, tableName, tempTableName string, serializer serialization.Serializer) error {
	ch := make(chan types.MonikerLocations)

	return util.InvokeAll(
		func() error { return readMonikerLocations(ctx, s, tableName, ch) },
		func() error { return batch.WriteMonikerLocationsChan(ctx, s, tempTableName, serializer, ch) },
	)
}

// swapTables deletes the original table and replaces it with the temporary table.
func swapTables(ctx context.Context, s *store.Store, tableName, tempTableName string, serializer serialization.Serializer) error {
	queries := []*sqlf.Query{
		sqlf.Sprintf(`DROP TABLE "` + tableName + `"`),
		sqlf.Sprintf(`ALTER TABLE "` + tempTableName + `" RENAME TO "` + tableName + `"`),
	}

	for _, query := range queries {
		if err := s.Exec(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

// readMonikerLocations reads all moniker locations from the given table name and writes the scanned results
// onto the given channel. If an error occurs during query or scanning, that error is returned and no future
// writes to the channel will be performed. The given channel is closed when the function exits.
func readMonikerLocations(ctx context.Context, s *store.Store, tableName string, ch chan<- types.MonikerLocations) (err error) {
	defer close(ch)

	rows, err := s.Query(
		ctx,
		sqlf.Sprintf(`
			SELECT
				scheme,
				identifier,
				group_concat(documentPath, %s),
				group_concat(startLine, %s),
				group_concat(startCharacter, %s),
				group_concat(endLine, %s),
				group_concat(endCharacter, %s)
			FROM "`+tableName+`"
			GROUP BY scheme, identifier
		`,
			Delimiter, Delimiter, Delimiter, Delimiter, Delimiter,
		),
	)
	if err != nil {
		return err
	}
	defer func() { err = store.CloseRows(rows, err) }()

	for rows.Next() {
		monikerLocations, err := scanDefinitionReferenceRow(rows)
		if err != nil {
			return err
		}

		ch <- monikerLocations
	}

	return nil
}

// Delimiter is the text used to separate values of distinct rows when performing a the GROUP BY query.
const Delimiter = ":"

// GroupedDefinitionReferenceRow is a row of all moniker locations grouped by scheme and identifier. The
// remaining columns are string values concatenated by the delimiter defined above.
type GroupedDefinitionReferenceRow struct {
	Scheme          string
	Identifier      string
	URIs            string
	StartLines      string
	StartCharacters string
	EndLines        string
	EndCharacters   string
}

// scanDefinitionReferenceRow reads a row that describes the GroupedDefinitionReferenceRow and converts it
// into a moniker location. The uri and range data for each location is extracted by splitting the concatenated
// string values of each row by the delimiter defined above.
func scanDefinitionReferenceRow(rows *sql.Rows) (types.MonikerLocations, error) {
	var row GroupedDefinitionReferenceRow
	if err := rows.Scan(
		&row.Scheme,
		&row.Identifier,
		&row.URIs,
		&row.StartLines,
		&row.StartCharacters,
		&row.EndLines,
		&row.EndCharacters,
	); err != nil {
		return types.MonikerLocations{}, err
	}

	v1 := strings.Split(row.URIs, Delimiter)
	v2 := strings.Split(row.StartLines, Delimiter)
	v3 := strings.Split(row.StartCharacters, Delimiter)
	v4 := strings.Split(row.EndLines, Delimiter)
	v5 := strings.Split(row.EndCharacters, Delimiter)

	// Ensure that all slices have the same length so that we don't panic if we
	// index a short slice because some document path included the delimiter. This
	// REALLY should never happen as the delimilter is illegal in Unix/Windows paths.
	if n := len(v1); len(v2) != n || len(v3) != n || len(v4) != n || len(v5) != n {
		return types.MonikerLocations{}, fmt.Errorf("unexpected '%s' in path", Delimiter)
	}
	locations := make([]types.Location, 0, len(v1))
	for i := range v1 {
		i2, _ := strconv.Atoi(v2[i])
		i3, _ := strconv.Atoi(v3[i])
		i4, _ := strconv.Atoi(v4[i])
		i5, _ := strconv.Atoi(v5[i])

		locations = append(locations, types.Location{
			URI:            v1[i],
			StartLine:      i2,
			StartCharacter: i3,
			EndLine:        i4,
			EndCharacter:   i5,
		})
	}

	return types.MonikerLocations{
		Scheme:     row.Scheme,
		Identifier: row.Identifier,
		Locations:  locations,
	}, nil
}
