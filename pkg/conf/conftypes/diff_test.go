package conftypes

import (
	"reflect"
	"sort"
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestDiff(t *testing.T) {
	tests := []struct {
		name          string
		before, after SiteConfiguration
		want          []string
	}{
		{
			name: "diff",
			before: SiteConfiguration{
				CoreSiteConfiguration: schema.CoreSiteConfiguration{
					AppURL: "a",
				},
			},
			after: SiteConfiguration{
				CoreSiteConfiguration: schema.CoreSiteConfiguration{
					AppURL: "b",
				},
			},
			want: []string{"appURL"},
		},
		{
			name: "nodiff",
			before: SiteConfiguration{
				CoreSiteConfiguration: schema.CoreSiteConfiguration{
					AppURL: "a",
				},
			},
			after: SiteConfiguration{
				CoreSiteConfiguration: schema.CoreSiteConfiguration{
					AppURL: "a",
				},
			},
			want: nil,
		},
		{
			name: "slice_diff",
			before: SiteConfiguration{
				CoreSiteConfiguration: schema.CoreSiteConfiguration{
					AppURL: "a",
				}, BasicSiteConfiguration: schema.BasicSiteConfiguration{
					Langservers: []*schema.Langservers{{Address: "a"}},
				}},
			after: SiteConfiguration{
				CoreSiteConfiguration: schema.CoreSiteConfiguration{
					AppURL: "a",
				}, BasicSiteConfiguration: schema.BasicSiteConfiguration{
					Langservers: []*schema.Langservers{{Address: "b"}},
				}},
			want: []string{"langservers"},
		},
		{
			name: "slice_nodiff",
			before: SiteConfiguration{
				CoreSiteConfiguration: schema.CoreSiteConfiguration{
					AppURL: "a",
				}, BasicSiteConfiguration: schema.BasicSiteConfiguration{
					Langservers: []*schema.Langservers{{Address: "a"}},
				}},
			after: SiteConfiguration{
				CoreSiteConfiguration: schema.CoreSiteConfiguration{
					AppURL: "a",
				}, BasicSiteConfiguration: schema.BasicSiteConfiguration{
					Langservers: []*schema.Langservers{{Address: "a"}},
				}},
			want: nil,
		},
		{
			name: "multi_diff",
			before: SiteConfiguration{
				CoreSiteConfiguration: schema.CoreSiteConfiguration{
					AppURL: "a",
				}, BasicSiteConfiguration: schema.BasicSiteConfiguration{
					Langservers: []*schema.Langservers{{Address: "b"}},
				}},
			after: SiteConfiguration{
				CoreSiteConfiguration: schema.CoreSiteConfiguration{
					AppURL: "b",
				}, BasicSiteConfiguration: schema.BasicSiteConfiguration{
					Langservers: []*schema.Langservers{{Address: "a"}},
				}},
			want: []string{"appURL", "langservers"},
		},
		{
			name: "experimental_features",
			before: SiteConfiguration{
				BasicSiteConfiguration: schema.BasicSiteConfiguration{
					ExperimentalFeatures: &schema.ExperimentalFeatures{
						Discussions: "enabled",
					}},
			},
			after: SiteConfiguration{
				BasicSiteConfiguration: schema.BasicSiteConfiguration{
					ExperimentalFeatures: &schema.ExperimentalFeatures{
						Discussions: "disabled",
					}},
			},
			want: []string{"experimentalFeatures::discussions"},
		},
		{
			name: "experimental_features_noop",
			before: SiteConfiguration{
				BasicSiteConfiguration: schema.BasicSiteConfiguration{
					ExperimentalFeatures: &schema.ExperimentalFeatures{},
				},
			},
			after: SiteConfiguration{
				BasicSiteConfiguration: schema.BasicSiteConfiguration{
					ExperimentalFeatures: &schema.ExperimentalFeatures{},
				}},
			want: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := toSlice(diff(&test.before, &test.after))
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
