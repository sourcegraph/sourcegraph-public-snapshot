package dbutil

import (
	"database/sql"
	"log"
	"time"

	"github.com/kr/text"
	"github.com/sqs/modl"
)

type LoggedSQLExecutor struct {
	modl.SqlExecutor
	Logger *log.Logger
}

func (x LoggedSQLExecutor) log(started time.Time, query string, args []interface{}) {
	if x.Logger != nil {
		query = UnbindQuery(query, args...)
		x.Logger.Printf("%s elapsed\n%s\n", time.Since(started), text.Indent(query, "\t"))
	}
}

func (x LoggedSQLExecutor) Exec(query string, args ...interface{}) (sql.Result, error) {
	defer x.log(time.Now(), query, args)
	return x.SqlExecutor.Exec(query, args...)
}

func (x LoggedSQLExecutor) Select(dest interface{}, query string, args ...interface{}) error {
	defer x.log(time.Now(), query, args)
	return x.SqlExecutor.Select(dest, query, args...)
}

// UnderlyingSQLExecutor implements db.SQLExecutorWrapper.
func (x LoggedSQLExecutor) UnderlyingSQLExecutor() modl.SqlExecutor { return x.SqlExecutor }

type SQLExecutorWrapper interface {
	UnderlyingSQLExecutor() modl.SqlExecutor
}

func GetUnderlyingSQLExecutor(x modl.SqlExecutor) modl.SqlExecutor {
	if w, ok := x.(SQLExecutorWrapper); ok {
		x = GetUnderlyingSQLExecutor(w.UnderlyingSQLExecutor())
	}
	return x
}
