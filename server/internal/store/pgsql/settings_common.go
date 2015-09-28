package pgsql

import (
	"fmt"
	"strings"

	"github.com/sqs/modl"
)

// insertOrUpdateSQL creates an SQL query string that creates or
// updates a row in tbl (with primary key column pkeyColumn). The only
// fields inserted or updated are those in setFields (and the primary
// key column, in an insert). The setFields map maps column names to
// the desired column values.
//
// The returned SQL and args can be passed to an SQL Exec function
// directly.
func insertOrUpdateSQL(tbl, pkeyColumn string, pkeyVal interface{}, setFields map[string]interface{}) (sql string, args []interface{}) {
	if len(setFields) == 0 {
		// Nothing to update or create.
		return "", nil
	}

	// arg is a helper function that returns the bind variable that
	// refers to a (and adds a to the args list).
	arg := func(a interface{}) string {
		v := modl.PostgresDialect{}.BindVar(len(args))
		args = append(args, a)
		return v
	}

	// Generate SQL fragments for the fields we need to set.
	cols := make([]string, len(setFields))
	updates := make([]string, len(setFields))
	inserts := make([]string, len(setFields))
	i := 0
	for col, val := range setFields {
		cols[i] = col
		bv := arg(val)
		updates[i] = fmt.Sprintf(`"%s"=%s`, col, bv)
		inserts[i] = fmt.Sprintf(`%s AS "%s"`, bv, col)
		i++
	}

	// Substititions to make using the SQL template below.
	subst := map[string]string{
		"tbl":        tbl,
		"pkeyColumn": pkeyColumn,
		"updates":    strings.Join(updates, ", "),
		"inserts":    strings.Join(inserts, ", "),
		"cols":       strings.Join(cols, ", "),
		"pkeyBind":   arg(pkeyVal),
	}

	sql = `
WITH update_result AS (
  UPDATE %(tbl) SET %(updates) WHERE %(pkeyColumn)=%(pkeyBind)
  RETURNING 1
),
insert_data AS (
  SELECT %(pkeyBind) AS %(pkeyColumn), %(inserts)
)
INSERT INTO %(tbl)(%(pkeyColumn), %(cols))
SELECT * FROM insert_data
WHERE NOT EXISTS (SELECT NULL FROM update_result);
`

	// Perform substitutions.
	for k, v := range subst {
		sql = strings.Replace(sql, "%("+k+")", v, -1)
	}
	return sql, args
}
