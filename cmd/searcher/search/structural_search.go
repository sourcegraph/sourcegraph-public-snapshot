package search

import (
	"context"
	"fmt"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
)

func combyFind(ctx context.Context, zipPath string, fileMatchLimit int, onlyFiles []string) (matches []protocol.FileMatch, limitHit bool, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CombyFind")
	ext.Component.Set(span, "structural_search")

	// XXX shares const with regex_search.
	if fileMatchLimit > maxFileMatches || fileMatchLimit <= 0 {
		fileMatchLimit = maxFileMatches
	}

	// read stdout and kill comby if we reach fileMatchLimit

	// How to send the filepaths to search thanks to indexed search?
	fmt.Println("Hi, yes, I have nothing for you")
	fmt.Println("Hi, yes, I have nothing for you")
	fmt.Println("Hi, yes, I have nothing for you")
	fmt.Println("Hi, yes, I have nothing for you")
	fmt.Println("Hi, yes, I have nothing for you")
	fmt.Println("Hi, yes, I have nothing for you")

	return matches, false, err
}
