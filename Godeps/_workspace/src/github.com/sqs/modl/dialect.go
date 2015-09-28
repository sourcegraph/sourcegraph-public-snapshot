package modl

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Dialect is an interface that encapsulates behaviors that differ across
// SQL databases.
type Dialect interface {

	// ToSqlType returns the SQL column type to use when creating a table of the
	// given Go Type.  maxsize can be used to switch based on size.  For example,
	// in MySQL []byte could map to BLOB, MEDIUMBLOB, or LONGBLOB depending on the
	// maxsize
	ToSqlType(col *ColumnMap) string

	// AutoIncrStr returns a string to append to primary key column definitions
	AutoIncrStr() string

	//AutoIncrBindValue returns the default value for an auto increment row in an insert.
	AutoIncrBindValue() string

	// AutoIncrInsertSuffix returns a string to append to an insert statement which
	// will make it return the auto-increment key on insert.
	AutoIncrInsertSuffix(col *ColumnMap) string

	// CreateTableSuffix returns a string to append to "create table" statement for
	// vendor specific table attributes, eg. MySQL engine
	CreateTableSuffix() string

	// InsertAutoIncr
	InsertAutoIncr(e SqlExecutor, insertSql string, params ...interface{}) (int64, error)
	// InsertAutIncrAny takes a destination for non-integer auto-incr, like
	// uuids which scan to strings, hashes, etc.
	InsertAutoIncrAny(e SqlExecutor, insertSql string, dest interface{}, params ...interface{}) error

	// BindVar returns the variable string to use when forming SQL statements
	// in many dbs it is "?", but Postgres requires '$#'
	//
	// The supplied int is a zero based index of the bindvar in the statement
	BindVar(i int) string

	// QuoteField returns a quoted version of the field name.
	QuoteField(field string) string

	// TruncateClause is a string used to truncate tables.
	TruncateClause() string

	// RestartIdentityClause returns a string used to reset the identity counter
	// when truncating tables.  If the string starts with a ';', it is assumed to
	// be a separate query and is executed separately.
	RestartIdentityClause(table string) string

	// DriverName returns the driver name for a dialect.
	DriverName() string
}

