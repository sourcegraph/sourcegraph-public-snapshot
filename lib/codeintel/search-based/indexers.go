package search_based

import (
	"github.com/sourcegraph/sourcegraph/lib/codeintel/search-based/api"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/search-based/golang"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/search-based/java"
)

var AllIndexers = []api.Indexer{golang.Indexer{}, java.Indexer{}}
