// Package modl provides a non-declarative database modelling layer to ease
// the use of frequently repeated patterns in database-backed applications
// and centralize database use to ease profiling and reporting.
//
// It is a fork of the wonderful github.com/coopernurse/gorp package, but is
// rewritten to use github.com/jmoiron/sqlx as a base.
//
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
//
package modl

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

// TableNameMapper is the function used by AddTable to map struct names to database table names, in analogy
// to sqlx.NameMapper which does the same for struct field name to database column names.
var TableNameMapper = strings.ToLower

// DbMap is the root modl mapping object. Create one of these for each
// database schema you wish to map.  Each DbMap contains a list of
// mapped tables.
//
// Example:
//
//     dialect := modl.MySQLDialect{"InnoDB", "UTF8"}
//     dbmap := &modl.DbMap{Db: db, Dialect: dialect}
//
type DbMap struct {
	// Db handle to use with this map
	Db  *sql.DB
	Dbx *sqlx.DB

	// Dialect implementation to use with this map
	Dialect Dialect

	tables    []*TableMap
	logger    *log.Logger
	logPrefix string
	mapper    *reflectx.Mapper
}

// NewDbMap returns a new DbMap using the db connection and dialect.
func NewDbMap(db *sql.DB, dialect Dialect) *DbMap {
	return &DbMap{
		Db:      db,
		Dialect: dialect,
		Dbx:     sqlx.NewDb(db, dialect.DriverName()),
		mapper:  reflectx.NewMapperFunc("db", sqlx.NameMapper),
	}
}

// TraceOn turns on SQL statement logging for this DbMap.  After this is
// called, all SQL statements will be sent to the logger.  If prefix is
// a non-empty string, it will be written to the front of all logged
// strings, which can aid in filtering log lines.
//
// Use TraceOn if you want to spy on the SQL statements that modl
// generates.
func (m *DbMap) TraceOn(prefix string, logger *log.Logger) {
	m.logger = logger
	if len(prefix) == 0 {
		m.logPrefix = prefix
	} else {
		m.logPrefix = prefix + " "
	}
}

// TraceOff turns off tracing. It is idempotent.
func (m *DbMap) TraceOff() {
	m.logger = nil
	m.logPrefix = ""
}

// AddTable registers the given interface type with modl. The table name
// will be given the name of the TypeOf(i), lowercased.
//
// This operation is idempotent. If i's type is already mapped, the
// existing *TableMap is returned.
func (m *DbMap) AddTable(i interface{}, name ...string) *TableMap {
	Name := ""
	if len(name) > 0 {
		Name = name[0]
	}

	t := reflect.TypeOf(i)
	// Use sqlx's NameMapper function if no name is supplied
	if len(Name) == 0 {
		Name = TableNameMapper(t.Name())
	}

	// check if we have a table for this type already
	// if so, update the name and return the existing pointer
	for i := range m.tables {
		table := m.tables[i]
		if table.gotype == t {
			table.TableName = Name
			return table
		}
	}

	tmap := &TableMap{gotype: t, TableName: Name, dbmap: m, mapper: m.mapper}
	tmap.setupHooks(i)

	tmap.Columns = columnMaps(t, tmap, nil)
	for _, cm := range tmap.Columns {
		if cm.fieldName == "Version" {
			tmap.version = cm
		}
	}
	m.tables = append(m.tables, tmap)

	return tmap

}

func columnMaps(t reflect.Type, tmap *TableMap, parentFieldIdx []int) []*ColumnMap {
	var cols []*ColumnMap
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name := f.Tag.Get("db")
		if f.Anonymous {
			cols = append(cols, columnMaps(f.Type, tmap, makeFieldIdx(parentFieldIdx, i))...)
		} else {
			if name == "" {
				name = sqlx.NameMapper(f.Name)
			}
			cols = append(cols, &ColumnMap{
				ColumnName: name,
				Transient:  name == "-",
				fieldName:  f.Name,
				fieldIdx:   makeFieldIdx(parentFieldIdx, i),
				gotype:     f.Type,
				table:      tmap,
			})
		}
	}
	return cols
}

// makeFieldIdx returns a new slice whose elements are equal to
// append(parent, i).
func makeFieldIdx(parent []int, i int) []int {
	s := make([]int, len(parent)+1)
	copy(s, parent)
	s[len(s)-1] = i
	return s
}

// AddTableWithName adds a new mapping of the interface to a table name.
func (m *DbMap) AddTableWithName(i interface{}, name string) *TableMap {
	return m.AddTable(i, name)
}

// CreateTablesSql returns create table SQL as a map of table names to
// their associated CREATE TABLE statements.
func (m *DbMap) CreateTablesSql() (map[string]string, error) {
	return m.createTables(false, false)
}

// CreateTables iterates through TableMaps registered to this DbMap and
// executes "create table" statements against the database for each.
//
// This is particularly useful in unit tests where you want to create
// and destroy the schema automatically.
func (m *DbMap) CreateTables() error {
	_, err := m.createTables(false, true)
	return err
}

