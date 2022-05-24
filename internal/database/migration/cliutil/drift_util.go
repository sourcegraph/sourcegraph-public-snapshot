package cliutil

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// getSchemaJSONFilename returns the basename of the JSON-serialized schema in the sg/sg repository.
func getSchemaJSONFilename(schemaName string) (string, error) {
	switch schemaName {
	case "frontend":
		return "internal/database/schema.json", nil
	case "codeintel":
		fallthrough
	case "codeinsights":
		return fmt.Sprintf("internal/database/schema.%s.json", schemaName), nil
	}

	return "", errors.Newf("unknown schema name %q", schemaName)
}

var errOutOfSync = errors.Newf("database schema is out of sync")

func compareSchemaDescriptions(out *output.Output, actual, expected schemas.SchemaDescription) (err error) {
	for _, f := range []func(out *output.Output, actual, expected schemas.SchemaDescription) bool{
		compareExtensions,
		compareEnums,
		compareFunctions,
		compareSequences,
		compareTables,
		compareViews,
	} {
		if f(out, actual, expected) {
			err = errOutOfSync
		}
	}

	if err == nil {
		out.Write("No drift detected")
	}
	return err
}

func compareExtensions(out *output.Output, actual, expected schemas.SchemaDescription) (outOfSync bool) {
	compareExtension := func(extension *stringNamer, expectedExtension stringNamer) {
		outOfSync = true

		if extension == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing extension %q", expectedExtension)))
			out.WriteLine(output.Line(output.EmojiLightbulb, output.StyleItalic, fmt.Sprintf("Suggested action: install the extension:")))
			out.WriteMarkdown(fmt.Sprintf("```sql\nCREATE EXTENSION %s;\n```", expectedExtension))
		}
	}

	compareNamedLists(wrapStrings(actual.Extensions), wrapStrings(expected.Extensions), compareExtension)
	return
}

func compareEnums(out *output.Output, actual, expected schemas.SchemaDescription) (outOfSync bool) {
	compareEnum := func(enum *schemas.EnumDescription, expectedEnum schemas.EnumDescription) {
		outOfSync = true

		quotedLabels := make([]string, 0, len(expectedEnum.Labels))
		for _, label := range expectedEnum.Labels {
			quotedLabels = append(quotedLabels, fmt.Sprintf("'%s'", label))
		}

		if enum == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing enum %q", expectedEnum.Name)))
			out.WriteLine(output.Line(output.EmojiLightbulb, output.StyleItalic, fmt.Sprintf("Suggested action: create the type:")))
			out.WriteMarkdown(fmt.Sprintf("```sql\nCREATE TYPE %s AS ENUM (%s);\n```", expectedEnum.Name, strings.Join(quotedLabels, ", ")))
		} else {
			labels := groupByName(wrapStrings(enum.Labels))
			expectedLabels := groupByName(wrapStrings(expectedEnum.Labels))

			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected labels for enum %q", expectedEnum.Name)))
			out.WriteMarkdown(fmt.Sprintf("```diff\n%s\n```", cmp.Diff(enum.Labels, expectedEnum.Labels)))

			for label := range labels {
				if _, ok := expectedLabels[label]; !ok {
					out.WriteLine(output.Line(output.EmojiWarningSign, output.StyleBold, fmt.Sprintf("Type contains additional values - this type will need to be dropped and re-created")))
					out.WriteLine(output.Line(output.EmojiLightbulb, output.StyleItalic, fmt.Sprintf("Suggested action: drop and re-create the type:")))
					out.WriteMarkdown(fmt.Sprintf("```sql\nDROP TYPE %s;\nCREATE TYPE %s AS ENUM (%s);\n```", expectedEnum.Name, expectedEnum.Name, strings.Join(quotedLabels, ", ")))
					return
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

			out.WriteLine(output.Line(output.EmojiLightbulb, output.StyleItalic, fmt.Sprintf("Suggested action: add the missing enum labels")))
			out.WriteMarkdown(fmt.Sprintf("```sql\n%s\n```", strings.Join(ordered, "\n")))
		}
	}

	compareNamedLists(actual.Enums, expected.Enums, compareEnum)
	return outOfSync
}

func compareFunctions(out *output.Output, actual, expected schemas.SchemaDescription) (outOfSync bool) {
	compareFunction := func(function *schemas.FunctionDescription, expectedFunction schemas.FunctionDescription) {
		outOfSync = true

		if function == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing function %q", expectedFunction.Name)))
			out.WriteLine(output.Line(output.EmojiLightbulb, output.StyleItalic, fmt.Sprintf("Suggested action: define the function:")))
			// TODO - remove leading from function body; it doesn't work with the
			out.WriteMarkdown(fmt.Sprintf("```sql\n%s;\n```", strings.ReplaceAll(expectedFunction.Definition, "\\n", "\n")))
		} else {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected definition of function %q", expectedFunction.Name)))
			out.WriteMarkdown(fmt.Sprintf("```diff\n%s\n```", cmp.Diff(expectedFunction.Definition, function.Definition)))
			out.WriteLine(output.Line(output.EmojiLightbulb, output.StyleItalic, fmt.Sprintf("Suggested action: replace the function definition:")))
			// TODO - remove leading from function body; it doesn't work with the
			out.WriteMarkdown(fmt.Sprintf("```sql\n%s;\n```", strings.ReplaceAll(expectedFunction.Definition, "\\n", "\n")))
		}
	}

	compareNamedLists(actual.Functions, expected.Functions, compareFunction)
	return outOfSync
}

