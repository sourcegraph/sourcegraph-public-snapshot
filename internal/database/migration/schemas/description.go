package schemas

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

type SchemaDescription struct {
	Extensions []string
	Enums      []EnumDescription
	Functions  []FunctionDescription
	Sequences  []SequenceDescription
	Tables     []TableDescription
	Views      []ViewDescription
}

func (d SchemaDescription) WrappedExtensions() []ExtensionDescription {
	extensions := make([]ExtensionDescription, 0, len(d.Extensions))
	for _, name := range d.Extensions {
		extensions = append(extensions, ExtensionDescription{Name: name})
	}

	return extensions
}

type ExtensionDescription struct {
	Name string
}

func (d ExtensionDescription) CreateStatement() string {
	return fmt.Sprintf("CREATE EXTENSION %s;", d.Name)
}

type EnumDescription struct {
	Name   string
	Labels []string
}

func (d EnumDescription) CreateStatement() string {
	quotedLabels := make([]string, 0, len(d.Labels))
	for _, label := range d.Labels {
		quotedLabels = append(quotedLabels, fmt.Sprintf("'%s'", label))
	}

	return fmt.Sprintf("CREATE TYPE %s AS ENUM (%s);", d.Name, strings.Join(quotedLabels, ", "))
}

func (d EnumDescription) DropStatement() string {
	return fmt.Sprintf("DROP TYPE IF EXISTS %s;", d.Name)
}

// AlterToTarget returns a set of `ALTER ENUM ADD VALUE` statements to make the given enum equivalent to
// the expected enum, then additive statements cannot bring the enum to the expected state and we return
// a false-valued flag. In this case the existing type must be dropped and re-created as there's currently
// no way to *remove* values from an enum type.
func (d EnumDescription) AlterToTarget(target EnumDescription) ([]string, bool) {
	labels := GroupByName(wrapStrings(d.Labels))
	expectedLabels := GroupByName(wrapStrings(target.Labels))

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
	missingLabels := make([]missingLabel, 0, len(target.Labels))

	after := ""
	for _, label := range target.Labels {
		if _, ok := labels[label]; !ok && after != "" {
			missingLabels = append(missingLabels, missingLabel{label: label, neighbor: after, before: false})
		}
		after = label
	}

	before := ""
	for i := len(target.Labels) - 1; i >= 0; i-- {
		label := target.Labels[i]

		if _, ok := labels[label]; !ok && before != "" {
			missingLabels = append(missingLabels, missingLabel{label: label, neighbor: before, before: true})
		}
		before = label
	}

	var (
		ordered   []string
		reachable = GroupByName(wrapStrings(d.Labels))
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
			ordered = append(ordered, fmt.Sprintf("ALTER TYPE %s ADD VALUE '%s' %s '%s';", target.GetName(), s.label, rel, s.neighbor))
			continue outer
		}

		panic("Infinite loop")
	}

	return ordered, true
}

type FunctionDescription struct {
	Name       string
	Definition string
}

func (d FunctionDescription) CreateOrReplaceStatement() string {
	return fmt.Sprintf("%s;", d.Definition)
}

type SequenceDescription struct {
	Name         string
	TypeName     string
	StartValue   int
	MinimumValue int
	MaximumValue int
	Increment    int
	CycleOption  string
}

func (d SequenceDescription) CreateStatement() string {
	minValue := "NO MINVALUE"
	if d.MinimumValue != 0 {
		minValue = fmt.Sprintf("MINVALUE %d", d.MinimumValue)
	}
	maxValue := "NO MAXVALUE"
	if d.MaximumValue != 0 {
		maxValue = fmt.Sprintf("MAXVALUE %d", d.MaximumValue)
	}

	return fmt.Sprintf(
		"CREATE SEQUENCE %s AS %s INCREMENT BY %d %s %s START WITH %d %s CYCLE;",
		d.Name,
		d.TypeName,
		d.Increment,
		minValue,
		maxValue,
		d.StartValue,
		d.CycleOption,
	)
}

