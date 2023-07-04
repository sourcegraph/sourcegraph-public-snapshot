package conf

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGetDeduplicatedForksIndex(t *testing.T) {
	testCases := []struct {
		name       string
		haveConfig *schema.Repositories
		wantIndex  map[api.RepoName]struct{}
	}{
		{
			name:      "config not set",
			wantIndex: map[api.RepoName]struct{}{},
		},
		{
			name:       "repositories set, but deduplicated forks is empty",
			haveConfig: &schema.Repositories{},
			wantIndex:  map[api.RepoName]struct{}{},
		},
		{
			name: "deduplicated forks is not empty",
			haveConfig: &schema.Repositories{
				DeduplicateForks: []string{
					"abc",
					"def",
				},
			},
			wantIndex: map[api.RepoName]struct{}{
				"abc": struct{}{},
				"def": struct{}{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Mock(&Unified{
				SiteConfiguration: schema.SiteConfiguration{
					Repositories: tc.haveConfig,
				},
			})

			gotIndex := GetDeduplicatedForksIndex()
			if diff := cmp.Diff(gotIndex, tc.wantIndex); diff != "" {
				t.Errorf("mismatched deduplicated repos index: (-want, +got)\n%s", diff)
			}
		})
	}
}
