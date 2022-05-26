package cliutil

import (
	"fmt"
	"net/url"
	"regexp"
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

func compareSchemaDescriptions(out *output.Output, schemaName string, actual, expected schemas.SchemaDescription) (err error) {
	for _, f := range []func(out *output.Output, schemaName string, actual, expected schemas.SchemaDescription) bool{
		compareExtensions,
		compareEnums,
		compareFunctions,
		compareSequences,
		compareTables,
		compareViews,
	} {
		if f(out, schemaName, actual, expected) {
			err = errOutOfSync
		}
	}

	if err == nil {
		out.WriteLine(output.Line(output.EmojiSuccess, output.StyleSuccess, "No drift detected"))
	}
	return err
}

func compareExtensions(out *output.Output, schemaName string, actual, expected schemas.SchemaDescription) (outOfSync bool) {
	compareExtension := func(extension *stringNamer, expectedExtension stringNamer) {
		outOfSync = true

		if extension == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing extension %q", expectedExtension)))
			writeSQLSolution(out, "install the extension", fmt.Sprintf("CREATE EXTENSION %s;", expectedExtension))
		}
	}

	compareNamedLists(wrapStrings(actual.Extensions), wrapStrings(expected.Extensions), compareExtension)
	return
}

func compareEnums(out *output.Output, schemaName string, actual, expected schemas.SchemaDescription) (outOfSync bool) {
	compareEnum := func(enum *schemas.EnumDescription, expectedEnum schemas.EnumDescription) {
		outOfSync = true

		quotedLabels := make([]string, 0, len(expectedEnum.Labels))
		for _, label := range expectedEnum.Labels {
			quotedLabels = append(quotedLabels, fmt.Sprintf("'%s'", label))
		}
		createEnumStmt := fmt.Sprintf("CREATE TYPE %s AS ENUM (%s);", expectedEnum.Name, strings.Join(quotedLabels, ", "))
		dropEnumStmt := fmt.Sprintf("DROP TYPE %s;", expectedEnum.Name)

		if enum == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing enum %q", expectedEnum.Name)))
			writeSQLSolution(out, "create the type", createEnumStmt)
		} else {
			if ordered, ok := constructEnumRepairStatements(*enum, expectedEnum); ok {
				out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing %d labels for enum %q", len(ordered), expectedEnum.Name)))
				writeSQLSolution(out, "add the missing enum labels", ordered...)
				return
			}

			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected labels for enum %q", expectedEnum.Name)))
			writeDiff(out, enum.Labels, expectedEnum.Labels)
			writeSQLSolution(out, "drop and re-create the type", dropEnumStmt, createEnumStmt)
		}
	}

	compareNamedLists(actual.Enums, expected.Enums, compareEnum)
	return outOfSync
}

func compareFunctions(out *output.Output, schemaName string, actual, expected schemas.SchemaDescription) (outOfSync bool) {
	compareFunction := func(function *schemas.FunctionDescription, expectedFunction schemas.FunctionDescription) {
		outOfSync = true

		definitionStmt := fmt.Sprintf("%s;", strings.TrimSpace(expectedFunction.Definition))

		if function == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing function %q", expectedFunction.Name)))
			writeSQLSolution(out, "define the function", definitionStmt)
		} else {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected definition of function %q", expectedFunction.Name)))
			writeDiff(out, expectedFunction.Definition, function.Definition)
			writeSQLSolution(out, "replace the function definition", definitionStmt)
		}
	}

	compareNamedLists(actual.Functions, expected.Functions, compareFunction)
	return outOfSync
}

func compareSequences(out *output.Output, schemaName string, actual, expected schemas.SchemaDescription) (outOfSync bool) {
	compareSequence := func(sequence *schemas.SequenceDescription, expectedSequence schemas.SequenceDescription) {
		outOfSync = true

		if sequence == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing sequence %q", expectedSequence.Name)))
		} else {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected properties of sequence %q", expectedSequence.Name)))
			writeDiff(out, expectedSequence, *sequence)
		}

		url := makeSearchURL(schemaName, fmt.Sprintf("CREATE SEQUENCE %s", expectedSequence.Name))
		writeSearchHint(out, fmt.Sprintf("define or redefine the sequence given the definition provided at %s", url))
	}

	compareNamedLists(actual.Sequences, expected.Sequences, compareSequence)
	return outOfSync
}

