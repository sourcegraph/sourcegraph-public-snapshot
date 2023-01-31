package cliutil

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/regexp"

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

func compareSchemaDescriptions(rawOut *output.Output, schemaName, version string, actual, expected schemas.SchemaDescription) (err error) {
	out := &preambledOutput{rawOut, false}

	for _, f := range []func(out *preambledOutput, schemaName, version string, actual, expected schemas.SchemaDescription) bool{
		compareExtensions,
		compareEnums,
		compareFunctions,
		compareSequences,
		compareTables,
		compareViews,
	} {
		if f(out, schemaName, version, actual, expected) {
			err = errOutOfSync
		}
	}

	if err == nil {
		rawOut.WriteLine(output.Line(output.EmojiSuccess, output.StyleSuccess, "No drift detected"))
	}
	return err
}

func compareExtensions(out *preambledOutput, schemaName, version string, actual, expected schemas.SchemaDescription) bool {
	return compareNamedLists(wrapStrings(actual.Extensions), wrapStrings(expected.Extensions), func(extension *stringNamer, expectedExtension stringNamer) bool {
		if extension == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing extension %q", expectedExtension)))
			writeSQLSolution(out, "install the extension", fmt.Sprintf("CREATE EXTENSION %s;", expectedExtension))
			return true
		}

		return false
	}, noopAdditionalCallback[stringNamer])
}

func compareEnums(out *preambledOutput, schemaName, version string, actual, expected schemas.SchemaDescription) bool {
	return compareNamedLists(actual.Enums, expected.Enums, func(enum *schemas.EnumDescription, expectedEnum schemas.EnumDescription) bool {
		quotedLabels := make([]string, 0, len(expectedEnum.Labels))
		for _, label := range expectedEnum.Labels {
			quotedLabels = append(quotedLabels, fmt.Sprintf("'%s'", label))
		}
		createEnumStmt := fmt.Sprintf("CREATE TYPE %s AS ENUM (%s);", expectedEnum.Name, strings.Join(quotedLabels, ", "))
		dropEnumStmt := fmt.Sprintf("DROP TYPE %s;", expectedEnum.Name)

		if enum == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing enum %q", expectedEnum.Name)))
			writeSQLSolution(out, "create the type", createEnumStmt)
			return true
		}

		if ordered, ok := constructEnumRepairStatements(*enum, expectedEnum); ok {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing %d labels for enum %q", len(ordered), expectedEnum.Name)))
			writeSQLSolution(out, "add the missing enum labels", ordered...)
			return true
		}

		out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected labels for enum %q", expectedEnum.Name)))
		writeDiff(out, enum.Labels, expectedEnum.Labels)
		writeSQLSolution(out, "drop and re-create the type", dropEnumStmt, createEnumStmt)
		return true
	}, noopAdditionalCallback[schemas.EnumDescription])
}

func compareFunctions(out *preambledOutput, schemaName, version string, actual, expected schemas.SchemaDescription) bool {
	return compareNamedLists(actual.Functions, expected.Functions, func(function *schemas.FunctionDescription, expectedFunction schemas.FunctionDescription) bool {
		definitionStmt := fmt.Sprintf("%s;", strings.TrimSpace(expectedFunction.Definition))

		if function == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing function %q", expectedFunction.Name)))
			writeSQLSolution(out, "define the function", definitionStmt)
			return true
		}

		out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected definition of function %q", expectedFunction.Name)))
		writeDiff(out, expectedFunction.Definition, function.Definition)
		writeSQLSolution(out, "replace the function definition", definitionStmt)
		return true
	}, noopAdditionalCallback[schemas.FunctionDescription])
}

func compareSequences(out *preambledOutput, schemaName, version string, actual, expected schemas.SchemaDescription) bool {
	return compareNamedLists(actual.Sequences, expected.Sequences, func(sequence *schemas.SequenceDescription, expectedSequence schemas.SequenceDescription) bool {
		definitionStmt := makeSearchURL(schemaName, version,
			fmt.Sprintf("CREATE SEQUENCE %s", expectedSequence.Name),
			fmt.Sprintf("nextval('%s'::regclass);", expectedSequence.Name),
		)

		if sequence == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing sequence %q", expectedSequence.Name)))
			writeSearchHint(out, "define the sequence", definitionStmt)
			return true
		}

		out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected properties of sequence %q", expectedSequence.Name)))
		writeDiff(out, expectedSequence, *sequence)
		writeSearchHint(out, "redefine the sequence", definitionStmt)
		return true
	}, noopAdditionalCallback[schemas.SequenceDescription])
}

