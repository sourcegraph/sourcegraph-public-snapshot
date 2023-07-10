package embeddings

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

var nonAlphanumericCharsRegexp = lazyregexp.New(`[^0-9a-zA-Z]`)

var CONTEXT_DETECTION_INDEX_NAME = "context_detection.embeddingindex"

type RepoEmbeddingIndexName string

func GetRepoEmbeddingIndexNameDeprecated(repoName api.RepoName) RepoEmbeddingIndexName {
	fsSafeRepoName := nonAlphanumericCharsRegexp.ReplaceAllString(string(repoName), "_")
	// Add a hash as well to avoid name collisions
	hash := md5.Sum([]byte(repoName))
	return RepoEmbeddingIndexName(fmt.Sprintf(`%s_%s.embeddingindex`, fsSafeRepoName, hex.EncodeToString(hash[:])))
}
func GetRepoEmbeddingIndexName(repoID api.RepoID) RepoEmbeddingIndexName {
	return RepoEmbeddingIndexName(fmt.Sprintf(`%d.embeddingindex`, repoID))
}
