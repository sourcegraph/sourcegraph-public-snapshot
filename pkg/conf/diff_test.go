package conf

import (
	"reflect"
	"sort"
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestDiff(t *testing.T) {
	tests := []struct {
		name          string
		before, after *UnifiedConfiguration
		want          []string
	}{
		{
			name:   "diff",
			before: &UnifiedConfiguration{AppURL: "a"},
			after:  &UnifiedConfiguration{AppURL: "b"},
			want:   []string{"appURL"},
		},
		{
			name:   "nodiff",
			before: &UnifiedConfiguration{AppURL: "a"},
			after:  &UnifiedConfiguration{AppURL: "a"},
			want:   nil,
		},
		{
			name:   "slice_diff",
			before: &UnifiedConfiguration{AppURL: "a", ReposList: []*schema.Repository{{Path: "a"}}},
			after:  &UnifiedConfiguration{AppURL: "a", ReposList: []*schema.Repository{{Path: "b"}}},
			want:   []string{"repos.list"},
		},
		{
			name:   "slice_nodiff",
			before: &UnifiedConfiguration{AppURL: "a", ReposList: []*schema.Repository{{Path: "a"}}},
			after:  &UnifiedConfiguration{AppURL: "a", ReposList: []*schema.Repository{{Path: "a"}}},
			want:   nil,
		},
		{
			name:   "multi_diff",
			before: &UnifiedConfiguration{AppURL: "a", ReposList: []*schema.Repository{{Path: "b"}}},
			after:  &UnifiedConfiguration{AppURL: "b", ReposList: []*schema.Repository{{Path: "a"}}},
			want:   []string{"appURL", "repos.list"},
		},
		{
			name: "experimental_features",
			before: &UnifiedConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{
				Discussions: "enabled",
			}},
			after: &UnifiedConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{
				Discussions: "disabled",
			}},
			want: []string{"experimentalFeatures::discussions"},
		},
		{
			name:   "experimental_features_noop",
			before: &UnifiedConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{}},
			after:  &UnifiedConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{}},
			want:   nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := toSlice(diff(test.before, test.after))
			sort.Strings(got)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("got %#v want %#v", got, test.want)
			}
		})
	}
}

func toSlice(m map[string]struct{}) []string {
	var s []string
	for v := range m {
		s = append(s, v)
	}
	return s
}
