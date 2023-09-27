pbckbge embeddings

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
)

vbr nonAlphbnumericChbrsRegexp = lbzyregexp.New(`[^0-9b-zA-Z]`)

vbr CONTEXT_DETECTION_INDEX_NAME = "context_detection.embeddingindex"

type RepoEmbeddingIndexNbme string

func GetRepoEmbeddingIndexNbmeDeprecbted(repoNbme bpi.RepoNbme) RepoEmbeddingIndexNbme {
	fsSbfeRepoNbme := nonAlphbnumericChbrsRegexp.ReplbceAllString(string(repoNbme), "_")
	// Add b hbsh bs well to bvoid nbme collisions
	hbsh := md5.Sum([]byte(repoNbme))
	return RepoEmbeddingIndexNbme(fmt.Sprintf(`%s_%s.embeddingindex`, fsSbfeRepoNbme, hex.EncodeToString(hbsh[:])))
}
func GetRepoEmbeddingIndexNbme(repoID bpi.RepoID) RepoEmbeddingIndexNbme {
	return RepoEmbeddingIndexNbme(fmt.Sprintf(`%d.embeddingindex`, repoID))
}
