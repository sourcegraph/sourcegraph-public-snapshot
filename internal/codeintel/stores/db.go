package stores

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type CodeIntelDB interface {
	basestore.ShareableStore
}

type codeIntelDB struct {
	*basestore.Store
}

func NewCodeIntelDB(inner *sql.DB) CodeIntelDB {
	return &codeIntelDB{basestore.NewWithHandle(basestore.NewHandleWithDB(inner, sql.TxOptions{}))}
}