func compareTables(out *output.Output, schemaName string, actual, expected schemas.SchemaDescription) (outOfSync bool) {
	compareTables := func(table *schemas.TableDescription, expectedTable schemas.TableDescription) {
		outOfSync = true

		if table == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing table %q", expectedTable.Name)))

			url := makeSearchURL(schemaName, fmt.Sprintf("CREATE TABLE %s", expectedTable.Name))
			writeSearchHint(out, fmt.Sprintf("define the table given the definition provided at %s", url))
		} else {
			compareColumns(out, schemaName, *table, expectedTable)
			compareConstraints(out, schemaName, *table, expectedTable)
			compareIndexes(out, schemaName, *table, expectedTable)
			compareTriggers(out, schemaName, *table, expectedTable)
		}
	}

	compareNamedLists(actual.Tables, expected.Tables, compareTables)
	return outOfSync
}

func compareColumns(out *output.Output, schemaName string, actualTable, expectedTable schemas.TableDescription) {
	compareNamedLists(actualTable.Columns, expectedTable.Columns, func(column *schemas.ColumnDescription, expectedColumn schemas.ColumnDescription) {
		if column == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing column %q.%q", expectedTable.Name, expectedColumn.Name)))
		} else {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected properties of column %q.%q", expectedTable.Name, expectedColumn.Name)))
			writeDiff(out, expectedColumn, *column)
		}

		url := makeSearchURL(schemaName, fmt.Sprintf("CREATE TABLE %s", expectedTable.Name))
		writeSearchHint(out, fmt.Sprintf("define or redefine the column given the definition provided at %s", url))
	})
}

func compareConstraints(out *output.Output, schemaName string, actualTable, expectedTable schemas.TableDescription) {
	compareNamedLists(actualTable.Constraints, expectedTable.Constraints, func(constraint *schemas.ConstraintDescription, expectedConstraint schemas.ConstraintDescription) {
		createConstraintStmt := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s;", expectedTable.Name, expectedConstraint.Name, expectedConstraint.ConstraintDefinition)
		dropConstraintStmt := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;", expectedTable.Name, expectedConstraint.Name)

		if constraint == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing constraint %q.%q", expectedTable.Name, expectedConstraint.Name)))
			writeSQLSolution(out, "define the constraint", createConstraintStmt)
		} else {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected properties of constraint %q.%q", expectedTable.Name, expectedConstraint.Name)))
			writeDiff(out, expectedConstraint, *constraint)
			writeSQLSolution(out, "redefine the constraint", dropConstraintStmt, createConstraintStmt)
		}
	})
}

func compareIndexes(out *output.Output, schemaName string, actualTable, expectedTable schemas.TableDescription) {
	compareNamedLists(actualTable.Indexes, expectedTable.Indexes, func(index *schemas.IndexDescription, expectedIndex schemas.IndexDescription) {
		var createIndexStmt string
		var dropIndexStmt string

		switch expectedIndex.ConstraintType {
		case "u":
			fallthrough
		case "p":
			createIndexStmt = fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s;", actualTable.Name, expectedIndex.Name, expectedIndex.ConstraintDefinition)
			dropIndexStmt = fmt.Sprintf("DROP INDEX %s;", expectedIndex.Name)
		default:
			createIndexStmt = fmt.Sprintf("%s;", expectedIndex.IndexDefinition)
			dropIndexStmt = fmt.Sprintf("DROP INDEX %s;", expectedIndex.Name)
		}

		if index == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing index %q.%q", expectedTable.Name, expectedIndex.Name)))
			writeSQLSolution(out, "define the index", createIndexStmt)
		} else {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected properties of index %q.%q", expectedTable.Name, expectedIndex.Name)))
			writeDiff(out, expectedIndex, *index)
			writeSQLSolution(out, "redefine the index", dropIndexStmt, createIndexStmt)
		}
	})
}

