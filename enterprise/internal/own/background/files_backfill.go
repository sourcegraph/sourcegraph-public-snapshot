package background

import (
	"context"
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func handleFilesBackfill(ctx context.Context, lgr logger.Logger, repoId api.RepoID, db database.DB) error {
	// 🚨 SECURITY: we use the internal actor because the background indexer is not associated with any user, and needs
	// to see all repos and files
	internalCtx := actor.WithInternalActor(ctx)
	lgr.Info("backfilling files for repository")
	indexer := newFilesBackfillIndexer(gitserver.NewClient(), db, lgr)
	return indexer.indexRepo(internalCtx, repoId)
}

type filesBackfillIndexer struct {
	client gitserver.Client
	db     database.DB
	logger logger.Logger
}

func newFilesBackfillIndexer(client gitserver.Client, db database.DB, lgr logger.Logger) *filesBackfillIndexer {
	return &filesBackfillIndexer{client: client, db: db, logger: lgr}
}

var filesCounter = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: "src",
	Name:      "own_files_backfill_files_indexed_total",
})

func (r *filesBackfillIndexer) indexRepo(ctx context.Context, repoId api.RepoID) error {
	repoStore := r.db.Repos()
	repo, err := repoStore.Get(ctx, repoId)
	if err != nil {
		return errors.Wrap(err, "repoStore.Get")
	}
	r.logger.Info("LsFines", logger.String("repo_name", string(repo.Name)))
	// TODO: can shard by pathspecs here:
	files, err := r.client.LsFiles(ctx, nil, repo.Name, "HEAD")
	if err != nil {
		r.logger.Error("ls-files failed", logger.String("msg", err.Error()))
		return errors.Wrap(err, "LsFiles")
	}
	ids, err := r.db.RepoPaths().EnsureExist(ctx, repo.ID, files)
	if err != nil {
		r.logger.Error("inserting backfill files failed", logger.String("msg", err.Error()))
		return errors.Wrap(err, "EnsureExist")
	}
	root := &fileNode{}
	for f, id := range ids {
		root.add(f, id)
	}
	inserter := batch.NewInserterWithConflict(
		ctx,
		r.db.Handle(),
		"repo_paths",
		batch.MaxNumPostgresParameters,
		"ON CONFLICT SET deep_file_count = EXCLUDED.deep_file_count",
		"id",
		"deep_file_count",
	)
	type file struct {
		name string
		node *fileNode
	}
	stack := []file{{name: "", node: root}}
	pop := func() file {
		lastIndex := len(stack) - 1
		f := stack[lastIndex]
		stack = stack[:lastIndex]
		return f
	}
	push := func(f file) {
		stack = append(stack, f)
	}
	for len(stack) > 0 {
		f := pop()
		if err := inserter.Insert(ctx, f.node.id, f.node.deepCount); err != nil {
			return err
		}
		for _, n := range f.node.nodes {
			push(file{name: fmt.Sprintf("%s/%s", f.name, n.baseName), node: n})
		}
	}
	if err := inserter.Flush(ctx); err != nil {
		return err
	}
	r.logger.Info("files", logger.Int("total", len(files)), logger.String("repo_name", string(repo.Name)))
	filesCounter.Add(float64(len(files)))
	return nil
}

type fileNode struct {
	baseName  string
	id        int
	deepCount int
	nodes     []*fileNode
}

func (t *fileNode) add(path string, id int) {
	for path != "" {
		sep := strings.Index(path, "/")
		var base, rest string
		if sep == -1 {
			base = path
		} else {
			base = path[:sep]
			rest = path[sep+1:]
		}
		var node *fileNode
		for _, n := range t.nodes {
			if node.baseName == base {
				node = n
				break
			}
		}
		if node == nil {
			node = &fileNode{baseName: base}
			t.nodes = append(t.nodes, node)
		}
		t.deepCount++
		// tail-recurse
		t = node
		path = rest
	}
	t.id = id
}
