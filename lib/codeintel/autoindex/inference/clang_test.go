package inference

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestInferClangJobHints(t *testing.T) {
	paths := []string{
		"CMakeLists.txt",
		"dir/cmakelists.txt",
		"other/test.cpp",
		"other/test.h",
	}

	expectedHints := []config.IndexJobHint{
		{
			Root:           ".",
			Indexer:        "sourcegraph/lsif-clang",
			HintConfidence: config.HintConfidenceProjectStructureSupported,
		},
		{
			Root:           "dir",
			Indexer:        "sourcegraph/lsif-clang",
			HintConfidence: config.HintConfidenceProjectStructureSupported,
		},
		{
			Root:           "other",
			Indexer:        "sourcegraph/lsif-clang",
			HintConfidence: config.HintConfidenceLanguageSupport,
		},
	}

	if diff := cmp.Diff(expectedHints, InferClangIndexJobHints(NewMockGitClient(), paths)); diff != "" {
		t.Errorf("unexpected index job hints (-want +got)\n%s", diff)
	}
}
