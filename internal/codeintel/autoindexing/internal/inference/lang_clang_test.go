package inference

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference/libs"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestClangHinter(t *testing.T) {
	expectedIndexerImage, _ := libs.DefaultIndexerForLang("clang")

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
