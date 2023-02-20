package repo

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/sourcegraph/log"

	"github.com/grafana/regexp"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	repoembeddingsbg "github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/embed"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type handler struct {
	db              edb.EnterpriseDB
	uploadStore     uploadstore.Store
	gitserverClient gitserver.Client
}

var _ workerutil.Handler[*repoembeddingsbg.RepoEmbeddingJob] = &handler{}

var matchEverythingRegexp = regexp.MustCompile(``)

const MAX_FILE_SIZE = 1000000 // 1MB

func (h *handler) Handle(ctx context.Context, logger log.Logger, record *repoembeddingsbg.RepoEmbeddingJob) error {
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

	config := conf.Get()

	repoEmbeddingIndex, err := embed.EmbedRepo(ctx, repo.Name, record.Revision, validFiles, config.Embeddings, func(fileName string) ([]byte, error) {
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
	_, err = h.uploadStore.Upload(ctx, string(embeddings.GetRepoEmbeddingIndexName(repo.Name)), bytesReader)
	return err
}
