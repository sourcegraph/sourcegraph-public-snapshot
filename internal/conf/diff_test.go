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
		before, after *Unified
		want          []string
	}{
		{
			name:   "diff",
			before: &Unified{SiteConfiguration: schema.SiteConfiguration{ExternalURL: "a"}},
			after:  &Unified{SiteConfiguration: schema.SiteConfiguration{ExternalURL: "b"}},
			want:   []string{"externalURL"},
		},
		{
			name:   "nodiff",
			before: &Unified{SiteConfiguration: schema.SiteConfiguration{ExternalURL: "a"}},
			after:  &Unified{SiteConfiguration: schema.SiteConfiguration{ExternalURL: "a"}},
			want:   nil,
		},
		{
			name: "slice_diff",
			before: &Unified{
				SiteConfiguration: schema.SiteConfiguration{GitCloneURLToRepositoryName: []*schema.CloneURLToRepositoryName{{From: "a"}}, ExternalURL: "a"},
			},
			after: &Unified{
				SiteConfiguration: schema.SiteConfiguration{GitCloneURLToRepositoryName: []*schema.CloneURLToRepositoryName{{From: "b"}}, ExternalURL: "a"},
			},
			want: []string{"git.cloneURLToRepositoryName"},
		},
		{
			name: "slice_nodiff",
			before: &Unified{
				SiteConfiguration: schema.SiteConfiguration{GitCloneURLToRepositoryName: []*schema.CloneURLToRepositoryName{{From: "a"}}, ExternalURL: "a"},
			},
			after: &Unified{
				SiteConfiguration: schema.SiteConfiguration{GitCloneURLToRepositoryName: []*schema.CloneURLToRepositoryName{{From: "a"}}, ExternalURL: "a"},
			},
		},
		{
			name: "multi_diff",
			before: &Unified{
				SiteConfiguration: schema.SiteConfiguration{GitCloneURLToRepositoryName: []*schema.CloneURLToRepositoryName{{From: "b"}}, ExternalURL: "a"},
			},
			after: &Unified{
				SiteConfiguration: schema.SiteConfiguration{GitCloneURLToRepositoryName: []*schema.CloneURLToRepositoryName{{From: "a"}}, ExternalURL: "b"},
			},
			want: []string{"externalURL", "git.cloneURLToRepositoryName"},
		},
		{
			name: "experimental_features",
			before: &Unified{SiteConfiguration: schema.SiteConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{
				StructuralSearch: "enabled",
			}}},
			after: &Unified{SiteConfiguration: schema.SiteConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{
				StructuralSearch: "disabled",
			}}},
			want: []string{"experimentalFeatures::structuralSearch"},
		},
		{
			name:   "experimental_features_noop",
			before: &Unified{SiteConfiguration: schema.SiteConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{}}},
			after:  &Unified{SiteConfiguration: schema.SiteConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{}}},
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