// CreateTablesIfNotExists is similar to CreateTables, but starts
// each statement with "create table if not exists" so that existing
// tables do not raise errors.
func (m *DbMap) CreateTablesIfNotExists() error {
	_, err := m.createTables(true, true)
	return err
}

func writeColumnSql(sql *bytes.Buffer, col *ColumnMap) {
	if len(col.createSql) > 0 {
		sql.WriteString(col.createSql)
		return
	}
	sqltype := col.sqltype
	if len(sqltype) == 0 {
		sqltype = col.table.dbmap.Dialect.ToSqlType(col)
	}
	sql.WriteString(fmt.Sprintf("%s %s", col.table.dbmap.Dialect.QuoteField(col.ColumnName), sqltype))
	if col.isPK {
		sql.WriteString(" not null")
		if len(col.table.Keys) == 1 {
			sql.WriteString(" primary key")
		}
	}
	if col.Unique {
		sql.WriteString(" unique")
	}
	if col.isAutoIncr {
		sql.WriteString(" " + col.table.dbmap.Dialect.AutoIncrStr())
	}
}

func (m *DbMap) createTables(ifNotExists, exec bool) (map[string]string, error) {
	var err error
	ret := map[string]string{}

	sep := ", "
	prefix := ""
	if !exec {
		sep = ",\n"
		prefix = "    "
	}

	for i := range m.tables {
		table := m.tables[i]

		s := bytes.Buffer{}
		s.WriteString("create table ")
		if ifNotExists {
			s.WriteString("if not exists ")
		}
		s.WriteString(m.Dialect.QuoteField(table.TableName))
		s.WriteString(" (")
		if !exec {
			s.WriteString("\n")
		}
		x := 0
		for _, col := range table.Columns {
			if !col.Transient {
				if x > 0 {
					s.WriteString(sep)
				}
				s.WriteString(prefix)
				writeColumnSql(&s, col)
				x++
			}
		}
		if len(table.Keys) > 1 {
			s.WriteString(", primary key (")
			for x := range table.Keys {
				if x > 0 {
					s.WriteString(", ")
				}
				s.WriteString(m.Dialect.QuoteField(table.Keys[x].ColumnName))
			}
			s.WriteString(")")
		}
		s.WriteString(fmt.Sprintf(")%s;", m.Dialect.CreateTableSuffix()))
		if exec {
			_, err = m.Exec(s.String())
			if err != nil {
				break
			}
		} else {
			ret[table.TableName] = s.String()
		}
	}
	return ret, err
}

// DropTables iterates through TableMaps registered to this DbMap and
// executes "drop table" statements against the database for each.
func (m *DbMap) DropTables() error {
	var err error
	for i := range m.tables {
		table := m.tables[i]
		_, e := m.Exec(fmt.Sprintf("drop table %s;", m.Dialect.QuoteField(table.TableName)))
		if e != nil {
			err = e
		}
	}
	return err
}

// DropTablesIfExists iterates through TableMaps registered to this DbMap and
// executes "drop table if exists" statements against the database for each.
func (m *DbMap) DropTablesIfExists() error {
	var err error
	for i := range m.tables {
		table := m.tables[i]
		_, e := m.Exec(fmt.Sprintf("drop table if exists %s;", m.Dialect.QuoteField(table.TableName)))
		if e != nil {
			err = e
		}
	}
	return err
}

// Insert runs a SQL INSERT statement for each element in list.  List
// items must be pointers, because any interface whose TableMap has an
// auto-increment PK will have its insert Id bound to the PK struct field.
//
// Hook functions PreInsert() and/or PostInsert() will be executed
// before/after the INSERT statement if the interface defines them.
func (m *DbMap) Insert(list ...interface{}) error {
	return insert(m, m, list...)
}

// Update runs a SQL UPDATE statement for each element in list.  List
// items must be pointers.
//
// Hook functions PreUpdate() and/or PostUpdate() will be executed
// before/after the UPDATE statement if the interface defines them.
//
// Returns number of rows updated.
//
// Returns an error if SetKeys has not been called on the TableMap or if
// any interface in the list has not been registered with AddTable.
func (m *DbMap) Update(list ...interface{}) (int64, error) {
	return update(m, m, list...)
}

// Delete runs a SQL DELETE statement for each element in list.  List
// items must be pointers.
//
// Hook functions PreDelete() and/or PostDelete() will be executed
// before/after the DELETE statement if the interface defines them.
//
// Returns number of rows deleted.
//
// Returns an error if SetKeys has not been called on the TableMap or if
// any interface in the list has not been registered with AddTable.
func (m *DbMap) Delete(list ...interface{}) (int64, error) {
	return deletes(m, m, list...)
}

