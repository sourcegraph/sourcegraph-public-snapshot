package multiversion

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/drift"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// CheckDrift uses given runner to check whether schema drift exists for any
// non-empty database. It returns ErrDatabaseDriftDetected when the schema drift
// exists, and nil error when not.
//
//   - The `verbose` indicates whether to collect drift details in the output.
//   - The `schemaNames` is the list of schema names to check for drift.
//   - The `expectedSchemaFactories` is the means to retrieve the schema.
//     definitions at the target version.
func CheckDrift(ctx context.Context, r *runner.Runner, version string, out *output.Output, verbose bool, schemaNames []string, expectedSchemaFactories []schemas.ExpectedSchemaFactory) error {
	type schemaWithDrift struct {
		name  string
		drift *bytes.Buffer
	}
	schemasWithDrift := make([]*schemaWithDrift, 0, len(schemaNames))
	for _, schemaName := range schemaNames {
		store, err := r.Store(ctx, schemaName)
		if err != nil {
			return errors.Wrap(err, "get migration store")
		}
		schemaDescriptions, err := store.Describe(ctx)
		if err != nil {
			return err
		}
		schema := schemaDescriptions["public"]

		var driftBuff bytes.Buffer
		driftOut := output.NewOutput(&driftBuff, output.OutputOpts{})

		expectedSchema, err := FetchExpectedSchema(ctx, schemaName, version, driftOut, expectedSchemaFactories)
		if err != nil {
			return err
		}
		if err := drift.DisplaySchemaSummaries(driftOut, drift.CompareSchemaDescriptions(schemaName, version, Canonicalize(schema), Canonicalize(expectedSchema))); err != nil {
			schemasWithDrift = append(schemasWithDrift,
				&schemaWithDrift{
					name:  schemaName,
					drift: &driftBuff,
				},
			)
		}
	}

	drift := false
	for _, schemaWithDrift := range schemasWithDrift {
		empty, err := isEmptySchema(ctx, r, schemaWithDrift.name)
		if err != nil {
			return err
		}
		if empty {
			continue
		}

		drift = true
		out.WriteLine(output.Linef(output.EmojiFailure, output.StyleFailure, "Schema drift detected for %s", schemaWithDrift.name))
		if verbose {
			out.Write(schemaWithDrift.drift.String())
		}
	}
	if !drift {
		return nil
	}

	out.WriteLine(output.Linef(
		output.EmojiLightbulb,
		output.StyleItalic,
		""+
			"Before continuing with this operation, run the migrator's drift command and follow instructions to repair the schema to the expected current state."+
			" "+
			"See https://docs.sourcegraph.com/admin/updates/migrator/schema-drift for additional instructions."+
			"\n",
	))

	return ErrDatabaseDriftDetected
}

var ErrDatabaseDriftDetected = errors.New("database drift detected")

func isEmptySchema(ctx context.Context, r *runner.Runner, schemaName string) (bool, error) {
	store, err := r.Store(ctx, schemaName)
	if err != nil {
		return false, err
	}

	appliedVersions, _, _, err := store.Versions(ctx)
	if err != nil {
		return false, err
	}

	return len(appliedVersions) == 0, nil
}

func FetchExpectedSchema(
	ctx context.Context,
	schemaName string,
	version string,
	out *output.Output,
	expectedSchemaFactories []schemas.ExpectedSchemaFactory,
) (schemas.SchemaDescription, error) {
	filename, err := schemas.GetSchemaJSONFilename(schemaName)
	if err != nil {
		return schemas.SchemaDescription{}, err
	}

	out.WriteLine(output.Line(output.EmojiInfo, output.StyleReset, "Locating schema description"))

	for i, factory := range expectedSchemaFactories {
		matches := false
		patterns := factory.VersionPatterns()
		for _, pattern := range patterns {
			if pattern.MatchString(version) {
				matches = true
				break
			}
		}
		if len(patterns) > 0 && !matches {
			continue
		}

		resourcePath := factory.ResourcePath(filename, version)
		expectedSchema, err := factory.CreateFromPath(ctx, resourcePath)
		if err != nil {
			suffix := ""
			if i < len(expectedSchemaFactories)-1 {
				suffix = " Will attempt a fallback source."
			}

			out.WriteLine(output.Linef(output.EmojiInfo, output.StyleReset, "Reading schema definition in %s (%s)... Schema not found (%s).%s", factory.Name(), resourcePath, err, suffix))
			continue
		}

		out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleReset, "Schema found in %s (%s).", factory.Name(), resourcePath))
		return expectedSchema, nil
	}

	exampleMap := map[string]struct{}{}
	failedPaths := map[string]struct{}{}
	for _, factory := range expectedSchemaFactories {
		for _, pattern := range factory.VersionPatterns() {
			if !pattern.MatchString(version) {
				exampleMap[pattern.Example()] = struct{}{}
			} else {
				failedPaths[factory.ResourcePath(filename, version)] = struct{}{}
			}
		}
	}

	versionExamples := make([]string, 0, len(exampleMap))
	for pattern := range exampleMap {
		versionExamples = append(versionExamples, pattern)
	}
	sort.Strings(versionExamples)

	paths := make([]string, 0, len(exampleMap))
	for path := range failedPaths {
		if u, err := url.Parse(path); err == nil && (u.Scheme == "http" || u.Scheme == "https") {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)

	if len(paths) > 0 {
		var additionalHints string
		if len(versionExamples) > 0 {
			additionalHints = fmt.Sprintf(
				"Alternative, provide a different version that matches one of the following patterns: \n  - %s\n", strings.Join(versionExamples, "\n  - "),
			)
		}

		out.WriteLine(output.Linef(
			output.EmojiLightbulb,
			output.StyleFailure,
			"Schema not found. "+
				"Check if the following resources exist. "+
				"If they do, then the context in which this migrator is being run may not be permitted to reach the public internet."+
				"\n  - %s\n%s",
			strings.Join(paths, "\n  - "),
			additionalHints,
		))
	} else if len(versionExamples) > 0 {
		out.WriteLine(output.Linef(
			output.EmojiLightbulb,
			output.StyleFailure,
			"Schema not found. Ensure your supplied version matches one of the following patterns: \n  - %s\n", strings.Join(versionExamples, "\n  - "),
		))
	}

	return schemas.SchemaDescription{}, errors.New("failed to locate target schema description")
}

func Canonicalize(schemaDescription schemas.SchemaDescription) schemas.SchemaDescription {
	schemas.Canonicalize(schemaDescription)

	filtered := schemaDescription.Tables[:0]
	for i, table := range schemaDescription.Tables {
		if table.Name == "migration_logs" {
			continue
		}

		filtered = append(filtered, schemaDescription.Tables[i])
	}
	schemaDescription.Tables = filtered

	return schemaDescription
}
