package modl

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

// TableMap represents a mapping between a Go struct and a database table
// Use dbmap.AddTable() or dbmap.AddTableWithName() to create these
type TableMap struct {
	// Name of database table.
	TableName  string
	Keys       []*ColumnMap
	Columns    []*ColumnMap
	gotype     reflect.Type
	version    *ColumnMap
	updatePlan bindPlan
	deletePlan bindPlan
	getPlan    bindPlan
	dbmap      *DbMap
	mapper     *reflectx.Mapper
	// Cached capabilities for the struct mapped to this table
	CanPreInsert  bool
	CanPostInsert bool
	CanPostGet    bool
	CanPreUpdate  bool
	CanPostUpdate bool
	CanPreDelete  bool
	CanPostDelete bool

	insertPlan         *bindPlan
	insertAutoIncrPlan *bindPlan
}

// ResetSql removes cached insert/update/select/delete SQL strings
// associated with this TableMap.  Call this if you've modified
// any column names or the table name itself.
func (t *TableMap) ResetSql() {
	t.insertPlan = nil
	t.insertAutoIncrPlan = nil
	t.updatePlan = bindPlan{}
	t.deletePlan = bindPlan{}
	t.getPlan = bindPlan{}
}

// SetKeys lets you specify the fields on a struct that map to primary
// key columns on the table.  If isAutoIncr is set, result.LastInsertId()
// will be used after INSERT to bind the generated id to the Go struct.
//
// Automatically calls ResetSql() to ensure SQL statements are regenerated.
func (t *TableMap) SetKeys(isAutoIncr bool, fieldNames ...string) *TableMap {
	t.Keys = make([]*ColumnMap, 0)
	for _, name := range fieldNames {
		// FIXME: sqlx.NameMapper is a deprecated API.  modl should have its
		// own API which sets sqlx's mapping funcs as necessary
		colmap := t.ColMap(sqlx.NameMapper(name))
		colmap.isPK = true
		colmap.isAutoIncr = isAutoIncr
		t.Keys = append(t.Keys, colmap)
	}
	t.ResetSql()

	return t
}

// ColMap returns the ColumnMap pointer matching the given struct field
// name.  It panics if the struct does not contain a field matching this
// name.
func (t *TableMap) ColMap(field string) *ColumnMap {
	for _, col := range t.Columns {
		if col.fieldName == field || col.ColumnName == field {
			return col
		}
	}
	panic(fmt.Sprintf("No ColumnMap in table %s type %s with field %s",
		t.TableName, t.gotype.Name(), field))
}

// SetVersionCol sets the column to use as the Version field.  By default
// the "Version" field is used.  Returns the column found, or panics
// if the struct does not contain a field matching this name.
//
// Automatically calls ResetSql() to ensure SQL statements are regenerated.
func (t *TableMap) SetVersionCol(field string) *ColumnMap {
	c := t.ColMap(field)
	t.version = c
	t.ResetSql()
	return c
}

func (t *TableMap) bindGet() bindPlan {
	plan := t.getPlan
	if plan.query == "" {

		s := bytes.Buffer{}
		s.WriteString("select ")

		x := 0
		for _, col := range t.Columns {
			if !col.Transient {
				if x > 0 {
					s.WriteString(",")
				}
				s.WriteString(t.dbmap.Dialect.QuoteField(col.ColumnName))
				plan.argFields = append(plan.argFields, col.fieldName)
				x++
			}
		}
		s.WriteString(" from ")
		s.WriteString(t.dbmap.Dialect.QuoteField(t.TableName))
		s.WriteString(" where ")
		for x := range t.Keys {
			col := t.Keys[x]
			if x > 0 {
				s.WriteString(" and ")
			}
			s.WriteString(t.dbmap.Dialect.QuoteField(col.ColumnName))
			s.WriteString("=")
			s.WriteString(t.dbmap.Dialect.BindVar(x))

			plan.keyFields = append(plan.keyFields, col.fieldName)
		}
		s.WriteString(";")

		plan.query = s.String()
		t.getPlan = plan
	}

	return plan
}

