package dbutil

import "github.com/sqs/modl"

type SQLExecutorWrapper interface {
	UnderlyingSQLExecutor() modl.SqlExecutor
}

func GetUnderlyingSQLExecutor(x modl.SqlExecutor) modl.SqlExecutor {
	if w, ok := x.(SQLExecutorWrapper); ok {
		x = GetUnderlyingSQLExecutor(w.UnderlyingSQLExecutor())
	}
	return x
}
