package search

import (
	"github.com/sourcegraph/sourcegraph/cmd/searcher/proto"
)

type Server struct {
	proto.UnimplementedSearcherServer
}

func (s *Server) Search(req *proto.SearchRequest, stream proto.Searcher_SearchServer) error {
	println("SEARCHED!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	return nil
}
