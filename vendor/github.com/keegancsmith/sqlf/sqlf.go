// Package sqlf generates parameterized SQL statements in Go, sprintf style.
//
// A simple example:
//
//   q := sqlf.Sprintf("SELECT * FROM users WHERE country = %s AND age > %d", "US", 27);
//   rows, err := db.Query(q.Query(sqlf.SimpleBindVar), q.Args()...) // db is a database/sql.DB
//
// sqlf.Sprintf does not return a string. It returns *sqlf.Query which has
// methods for a parameterized SQL query and arguments. You then pass that to
// db.Query, db.Exec, etc. This is not like using fmt.Sprintf, which could
// expose you to malformed SQL or SQL injection attacks.
//
// sqlf.Query can be passed as an argument to sqlf.Sprintf. It will "flatten"
// the query string, while preserving the correct variable binding. This
// allows you to easily compose and build SQL queries. See the below examples
// to find out more.
package sqlf

import (
	"fmt"
	"io"
	"strings"
)

// Query stores a SQL expression and arguments for passing on to
// database/sql/db.Query or gorp.SqlExecutor.
type Query struct {
	fmt  string
	args []interface{}
}

// Sprintf formats according to a format specifier and returns the resulting
// Query.
func Sprintf(format string, args ...interface{}) *Query {
	f := make([]interface{}, len(args))
	a := make([]interface{}, 0, len(args))
	for i, arg := range args {
		if q, ok := arg.(*Query); ok {
			f[i] = ignoreFormat{q.fmt}
			a = append(a, q.args...)
		} else {
			f[i] = ignoreFormat{"%s"}
			a = append(a, arg)
		}
	}
	return &Query{
		fmt:  fmt.Sprintf(format, f...),
		args: a,
	}
}

// Query returns a string for use in database/sql/db.Query. binder is used to
// update the format specifiers with the relevant BindVar format
func (q *Query) Query(binder BindVar) string {
	a := make([]interface{}, len(q.args))
	for i := range a {
		a[i] = ignoreFormat{binder.BindVar(i)}
	}
	return fmt.Sprintf(q.fmt, a...)
}

// Args returns the args for use in database/sql/db.Query along with
// q.Query()
func (q *Query) Args() []interface{} {
	return q.args
}

// Join concatenates the elements of queries to create a single Query. The
// separator string sep is placed between elements in the resulting Query.
//
// This is commonly used to join clauses in a WHERE query. As such sep is
// usually "AND" or "OR".
func Join(queries []*Query, sep string) *Query {
	f := make([]string, 0, len(queries))
	var a []interface{}
	for _, q := range queries {
		f = append(f, q.fmt)
		a = append(a, q.args...)
	}
	return &Query{
		fmt:  strings.Join(f, " "+sep+" "),
		args: a,
	}
}

type ignoreFormat struct{ s string }

func (e ignoreFormat) Format(f fmt.State, c rune) {
	io.WriteString(f, e.s)
}
