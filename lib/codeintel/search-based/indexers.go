package search_based

import (
	"github.com/sourcegraph/sourcegraph/lib/codeintel/search-based/api"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/search-based/golang"
)

var AllIndexers = []api.Indexer{golang.GoIndexer{}}
