package inference

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestClangHinter(t *testing.T) {
	testHinters(t,
		hinterTestCase{
			description: "basic hints",
			repositoryContents: map[string]string{
				"CMakeLists.txt":     "",
				"dir/cmakelists.txt": "",
				"other/test.cpp":     "",
				"other/test.h":       "",
			},
			expected: []config.IndexJobHint{
				{
					Root:           "",
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
			},
		},
	)
}