func standardInsertAutoIncr(e SqlExecutor, insertSql string, params ...interface{}) (int64, error) {
	res, err := e.handle().Exec(insertSql, params...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func standardAutoIncrAny(e SqlExecutor, insertSql string, dest interface{}, params ...interface{}) error {
	rows, err := e.handle().Queryx(insertSql, params...)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return rows.Scan(dest)
	}
	return fmt.Errorf("No auto-incr value returned for insert: `%s` error: %s", insertSql, rows.Err())
}

// -- sqlite3

// SqliteDialect implements the Dialect interface for Sqlite3.
type SqliteDialect struct {
	suffix string
}

// DriverName returns the driverName for sqlite.
func (d SqliteDialect) DriverName() string {
	return "sqlite"
}

// ToSqlType maps go types to sqlite types.
func (d SqliteDialect) ToSqlType(col *ColumnMap) string {
	switch col.gotype.Kind() {
	case reflect.Bool:
		return "integer"
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float64, reflect.Float32:
		return "real"
	case reflect.Slice:
		if col.gotype.Elem().Kind() == reflect.Uint8 {
			return "blob"
		}
	}

	switch col.gotype.Name() {
	case "NullableInt64":
		return "integer"
	case "NullableFloat64":
		return "real"
	case "NullableBool":
		return "integer"
	case "NullableBytes":
		return "blob"
	case "Time", "NullTime":
		return "datetime"
	}

	// sqlite ignores maxsize, so we will do that here too
	return fmt.Sprintf("text")
}

// AutoIncrStr returns autoincrement clause for sqlite.
func (d SqliteDialect) AutoIncrStr() string {
	return "autoincrement"
}

// AutoIncrBindValue returns the bind value for auto incr fields in sqlite.
func (d SqliteDialect) AutoIncrBindValue() string {
	return "null"
}

// AutoIncrInsertSuffix returns the auto incr insert suffix for sqlite.
func (d SqliteDialect) AutoIncrInsertSuffix(col *ColumnMap) string {
	return ""
}

// CreateTableSuffix returns d.suffix
func (d SqliteDialect) CreateTableSuffix() string {
	return d.suffix
}

// BindVar returns "?", the simpler of the sqlite bindvars.
func (d SqliteDialect) BindVar(i int) string {
	return "?"
}

// InsertAutoIncr runs the standard
func (d SqliteDialect) InsertAutoIncr(e SqlExecutor, insertSql string, params ...interface{}) (int64, error) {
	return standardInsertAutoIncr(e, insertSql, params...)
}

func (d SqliteDialect) InsertAutoIncrAny(e SqlExecutor, insertSql string, dest interface{}, params ...interface{}) error {
	return standardAutoIncrAny(e, insertSql, dest, params...)
}

// QuoteField quotes f with "" for sqlite
func (d SqliteDialect) QuoteField(f string) string {
	return `"` + f + `"`
}

// TruncateClause returns the truncate clause for sqlite.  There is no TRUNCATE
// statement in sqlite3, but DELETE FROM uses a truncate optimization:
// http://www.sqlite.org/lang_delete.html
func (d SqliteDialect) TruncateClause() string {
	return "delete from"
}

// RestartIdentityClause restarts the sqlite_sequence for the provided table.
// It is executed by TruncateTable as a separate query.
func (d SqliteDialect) RestartIdentityClause(table string) string {
	return "; DELETE FROM sqlite_sequence WHERE name='" + table + "'"
}

// -- PostgreSQL

// PostgresDialect implements the Dialect interface for PostgreSQL.
type PostgresDialect struct {
	suffix string
}

// DriverName returns the driverName for postgres.
func (d PostgresDialect) DriverName() string {
	return "postgres"
}

// ToSqlType maps go types to postgres types.
func (d PostgresDialect) ToSqlType(col *ColumnMap) string {

	switch col.gotype.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Uint16, reflect.Uint32:
		if col.isAutoIncr {
			return "serial"
		}
		return "integer"
	case reflect.Int64, reflect.Uint64:
		if col.isAutoIncr {
			return "bigserial"
		}
		return "bigint"
	case reflect.Float64, reflect.Float32:
		return "real"
	case reflect.Slice:
		if col.gotype.Elem().Kind() == reflect.Uint8 {
			return "bytea"
		}
	}

	switch col.gotype.Name() {
	case "NullableInt64":
		return "bigint"
	case "NullableFloat64":
		return "double"
	case "NullableBool":
		return "smallint"
	case "NullableBytes":
		return "bytea"
	case "Time", "Nulltime":
		return "timestamp with time zone"
	}

	maxsize := col.MaxSize
	if col.MaxSize < 1 {
		maxsize = 255
	}
	return fmt.Sprintf("varchar(%d)", maxsize)
}

// AutoIncrStr returns empty string, as it's not used in postgres.
func (d PostgresDialect) AutoIncrStr() string {
	return ""
}

// AutoIncrBindValue returns 'default' for default auto incr bind values.
func (d PostgresDialect) AutoIncrBindValue() string {
	return "default"
}

// AutoIncrInsertSuffix appnds 'returning colname' to a query so that the
// new autoincrement value for the column name is returned by Insert.
func (d PostgresDialect) AutoIncrInsertSuffix(col *ColumnMap) string {
	return " returning " + col.ColumnName
}

// CreateTableSuffix returns the configured suffix.
func (d PostgresDialect) CreateTableSuffix() string {
	return d.suffix
}

// BindVar returns "$(i+1)"
func (d PostgresDialect) BindVar(i int) string {
	return fmt.Sprintf("$%d", i+1)
}

