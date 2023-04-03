package repo

import (
	"archive/tar"
	"context"
	"io"
	"os"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	repoembeddingsbg "github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/embed"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/split"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
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

const MAX_FILE_SIZE = 1_000_000 // 1MB

// The threshold to embed the entire file is slightly larger than the chunk threshold to
// avoid splitting small files unnecessarily.
const (
	embedEntireFileTokensThreshold          = 384
	embeddingChunkTokensThreshold           = 256
	embeddingChunkEarlySplitTokensThreshold = embeddingChunkTokensThreshold - 32
)

var splitOptions = split.SplitOptions{
	NoSplitTokensThreshold:         embedEntireFileTokensThreshold,
	ChunkTokensThreshold:           embeddingChunkTokensThreshold,
	ChunkEarlySplitTokensThreshold: embeddingChunkEarlySplitTokensThreshold,
}

func (h *handler) Handle(ctx context.Context, logger log.Logger, record *repoembeddingsbg.RepoEmbeddingJob) error {
	if !conf.EmbeddingsEnabled() {
		return errors.New("embeddings are not configured or disabled")
	}

	repo, err := h.db.Repos().Get(ctx, record.RepoID)
	if err != nil {
		return err
	}

	documentRanks, err := getDocumentRanks(ctx, repo.Name)
	if err != nil {
		return errors.Wrap(err, "failed to get document ranks")
	}

	repoTar, err := fetchTarToDisk(ctx, h.gitserverClient, repo.Name, record.Revision)
	if err != nil {
		return errors.Wrap(err, "fetching repo tar to disk")
	}
	defer func() {
		_ = repoTar.Close()
		_ = os.Remove(repoTar.Name())
	}()

	config := conf.Get().Embeddings
	excludedGlobPatterns := embed.GetDefaultExcludedFilePathPatterns()
	excludedGlobPatterns = append(excludedGlobPatterns, embed.CompileGlobPatterns(config.ExcludedFilePathPatterns)...)

	repoEmbeddingIndex, err := embed.EmbedRepo(
		ctx,
		repo.Name,
		record.Revision,
		embed.NewEmbeddingsClient(),
		splitOptions,
		// TODO: Instead of iterating over the tar twice, and resetting the read header
		// of the file, we should probably produce the text and code indexes
		// at the same time, instead of after each other.
		func(cb func(fileName string, fileReader io.Reader) error) error {
			t := tar.NewReader(repoTar)
			for {
				if err := ctx.Err(); err != nil {
					return err
				}

				thr, err := t.Next()
				if err != nil {
					if err == io.EOF {
						return nil
					}
					return err
				}
				switch thr.Typeflag {
				case tar.TypeReg, tar.TypeRegA:
					// Skip files that are too big.
					if thr.Size >= MAX_FILE_SIZE {
						continue
					}
					if thr.Size < embed.MIN_EMBEDDABLE_FILE_SIZE {
						continue
					}
					// Skip files that are excluded.
					if embed.IsExcludedFilePath(thr.Name, excludedGlobPatterns) {
						continue
					}
					// Only emit code files.
					if embed.IsValidTextFile(thr.Name) {
						continue
					}
					if err := cb(thr.Name, t); err != nil {
						return err
					}
				default:
					continue
				}
			}
		},
		func(cb func(fileName string, fileReader io.Reader) error) error {
			if _, err := repoTar.Seek(0, 0); err != nil {
				return err
			}
			t := tar.NewReader(repoTar)
			for {
				if err := ctx.Err(); err != nil {
					return err
				}

				thr, err := t.Next()
				if err != nil {
					if err == io.EOF {
						return nil
					}
					return err
				}
				switch thr.Typeflag {
				case tar.TypeReg, tar.TypeRegA:
					// Skip files that are too big.
					if thr.Size >= MAX_FILE_SIZE {
						continue
					}
					if thr.Size < embed.MIN_EMBEDDABLE_FILE_SIZE {
						continue
					}
					// Skip files that are excluded.
					if embed.IsExcludedFilePath(thr.Name, excludedGlobPatterns) {
						continue
					}
					// Only emit text files.
					if !embed.IsValidTextFile(thr.Name) {
						continue
					}
					if err := cb(thr.Name, t); err != nil {
						return err
					}
				default:
					continue
				}
			}
		},
		documentRanks,
	)
	if err != nil {
		return err
	}

	return embeddings.UploadRepoEmbeddingIndex(ctx, h.uploadStore, string(embeddings.GetRepoEmbeddingIndexName(repo.Name)), repoEmbeddingIndex)
}

// fetchTarToDisk retrieves a tar for the given repo at the given commit and writes
// it to a temporary file.
func fetchTarToDisk(ctx context.Context, gitserverClient gitserver.Client, repoName api.RepoName, commit api.CommitID) (*os.File, error) {
	r, err := gitserverClient.ArchiveReader(ctx, authz.DefaultSubRepoPermsChecker, repoName, gitserver.ArchiveOptions{
		Treeish: string(commit),
		Format:  "tar",
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create archive reader for repo")
	}
	defer r.Close()

	tmpFile, err := os.CreateTemp(os.TempDir(), "embeddings-repo-*")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temp dir to extract repo")
	}

	_, err = io.Copy(tmpFile, r)
	if err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		return nil, errors.Wrap(err, "failed to write tar file to disk")
	}

	// Reset read head to beginning of the file.
	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		return nil, errors.Wrap(err, "failed to reset file head")
	}

	return tmpFile, nil
}
