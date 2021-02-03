package searchcontexts

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestResolvingValidSearchContextSpecs(t *testing.T) {
	orig := envvar.SourcegraphDotComMode()
	envvar.MockSourcegraphDotComMode(true)
	defer envvar.MockSourcegraphDotComMode(orig)

	tests := []struct {
		name                  string
		searchContextSpec     string
		wantSearchContextName string
	}{
		{name: "resolve user search context", searchContextSpec: "@user", wantSearchContextName: "user"},
		{name: "resolve global search context", searchContextSpec: "global", wantSearchContextName: "global"},
		{name: "resolve empty search context as global", searchContextSpec: "", wantSearchContextName: "global"},
	}

	getNamespaceByName := func(ctx context.Context, name string) (*database.Namespace, error) {
		return &database.Namespace{Name: name, User: 1}, nil
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searchContext, err := ResolveSearchContextSpec(context.Background(), tt.searchContextSpec, getNamespaceByName)
			if err != nil {
				t.Fatal(err)
			}
			if searchContext.Name != tt.wantSearchContextName {
				t.Fatalf("got %q, expected %q", searchContext.Name, tt.wantSearchContextName)
			}
		})
	}
}

func TestResolvingInvalidSearchContextSpecs(t *testing.T) {
	orig := envvar.SourcegraphDotComMode()
	envvar.MockSourcegraphDotComMode(true)
	defer envvar.MockSourcegraphDotComMode(orig)

	tests := []struct {
		name              string
		searchContextSpec string
		wantErr           string
	}{
		{name: "invalid format", searchContextSpec: "+user", wantErr: "search context '+user' does not have the correct format (global or @username)"},
		{name: "user not found", searchContextSpec: "@user", wantErr: "search context '@user' not found"},
		{name: "empty user not found", searchContextSpec: "@", wantErr: "search context '@' not found"},
	}

	getNamespaceByName := func(ctx context.Context, name string) (*database.Namespace, error) { return &database.Namespace{}, nil }
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ResolveSearchContextSpec(context.Background(), tt.searchContextSpec, getNamespaceByName)
			if err == nil {
				t.Error("Expected error, but there was none")
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("err: got %q, expected %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestConstructingSearchContextSpecs(t *testing.T) {
	tests := []struct {
		name                  string
		searchContext         *types.SearchContext
		wantSearchContextSpec string
	}{
		{name: "global search context", searchContext: GetGlobalSearchContext(), wantSearchContextSpec: "global"},
		{name: "user search context", searchContext: &types.SearchContext{Name: "user"}, wantSearchContextSpec: "@user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searchContextSpec := GetSearchContextSpec(tt.searchContext)
			if searchContextSpec != tt.wantSearchContextSpec {
				t.Fatalf("got %q, expected %q", searchContextSpec, tt.wantSearchContextSpec)
			}
		})
	}
}
