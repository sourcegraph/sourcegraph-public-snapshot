package main

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

// computeClient is an interface so we can swap out a streaming vs grahql based
// compute API. It only supports the methods that streaming supports.
type computeClient interface {
	Compute(query string) ([]*gqltestutil.ComputeResult, error)
}

func testComputeClient(t *testing.T, client computeClient) {
	t.Run("errors", func(t *testing.T) {
		tests := []struct {
			query string
		}{
			{
				// Need an erroring query.
				query: "",
			},
		}
		for _, test := range tests {
			t.Run(test.query, func(t *testing.T) {
				// TODO: not actually sure how graphQL compute endpoint handles errors.
				results, err := client.Compute(test.query)
				if len(results) != 0 {
					t.Errorf("Expected err, got results: %v", results)
				}
				if err == nil {
					t.Error("Expected err, got nil")
				}
			})
		}
	})

}

func TestCompute(t *testing.T) {
	// TODO: clone some github repos.

	t.Run("graphql", func(t *testing.T) {
		testComputeClient(t, client)
	})

	streamClient := &gqltestutil.ComputeStreamClient{Client: client}
	t.Run("stream", func(t *testing.T) {
		testComputeClient(t, streamClient)
	})
}
