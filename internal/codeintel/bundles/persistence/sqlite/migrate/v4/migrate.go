package v4

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization"
	jsonserializer "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization/json"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/sqliteutil"
)

type MigrationStep func(context.Context, *store.Store, string, string, serialization.Serializer) error

// Migrate v4: Modify the storage of definition and references. Prior to this version, both
// tables stored scheme, identifier, and location fields as normalized rows. This version
// modifies the tables to store arrays of encoded locations keyed by (scheme, identifier)
// pairs. This makes the storage more uniform with documents and result chunks, and tends
// to save a good amount of space on disk due to the reduce number of tuples.
func Migrate(ctx context.Context, s *store.Store, serializer serialization.Serializer) error {
	steps := []MigrationStep{
		createTempTable,
		populateTable,
		swapTables,
	}

	// NOTE: We need to serialize with the JSON serializer, NOT the current serializer. This is
	// because future migrations assume that v4 was written with the most current serializer at
	// that time. Using the current serializer will cause future migrations to fail to read the
	// encoded data.
	serializer = jsonserializer.New()

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

const Delimiter = ":"

// populateTable pulls data from the old definition or reference table and inserts the data into the temporary table.
func populateTable(ctx context.Context, s *store.Store, tableName, tempTableName string, serializer serialization.Serializer) error {
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
	defer func() {
		err = store.CloseRows(rows, err)
	}()

	inserter := sqliteutil.NewBatchInserter(s, tempTableName, "scheme", "identifier", "data")

	for rows.Next() {
		monikerLocations, err := scanDefinitionReferenceRow(rows)
		if err != nil {
			return err
		}

		data, err := serializer.MarshalLocations(monikerLocations.Locations)
		if err != nil {
			return err
		}

		if err := inserter.Insert(ctx, monikerLocations.Scheme, monikerLocations.Identifier, data); err != nil {
			return err
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		return err
	}

	return nil
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

type GroupedDefinitionReferenceRow struct {
	Scheme         string
	Identifier     string
	URI            string
	StartLine      string
	StartCharacter string
	EndLine        string
	EndCharacter   string
}

func scanDefinitionReferenceRow(rows *sql.Rows) (types.MonikerLocations, error) {
	var row GroupedDefinitionReferenceRow
	if err := rows.Scan(
		&row.Scheme,
		&row.Identifier,
		&row.URI,
		&row.StartLine,
		&row.StartCharacter,
		&row.EndLine,
		&row.EndCharacter,
	); err != nil {
		return types.MonikerLocations{}, err
	}

	uriParts := strings.Split(row.URI, Delimiter)
	startLineParts := strings.Split(row.StartLine, Delimiter)
	startCharacterParts := strings.Split(row.StartCharacter, Delimiter)
	endLineParts := strings.Split(row.EndLine, Delimiter)
	endCharacterParts := strings.Split(row.EndCharacter, Delimiter)

	// Ensure that all slices have the same length so that we don't panic if we
	// index a short slice because some document path included the delimiter.
	// This REALLY should never happen as the delimilter is illegal in both Unix
	// and Windows paths.
	if n := len(uriParts); len(startLineParts) != n || len(startCharacterParts) != n || len(endLineParts) != n || len(endCharacterParts) != n {
		return types.MonikerLocations{}, fmt.Errorf("unexpected '%s' in path", Delimiter)
	}

	var locations []types.Location
	for i, uriPart := range uriParts {
		startLinePart, _ := strconv.Atoi(startLineParts[i])
		startCharacterPart, _ := strconv.Atoi(startCharacterParts[i])
		endLinePart, _ := strconv.Atoi(endLineParts[i])
		endCharacterPart, _ := strconv.Atoi(endCharacterParts[i])

		locations = append(locations, types.Location{
			URI:            uriPart,
			StartLine:      startLinePart,
			StartCharacter: startCharacterPart,
			EndLine:        endLinePart,
			EndCharacter:   endCharacterPart,
		})
	}

	return types.MonikerLocations{
		Scheme:     row.Scheme,
		Identifier: row.Identifier,
		Locations:  locations,
	}, nil
}