func (d SequenceDescription) AlterToTarget(target SequenceDescription) ([]string, bool) {
	statements := []string{}

	if d.TypeName != target.TypeName {
		statements = append(statements, fmt.Sprintf("ALTER SEQUENCE %s AS %s MAXVALUE %d;", d.Name, target.TypeName, target.MaximumValue))

		// Remove from diff below
		d.TypeName = target.TypeName
		d.MaximumValue = target.MaximumValue
	}

	// Abort if there are other fields we haven't addressed
	hasAdditionalDiff := cmp.Diff(d, target) != ""
	return statements, !hasAdditionalDiff
}

type TableDescription struct {
	Name        string
	Comment     string
	Columns     []ColumnDescription
	Indexes     []IndexDescription
	Constraints []ConstraintDescription
	Triggers    []TriggerDescription
}

type ColumnDescription struct {
	Name                   string
	Index                  int
	TypeName               string
	IsNullable             bool
	Default                string
	CharacterMaximumLength int
	IsIdentity             bool
	IdentityGeneration     string
	IsGenerated            string
	GenerationExpression   string
	Comment                string
}

func (d ColumnDescription) CreateStatement(table TableDescription) string {
	nullableExpr := ""
	if !d.IsNullable {
		nullableExpr = " NOT NULL"
	}
	defaultExpr := ""
	if d.Default != "" {
		defaultExpr = fmt.Sprintf(" DEFAULT %s", d.Default)
	}

	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s%s%s;", table.Name, d.Name, d.TypeName, nullableExpr, defaultExpr)
}

func (d ColumnDescription) DropStatement(table TableDescription) string {
	return fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s;", table.Name, d.Name)
}

