package embeddings

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/sourcegraph/log"

	"github.com/grafana/regexp"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	embeddingsbg "github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/embed"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type handler struct {
	db              edb.EnterpriseDB
	uploadStore     uploadstore.Store
	gitserverClient gitserver.Client
}

var _ workerutil.Handler[*embeddingsbg.RepoEmbeddingJob] = &handler{}

var matchEverythingRegexp = regexp.MustCompile(``)

const MAX_FILE_SIZE = 1000000 // 1MB

func (h *handler) Handle(ctx context.Context, logger log.Logger, record *embeddingsbg.RepoEmbeddingJob) error {
	repo, err := h.db.Repos().Get(ctx, record.RepoID)
	if err != nil {
		return err
	}

	files, err := h.gitserverClient.ListFiles(ctx, nil, repo.Name, record.Revision, matchEverythingRegexp)
	if err != nil {
		return err
	}

	validFiles := []string{}
	for _, file := range files {
		stat, err := h.gitserverClient.Stat(ctx, nil, repo.Name, record.Revision, file)
		if err != nil {
			return err
		}

		if !stat.IsDir() && stat.Size() <= MAX_FILE_SIZE {
			validFiles = append(validFiles, file)
		}
	}

	repoEmbeddingIndex, err := embed.EmbedRepo(ctx, repo.Name, record.Revision, validFiles, func(fileName string) ([]byte, error) {
		return h.gitserverClient.ReadFile(ctx, nil, repo.Name, record.Revision, fileName)
	})
	if err != nil {
		return err
	}

	indexJsonBytes, err := json.Marshal(repoEmbeddingIndex)
	if err != nil {
		return err
	}

	bytesReader := bytes.NewReader(indexJsonBytes)
	_, err = h.uploadStore.Upload(ctx, "index", bytesReader)
	return err
}