func compareTables(out *preambledOutput, schemaName, version string, actual, expected schemas.SchemaDescription) bool {
	return compareNamedLists(actual.Tables, expected.Tables, func(table *schemas.TableDescription, expectedTable schemas.TableDescription) bool {
		if table == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing table %q", expectedTable.Name)))
			writeSearchHint(out, "define the table", makeSearchURL(schemaName, version,
				fmt.Sprintf("CREATE TABLE %s", expectedTable.Name),
				fmt.Sprintf("ALTER TABLE ONLY %s", expectedTable.Name),
				fmt.Sprintf("CREATE .*(INDEX|TRIGGER).* ON %s", expectedTable.Name),
			))
			return true
		}

		outOfSync := false
		outOfSync = compareColumns(out, schemaName, version, *table, expectedTable) || outOfSync
		outOfSync = compareConstraints(out, *table, expectedTable) || outOfSync
		outOfSync = compareIndexes(out, *table, expectedTable) || outOfSync
		outOfSync = compareTriggers(out, *table, expectedTable) || outOfSync
		outOfSync = compareTableComments(out, *table, expectedTable) || outOfSync
		return outOfSync
	}, noopAdditionalCallback[schemas.TableDescription])
}

func compareColumns(out *preambledOutput, schemaName, version string, actualTable, expectedTable schemas.TableDescription) bool {
	return compareNamedLists(actualTable.Columns, expectedTable.Columns, func(column *schemas.ColumnDescription, expectedColumn schemas.ColumnDescription) bool {
		if column == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing column %q.%q", expectedTable.Name, expectedColumn.Name)))
			writeSearchHint(out, "define the column", makeSearchURL(schemaName, version,
				fmt.Sprintf("CREATE TABLE %s", expectedTable.Name),
				fmt.Sprintf("ALTER TABLE ONLY %s", expectedTable.Name),
			))
			return true
		}

		out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected properties of column %q.%q", expectedTable.Name, expectedColumn.Name)))
		writeDiff(out, expectedColumn, *column)

		equivIf := func(f func(*schemas.ColumnDescription)) bool {
			c := *column
			f(&c)
			return cmp.Diff(c, expectedColumn) == ""
		}

		// TODO
		// if equivIf(func(s *schemas.ColumnDescription) { s.TypeName = expectedColumn.TypeName }) {}
		if equivIf(func(s *schemas.ColumnDescription) { s.IsNullable = expectedColumn.IsNullable }) {
			var verb string
			if expectedColumn.IsNullable {
				verb = "DROP"
			} else {
				verb = "SET"
			}

			nullabilityStmt := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s %s NOT NULL;", expectedTable.Name, expectedColumn.Name, verb)
			writeSQLSolution(out, "change the column nullability constraint", nullabilityStmt)
			return true
		}
		if equivIf(func(s *schemas.ColumnDescription) { s.Default = expectedColumn.Default }) {
			setDefaultStmt := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s;", expectedTable.Name, expectedColumn.Name, expectedColumn.Default)
			writeSQLSolution(out, "change the column default", setDefaultStmt)
			return true
		}
		if equivIf(func(s *schemas.ColumnDescription) { s.Comment = expectedColumn.Comment }) {
			setDefaultStmt := fmt.Sprintf("COMMENT ON COLUMN %s.%s IS '%s';", expectedTable.Name, expectedColumn.Name, expectedColumn.Comment)
			writeSQLSolution(out, "change the column comment", setDefaultStmt)
			return true
		}

		writeSearchHint(out, "redefine the column", makeSearchURL(schemaName, version,
			fmt.Sprintf("CREATE TABLE %s", expectedTable.Name),
			fmt.Sprintf("ALTER TABLE ONLY %s", expectedTable.Name),
		))
		return true
	}, func(additional []schemas.ColumnDescription) bool {
		for _, column := range additional {
			dropColumnStmt := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;", expectedTable.Name, column.Name)
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected column %q.%q", expectedTable.Name, column.Name)))
			writeSQLSolution(out, "drop the column", dropColumnStmt)
		}

		return true
	})
}

func compareConstraints(out *preambledOutput, actualTable, expectedTable schemas.TableDescription) bool {
	return compareNamedLists(actualTable.Constraints, expectedTable.Constraints, func(constraint *schemas.ConstraintDescription, expectedConstraint schemas.ConstraintDescription) bool {
		createConstraintStmt := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s;", expectedTable.Name, expectedConstraint.Name, expectedConstraint.ConstraintDefinition)
		dropConstraintStmt := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;", expectedTable.Name, expectedConstraint.Name)

		if constraint == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing constraint %q.%q", expectedTable.Name, expectedConstraint.Name)))
			writeSQLSolution(out, "define the constraint", createConstraintStmt)
			return true
		}

		out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected properties of constraint %q.%q", expectedTable.Name, expectedConstraint.Name)))
		writeDiff(out, expectedConstraint, *constraint)
		writeSQLSolution(out, "redefine the constraint", dropConstraintStmt, createConstraintStmt)
		return true
	}, func(additional []schemas.ConstraintDescription) bool {
		for _, constraint := range additional {
			dropConstraintStmt := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;", expectedTable.Name, constraint.Name)
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected constraint %q.%q", expectedTable.Name, constraint.Name)))
			writeSQLSolution(out, "drop the constraint", dropConstraintStmt)
		}

		return true
	})
}

