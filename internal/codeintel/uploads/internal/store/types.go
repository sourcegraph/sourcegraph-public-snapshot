package store

import (
	"github.com/keegancsmith/sqlf"
)

type cteDefinition struct {
	name       string
	definition *sqlf.Query
}
