package search

import (
	"github.com/sourcegraph/sourcegraph/cmd/searcher/proto"
)

type Server struct {
	proto.UnimplementedSearcherServer
}