func compareIndexes(out *preambledOutput, actualTable, expectedTable schemas.TableDescription) bool {
	return compareNamedLists(actualTable.Indexes, expectedTable.Indexes, func(index *schemas.IndexDescription, expectedIndex schemas.IndexDescription) bool {
		var createIndexStmt string
		switch expectedIndex.ConstraintType {
		case "u":
			fallthrough
		case "p":
			createIndexStmt = fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s;", actualTable.Name, expectedIndex.Name, expectedIndex.ConstraintDefinition)
		default:
			createIndexStmt = fmt.Sprintf("%s;", expectedIndex.IndexDefinition)
		}

		if index == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing index %q.%q", expectedTable.Name, expectedIndex.Name)))
			writeSQLSolution(out, "define the index", createIndexStmt)
			return true
		}

		dropIndexStmt := fmt.Sprintf("DROP INDEX %s;", expectedIndex.Name)
		out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected properties of index %q.%q", expectedTable.Name, expectedIndex.Name)))
		writeDiff(out, expectedIndex, *index)
		writeSQLSolution(out, "redefine the index", dropIndexStmt, createIndexStmt)
		return true
	}, func(additional []schemas.IndexDescription) bool {
		for _, index := range additional {
			dropIndexStmt := fmt.Sprintf("DROP INDEX %s;", index.Name)
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected index %q.%q", expectedTable.Name, index.Name)))
			writeSQLSolution(out, "drop the index", dropIndexStmt)
		}

		return true
	})
}

func compareTriggers(out *preambledOutput, actualTable, expectedTable schemas.TableDescription) bool {
	return compareNamedLists(actualTable.Triggers, expectedTable.Triggers, func(trigger *schemas.TriggerDescription, expectedTrigger schemas.TriggerDescription) bool {
		createTriggerStmt := fmt.Sprintf("%s;", expectedTrigger.Definition)
		dropTriggerStmt := fmt.Sprintf("DROP TRIGGER %s ON %s;", expectedTrigger.Name, expectedTable.Name)

		if trigger == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing trigger %q.%q", expectedTable.Name, expectedTrigger.Name)))
			writeSQLSolution(out, "define the trigger", createTriggerStmt)
			return true
		}

		out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected properties of trigger %q.%q", expectedTable.Name, expectedTrigger.Name)))
		writeDiff(out, expectedTrigger, *trigger)
		writeSQLSolution(out, "redefine the trigger", dropTriggerStmt, createTriggerStmt)
		return true
	}, func(additional []schemas.TriggerDescription) bool {
		for _, trigger := range additional {
			dropTriggerStmt := fmt.Sprintf("DROP TRIGGER %s ON %s;", trigger.Name, expectedTable.Name)
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected trigger %q.%q", expectedTable.Name, trigger.Name)))
			writeSQLSolution(out, "drop the trigger", dropTriggerStmt)
		}

		return true

	})
}

func compareTableComments(out *preambledOutput, actualTable, expectedTable schemas.TableDescription) bool {
	if actualTable.Comment != expectedTable.Comment {
		out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected comment of table %q", expectedTable.Name)))
		setDefaultStmt := fmt.Sprintf("COMMENT ON TABLE %s IS '%s';", expectedTable.Name, expectedTable.Comment)
		writeSQLSolution(out, "change the table comment", setDefaultStmt)
		return true
	}

	return false
}

func compareViews(out *preambledOutput, schemaName, version string, actual, expected schemas.SchemaDescription) bool {
	return compareNamedLists(actual.Views, expected.Views, func(view *schemas.ViewDescription, expectedView schemas.ViewDescription) bool {
		// pgsql has weird indents here
		viewDefinition := strings.TrimSpace(stripIndent(" " + expectedView.Definition))
		createViewStmt := fmt.Sprintf("CREATE VIEW %s AS %s", expectedView.Name, viewDefinition)
		dropViewStmt := fmt.Sprintf("DROP VIEW %s;", expectedView.Name)

		if view == nil {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Missing view %q", expectedView.Name)))
			writeSQLSolution(out, "define the view", createViewStmt)
			return true
		}

		out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, fmt.Sprintf("Unexpected definition of view %q", expectedView.Name)))
		writeDiff(out, expectedView.Definition, view.Definition)
		writeSQLSolution(out, "redefine the view", dropViewStmt, createViewStmt)
		return true
	}, noopAdditionalCallback[schemas.ViewDescription])
}

