package main

import (
	"fmt"
	"os"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	querytypes "github.com/sourcegraph/sourcegraph/internal/search/query/types"
)

func isEquivalent(q string) {
	if query.ContainsAndOrKeyword(q) {
		// Things containing operators will always differ--don't test.
		return
	}

	var oldQuery, newQuery query.QueryInfo

	newQuery, newErr := query.ProcessAndOr(q, query.ParserOptions{SearchType: query.SearchTypeRegex})
	oldQuery, oldErr := query.Process(q, query.SearchTypeRegex)

	if newErr != nil && oldErr == nil {
		fmt.Printf("New parser has stricter validation: new parser reports %s\n", newErr)
		return
	}

	if newErr == nil && oldErr != nil {
		fmt.Printf("New parser has weaker validation: old parser reports %s\n", oldErr)
		return
	}

	var oldFields, newFields querytypes.Fields

	oldFields = oldQuery.Fields()
	newFields = newQuery.Fields()

	if diff := cmp.Diff(oldFields.String(), newFields.String()); diff != "" {
		fmt.Printf("Input: %s\n", q)
		panic(fmt.Sprintf("-old, +new: %s", diff))
	}
}

func main() {
	isEquivalent(os.Args[1])
}
