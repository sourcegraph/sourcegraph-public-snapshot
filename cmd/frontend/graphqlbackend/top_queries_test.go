package graphqlbackend

import (
	"context"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"reflect"
	"testing"
)

func Test_schemaResolver_TopQueries(t *testing.T) {
	tests := []struct {
		name    string
		queries []string
		want    []*queryCountResolver
	}{
		{
			name:    "empty case",
			queries: nil,
			want:    nil,
		},
		{
			name:    "single query",
			queries: []string{""},
			want: []*queryCountResolver{
				{query: "", count: 1},
			},
		},
		{
			name:    "two of the same query",
			queries: []string{"", ""},
			want: []*queryCountResolver{
				{query: "", count: 2},
			},
		},
		{
			name:    "two different queries",
			queries: []string{"a", "b"},
			want: []*queryCountResolver{
				{query: "a", count: 1},
				{query: "b", count: 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &schemaResolver{}
			db.Mocks.RecentSearches.Get = func(ctx context.Context) ([]string, error) {
				return tt.queries, nil
			}
			got, err := s.TopQueries(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("schemaResolver.TopQueries() = %v, want %v", got, tt.want)
			}
		})
	}
}
