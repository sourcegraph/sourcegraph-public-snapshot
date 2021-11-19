package search

import (
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/parser"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/sqlite"
)

type GitserverClient interface {
	sqlite.GitserverClient
	parser.GitserverClient
}
