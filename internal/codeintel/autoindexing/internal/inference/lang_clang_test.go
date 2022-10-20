package inference

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestClangHinter(t *testing.T) {
	expectedIndexerImage := "sourcegraph/lsif-clang@sha256:5ef2334ac9d58f1f947651812aa8d8ba0ed584913f2429cc9952cb25f94976d8"

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
					Indexer:        expectedIndexerImage,
					HintConfidence: config.HintConfidenceProjectStructureSupported,
				},
				{
					Root:           "dir",
					Indexer:        expectedIndexerImage,
					HintConfidence: config.HintConfidenceProjectStructureSupported,
				},
				{
					Root:           "other",
					Indexer:        expectedIndexerImage,
					HintConfidence: config.HintConfidenceLanguageSupport,
				},
			},
		},
	)
}