func (d ColumnDescription) AlterToTarget(table TableDescription, target ColumnDescription) ([]string, bool) {
	statements := []string{}

	if d.TypeName != target.TypeName {
		statements = append(statements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DATA TYPE %s;", table.Name, target.Name, target.TypeName))

		// Remove from diff below
		d.TypeName = target.TypeName
	}
	if d.IsNullable != target.IsNullable {
		var verb string
		if target.IsNullable {
			verb = "DROP"
		} else {
			verb = "SET"
		}

		statements = append(statements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s %s NOT NULL;", table.Name, target.Name, verb))

		// Remove from diff below
		d.IsNullable = target.IsNullable
	}
	if d.Default != target.Default {
		if target.Default == "" {
			statements = append(statements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT;", table.Name, target.Name))
		} else {
			statements = append(statements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s;", table.Name, target.Name, target.Default))
		}

		// Remove from diff below
		d.Default = target.Default
	}

	// Abort if there are other fields we haven't addressed
	hasAdditionalDiff := cmp.Diff(d, target) != ""
	return statements, !hasAdditionalDiff
}

type IndexDescription struct {
	Name                 string
	IsPrimaryKey         bool
	IsUnique             bool
	IsExclusion          bool
	IsDeferrable         bool
	IndexDefinition      string
	ConstraintType       string
	ConstraintDefinition string
}

func (d IndexDescription) CreateStatement(table TableDescription) string {
	if d.ConstraintType == "u" || d.ConstraintType == "p" {
		return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s;", table.Name, d.Name, d.ConstraintDefinition)
	}

	return fmt.Sprintf("%s;", d.IndexDefinition)
}

func (d IndexDescription) DropStatement(table TableDescription) string {
	if d.ConstraintType == "u" || d.ConstraintType == "p" {
		return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;", table.Name, d.Name)
	}

	return fmt.Sprintf("DROP INDEX IF EXISTS %s;", d.GetName())
}

type ConstraintDescription struct {
	Name                 string
	ConstraintType       string
	RefTableName         string
	IsDeferrable         bool
	ConstraintDefinition string
}

func (d ConstraintDescription) CreateStatement(table TableDescription) string {
	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s;", table.Name, d.Name, d.ConstraintDefinition)
}

func (d ConstraintDescription) DropStatement(table TableDescription) string {
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;", table.Name, d.Name)
}

type TriggerDescription struct {
	Name       string
	Definition string
}

func (d TriggerDescription) CreateStatement() string {
	return fmt.Sprintf("%s;", d.Definition)
}

func (d TriggerDescription) DropStatement(table TableDescription) string {
	return fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s;", d.Name, table.Name)
}

type ViewDescription struct {
	Name       string
	Definition string
}

func (d ViewDescription) CreateStatement() string {
	// pgsql indents definitions strangely; we copy that
	return fmt.Sprintf("CREATE VIEW %s AS %s", d.Name, strings.TrimSpace(stripIndent(" "+d.Definition)))
}

func (d ViewDescription) DropStatement() string {
	return fmt.Sprintf("DROP VIEW IF EXISTS %s;", d.Name)
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

func Canonicalize(schemaDescription SchemaDescription) {
	for i := range schemaDescription.Tables {
		sortColumnsByName(schemaDescription.Tables[i].Columns)
		sortIndexes(schemaDescription.Tables[i].Indexes)
		sortConstraints(schemaDescription.Tables[i].Constraints)
		sortTriggers(schemaDescription.Tables[i].Triggers)
	}

	sortEnums(schemaDescription.Enums)
	sortFunctions(schemaDescription.Functions)
	sortSequences(schemaDescription.Sequences)
	sortTables(schemaDescription.Tables)
	sortViews(schemaDescription.Views)
}

type Namer interface{ GetName() string }

func GroupByName[T Namer](ts []T) map[string]T {
	m := make(map[string]T, len(ts))
	for _, t := range ts {
		m[t.GetName()] = t
	}

	return m
}

type stringNamer string

func wrapStrings(ss []string) []Namer {
	sn := make([]Namer, 0, len(ss))
	for _, s := range ss {
		sn = append(sn, stringNamer(s))
	}

	return sn
}

func (n stringNamer) GetName() string           { return string(n) }
func (d ExtensionDescription) GetName() string  { return d.Name }
func (d EnumDescription) GetName() string       { return d.Name }
func (d FunctionDescription) GetName() string   { return d.Name }
func (d SequenceDescription) GetName() string   { return d.Name }
func (d TableDescription) GetName() string      { return d.Name }
func (d ColumnDescription) GetName() string     { return d.Name }
func (d IndexDescription) GetName() string      { return d.Name }
func (d ConstraintDescription) GetName() string { return d.Name }
func (d TriggerDescription) GetName() string    { return d.Name }
func (d ViewDescription) GetName() string       { return d.Name }

type (
	Normalizer[T any]              interface{ Normalize() T }
	PreComparisonNormalizer[T any] interface{ PreComparisonNormalize() T }
)

func (d FunctionDescription) PreComparisonNormalize() FunctionDescription {
	d.Definition = normalizeFunction(d.Definition)
	return d
}

func (d TableDescription) Normalize() TableDescription {
	d.Comment = ""
	return d
}

func (d ColumnDescription) Normalize() ColumnDescription {
	d.Index = -1
	d.Comment = ""
	return d
}

func Normalize[T any](v T) T {
	if normalizer, ok := any(v).(Normalizer[T]); ok {
		return normalizer.Normalize()
	}

	return v
}

func PreComparisonNormalize[T any](v T) T {
	if normalizer, ok := any(v).(PreComparisonNormalizer[T]); ok {
		return normalizer.PreComparisonNormalize()
	}

	return v
}

var whitespacePattern = lazyregexp.New(`\s+`)

func normalizeFunction(definition string) string {
	lines := strings.Split(definition, "\n")
	for i, line := range lines {
		lines[i] = strings.Split(line, "--")[0]
	}

	return strings.TrimSpace(whitespacePattern.ReplaceAllString(strings.Join(lines, "\n"), " "))
}