// InsertAutoIncr inserts via a query and reads the resultant rows for the new
// auto increment ID, as it's not returned with the result in PostgreSQL.
func (d PostgresDialect) InsertAutoIncr(e SqlExecutor, insertSql string, params ...interface{}) (int64, error) {
	rows, err := e.handle().Queryx(insertSql, params...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if rows.Next() {
		var id int64
		err := rows.Scan(&id)
		return id, err
	}

	return 0, errors.New("No serial value returned for insert: " + insertSql + ", error: " + rows.Err().Error())
}

func (d PostgresDialect) InsertAutoIncrAny(e SqlExecutor, insertSql string, dest interface{}, params ...interface{}) error {
	return standardAutoIncrAny(e, insertSql, dest, params...)
}

// QuoteField quotes f with ""
func (d PostgresDialect) QuoteField(f string) string {
	return `"` + sqlx.NameMapper(f) + `"`
}

// TruncateClause returns 'truncate'
func (d PostgresDialect) TruncateClause() string {
	return "truncate"
}

// RestartIdentityClause returns 'restart identity', which will restart serial
// sequences for this table at the same time as a truncation is performed.
func (d PostgresDialect) RestartIdentityClause(table string) string {
	return "restart identity"
}

// -- MySQL

// MySQLDialect is an implementation of Dialect for MySQL databases.
type MySQLDialect struct {

	// Engine is the storage engine to use "InnoDB" vs "MyISAM" for example
	Engine string

	// Encoding is the character encoding to use for created tables
	Encoding string
}

// DriverName returns "mysql", used by the ziutek and the go-sql-driver
// versions of the MySQL driver.
func (d MySQLDialect) DriverName() string {
	return "mysql"
}

// ToSqlType maps go types to MySQL types.
func (d MySQLDialect) ToSqlType(col *ColumnMap) string {
	switch col.gotype.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Uint16, reflect.Uint32:
		return "int"
	case reflect.Int64, reflect.Uint64:
		return "bigint"
	case reflect.Float64, reflect.Float32:
		return "double"
	case reflect.Slice:
		if col.gotype.Elem().Kind() == reflect.Uint8 {
			return "mediumblob"
		}
	}

	switch col.gotype.Name() {
	case "NullableInt64":
		return "bigint"
	case "NullableFloat64":
		return "double"
	case "NullableBool":
		return "tinyint"
	case "NullableBytes":
		return "mediumblob"
	case "Time", "NullTime":
		return "datetime"
	}

	maxsize := col.MaxSize
	if col.MaxSize < 1 {
		maxsize = 255
	}
	return fmt.Sprintf("varchar(%d)", maxsize)
}

// AutoIncrStr returns "auto_increment".
func (d MySQLDialect) AutoIncrStr() string {
	return "auto_increment"
}

// AutoIncrBindValue returns "null", default for auto increment fields in MySQL.
func (d MySQLDialect) AutoIncrBindValue() string {
	return "null"
}

// AutoIncrInsertSuffix returns "".
func (d MySQLDialect) AutoIncrInsertSuffix(col *ColumnMap) string {
	return ""
}

// CreateTableSuffix returns engine=%s charset=%s  based on values stored on struct
func (d MySQLDialect) CreateTableSuffix() string {
	return fmt.Sprintf(" engine=%s charset=%s", d.Engine, d.Encoding)
}

// BindVar returns "?"
func (d MySQLDialect) BindVar(i int) string {
	return "?"
}

// InsertAutoIncr runs the standard Insert Exec, which uses LastInsertId to get
// the value of the auto increment column.
func (d MySQLDialect) InsertAutoIncr(e SqlExecutor, insertSql string, params ...interface{}) (int64, error) {
	return standardInsertAutoIncr(e, insertSql, params...)
}

func (d MySQLDialect) InsertAutoIncrAny(e SqlExecutor, insertSql string, dest interface{}, params ...interface{}) error {
	return standardAutoIncrAny(e, insertSql, dest, params...)
}

// QuoteField quotes f using ``.
func (d MySQLDialect) QuoteField(f string) string {
	return "`" + f + "`"
}

// FIXME: use sqlx's rebind, which was written after it had been created for modl

// ReBind formats the bindvars in the query string (these are '?') for the dialect.
func ReBind(query string, dialect Dialect) string {

	binder := dialect.BindVar(0)
	if binder == "?" {
		return query
	}

	for i, j := 0, strings.Index(query, "?"); j >= 0; i++ {
		query = strings.Replace(query, "?", dialect.BindVar(i), 1)
		j = strings.Index(query, "?")
	}
	return query
}

// TruncateClause returns 'truncate'.
func (d MySQLDialect) TruncateClause() string {
	return "truncate"
}

// RestartIdentityClause alters the table's AUTO_INCREMENT value after truncation,
// as MySQL doesn't have an identity clause for the truncate statement.
func (d MySQLDialect) RestartIdentityClause(table string) string {
	return "; alter table " + table + " AUTO_INCREMENT = 1"
}
