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

func compareAndDisplaySchemaDescriptions(rawOut *output.Output, schemaName, version string, actual, expected schemas.SchemaDescription) (err error) {
	out := &preambledOutput{rawOut, false}
	for _, drift := range compareSchemaDescriptions(schemaName, version, actual, expected) {
		drift.Display(out)
		err = errOutOfSync
	}

	if err == nil {
		rawOut.WriteLine(output.Line(output.EmojiSuccess, output.StyleSuccess, "No drift detected"))
	}
	return err
}

type DriftSummary interface {
	Display(out OutputWriter)
}

type driftSummary struct {
	name     string
	problem  string
	solution string

	hasDiff bool
	a, b    any

	hasStatements bool
	statements    []string

	hasURLHint bool
	url        string
}

func wrap(summary DriftSummary) []DriftSummary {
	return []DriftSummary{summary}
}

func (s *driftSummary) Display(out OutputWriter) {
	out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, s.problem))
	if s.hasDiff {
		_ = out.WriteCode("diff", strings.TrimSpace(cmp.Diff(s.a, s.b)))
	}

	out.WriteLine(output.Line(output.EmojiLightbulb, output.StyleItalic, fmt.Sprintf("Suggested action: %s.", s.solution)))

	if s.hasStatements {
		_ = out.WriteCode("sql", strings.Join(s.statements, "\n"))
	}

	if s.hasURLHint {
		out.WriteLine(output.Line(output.EmojiLightbulb, output.StyleItalic, fmt.Sprintf("Hint: Reproduce %s as defined at the following URL:", s.name)))
		out.Write("")
		out.WriteLine(output.Line(output.EmojiFingerPointRight, output.StyleUnderline, s.url))
		out.Write("")
	}
}

func newDriftSummary(name string, problem, solution string) *driftSummary {
	return &driftSummary{
		name:     name,
		problem:  problem,
		solution: solution,
	}
}

func (s *driftSummary) withDiff(a, b any) *driftSummary {
	s.hasDiff = true
	s.a, s.b = a, b
	return s
}

func (s *driftSummary) withStatements(statements ...string) *driftSummary {
	s.hasStatements = true
	s.statements = statements
	return s
}

func (s *driftSummary) withURLHint(url string) *driftSummary {
	s.hasURLHint = true
	s.url = url
	return s
}

func compareSchemaDescriptions(schemaName, version string, actual, expected schemas.SchemaDescription) []DriftSummary {
	s := []DriftSummary{}
	for _, f := range []func(schemaName, version string, actual, expected schemas.SchemaDescription) []DriftSummary{
		compareExtensions,
		compareEnums,
		compareFunctions,
		compareSequences,
		compareTables,
		compareViews,
	} {
		s = append(s, f(schemaName, version, actual, expected)...)
	}

	return s
}

func compareExtensions(schemaName, version string, actual, expected schemas.SchemaDescription) []DriftSummary {
	return compareNamedLists(wrapStrings(actual.Extensions), wrapStrings(expected.Extensions), func(extension *stringNamer, expectedExtension stringNamer) DriftSummary {
		if extension == nil {
			createExtensionStmt := fmt.Sprintf("CREATE EXTENSION %s;", expectedExtension)

			return newDriftSummary(
				expectedExtension.GetName(),
				fmt.Sprintf("Missing extension %q", expectedExtension),
				"install the extension",
			).withStatements(createExtensionStmt)
		}

		return nil
	}, noopAdditionalCallback[stringNamer])
}

