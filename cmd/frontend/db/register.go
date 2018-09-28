package db

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
)

var (
	Pkgs       = &pkgs{}
	GlobalDeps = &globalDeps{}
)

func init() {
	db.Pkgs = Pkgs
	db.GlobalDeps = GlobalDeps
}