func noopAdditionalCallback[T schemas.Namer](_ []T) bool {
	return false
}

// writeDiff writes a colorized diff of the given objects.
func writeDiff(out *preambledOutput, a, b any) {
	out.WriteCode("diff", strings.TrimSpace(cmp.Diff(a, b)))
}

// writeSQLSolution writes a block of text containing the given solution deescription
// and the given SQL statements formatted (and colorized) as code.
func writeSQLSolution(out *preambledOutput, description string, statements ...string) {
	out.WriteLine(output.Line(output.EmojiLightbulb, output.StyleItalic, fmt.Sprintf("Suggested action: %s.", description)))
	out.WriteCode("sql", strings.Join(statements, "\n"))
}

// writeSearchHint writes a block of text containing the given hint description and
// a link to a set of Sourcegraph search results relevant to the missing or unexpected
// object definition.
func writeSearchHint(out *preambledOutput, description, url string) {
	out.WriteLine(output.Line(output.EmojiLightbulb, output.StyleItalic, fmt.Sprintf("Hint: %s using the definition at the following URL:", description)))
	out.Write("")
	out.WriteLine(output.Line(output.EmojiFingerPointRight, output.StyleUnderline, url))
	out.Write("")
}

// makeSearchURL returns a URL to a sourcegraph.com search query within the squashed
// definition of the given schema.
func makeSearchURL(schemaName, version string, searchTerms ...string) string {
	terms := make([]string, 0, len(searchTerms))
	for _, searchTerm := range searchTerms {
		terms = append(terms, quoteTerm(searchTerm))
	}

	queryParts := []string{
		fmt.Sprintf(`repo:^github\.com/sourcegraph/sourcegraph$@%s`, version),
		fmt.Sprintf(`file:^migrations/%s/squashed\.sql$`, schemaName),
		strings.Join(terms, " OR "),
	}

	qs := url.Values{}
	qs.Add("patternType", "regexp")
	qs.Add("q", strings.Join(queryParts, " "))

	searchUrl, _ := url.Parse("https://sourcegraph.com/search")
	searchUrl.RawQuery = qs.Encode()
	return searchUrl.String()
}

// quoteTerm converts the given literal search term into a regular expression.
func quoteTerm(searchTerm string) string {
	terms := strings.Split(searchTerm, " ")
	for i, term := range terms {
		terms[i] = regexp.QuoteMeta(term)
	}

	return "(^|\\b)" + strings.Join(terms, "\\s") + "($|\\b)"
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

// compareNamedLists invokes the given primary callback with a pair of differing elements from slices
// `as` and `bs`, respectively, with the same name. If there is a missing element from `as`, there will
// be an invocation of this callback with a nil value for its first parameter. Elements for which there
// is no analog in `bs` will be collected and sent to an invocation of the additions callback. If any
// invocation of either function returns true, the output of this function will be true.
func compareNamedLists[T schemas.Namer](
	as []T,
	bs []T,
	primaryCallback func(a *T, b T) bool,
	additionsCallback func(additional []T) bool,
) (outOfSync bool) {
	am := groupByName(as)
	bm := groupByName(bs)
	additional := make([]T, 0, len(am))

	for _, k := range keys(am) {
		av := am[k]

		if bv, ok := bm[k]; ok {
			if cmp.Diff(schemas.Normalize(av), schemas.Normalize(bv)) != "" {
				if primaryCallback(&av, bv) {
					outOfSync = true
				}
			}
		} else {
			additional = append(additional, av)
		}
	}
	for _, k := range keys(bm) {
		bv := bm[k]

		if _, ok := am[k]; !ok {
			if primaryCallback(nil, bv) {
				outOfSync = true
			}
		}
	}

	if len(additional) > 0 {
		if additionsCallback(additional) {
			outOfSync = true
		}
	}

	return
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

//
// Output

type preambledOutput struct {
	out     *output.Output
	emitted bool
}

func (o *preambledOutput) check() {
	if o.emitted {
		return
	}

	o.out.WriteLine(output.Line(output.EmojiFailure, output.StyleFailure, "Drift detected!"))
	o.out.Write("")
	o.emitted = true
}

func (o *preambledOutput) Write(s string) {
	o.check()
	o.out.Write(s)
}

func (o *preambledOutput) Writef(format string, args ...any) {
	o.check()
	o.out.Writef(format, args...)
}

func (o *preambledOutput) WriteLine(line output.FancyLine) {
	o.check()
	o.out.WriteLine(line)
}

func (o *preambledOutput) WriteCode(languageName, str string) error {
	o.check()
	return o.out.WriteCode(languageName, str)
}