func (t *TableMap) bindDelete(elem reflect.Value) bindInstance {
	plan := t.deletePlan
	if plan.query == "" {

		s := bytes.Buffer{}
		s.WriteString(fmt.Sprintf("delete from %s", t.dbmap.Dialect.QuoteField(t.TableName)))

		for y := range t.Columns {
			col := t.Columns[y]
			if !col.Transient {
				if col == t.version {
					plan.versField = col.fieldName
				}
			}
		}

		s.WriteString(" where ")
		for x := range t.Keys {
			k := t.Keys[x]
			if x > 0 {
				s.WriteString(" and ")
			}
			s.WriteString(t.dbmap.Dialect.QuoteField(k.ColumnName))
			s.WriteString("=")
			s.WriteString(t.dbmap.Dialect.BindVar(x))

			plan.keyFields = append(plan.keyFields, k.fieldName)
			plan.argFields = append(plan.argFields, k.fieldName)
		}
		if plan.versField != "" {
			s.WriteString(" and ")
			s.WriteString(t.dbmap.Dialect.QuoteField(t.version.ColumnName))
			s.WriteString("=")
			s.WriteString(t.dbmap.Dialect.BindVar(len(plan.argFields)))

			plan.argFields = append(plan.argFields, plan.versField)
		}
		s.WriteString(";")

		plan.query = s.String()
		t.deletePlan = plan
	}

	return plan.createBindInstance(elem)
}

func (t *TableMap) bindUpdate(elem reflect.Value) bindInstance {
	plan := t.updatePlan
	if plan.query == "" {

		s := bytes.Buffer{}
		s.WriteString(fmt.Sprintf("update %s set ", t.dbmap.Dialect.QuoteField(t.TableName)))
		x := 0

		for y := range t.Columns {
			col := t.Columns[y]
			if !col.isPK && !col.Transient {
				if x > 0 {
					s.WriteString(", ")
				}
				s.WriteString(t.dbmap.Dialect.QuoteField(col.ColumnName))
				s.WriteString("=")
				s.WriteString(t.dbmap.Dialect.BindVar(x))

				if col == t.version {
					plan.versField = col.fieldName
					plan.argFields = append(plan.argFields, versFieldConst)
				} else {
					plan.argFields = append(plan.argFields, col.fieldName)
				}
				x++
			}
		}

		s.WriteString(" where ")
		for y := range t.Keys {
			col := t.Keys[y]
			if y > 0 {
				s.WriteString(" and ")
			}
			s.WriteString(t.dbmap.Dialect.QuoteField(col.ColumnName))
			s.WriteString("=")
			s.WriteString(t.dbmap.Dialect.BindVar(x))

			plan.argFields = append(plan.argFields, col.fieldName)
			plan.keyFields = append(plan.keyFields, col.fieldName)
			x++
		}
		if plan.versField != "" {
			s.WriteString(" and ")
			s.WriteString(t.dbmap.Dialect.QuoteField(t.version.ColumnName))
			s.WriteString("=")
			s.WriteString(t.dbmap.Dialect.BindVar(x))
			plan.argFields = append(plan.argFields, plan.versField)
		}
		s.WriteString(";")

		plan.query = s.String()
		t.updatePlan = plan
	}

	return plan.createBindInstance(elem)
}

func isZeroVal(v reflect.Value) bool {
	return v.Interface() == reflect.Zero(v.Type()).Interface()
}

