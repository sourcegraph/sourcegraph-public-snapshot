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
		before, after *schema.SiteConfiguration
		want          []string
	}{
		{
			name:   "diff",
			before: &schema.SiteConfiguration{AppURL: "a"},
			after:  &schema.SiteConfiguration{AppURL: "b"},
			want:   []string{"appURL"},
		},
		{
			name:   "nodiff",
			before: &schema.SiteConfiguration{AppURL: "a"},
			after:  &schema.SiteConfiguration{AppURL: "a"},
			want:   nil,
		},
		{
			name:   "slice_diff",
			before: &schema.SiteConfiguration{AppURL: "a", Langservers: []*schema.Langservers{{Address: "a"}}},
			after:  &schema.SiteConfiguration{AppURL: "a", Langservers: []*schema.Langservers{{Address: "b"}}},
			want:   []string{"langservers"},
		},
		{
			name:   "slice_nodiff",
			before: &schema.SiteConfiguration{AppURL: "a", Langservers: []*schema.Langservers{{Address: "a"}}},
			after:  &schema.SiteConfiguration{AppURL: "a", Langservers: []*schema.Langservers{{Address: "a"}}},
			want:   nil,
		},
		{
			name:   "multi_diff",
			before: &schema.SiteConfiguration{AppURL: "a", Langservers: []*schema.Langservers{{Address: "b"}}},
			after:  &schema.SiteConfiguration{AppURL: "b", Langservers: []*schema.Langservers{{Address: "a"}}},
			want:   []string{"appURL", "langservers"},
		},
		{
			name: "experimental_features",
			before: &schema.SiteConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{
				Discussions: "enabled",
			}},
			after: &schema.SiteConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{
				Discussions: "disabled",
			}},
			want: []string{"experimentalFeatures::discussions"},
		},
		{
			name:   "experimental_features_noop",
			before: &schema.SiteConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{}},
			after:  &schema.SiteConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{}},
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
