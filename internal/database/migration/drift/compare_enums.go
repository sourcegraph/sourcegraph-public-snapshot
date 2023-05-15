package drift

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func compareEnums(schemaName, version string, actual, expected schemas.SchemaDescription) []Summary {
	return compareNamedLists(actual.Enums, expected.Enums, func(enum *schemas.EnumDescription, expectedEnum schemas.EnumDescription) Summary {
		quotedLabels := make([]string, 0, len(expectedEnum.Labels))
		for _, label := range expectedEnum.Labels {
			quotedLabels = append(quotedLabels, fmt.Sprintf("'%s'", label))
		}
		createEnumStmt := fmt.Sprintf("CREATE TYPE %s AS ENUM (%s);", expectedEnum.Name, strings.Join(quotedLabels, ", "))
		dropEnumStmt := fmt.Sprintf("DROP TYPE %s;", expectedEnum.Name)

		if enum == nil {
			return newDriftSummary(
				expectedEnum.Name,
				fmt.Sprintf("Missing enum %q", expectedEnum.Name),
				"create the type",
			).withStatements(createEnumStmt)
		}

		if ordered, ok := constructEnumRepairStatements(*enum, expectedEnum); ok {
			return newDriftSummary(
				expectedEnum.Name,
				fmt.Sprintf("Missing %d labels for enum %q", len(ordered), expectedEnum.Name),
				"add the missing enum labels",
			).withStatements(ordered...)
		}

		return newDriftSummary(
			expectedEnum.Name,
			fmt.Sprintf("Unexpected labels for enum %q", expectedEnum.Name),
			"drop and re-create the type",
		).withDiff(enum.Labels, expectedEnum.Labels).withStatements(dropEnumStmt, createEnumStmt)
	}, noopAdditionalCallback[schemas.EnumDescription])
}

// constructEnumRepairStatements returns a set of `ALTER ENUM ADD VALUE` statements to make
// the given enum equivalent to the given expected enum. If the given enum is not a subset of
// the expected enum, then additive statements cannot bring the enum to the expected state and
// we return a false-valued flag. In this case the existing type must be dropped and re-created
// as there's currently no way to *remove* values from an enum type.
func constructEnumRepairStatements(enum, expectedEnum schemas.EnumDescription) ([]string, bool) {
	labels := groupByName(wrapStrings(enum.Labels))
	expectedLabels := groupByName(wrapStrings(expectedEnum.Labels))

	for label := range labels {
		if _, ok := expectedLabels[label]; !ok {
			return nil, false
		}
	}

	// If we're here then we're strictly missing labels and can add them in-place.
	// Try to reconstruct the data we need to make the proper create type statement.

	type missingLabel struct {
		label    string
		neighbor string
		before   bool
	}
	missingLabels := make([]missingLabel, 0, len(expectedEnum.Labels))

	after := ""
	for _, label := range expectedEnum.Labels {
		if _, ok := labels[label]; !ok && after != "" {
			missingLabels = append(missingLabels, missingLabel{label: label, neighbor: after, before: false})
		}
		after = label
	}

	before := ""
	for i := len(expectedEnum.Labels) - 1; i >= 0; i-- {
		label := expectedEnum.Labels[i]

		if _, ok := labels[label]; !ok && before != "" {
			missingLabels = append(missingLabels, missingLabel{label: label, neighbor: before, before: true})
		}
		before = label
	}

	var (
		ordered   []string
		reachable = groupByName(wrapStrings(enum.Labels))
	)

outer:
	for len(missingLabels) > 0 {
		for _, s := range missingLabels {
			// Neighbor doesn't exist yet, blocked from creating
			if _, ok := reachable[s.neighbor]; !ok {
				continue
			}

			rel := "AFTER"
			if s.before {
				rel = "BEFORE"
			}

			filtered := missingLabels[:0]
			for _, l := range missingLabels {
				if l.label != s.label {
					filtered = append(filtered, l)
				}
			}

			missingLabels = filtered
			reachable[s.label] = stringNamer(s.label)
			ordered = append(ordered, fmt.Sprintf("ALTER TYPE %s ADD VALUE '%s' %s '%s';", expectedEnum.Name, s.label, rel, s.neighbor))
			continue outer
		}

		panic("Infinite loop")
	}

	return ordered, true
}
