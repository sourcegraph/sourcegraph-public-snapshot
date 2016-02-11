package dbutil

import "gopkg.in/gorp.v1"

type SQLExecutorWrapper interface {
	UnderlyingSQLExecutor() gorp.SqlExecutor
}

func GetUnderlyingSQLExecutor(x gorp.SqlExecutor) gorp.SqlExecutor {
	if w, ok := x.(SQLExecutorWrapper); ok {
		x = GetUnderlyingSQLExecutor(w.UnderlyingSQLExecutor())
	}
	return x
}
