package cliutil

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/drift"
	descriptions "github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const maxAutofixAttempts = 3

// attemptAutofix tries to apply the suggestions in the given diff summary on the store.
// Multiple attempts to apply drift may occur, and partial recovery of the schema may be
// applied. New drift may be found after an attempted autofix.
//
// This function returns a fresh drift description of the target schema if any SQL queries
// modifying the Postgres catalog have been attempted. This function returns an error only
// on failure to describe the current schema, not on failure to apply the drift.
func attemptAutofix(
	ctx context.Context,
	out *output.Output,
	store Store,
	summaries []drift.Summary,
	compareDescriptionAgainstTarget func(descriptions.SchemaDescription) []drift.Summary,
) (_ []drift.Summary, err error) {
	for attempts := maxAutofixAttempts; attempts > 0 && len(summaries) > 0 && err == nil; attempts-- {
		if !runAutofix(ctx, out, store, summaries) {
			break
		}

		out.WriteLine(output.Linef(output.EmojiInfo, output.StyleReset, "Re-checking drift"))

		schemas, err := store.Describe(ctx)
		if err != nil {
			return nil, err
		}
		schema := schemas["public"]
		summaries = compareDescriptionAgainstTarget(schema)
	}

	return summaries, nil
}

func runAutofix(
	ctx context.Context,
	out *output.Output,
	store Store,
	summaries []drift.Summary,
) bool {
	allStatements := []string{}
	for _, summary := range summaries {
		if statements, ok := summary.Statements(); ok {
			allStatements = append(allStatements, statements...)
		}
	}
	if len(allStatements) == 0 {
		out.WriteLine(output.Linef(output.EmojiInfo, output.StyleReset, "No autofix to apply"))
		return false
	}

	if err := store.RunDDLStatements(ctx, allStatements); err != nil {
		out.WriteLine(output.Linef(output.EmojiFailure, output.StyleFailure, "Failed to apply autofix: %s", err))
	} else {
		out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Successfully applied autofix"))
	}

	return true
}