func compareSequences(out *output.Output, actual, expected schemas.SchemaDescription) (outOfSync bool) {
	compareSequence := func(sequence *schemas.SequenceDescription, expectedSequence schemas.SequenceDescription) {
		outOfSync = true

		if sequence == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing sequence %q", expectedSequence.Name)))
			out.WriteMarkdown(fmt.Sprintf("```diff\n%s\n```", cmp.Diff(expectedSequence, nil)))
			// TODO - suggest an action
		} else {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected properties of sequence %q", expectedSequence.Name)))
			out.WriteMarkdown(fmt.Sprintf("```diff\n%s\n```", cmp.Diff(expectedSequence, *sequence)))
			// TODO - suggest an action
		}
	}

	compareNamedLists(actual.Sequences, expected.Sequences, compareSequence)
	return outOfSync
}

func compareTables(out *output.Output, actual, expected schemas.SchemaDescription) (outOfSync bool) {
	compareTables := func(table *schemas.TableDescription, expectedTable schemas.TableDescription) {
		outOfSync = true

		if table == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing table %q", expectedTable.Name)))
			// TODO - suggest an action
		} else {
			compareColumns(out, *table, expectedTable)
			compareConstraints(out, *table, expectedTable)
			compareIndexes(out, *table, expectedTable)
			compareTriggers(out, *table, expectedTable)
		}
	}

	compareNamedLists(actual.Tables, expected.Tables, compareTables)
	return outOfSync
}

func compareColumns(out *output.Output, actualTable, expectedTable schemas.TableDescription) {
	compareNamedLists(actualTable.Columns, expectedTable.Columns, func(column *schemas.ColumnDescription, expectedColumn schemas.ColumnDescription) {
		if column == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing column %q.%q", expectedTable.Name, expectedColumn.Name)))
			out.WriteMarkdown(fmt.Sprintf("```diff\n%s\n```", cmp.Diff(expectedColumn, nil)))
			// TODO - suggest an action
		} else {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected properties of column %q.%q", expectedTable.Name, expectedColumn.Name)))
			out.WriteMarkdown(fmt.Sprintf("```diff\n%s\n```", cmp.Diff(expectedColumn, *column)))
			// TODO - suggest an action
		}
	})
}

func compareConstraints(out *output.Output, actualTable, expectedTable schemas.TableDescription) {
	compareNamedLists(actualTable.Constraints, expectedTable.Constraints, func(constraint *schemas.ConstraintDescription, expectedConstraint schemas.ConstraintDescription) {
		if constraint == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing constraint %q.%q", expectedTable.Name, expectedConstraint.Name)))
			out.WriteMarkdown(fmt.Sprintf("```diff\n%s\n```", cmp.Diff(expectedConstraint, nil)))
			// TODO - suggest an action
		} else {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected properties of constraint %q.%q", expectedTable.Name, expectedConstraint.Name)))
			out.WriteMarkdown(fmt.Sprintf("```diff\n%s\n```", cmp.Diff(expectedConstraint, *constraint)))
			// TODO - suggest an action
		}
	})
}