func compareTriggers(out *output.Output, schemaName string, actualTable, expectedTable schemas.TableDescription) {
	compareNamedLists(actualTable.Triggers, expectedTable.Triggers, func(trigger *schemas.TriggerDescription, expectedTrigger schemas.TriggerDescription) {
		createTriggerStmt := fmt.Sprintf("%s;", expectedTrigger.Definition)
		dropTriggerStmt := fmt.Sprintf("DROP TRIGGER %s ON %s;", expectedTrigger.Name, expectedTable.Name)

		if trigger == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing trigger %q.%q", expectedTable.Name, expectedTrigger.Name)))
			writeSQLSolution(out, "define the trigger", createTriggerStmt)
		} else {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected properties of trigger %q.%q", expectedTable.Name, expectedTrigger.Name)))
			writeDiff(out, expectedTrigger, *trigger)
			writeSQLSolution(out, "redefine the trigger", dropTriggerStmt, createTriggerStmt)
		}
	})
}

func compareViews(out *output.Output, schemaName string, actual, expected schemas.SchemaDescription) (outOfSync bool) {
	compareView := func(view *schemas.ViewDescription, expectedView schemas.ViewDescription) {
		outOfSync = true

		// pgsql has weird indents here
		viewDefinition := strings.TrimSpace(stripIndent(" " + expectedView.Definition))
		createViewStmt := fmt.Sprintf("CREATE VIEW %s AS %s", expectedView.Name, viewDefinition)
		dropViewStmt := fmt.Sprintf("DROP VIEW %s;", expectedView.Name)

		if view == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing view %q", expectedView.Name)))
			writeSQLSolution(out, "define the view", createViewStmt)
		} else {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected definition of view %q", expectedView.Name)))
			writeDiff(out, expectedView.Definition, view.Definition)
			writeSQLSolution(out, "redefine the view", dropViewStmt, createViewStmt)
		}
	}

	compareNamedLists(actual.Views, expected.Views, compareView)
	return outOfSync
}

// writeDiff writes a colorized diff of the given objects.
func writeDiff(out *output.Output, a, b any) {
	out.WriteCode("diff", strings.TrimSpace(cmp.Diff(a, b)))
}

// writeSQLSolution writes a block of text containing the given solution deescription
// and the given SQL statements formatted (and colorized) as code.
func writeSQLSolution(out *output.Output, description string, statements ...string) {
	out.WriteLine(output.Line(output.EmojiLightbulb, output.StyleItalic, fmt.Sprintf("Suggested action: %s.", description)))
	out.WriteCode("sql", strings.Join(statements, "\n"))
}

// writeSearchHint writes a block of text containing the given hint description and
// instructions.
func writeSearchHint(out *output.Output, description string, statements ...string) {
	out.WriteLine(output.Line(output.EmojiLightbulb, output.StyleItalic, fmt.Sprintf("Hint: %s.", description)))
	for _, s := range statements {
		out.Write(s)
	}
	out.Write("")
}

// makeSearchURL returns a URL to a sourcegraph.com search query within the squashed
// definition of the given schema.
func makeSearchURL(schemaName, searchTerm string) string {
	terms := strings.Split(searchTerm, " ")
	for i, term := range terms {
		terms[i] = regexp.QuoteMeta(term)
	}

	queryParts := []string{
		`repo:^github\.com/sourcegraph/sourcegraph$`,
		fmt.Sprintf(`file:^migrations/%s/squashed\.sql$`, schemaName),
		"^" + strings.Join(terms, "\\s") + "\\b",
	}

	qs := url.Values{}
	qs.Add("patternType", "regexp")
	qs.Add("q", strings.Join(queryParts, " "))

	url, _ := url.Parse("https://sourcegraph.com/search")
	url.RawQuery = qs.Encode()
	return url.String()
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

// stripIndent removes the largest common indent from the given text.
func stripIndent(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")

	min := len(lines[0])
	for _, line := range lines {
		if indent := len(line) - len(strings.TrimLeft(line, " ")); indent < min {
			min = indent
		}
	}
	for i, line := range lines {
		lines[i] = line[min:]
	}

	return strings.Join(lines, "\n")
}
