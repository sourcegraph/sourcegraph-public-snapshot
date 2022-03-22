package inference

import (
	"path/filepath"
	"strings"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func InferClangIndexJobHints(gitserver GitClient, paths []string) (hints []config.IndexJobHint) {
	for _, path := range paths {
		if strings.ToLower(filepath.Base(path)) == "cmakelists.txt" {
			hints = append(hints, config.IndexJobHint{
				Root:           filepath.Dir(path),
				Indexer:        "sourcegraph/lsif-clang",
				HintConfidence: config.ProjectStructureSupported,
			})
			continue
		}

		ext := filepath.Ext(path)
		if ext == ".cpp" || ext == ".c" || ext == ".h" || ext == ".hpp" || ext == ".cxx" || ext == ".cc" {
			hints = append(hints, config.IndexJobHint{
				Root:           filepath.Dir(path),
				Indexer:        "sourcegraph/lsif-clang",
				HintConfidence: config.LanguageSupport,
			})
		}
	}

	return
}

func ClangPatterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		suffixPattern(pathPattern(rawPattern("[cC][mM]ake[lL]ists.txt"))),
		extensionPattern(rawPattern("cpp")),
		extensionPattern(rawPattern("c")),
		extensionPattern(rawPattern("h")),
		extensionPattern(rawPattern("hpp")),
		extensionPattern(rawPattern("cxx")),
		extensionPattern(rawPattern("cc")),
	}
}
