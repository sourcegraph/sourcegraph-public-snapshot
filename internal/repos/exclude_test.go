package repos

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestBuildGitHubExcludeRule(t *testing.T) {
	assertExcluded := func(t *testing.T, rule *schema.ExcludedGitHubRepo, repo github.Repository, wantExcluded bool) {
		t.Helper()
		fn, err := buildGitHubExcludeRule(rule)
		assert.Nil(t, err)
		assert.Equal(
			t,
			wantExcluded,
			fn(repo),
			"rule.Stars=%q, rule.Size=%q, repo.StarGazerCount=%d, repo.DiskUsageKibibytes=%d",
			rule.Stars,
			rule.Size,
			repo.StargazerCount,
			repo.DiskUsageKibibytes,
		)
	}

	t.Run("stars", func(t *testing.T) {
		tests := []struct {
			rule         string
			stars        int
			wantExcluded bool
		}{
			{"< 100", 99, true},
			{"< 100", 100, false},

			{"<= 100", 100, true},
			{"<= 100", 99, true},
			{"<= 100", 101, false},

			{"> 100", 101, true},
			{"> 100", 100, false},

			{">= 100", 100, true},
			{">= 100", 101, true},
			{">= 100", 99, false},
		}
		for _, tt := range tests {
			excludeRule := &schema.ExcludedGitHubRepo{Stars: tt.rule}
			repo := github.Repository{StargazerCount: tt.stars}
			assertExcluded(t, excludeRule, repo, tt.wantExcluded)
		}
	})

	t.Run("size", func(t *testing.T) {
		tests := []struct {
			rule          string
			sizeKibibytes int
			wantExcluded  bool
		}{
			{"< 100 KiB", 99, true},
			{"< 100 KiB", 100, false},

			{"<= 100 KiB", 100, true},
			{"<= 100 KiB", 99, true},
			{"<= 100 KiB", 101, false},

			{"> 100 KiB", 101, true},
			{"> 100 KiB", 100, false},

			{">= 100 KiB", 100, true},
			{">= 100 KiB", 101, true},
			{">= 100 KiB", 99, false},

			{"< 1025 B", 1, true},
			{"< 1024 B", 1, false},
			{"< 1025 B", 2, false},

			{"< 100 KiB", 99, true},
			{"< 100 KiB", 100, false},
			{"< 100 KB", 99, false},
			{"< 102 KB", 99, true},

			{"< 100 MiB", 99, true},
			{"< 100 MiB", 102400, false},
			{"< 101 MiB", 102400, true},
			{"< 100 MB", 102400, false},
			{"< 105 MB", 102400, true},

			{"< 1 GiB", 1024*1024 - 1, true},
			{"< 1 GiB", 1024 * 1024, false},

			{"< 1 GB", 1024*1024 - 1, false},
			{"< 2 GB", 1024 * 1024, true},

			// Ignore repositories with 0 size
			{"< 1 MB", 0, false},
		}

		for _, tt := range tests {
			excludeRule := &schema.ExcludedGitHubRepo{Size: tt.rule}
			repo := github.Repository{DiskUsageKibibytes: tt.sizeKibibytes}
			assertExcluded(t, excludeRule, repo, tt.wantExcluded)
		}
	})

	t.Run("stars and size", func(t *testing.T) {
		rule := &schema.ExcludedGitHubRepo{Stars: "< 100", Size: ">= 1GB"}

		// Less than 100 stars, equal or greater than 1GB in size
		assertExcluded(t, rule, github.Repository{StargazerCount: 99, DiskUsageKibibytes: 976563}, true)
		assertExcluded(t, rule, github.Repository{StargazerCount: 99, DiskUsageKibibytes: 976563 + 1}, true)

		// Equal or greater than 100 stars, greater than 1GB in size
		assertExcluded(t, rule, github.Repository{StargazerCount: 100, DiskUsageKibibytes: 976563 + 1}, false)
		assertExcluded(t, rule, github.Repository{StargazerCount: 101, DiskUsageKibibytes: 976563 + 1}, false)

		// Greater than 100 stars, less than 1 GB
		assertExcluded(t, rule, github.Repository{StargazerCount: 101, DiskUsageKibibytes: 500}, false)
	})
}