// Get runs a SQL SELECT to fetch a single row from the table based on the
// primary key(s)
//
// dest should be an empty value for the struct to load.
// keys should be the primary key value(s) for the row to load.  If
// multiple keys exist on the table, the order should match the column
// order specified in SetKeys() when the table mapping was defined.
//
// Hook function PostGet() will be executed
// after the SELECT statement if the interface defines it.
//
// Returns a pointer to a struct that matches or nil if no row is found.
//
// Returns an error if SetKeys has not been called on the TableMap or
// if any interface in the list has not been registered with AddTable.
func (m *DbMap) Get(dest interface{}, keys ...interface{}) error {
	return get(m, m, dest, keys...)
}

// Select runs an arbitrary SQL query, binding the columns in the result
// to fields on the struct specified by dest.  args represent the bind
// parameters for the SQL statement.
//
// Column names on the SELECT statement should be aliased to the field names
// on the struct dest. Returns an error if one or more columns in the result
// do not match.  It is OK if fields on i are not part of the SQL
// statement.
//
// Hook function PostGet() will be executed
// after the SELECT statement if the interface defines it.
//
// Values are returned in one of two ways:
//
// 1. If dest is a struct or a pointer to a struct, returns a slice of pointers to
// matching rows of type dest.
//
// 2. If dest is a pointer to a slice, the results will be appended to that slice
// and nil returned.
//
// dest does NOT need to be registered with AddTable().
func (m *DbMap) Select(dest interface{}, query string, args ...interface{}) error {
	return hookedselect(m, m, dest, query, args...)
}

// SelectOne runs an arbitrary SQL Query, binding the columns in the result to
// fields on the struct specified by dest.
func (m *DbMap) SelectOne(dest interface{}, query string, args ...interface{}) error {
	return hookedget(m, m, dest, query, args...)
}

// Exec runs an arbitrary SQL statement.  args represent the bind parameters.
// This is equivalent to running Exec() using database/sql.
func (m *DbMap) Exec(query string, args ...interface{}) (sql.Result, error) {
	m.trace(query, args)
	//stmt, err := m.Db.Prepare(query)
	//if err != nil {
	//	return nil, err
	//}
	//fmt.Println("Exec", query, args)
	return m.Db.Exec(query, args...)
}

// Begin starts a modl Transaction.
func (m *DbMap) Begin() (*Transaction, error) {
	m.trace("begin;")
	tx, err := m.Dbx.Beginx()
	if err != nil {
		return nil, err
	}
	return &Transaction{m, tx}, nil
}

// FIXME: This is a poor interface.  Checking for nils is un-go-like, and this
// function should be TableFor(i interface{}) (*TableMap, error)
// FIXME: rewrite this in terms of sqlx's reflect helpers

// TableFor returns any matching tables for the interface i or nil if not found.
// If i is a slice, then the table is given for the base slice type.
func (m *DbMap) TableFor(i interface{}) *TableMap {
	var t reflect.Type
	v := reflect.ValueOf(i)
start:
	switch v.Kind() {
	case reflect.Ptr:
		// dereference pointer and try again;  we never want to store pointer
		// types anywhere, that way we always know how to do lookups
		v = v.Elem()
		goto start
	case reflect.Slice:
		// if this is a slice of X's, we're interested in the type of X
		t = v.Type().Elem()
	default:
		t = v.Type()
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return m.TableForType(t)
}

// FIXME: returning a nil pointer is not go-like;  return (*TableMap, err) instead.

// TableForType returns any matching tables for the type t or nil if not found.
func (m *DbMap) TableForType(t reflect.Type) *TableMap {
	for _, table := range m.tables {
		if table.gotype == t {
			return table
		}
	}
	return nil
}

// TruncateTables truncates all tables in the DbMap.
func (m *DbMap) TruncateTables() error {
	return m.truncateTables(false)
}

// TruncateTablesIdentityRestart truncates all tables in the DbMap and
// resets the identity counter.
func (m *DbMap) TruncateTablesIdentityRestart() error {
	return m.truncateTables(true)
}

func (m *DbMap) truncateTables(restartIdentity bool) error {
	var err error
	var restartClause string
	for i := range m.tables {
		table := m.tables[i]
		if restartIdentity {
			restartClause = m.Dialect.RestartIdentityClause(table.TableName)
		}

		// if the restart clause exists and starts with ';', then assume it's an
		// additional query to run after we truncate.  This is true with MySQL and
		// SQLite, which do not have extra clauses for this during table truncation.
		if len(restartClause) > 0 && restartClause[0] == ';' {
			_, err = m.Exec(fmt.Sprintf("%s %s;", m.Dialect.TruncateClause(),
				m.Dialect.QuoteField(table.TableName)))
			if err != nil {
				return err
			}
			_, err = m.Exec(restartClause[1:])
			if err != nil {
				return err
			}
		} else {
			_, err := m.Exec(fmt.Sprintf("%s %s %s;", m.Dialect.TruncateClause(), m.Dialect.QuoteField(table.TableName), restartClause))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *DbMap) handle() handle {
	return &tracingHandle{h: m.Dbx, d: m}
}

func (m *DbMap) trace(query string, args ...interface{}) {
	if m.logger != nil {
		m.logger.Printf("%s%s %v", m.logPrefix, query, args)
	}
}
