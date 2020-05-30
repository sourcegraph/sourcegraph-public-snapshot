package indexer

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/gitserver"
)

func fetchRepository(ctx context.Context, db db.DB, gitserverClient gitserver.Client, repositoryID int, commit string) (string, error) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(tempDir)
		}
	}()

	archive, err := gitserverClient.Archive(ctx, db, repositoryID, commit)
	if err != nil {
		return "", errors.Wrap(err, "gitserver.Archive")
	}

	if err := extractTarfile(tempDir, archive); err != nil {
		return "", errors.Wrap(err, "extractTarfile")
	}

	return tempDir, nil
}
