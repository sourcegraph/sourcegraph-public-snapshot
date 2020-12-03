package search

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Ensures graphqlbackend matches the interface we expect
func TestDefaultNewSearchResolver(t *testing.T) {
	_, err := defaultNewSearchResolver(context.Background(), &graphqlbackend.SearchArgs{
		Version:  "V2",
		Settings: &schema.Settings{},
	})
	if err != nil {
		t.Fatal(err)
	}
}
