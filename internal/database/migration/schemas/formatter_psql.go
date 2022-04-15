package schemas

import (
	"fmt"
	"html"
	"math"
	"sort"
	"strconv"
	"strings"
)

type psqlFormatter struct{}

func NewPSQLFormatter() SchemaFormatter {
	return psqlFormatter{}
}

func (f psqlFormatter) Format(schemaDescription SchemaDescription) string {
	docs := []string{}
	sortTables(schemaDescription.Tables)
	for _, table := range schemaDescription.Tables {
		docs = append(docs, f.formatTable(schemaDescription, table)...)
	}

	sortViews(schemaDescription.Views)
	for _, view := range schemaDescription.Views {
		docs = append(docs, f.formatView(schemaDescription, view)...)
	}

	types := map[string][]string{}
	for _, enum := range schemaDescription.Enums {
		types[enum.Name] = enum.Labels
	}
	if len(types) > 0 {
		docs = append(docs, f.formatTypes(schemaDescription, types)...)
	}

	return strings.Join(docs, "\n")
}

func (f psqlFormatter) formatTable(schemaDescription SchemaDescription, table TableDescription) []string {
	centerString := func(s string, n int) string {
		x := float64(n - len(s))
		i := int(math.Floor(x / 2))
		if i <= 0 {
			i = 1
		}
		j := int(math.Ceil(x / 2))
		if j <= 0 {
			j = 1
		}

		return strings.Repeat(" ", i) + s + strings.Repeat(" ", j)
	}

	formatColumns := func(headers []string, rows [][]string) []string {
		sizes := make([]int, len(headers))
		headerValues := make([]string, len(headers))
		sepValues := make([]string, len(headers))
		for i, header := range headers {
			sizes[i] = len(header)

			for _, row := range rows {
				if n := len(row[i]); n > sizes[i] {
					sizes[i] = n
				}
			}

			headerValues[i] = centerString(headers[i], sizes[i]+2)
			sepValues[i] = strings.Repeat("-", sizes[i]+2)
		}

		docs := make([]string, 0, len(rows)+2)
		docs = append(docs, strings.Join(headerValues, "|"))
		docs = append(docs, strings.Join(sepValues, "+"))

		for _, row := range rows {
			rowValues := make([]string, 0, len(headers))
			for i := range headers {
				if i == len(headers)-1 {
					rowValues = append(rowValues, row[i])
				} else {
					rowValues = append(rowValues, fmt.Sprintf("%-"+strconv.Itoa(sizes[i])+"s", row[i]))
				}
			}

			docs = append(docs, " "+strings.Join(rowValues, " | "))
		}

		return docs
	}

	docs := []string{}
	docs = append(docs, fmt.Sprintf("# Table \"public.%s\"", table.Name))
	docs = append(docs, "```")

	headers := []string{"Column", "Type", "Collation", "Nullable", "Default"}
	rows := [][]string{}
	sortColumns(table.Columns)
	for _, column := range table.Columns {
		nullConstraint := "not null"
		if column.IsNullable {
			nullConstraint = ""
		}

		defaultValue := column.Default
		if column.IsGenerated == "ALWAYS" {
			defaultValue = "generated always as (" + column.GenerationExpression + ") stored"
		}

		rows = append(rows, []string{
			column.Name,
			column.TypeName,
			"",
			nullConstraint,
			defaultValue,
		})
	}
	docs = append(docs, formatColumns(headers, rows)...)

	if len(table.Indexes) > 0 {
		docs = append(docs, "Indexes:")
		sortIndexes(table.Indexes)
		for _, index := range table.Indexes {
			if index.IsPrimaryKey {
				def := strings.TrimSpace(strings.Split(index.IndexDefinition, "USING")[1])
				docs = append(docs, fmt.Sprintf("    %q PRIMARY KEY, %s", index.Name, def))
			}
		}
		for _, index := range table.Indexes {
			if !index.IsPrimaryKey {
				uq := ""
				if index.IsUnique {
					c := ""
					if index.ConstraintType == "u" {
						c = " CONSTRAINT"
					}
					uq = " UNIQUE" + c + ","
				}
				deferrable := ""
				if index.IsDeferrable {
					deferrable = " DEFERRABLE"
				}
				def := strings.TrimSpace(strings.Split(index.IndexDefinition, "USING")[1])
				if index.IsExclusion {
					def = "EXCLUDE USING " + def
				}
				docs = append(docs, fmt.Sprintf("    %q%s %s%s", index.Name, uq, def, deferrable))
			}
		}
	}

	numCheckConstraints := 0
	numForeignKeyConstraints := 0
	for _, constraint := range table.Constraints {
		switch constraint.ConstraintType {
		case "c":
			numCheckConstraints++
		case "f":
			numForeignKeyConstraints++
		}
	}

	if numCheckConstraints > 0 {
		docs = append(docs, "Check constraints:")
		for _, constraint := range table.Constraints {
			if constraint.ConstraintType == "c" {
				docs = append(docs, fmt.Sprintf("    %q %s", constraint.Name, constraint.ConstraintDefinition))
			}
		}
	}
	if numForeignKeyConstraints > 0 {
		docs = append(docs, "Foreign-key constraints:")
		for _, constraint := range table.Constraints {
			if constraint.ConstraintType == "f" {
				docs = append(docs, fmt.Sprintf("    %q %s", constraint.Name, constraint.ConstraintDefinition))
			}
		}
	}

	type tableAndConstraint struct {
		TableDescription
		ConstraintDescription
	}
	tableAndConstraints := []tableAndConstraint{}
	for _, otherTable := range schemaDescription.Tables {
		for _, constraint := range otherTable.Constraints {
			if constraint.RefTableName == table.Name {
				tableAndConstraints = append(tableAndConstraints, tableAndConstraint{otherTable, constraint})
			}
		}
	}
	sort.Slice(tableAndConstraints, func(i, j int) bool {
		return tableAndConstraints[i].ConstraintDescription.Name < tableAndConstraints[j].ConstraintDescription.Name
	})
	if len(tableAndConstraints) > 0 {
		docs = append(docs, "Referenced by:")

		for _, tableAndConstraint := range tableAndConstraints {
			docs = append(docs, fmt.Sprintf("    TABLE %q CONSTRAINT %q %s", tableAndConstraint.TableDescription.Name, tableAndConstraint.ConstraintDescription.Name, tableAndConstraint.ConstraintDescription.ConstraintDefinition))
		}
	}

	if len(table.Triggers) > 0 {
		docs = append(docs, "Triggers:")
		for _, trigger := range table.Triggers {
			def := strings.TrimSpace(strings.SplitN(trigger.Definition, trigger.Name, 2)[1])
			docs = append(docs, fmt.Sprintf("    %s %s", trigger.Name, def))
		}
	}

	docs = append(docs, "\n```\n")

	if table.Comment != "" {
		docs = append(docs, html.EscapeString(table.Comment)+"\n")
	}

	sortColumnsByName(table.Columns)
	for _, column := range table.Columns {
		if column.Comment != "" {
			docs = append(docs, fmt.Sprintf("**%s**: %s\n", column.Name, html.EscapeString(column.Comment)))
		}
	}

	return docs
}

