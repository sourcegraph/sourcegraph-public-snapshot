package indexer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/codeintelutils"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

type Processor interface {
	Process(ctx context.Context, index store.Index) error
}

type processor struct {
	store           store.Store
	gitserverClient gitserver.Client
	frontendURL     string
}

func (p *processor) Process(ctx context.Context, index store.Index) error {
	repoDir, err := fetchRepository(ctx, p.store, p.gitserverClient, index.RepositoryID, index.Commit)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(repoDir)
	}()

	if err := p.index(ctx, repoDir, index); err != nil {
		return errors.Wrap(err, "failed to index repository")
	}

	if err := p.upload(ctx, repoDir, index); err != nil {
		return errors.Wrap(err, "failed to upload index")
	}

	return nil
}

func (p *processor) index(ctx context.Context, repoDir string, index store.Index) error {
	tag, exact, err := p.gitserverClient.Tags(ctx, p.store, index.RepositoryID, index.Commit)
	if err != nil {
		return err
	}
	if !exact {
		tag = fmt.Sprintf("%s-%s", tag, index.Commit[:12])
	}

	args := []string{
		"--repositoryRoot=.",
		fmt.Sprintf("--moduleVersion=%s", tag),
	}

	return command(repoDir, "lsif-go", args...)
}

func (p *processor) upload(ctx context.Context, repoDir string, index store.Index) error {
	repoName, err := p.store.RepoName(ctx, index.RepositoryID)
	if err != nil {
		return errors.Wrap(err, "store.RepoName")
	}

	opts := codeintelutils.UploadIndexOpts{
		Endpoint:            fmt.Sprintf("http://%s", p.frontendURL),
		Path:                "/.internal/lsif/upload",
		Repo:                repoName,
		Commit:              index.Commit,
		Root:                "",
		Indexer:             "lsif-go",
		File:                filepath.Join(repoDir, "dump.lsif"),
		MaxPayloadSizeBytes: 100 * 1000 * 1000, // 100Mb
		MaxRetries:          10,
		RetryInterval:       time.Second * 250,
	}

	if _, err := codeintelutils.UploadIndex(opts); err != nil {
		return errors.Wrap(err, "codeintelutils.UploadIndex")
	}

	return nil
}