func compareEnums(schemaName, version string, actual, expected schemas.SchemaDescription) []DriftSummary {
	return compareNamedLists(actual.Enums, expected.Enums, func(enum *schemas.EnumDescription, expectedEnum schemas.EnumDescription) DriftSummary {
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

func compareFunctions(schemaName, version string, actual, expected schemas.SchemaDescription) []DriftSummary {
	return compareNamedLists(actual.Functions, expected.Functions, func(function *schemas.FunctionDescription, expectedFunction schemas.FunctionDescription) DriftSummary {
		definitionStmt := fmt.Sprintf("%s;", strings.TrimSpace(expectedFunction.Definition))

		if function == nil {
			return newDriftSummary(
				expectedFunction.Name,
				fmt.Sprintf("Missing function %q", expectedFunction.Name),
				"define the function",
			).withStatements(definitionStmt)
		}

		return newDriftSummary(
			expectedFunction.Name,
			fmt.Sprintf("Unexpected definition of function %q", expectedFunction.Name),
			"replace the function definition",
		).withDiff(expectedFunction.Definition, function.Definition).withStatements(definitionStmt)
	}, noopAdditionalCallback[schemas.FunctionDescription])
}

func compareSequences(schemaName, version string, actual, expected schemas.SchemaDescription) []DriftSummary {
	return compareNamedLists(actual.Sequences, expected.Sequences, func(sequence *schemas.SequenceDescription, expectedSequence schemas.SequenceDescription) DriftSummary {
		definitionStmt := makeSearchURL(schemaName, version,
			fmt.Sprintf("CREATE SEQUENCE %s", expectedSequence.Name),
			fmt.Sprintf("nextval('%s'::regclass);", expectedSequence.Name),
		)

		if sequence == nil {
			return newDriftSummary(
				expectedSequence.Name,
				fmt.Sprintf("Missing sequence %q", expectedSequence.Name),
				"define the sequence",
			).withStatements(definitionStmt)
		}

		return newDriftSummary(
			expectedSequence.Name,
			fmt.Sprintf("Unexpected properties of sequence %q", expectedSequence.Name),
			"redefine the sequence",
		).withDiff(expectedSequence, *sequence).withStatements(definitionStmt)
	}, noopAdditionalCallback[schemas.SequenceDescription])
}

func compareTables(schemaName, version string, actual, expected schemas.SchemaDescription) []DriftSummary {
	return compareNamedListsMulti(actual.Tables, expected.Tables, func(table *schemas.TableDescription, expectedTable schemas.TableDescription) []DriftSummary {
		if table == nil {
			url := makeSearchURL(schemaName, version,
				fmt.Sprintf("CREATE TABLE %s", expectedTable.Name),
				fmt.Sprintf("ALTER TABLE ONLY %s", expectedTable.Name),
				fmt.Sprintf("CREATE .*(INDEX|TRIGGER).* ON %s", expectedTable.Name),
			)

			return wrap(newDriftSummary(
				expectedTable.Name,
				fmt.Sprintf("Missing table %q", expectedTable.Name),
				"define the table",
			).withURLHint(url))
		}

		summaries := []DriftSummary(nil)
		summaries = append(summaries, compareColumns(schemaName, version, *table, expectedTable)...)
		summaries = append(summaries, compareConstraints(*table, expectedTable)...)
		summaries = append(summaries, compareIndexes(*table, expectedTable)...)
		summaries = append(summaries, compareTriggers(*table, expectedTable)...)
		return summaries
	}, noopAdditionalCallback[schemas.TableDescription])
}

func compareColumns(schemaName, version string, actualTable, expectedTable schemas.TableDescription) []DriftSummary {
	return compareNamedLists(actualTable.Columns, expectedTable.Columns, func(column *schemas.ColumnDescription, expectedColumn schemas.ColumnDescription) DriftSummary {
		if column == nil {
			url := makeSearchURL(schemaName, version,
				fmt.Sprintf("CREATE TABLE %s", expectedTable.Name),
				fmt.Sprintf("ALTER TABLE ONLY %s", expectedTable.Name),
			)

			return newDriftSummary(
				fmt.Sprintf("%q.%q", expectedTable.Name, expectedColumn.Name),
				fmt.Sprintf("Missing column %q.%q", expectedTable.Name, expectedColumn.Name),
				"define the column",
			).withURLHint(url)
		}

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

			alterColumnStmt := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s %s NOT NULL;", expectedTable.Name, expectedColumn.Name, verb)

			return newDriftSummary(
				fmt.Sprintf("%q.%q", expectedTable.Name, expectedColumn.Name),
				fmt.Sprintf("Unexpected properties of column %q.%q", expectedTable.Name, expectedColumn.Name),
				"change the column nullability constraint",
			).withDiff(expectedColumn, *column).withStatements(alterColumnStmt)
		}
		if equivIf(func(s *schemas.ColumnDescription) { s.Default = expectedColumn.Default }) {
			alterColumnStmt := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s;", expectedTable.Name, expectedColumn.Name, expectedColumn.Default)

			return newDriftSummary(
				fmt.Sprintf("%q.%q", expectedTable.Name, expectedColumn.Name),
				fmt.Sprintf("Unexpected properties of column %q.%q", expectedTable.Name, expectedColumn.Name),
				"change the column default",
			).withDiff(expectedColumn, *column).withStatements(alterColumnStmt)
		}

		url := makeSearchURL(schemaName, version,
			fmt.Sprintf("CREATE TABLE %s", expectedTable.Name),
			fmt.Sprintf("ALTER TABLE ONLY %s", expectedTable.Name),
		)

		return newDriftSummary(
			fmt.Sprintf("%q.%q", expectedTable.Name, expectedColumn.Name),
			fmt.Sprintf("Unexpected properties of column %q.%q", expectedTable.Name, expectedColumn.Name),
			"redefine the column",
		).withDiff(expectedColumn, *column).withURLHint(url)
	}, func(additional []schemas.ColumnDescription) []DriftSummary {
		summaries := []DriftSummary{}
		for _, column := range additional {
			alterColumnStmt := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;", expectedTable.Name, column.Name)

			summary := newDriftSummary(
				fmt.Sprintf("%q.%q", expectedTable.Name, column.Name),
				fmt.Sprintf("Unexpected column %q.%q", expectedTable.Name, column.Name),
				"drop the column",
			).withStatements(alterColumnStmt)
			summaries = append(summaries, summary)
		}

		return summaries
	})
}

func compareConstraints(actualTable, expectedTable schemas.TableDescription) []DriftSummary {
	return compareNamedLists(actualTable.Constraints, expectedTable.Constraints, func(constraint *schemas.ConstraintDescription, expectedConstraint schemas.ConstraintDescription) DriftSummary {
		createConstraintStmt := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s;", expectedTable.Name, expectedConstraint.Name, expectedConstraint.ConstraintDefinition)
		dropConstraintStmt := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;", expectedTable.Name, expectedConstraint.Name)

		if constraint == nil {
			return newDriftSummary(
				fmt.Sprintf("%q.%q", expectedTable.Name, expectedConstraint.Name),
				fmt.Sprintf("Missing constraint %q.%q", expectedTable.Name, expectedConstraint.Name),
				"define the constraint",
			).withStatements(createConstraintStmt)
		}

		return newDriftSummary(
			fmt.Sprintf("%q.%q", expectedTable.Name, expectedConstraint.Name),
			fmt.Sprintf("Unexpected properties of constraint %q.%q", expectedTable.Name, expectedConstraint.Name),
			"redefine the constraint",
		).withDiff(expectedConstraint, *constraint).withStatements(dropConstraintStmt, createConstraintStmt)
	}, func(additional []schemas.ConstraintDescription) []DriftSummary {
		summaries := []DriftSummary{}
		for _, constraint := range additional {
			alterTableStmt := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;", expectedTable.Name, constraint.Name)

			summary := newDriftSummary(
				fmt.Sprintf("%q.%q", expectedTable.Name, constraint.Name),
				fmt.Sprintf("Unexpected constraint %q.%q", expectedTable.Name, constraint.Name),
				"drop the constraint",
			).withStatements(alterTableStmt)
			summaries = append(summaries, summary)
		}

		return summaries
	})
}

func compareIndexes(actualTable, expectedTable schemas.TableDescription) []DriftSummary {
	return compareNamedLists(actualTable.Indexes, expectedTable.Indexes, func(index *schemas.IndexDescription, expectedIndex schemas.IndexDescription) DriftSummary {
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
			return newDriftSummary(
				fmt.Sprintf("%q.%q", expectedTable.Name, expectedIndex.Name),
				fmt.Sprintf("Missing index %q.%q", expectedTable.Name, expectedIndex.Name),
				"define the index",
			).withStatements(createIndexStmt)
		}

		dropIndexStmt := fmt.Sprintf("DROP INDEX %s;", expectedIndex.Name)

		return newDriftSummary(
			fmt.Sprintf("%q.%q", expectedTable.Name, expectedIndex.Name),
			fmt.Sprintf("Unexpected properties of index %q.%q", expectedTable.Name, expectedIndex.Name),
			"redefine the index",
		).withDiff(expectedIndex, *index).withStatements(dropIndexStmt, createIndexStmt)
	}, func(additional []schemas.IndexDescription) []DriftSummary {
		summaries := []DriftSummary{}
		for _, index := range additional {
			dropIndexStmt := fmt.Sprintf("DROP INDEX %s;", index.Name)

			summary := newDriftSummary(
				fmt.Sprintf("%q.%q", expectedTable.Name, index.Name),
				fmt.Sprintf("Unexpected index %q.%q", expectedTable.Name, index.Name),
				"drop the index",
			).withStatements(dropIndexStmt)
			summaries = append(summaries, summary)
		}

		return summaries
	})
}

func compareTriggers(actualTable, expectedTable schemas.TableDescription) []DriftSummary {
	return compareNamedLists(actualTable.Triggers, expectedTable.Triggers, func(trigger *schemas.TriggerDescription, expectedTrigger schemas.TriggerDescription) DriftSummary {
		createTriggerStmt := fmt.Sprintf("%s;", expectedTrigger.Definition)
		dropTriggerStmt := fmt.Sprintf("DROP TRIGGER %s ON %s;", expectedTrigger.Name, expectedTable.Name)

		if trigger == nil {
			return newDriftSummary(
				fmt.Sprintf("%q.%q", expectedTable.Name, expectedTrigger.Name),
				fmt.Sprintf("Missing trigger %q.%q", expectedTable.Name, expectedTrigger.Name),
				"define the trigger",
			).withStatements(createTriggerStmt)
		}

		return newDriftSummary(
			fmt.Sprintf("%q.%q", expectedTable.Name, expectedTrigger.Name),
			fmt.Sprintf("Unexpected properties of trigger %q.%q", expectedTable.Name, expectedTrigger.Name),
			"redefine the trigger",
		).withDiff(expectedTrigger, *trigger).withStatements(dropTriggerStmt, createTriggerStmt)
	}, func(additional []schemas.TriggerDescription) []DriftSummary {
		summaries := []DriftSummary{}
		for _, trigger := range additional {
			dropTriggerStmt := fmt.Sprintf("DROP TRIGGER %s ON %s;", trigger.Name, expectedTable.Name)

			summary := newDriftSummary(
				fmt.Sprintf("%q.%q", expectedTable.Name, trigger.Name),
				fmt.Sprintf("Unexpected trigger %q.%q", expectedTable.Name, trigger.Name),
				"drop the trigger",
			).withStatements(dropTriggerStmt)
			summaries = append(summaries, summary)
		}

		return summaries
	})
}

func compareViews(schemaName, version string, actual, expected schemas.SchemaDescription) []DriftSummary {
	return compareNamedLists(actual.Views, expected.Views, func(view *schemas.ViewDescription, expectedView schemas.ViewDescription) DriftSummary {
		// pgsql has weird indents here
		viewDefinition := strings.TrimSpace(stripIndent(" " + expectedView.Definition))
		createViewStmt := fmt.Sprintf("CREATE VIEW %s AS %s", expectedView.Name, viewDefinition)
		dropViewStmt := fmt.Sprintf("DROP VIEW %s;", expectedView.Name)

		if view == nil {
			return newDriftSummary(
				expectedView.Name,
				fmt.Sprintf("Missing view %q", expectedView.Name),
				"define the view",
			).withStatements(createViewStmt)
		}

		return newDriftSummary(
			expectedView.Name,
			fmt.Sprintf("Unexpected definition of view %q", expectedView.Name),
			"redefine the view",
		).withDiff(expectedView.Definition, view.Definition).withStatements(dropViewStmt, createViewStmt)
	}, noopAdditionalCallback[schemas.ViewDescription])
}

func noopAdditionalCallback[T schemas.Namer](_ []T) []DriftSummary {
	return nil
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
	primaryCallback func(a *T, b T) DriftSummary,
	additionsCallback func(additional []T) []DriftSummary,
) []DriftSummary {
	return compareNamedListsMulti(
		as,
		bs,
		func(a *T, b T) []DriftSummary {
			v := primaryCallback(a, b)
			if v == nil {
				return nil
			}
			return wrap(v)
		},
		additionsCallback,
	)
}

func compareNamedListsMulti[T schemas.Namer](
	as []T,
	bs []T,
	primaryCallback func(a *T, b T) []DriftSummary,
	additionsCallback func(additional []T) []DriftSummary,
) []DriftSummary {
	am := groupByName(as)
	bm := groupByName(bs)
	additional := make([]T, 0, len(am))
	summaries := []DriftSummary(nil)

	for _, k := range keys(am) {
		av := schemas.Normalize(am[k])

		if bv, ok := bm[k]; ok {
			bv = schemas.Normalize(bv)

			if cmp.Diff(schemas.PreComparisonNormalize(av), schemas.PreComparisonNormalize(bv)) != "" {
				summaries = append(summaries, primaryCallback(&av, bv)...)
			}
		} else {
			additional = append(additional, av)
		}
	}
	for _, k := range keys(bm) {
		bv := schemas.Normalize(bm[k])

		if _, ok := am[k]; !ok {
			summaries = append(summaries, primaryCallback(nil, bv)...)
		}
	}

	if len(additional) > 0 {
		summaries = append(summaries, additionsCallback(additional)...)
	}

	return summaries
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

type OutputWriter interface {
	Write(s string)
	Writef(format string, args ...any)
	WriteLine(line output.FancyLine)
	WriteCode(languageName, str string) error
}

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
