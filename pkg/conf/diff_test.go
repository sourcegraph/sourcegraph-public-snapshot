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
			before: &UnifiedConfiguration{Core: schema.CoreSiteConfiguration{AppURL: "a"}},
			after:  &UnifiedConfiguration{Core: schema.CoreSiteConfiguration{AppURL: "b"}},
			want:   []string{"core::appURL"},
		},
		{
			name:   "nodiff",
			before: &UnifiedConfiguration{Core: schema.CoreSiteConfiguration{AppURL: "a"}},
			after:  &UnifiedConfiguration{Core: schema.CoreSiteConfiguration{AppURL: "a"}},
			want:   nil,
		},
		{
			name: "slice_diff",
			before: &UnifiedConfiguration{
				SiteConfiguration: schema.SiteConfiguration{ReposList: []*schema.Repository{{Path: "a"}}},
				Core:              schema.CoreSiteConfiguration{AppURL: "a"},
			},
			after: &UnifiedConfiguration{
				SiteConfiguration: schema.SiteConfiguration{ReposList: []*schema.Repository{{Path: "b"}}},
				Core:              schema.CoreSiteConfiguration{AppURL: "a"},
			},
			want: []string{"repos.list"},
		},
		{
			name: "slice_nodiff",
			before: &UnifiedConfiguration{
				SiteConfiguration: schema.SiteConfiguration{ReposList: []*schema.Repository{{Path: "a"}}},
				Core:              schema.CoreSiteConfiguration{AppURL: "a"},
			},
			after: &UnifiedConfiguration{
				SiteConfiguration: schema.SiteConfiguration{ReposList: []*schema.Repository{{Path: "a"}}},
				Core:              schema.CoreSiteConfiguration{AppURL: "a"},
			},
		},
		{
			name: "multi_diff",
			before: &UnifiedConfiguration{
				SiteConfiguration: schema.SiteConfiguration{ReposList: []*schema.Repository{{Path: "b"}}},
				Core:              schema.CoreSiteConfiguration{AppURL: "a"},
			},
			after: &UnifiedConfiguration{
				SiteConfiguration: schema.SiteConfiguration{ReposList: []*schema.Repository{{Path: "a"}}},
				Core:              schema.CoreSiteConfiguration{AppURL: "b"},
			},
			want: []string{"core::appURL", "repos.list"},
		},
		{
			name: "experimental_features",
			before: &UnifiedConfiguration{SiteConfiguration: schema.SiteConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{
				Discussions: "enabled",
			}}},
			after: &UnifiedConfiguration{SiteConfiguration: schema.SiteConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{
				Discussions: "disabled",
			}}},
			want: []string{"experimentalFeatures::discussions"},
		},
		{
			name:   "experimental_features_noop",
			before: &UnifiedConfiguration{SiteConfiguration: schema.SiteConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{}}},
			after:  &UnifiedConfiguration{SiteConfiguration: schema.SiteConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{}}},
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
