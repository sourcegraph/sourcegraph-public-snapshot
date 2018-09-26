package gocode

import (
	"go/importer"
	"go/types"
	"log"

	"github.com/sourcegraph/go-langserver/langserver/internal/gocode/gbimporter"
	"github.com/sourcegraph/go-langserver/langserver/internal/gocode/suggest"
)

type AutoCompleteRequest struct {
	Filename string
	Data     []byte
	Cursor   int
	Context  gbimporter.PackedContext
	Source   bool
	Builtin  bool
}

type AutoCompleteReply struct {
	Candidates []suggest.Candidate
	Len        int
}

func AutoComplete(req *AutoCompleteRequest) (res *AutoCompleteReply, err error) {
	res = &AutoCompleteReply{}
	defer func() {
		if err := recover(); err != nil {
			log.Printf("gocode panic: %s\n\n", err)

			res.Candidates = []suggest.Candidate{
				{Class: "PANIC", Name: "PANIC", Type: "PANIC"},
			}
		}
	}()

	var underlying types.ImporterFrom
	if req.Source {
		underlying = importer.For("source", nil).(types.ImporterFrom)
	} else {
		underlying = importer.Default().(types.ImporterFrom)
	}
	cfg := suggest.Config{
		Importer: gbimporter.New(&req.Context, req.Filename, underlying),
		Builtin:  req.Builtin,
	}

	candidates, d, err := cfg.Suggest(req.Filename, req.Data, req.Cursor)
	if err != nil {
		return nil, err
	}
	res.Candidates, res.Len = candidates, d
	return res, nil
}
