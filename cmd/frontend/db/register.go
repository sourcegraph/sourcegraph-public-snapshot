package db

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
)

func init() {
	db.GlobalDeps = &globalDeps{}
}
