package searchcontexts

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

func TestResolvingValidSearchContextSpecs(t *testing.T) {
	tests := []struct {
		name                  string
		searchContextSpec     string
		wantSearchContextName string
	}{
		{name: "resolve user search context", searchContextSpec: "@user", wantSearchContextName: "user"},
		{name: "resolve global search context", searchContextSpec: "global", wantSearchContextName: "global"},
		{name: "resolve empty search context as global", searchContextSpec: "", wantSearchContextName: "global"},
	}

	database.Mocks.Namespaces.GetByName = func(ctx context.Context, name string) (*database.Namespace, error) {
		return &database.Namespace{Name: name, User: 1}, nil
	}
	defer func() { database.Mocks.Namespaces.GetByName = nil }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searchContext, err := ResolveSearchContextSpec(context.Background(), tt.searchContextSpec)
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
	tests := []struct {
		name              string
		searchContextSpec string
		wantErr           string
	}{
		{name: "invalid format", searchContextSpec: "+user", wantErr: "search context spec does not have the correct format"},
		{name: "user not found", searchContextSpec: "@user", wantErr: "search context not found"},
		{name: "empty user not found", searchContextSpec: "@", wantErr: "search context not found"},
	}

	database.Mocks.Namespaces.GetByName = func(ctx context.Context, name string) (*database.Namespace, error) {
		return &database.Namespace{}, nil
	}
	defer func() { database.Mocks.Namespaces.GetByName = nil }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ResolveSearchContextSpec(context.Background(), tt.searchContextSpec)
			if err == nil {
				t.Error("Expected error, but there was none")
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("err: got %q, expected %q", err.Error(), tt.wantErr)
			}
		})
	}
}
