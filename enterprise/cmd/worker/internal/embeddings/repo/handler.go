package repo

import (
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/grafana/regexp"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	repoembeddingsbg "github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/embed"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/split"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type handler struct {
	db              edb.EnterpriseDB
	uploadStore     uploadstore.Store
	gitserverClient gitserver.Client
	sourcegraphURL  *url.URL
}

var _ workerutil.Handler[*repoembeddingsbg.RepoEmbeddingJob] = &handler{}

var matchEverythingRegexp = regexp.MustCompile(``)

const MAX_FILE_SIZE = 1000000 // 1MB

// The threshold to embed the entire file is slightly larger than the chunk threshold to
// avoid splitting small files unnecessarily.
const EMBED_ENTIRE_FILE_TOKENS_THRESHOLD = 384
const EMBEDDING_CHUNK_TOKENS_THRESHOLD = 256
const EMBEDDING_CHUNK_EARLY_SPLIT_TOKENS_THRESHOLD = EMBEDDING_CHUNK_TOKENS_THRESHOLD - 32

var splitOptions = split.SplitOptions{
	NoSplitTokensThreshold:         EMBED_ENTIRE_FILE_TOKENS_THRESHOLD,
	ChunkTokensThreshold:           EMBEDDING_CHUNK_TOKENS_THRESHOLD,
	ChunkEarlySplitTokensThreshold: EMBEDDING_CHUNK_EARLY_SPLIT_TOKENS_THRESHOLD,
}

func (h *handler) Handle(ctx context.Context, logger log.Logger, record *repoembeddingsbg.RepoEmbeddingJob) error {
	if !conf.EmbeddingsEnabled() {
		return errors.New("embeddings are not configured or disabled")
	}

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

	embeddingsClient := embed.NewEmbeddingsClient()

	repoEmbeddingIndex, err := embed.EmbedRepo(
		ctx,
		repo.Name,
		record.Revision,
		validFiles,
		embeddingsClient,
		splitOptions,
		func(fileName string) ([]byte, error) {
			return h.gitserverClient.ReadFile(ctx, nil, repo.Name, record.Revision, fileName)
		},
		func(ctx context.Context, repoName string) (types.RepoPathRanks, error) {
			return getDocumentRanksWithRetries(ctx, repoName, h.sourcegraphURL, 2)
		})

	if err != nil {
		return err
	}

	return embeddings.UploadIndex(ctx, h.uploadStore, string(embeddings.GetRepoEmbeddingIndexName(repo.Name)), repoEmbeddingIndex)
}
func getDocumentRanksWithRetries(ctx context.Context, repoName string, sourcegraphRoot *url.URL, retries int) (types.RepoPathRanks, error) {

	ranks, err := getDocumentRanks(ctx, repoName, sourcegraphRoot)
	if err == nil {
		return ranks, nil
	}

	for i := 0; i < retries; i++ {
		ranks, err := getDocumentRanks(ctx, repoName, sourcegraphRoot)
		if err == nil {
			return ranks, nil
		}
		delay := time.Duration(int(math.Pow(float64(2), float64(i))))
		time.Sleep(delay * time.Second)
	}

	return types.RepoPathRanks{}, err
}

func getDocumentRanks(ctx context.Context, repoName string, sourcegraphRoot *url.URL) (types.RepoPathRanks, error) {
	u := sourcegraphRoot.ResolveReference(&url.URL{
		Path: "/.internal/ranks/" + strings.Trim(repoName, "/") + "/documents",
	})

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return types.RepoPathRanks{}, err
	}

	resp, err := httpcli.InternalDoer.Do(req)
	if err != nil {
		return types.RepoPathRanks{}, err
	}

	if resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
		_ = resp.Body.Close()
		if err != nil {
			return types.RepoPathRanks{}, err
		}
		return types.RepoPathRanks{}, &url.Error{
			Op:  "Get",
			URL: u.String(),
			Err: errors.Errorf("%s: %s", resp.Status, string(b)),
		}
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.RepoPathRanks{}, err
	}

	ranks := types.RepoPathRanks{}
	err = json.Unmarshal(b, &ranks)
	if err != nil {
		return types.RepoPathRanks{}, err
	}

	return ranks, nil
}