func (t *TableMap) bindInsert(elem reflect.Value) bindInstance {
	// There are 2 possible insert plans: (1) if the autoincr field is
	// empty, then use a plan that inserts the AutoIncrBindValue; or
	// (2) if the autoincr field is not empty, use a plan that inserts
	// the row with the actual field value.

	insertsAutoIncrVal := false
	for _, col := range t.Keys {
		if col.isAutoIncr && isZeroVal(elem.FieldByIndex(col.fieldIdx)) {
			// This stmt needs to insert an auto-incremented value.
			insertsAutoIncrVal = true
			break
		}
	}

	var plan *bindPlan
	if insertsAutoIncrVal {
		plan = t.insertAutoIncrPlan
	} else {
		plan = t.insertPlan
	}

	if plan == nil {
		plan = &bindPlan{}

		s := bytes.Buffer{}
		s2 := bytes.Buffer{}
		s.WriteString(fmt.Sprintf("insert into %s (", t.dbmap.Dialect.QuoteField(t.TableName)))

		var autoIncrCol *ColumnMap
		x := 0
		first := true
		for y := range t.Columns {
			col := t.Columns[y]

			if !col.Transient {
				if !first {
					s.WriteString(",")
					s2.WriteString(",")
				}
				s.WriteString(t.dbmap.Dialect.QuoteField(col.ColumnName))

				if col.isAutoIncr && isZeroVal(elem.FieldByIndex(col.fieldIdx)) {
					s2.WriteString(t.dbmap.Dialect.AutoIncrBindValue())
					plan.autoIncrIdx = col.fieldIdx
					autoIncrCol = col
				} else {
					s2.WriteString(t.dbmap.Dialect.BindVar(x))
					if col == t.version {
						plan.versField = col.fieldName
						plan.argFields = append(plan.argFields, versFieldConst)
					} else {
						plan.argFields = append(plan.argFields, col.fieldName)
					}

					x++
				}

				first = false
			}
		}
		s.WriteString(") values (")
		s.WriteString(s2.String())
		s.WriteString(")")
		if autoIncrCol != nil {
			s.WriteString(t.dbmap.Dialect.AutoIncrInsertSuffix(autoIncrCol))
		}
		s.WriteString(";")

		plan.query = s.String()

		if insertsAutoIncrVal {
			t.insertAutoIncrPlan = plan
		} else {
			t.insertPlan = plan
		}
	}

	return plan.createBindInstance(elem)
}

// ColumnMap represents a mapping between a Go struct field and a single
// column in a table.
// Unique and MaxSize only inform the CreateTables() function and are not
// used for validation by Insert/Update/Delete/Get.
type ColumnMap struct {
	// Column name in db table
	ColumnName string

	// If true, this column is skipped in generated SQL statements
	Transient bool

	// If true, " unique" is added to create table statements.
	Unique bool

	// Passed to Dialect.ToSqlType() to assist in informing the
	// correct column type to map to in CreateTables()
	MaxSize int

	// the table this column belongs to
	table *TableMap

	fieldName  string
	fieldIdx   []int // Go struct field index from table's struct
	gotype     reflect.Type
	sqltype    string
	createSql  string
	isPK       bool
	isAutoIncr bool
}

// SetTransient allows you to mark the column as transient. If true
// this column will be skipped when SQL statements are generated
func (c *ColumnMap) SetTransient(b bool) *ColumnMap {
	c.Transient = b
	return c
}

// SetUnique sets the unqiue clause for this column.  If true, a unique clause
// will be added to create table statements for this column.
func (c *ColumnMap) SetUnique(b bool) *ColumnMap {
	c.Unique = b
	return c
}

// SetSqlCreate overrides the default create statement used when this column
// is created by CreateTable.  This will override all other options (like
// SetMaxSize, SetSqlType, etc).  To unset, call with the empty string.
func (c *ColumnMap) SetSqlCreate(s string) *ColumnMap {
	c.createSql = s
	return c
}

// SetSqlType sets an override for the column's sql type.  This is a string,
// such as 'varchar(32)' or 'text', which will be used by CreateTable and
// nothing else.  It is the caller's responsibility to ensure this will map
// cleanly to the underlying struct field via rows.Scan
func (c *ColumnMap) SetSqlType(t string) *ColumnMap {
	c.sqltype = t
	return c
}

// SetMaxSize specifies the max length of values of this column. This is
// passed to the dialect.ToSqlType() function, which can use the value
// to alter the generated type for "create table" statements
func (c *ColumnMap) SetMaxSize(size int) *ColumnMap {
	c.MaxSize = size
	return c
}

// Return a table for a pointer;  error if i is not a pointer or if the
// table is not found
func tableForPointer(m *DbMap, i interface{}, checkPk bool) (*TableMap, reflect.Value, error) {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Ptr {
		return nil, v, fmt.Errorf("value %v not a pointer", v)
	}
	v = v.Elem()
	t := m.TableForType(v.Type())
	if t == nil {
		return nil, v, fmt.Errorf("could not find table for %v", t)
	}
	if checkPk && len(t.Keys) < 1 {
		return t, v, &NoKeysErr{t}
	}
	return t, v, nil
}