func (f psqlFormatter) formatView(_ SchemaDescription, view ViewDescription) []string {
	docs := []string{}
	docs = append(docs, fmt.Sprintf("# View \"public.%s\"\n", view.Name))
	docs = append(docs, fmt.Sprintf("## View query:\n\n```sql\n%s\n```\n", view.Definition))
	return docs
}

func (f psqlFormatter) formatTypes(_ SchemaDescription, types map[string][]string) []string {
	typeNames := make([]string, 0, len(types))
	for typeName := range types {
		typeNames = append(typeNames, typeName)
	}
	sort.Strings(typeNames)

	docs := make([]string, 0, len(typeNames)*4)
	for _, name := range typeNames {
		docs = append(docs, fmt.Sprintf("# Type %s", name))
		docs = append(docs, "")
		docs = append(docs, "- "+strings.Join(types[name], "\n- "))
		docs = append(docs, "")
	}

	return docs
}

func sortTables(tables []TableDescription) {
	sort.Slice(tables, func(i, j int) bool { return tables[i].Name < tables[j].Name })
}

func sortViews(views []ViewDescription) {
	sort.Slice(views, func(i, j int) bool { return views[i].Name < views[j].Name })
}

func sortColumns(columns []ColumnDescription) {
	sort.Slice(columns, func(i, j int) bool { return columns[i].Index < columns[j].Index })
}

func sortColumnsByName(columns []ColumnDescription) {
	sort.Slice(columns, func(i, j int) bool { return columns[i].Name < columns[j].Name })
}

func sortIndexes(indexes []IndexDescription) {
	sort.Slice(indexes, func(i, j int) bool {
		if indexes[i].IsUnique && !indexes[j].IsUnique {
			return true
		}
		if !indexes[i].IsUnique && indexes[j].IsUnique {
			return false
		}

		return indexes[i].Name < indexes[j].Name
	})
}
