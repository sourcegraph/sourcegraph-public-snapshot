package search

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/zoekt"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/search/limits"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestZoektParameters(t *testing.T) {
	documentRanksWeight := 42.0

	cases := []struct {
		name                  string
		context               context.Context
		params                *ZoektParameters
		rankingFeatures       *schema.Ranking
		want                  *zoekt.SearchOptions
		enableDocumentRanking bool
	}{
		{
			name:    "test defaults",
			context: context.Background(),
			params: &ZoektParameters{
				FileMatchLimit: limits.DefaultMaxSearchResultsStreaming,
			},
			want: &zoekt.SearchOptions{
				ShardMaxMatchCount:  10000,
				TotalMaxMatchCount:  100000,
				MaxWallTime:         20000000000,
				MaxDocDisplayCount:  500,
				ChunkMatches:        true,
				FlushWallTime:       500000000,
				DocumentRanksWeight: 4500,
			},
		},
		{
			name:    "test defaults (with document ranking enabled)",
			context: context.Background(),
			params: &ZoektParameters{
				FileMatchLimit: limits.DefaultMaxSearchResultsStreaming,
			},
			enableDocumentRanking: true,
			want: &zoekt.SearchOptions{
				ShardMaxMatchCount:  10000,
				TotalMaxMatchCount:  100000,
				MaxWallTime:         20000000000,
				MaxDocDisplayCount:  500,
				ChunkMatches:        true,
				FlushWallTime:       500000000,
				UseDocumentRanks:    true,
				DocumentRanksWeight: 4500,
			},
		},
		{
			name:    "test repo search defaults",
			context: context.Background(),
			params: &ZoektParameters{
				Select:         []string{filter.Repository},
				FileMatchLimit: limits.DefaultMaxSearchResultsStreaming,
			},
			// Most important is ShardRepoMaxMatchCount=1. Otherwise we still
			// want to set normal limits so we respect things like low file
			// match limits.
			want: &zoekt.SearchOptions{
				ShardMaxMatchCount:     10_000,
				TotalMaxMatchCount:     100_000,
				ShardRepoMaxMatchCount: 1,
				MaxWallTime:            20000000000,
				MaxDocDisplayCount:     500,
				ChunkMatches:           true,
			},
		},
		{
			name:    "test repo search low match count",
			context: context.Background(),
			params: &ZoektParameters{
				Select:         []string{filter.Repository},
				FileMatchLimit: 5,
			},
			// This is like the above test, but we are testing
			// MaxDocDisplayCount is adjusted to 5.
			want: &zoekt.SearchOptions{
				ShardMaxMatchCount:     10_000,
				TotalMaxMatchCount:     100_000,
				ShardRepoMaxMatchCount: 1,
				MaxWallTime:            20000000000,
				MaxDocDisplayCount:     5,
				ChunkMatches:           true,
			},
		},
		{
			name:    "test large file match limit",
			context: context.Background(),
			params: &ZoektParameters{
				FileMatchLimit: 100_000,
			},
			want: &zoekt.SearchOptions{
				ShardMaxMatchCount:  100_000,
				TotalMaxMatchCount:  100_000,
				MaxWallTime:         20000000000,
				MaxDocDisplayCount:  100_000,
				ChunkMatches:        true,
				FlushWallTime:       500000000,
				DocumentRanksWeight: 4500,
			},
		},
		{
			name:    "test document ranks weight",
			context: context.Background(),
			rankingFeatures: &schema.Ranking{
				DocumentRanksWeight: &documentRanksWeight,
			},
			params: &ZoektParameters{
				FileMatchLimit: limits.DefaultMaxSearchResultsStreaming,
			},
			want: &zoekt.SearchOptions{
				ShardMaxMatchCount:  10000,
				TotalMaxMatchCount:  100000,
				MaxWallTime:         20000000000,
				FlushWallTime:       500000000,
				MaxDocDisplayCount:  500,
				ChunkMatches:        true,
				DocumentRanksWeight: 42,
			},
		},
		{
			name:    "test flush wall time",
			context: context.Background(),
			rankingFeatures: &schema.Ranking{
				FlushWallTimeMS: 3141,
			},
			params: &ZoektParameters{
				FileMatchLimit: limits.DefaultMaxSearchResultsStreaming,
			},
			want: &zoekt.SearchOptions{
				ShardMaxMatchCount:  10000,
				TotalMaxMatchCount:  100000,
				MaxWallTime:         20000000000,
				FlushWallTime:       3141000000,
				MaxDocDisplayCount:  500,
				ChunkMatches:        true,
				DocumentRanksWeight: 4500,
			},
		},
		{
			name:    "test keyword scoring",
			context: context.Background(),
			params: &ZoektParameters{
				FileMatchLimit: limits.DefaultMaxSearchResultsStreaming,
				PatternType:    query.SearchTypeKeyword,
			},
			want: &zoekt.SearchOptions{
				ShardMaxMatchCount:  100000,
				TotalMaxMatchCount:  1000000,
				MaxWallTime:         20000000000,
				FlushWallTime:       2000000000, // for keyword search, default is 2 sec
				MaxDocDisplayCount:  500,
				ChunkMatches:        true,
				DocumentRanksWeight: 4500,
				UseKeywordScoring:   true},
		},
	}

	enabled := true

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.enableDocumentRanking {
				conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{CodeIntelRankingDocumentReferenceCountsEnabled: &enabled}})
				defer conf.Mock(nil)
			}

			if tt.rankingFeatures != nil {
				cfg := conf.Get()
				cfg.ExperimentalFeatures.Ranking = tt.rankingFeatures
				conf.Mock(cfg)

				defer func() {
					cfg.ExperimentalFeatures.Ranking = nil
					conf.Mock(cfg)
				}()
			}

			got := tt.params.ToSearchOptions(tt.context)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("search params mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