func compareIndexes(out *output.Output, actualTable, expectedTable schemas.TableDescription) {
	compareNamedLists(actualTable.Indexes, expectedTable.Indexes, func(index *schemas.IndexDescription, expectedIndex schemas.IndexDescription) {
		if index == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing index %q.%q", expectedTable.Name, expectedIndex.Name)))
			out.WriteMarkdown(fmt.Sprintf("```diff\n%s\n```", cmp.Diff(expectedIndex, nil)))
			out.WriteLine(output.Line(output.EmojiLightbulb, output.StyleItalic, fmt.Sprintf("Suggested action: define the index:")))
			out.WriteMarkdown(fmt.Sprintf("```sql\n%s;\n```", expectedIndex.IndexDefinition))
		} else {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected properties of index %q.%q", expectedTable.Name, expectedIndex.Name)))
			out.WriteMarkdown(fmt.Sprintf("```diff\n%s\n```", cmp.Diff(expectedIndex, *index)))
			// TODO - suggest an action
		}
	})
}

func compareTriggers(out *output.Output, actualTable, expectedTable schemas.TableDescription) {
	compareNamedLists(actualTable.Triggers, expectedTable.Triggers, func(trigger *schemas.TriggerDescription, expectedTrigger schemas.TriggerDescription) {
		if trigger == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing trigger %q.%q", expectedTable.Name, expectedTrigger.Name)))
			out.WriteMarkdown(fmt.Sprintf("```diff\n%s\n```", cmp.Diff(expectedTrigger, nil)))
			// TODO - suggest an action
		} else {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected properties of trigger %q.%q", expectedTable.Name, expectedTrigger.Name)))
			out.WriteMarkdown(fmt.Sprintf("```diff\n%s\n```", cmp.Diff(expectedTrigger, *trigger)))
			// TODO - suggest an action
		}
	})
}

func compareViews(out *output.Output, actual, expected schemas.SchemaDescription) (outOfSync bool) {
	compareView := func(view *schemas.ViewDescription, expectedView schemas.ViewDescription) {
		outOfSync = true

		if view == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing view %q", expectedView.Name)))
			out.WriteLine(output.Line(output.EmojiLightbulb, output.StyleItalic, fmt.Sprintf("Suggested action: define the view:")))
			out.WriteMarkdown(fmt.Sprintf("```sql\n CREATE VIEW %s AS\n%s\n```", expectedView.Name, strings.ReplaceAll(expectedView.Definition, "\\n", "\n")))
		} else {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected definition of view %q", expectedView.Name)))
			out.WriteMarkdown(fmt.Sprintf("```diff\n%s\n```", cmp.Diff(expectedView.Definition, view.Definition)))
			out.WriteLine(output.Line(output.EmojiLightbulb, output.StyleItalic, fmt.Sprintf("Suggested action: redefine the view:")))
			out.WriteMarkdown(fmt.Sprintf("```sql\n DROP VIEW %s;\n CREATE VIEW %s AS\n%s\n```", expectedView.Name, expectedView.Name, strings.ReplaceAll(expectedView.Definition, "\\n", "\n")))
		}
	}

	compareNamedLists(actual.Views, expected.Views, compareView)
	return outOfSync
}

// compareNamedLists invokes the given callback with a pair of elements from slices
// `as` and `bs`, respectively, with the same name. If there is a missing element from
// `as`, there will be an invocation of `f` with a nil value for its first parameter.
func compareNamedLists[T schemas.Namer](as, bs []T, f func(a *T, b T)) {
	am := groupByName(as)
	bm := groupByName(bs)

	for _, k := range keys(am) {
		av := am[k]

		if bv, ok := bm[k]; ok {
			if cmp.Diff(av, bv) != "" {
				f(&av, bv)
			}
		} else {
			// f(&av, nil)
		}
	}
	for _, k := range keys(bm) {
		bv := bm[k]

		if _, ok := am[k]; !ok {
			f(nil, bv)
		}
	}
}

// groupByName converts the given element slice into a map indexed by
// each element's name.
func groupByName[T schemas.Namer](ts []T) map[string]T {
	m := make(map[string]T, len(ts))
	for _, t := range ts {
		m[t.GetName()] = t
	}

	return m
}

// keys returns the ordered keys of the given map.
func keys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

type stringNamer string

func (s stringNamer) GetName() string { return string(s) }

// wrapStrings converts a string slice into a string slice with GetName
// on each element.
func wrapStrings(ss []string) []stringNamer {
	sn := make([]stringNamer, 0, len(ss))
	for _, s := range ss {
		sn = append(sn, stringNamer(s))
	}

	return sn
}
