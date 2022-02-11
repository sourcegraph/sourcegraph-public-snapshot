package inference

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestRustPatterns(t *testing.T) {
	testLangPatterns(t, RustPatterns(), []PathTestCase{
		{"Cargo.toml", true},
		{"subdir/Cargo.toml", true},
		{"a.rs", false},
		{"subdir/a.rs", false},
	})
}

func TestInferRustIndexJobs(t *testing.T) {
	paths := []string{
		"dir1/src/a.rs",
		"dir1/Cargo.toml",
		"dir2/src/b.rs",
	}

	expectedIndexJobs := []config.IndexJob{
		{
			Indexer:     "sourcegraph/lsif-rust",
			IndexerArgs: []string{"lsif-rust", "index"},
			Outfile:     "dump.lsif",
			Root:        "",
			Steps:       []config.DockerStep{},
		},
	}
	if diff := cmp.Diff(expectedIndexJobs, InferRustIndexJobs(NewMockGitClient(), paths)); diff != "" {
		t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
	}
}
