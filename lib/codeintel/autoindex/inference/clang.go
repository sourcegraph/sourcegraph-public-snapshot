package inference

import (
	"path/filepath"
	"strings"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func InferClangIndexJobHints(gitserver GitClient, paths []string) (hints []config.IndexJobHint) {
	inferredDir := make(map[string]bool)

	for _, path := range paths {
		dir := filepath.Dir(path)
		if inferredDir[dir] {
			continue
		}
		if strings.ToLower(filepath.Base(path)) == "cmakelists.txt" {
			hints = append(hints, config.IndexJobHint{
				Root:           dir,
				Indexer:        "sourcegraph/lsif-clang",
				HintConfidence: config.HintConfidenceProjectStructureSupported,
			})
			inferredDir[dir] = true
			continue
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".cpp" || ext == ".c" || ext == ".h" || ext == ".hpp" || ext == ".cxx" || ext == ".cc" {
			hints = append(hints, config.IndexJobHint{
				Root:           dir,
				Indexer:        "sourcegraph/lsif-clang",
				HintConfidence: config.HintConfidenceLanguageSupport,
			})
			inferredDir[dir] = true
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
