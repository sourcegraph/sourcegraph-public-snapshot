package background

import (
	"context"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/own"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
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
	root := &RepoFileNode{Owners: map[string]int{}}
	for f, id := range ids {
		root.Add(f, id)
	}
	recounted, err := r.db.RepoPaths().UpdateCounts(ctx, repo.ID, root)
	if err != nil {
		return errors.Wrap(err, "UpdateCounts")
	}
	r.logger.Info("files", logger.Int("total", len(files)), logger.Int("recounted", recounted), logger.String("repo_name", string(repo.Name)))
	filesCounter.Add(float64(len(files)))
	// Try to compute ownership stats
	ownService := own.NewService(r.client, r.db)
	commitID, err := r.client.ResolveRevision(ctx, repo.Name, "HEAD", gitserver.ResolveRevisionOptions{NoEnsureRevision: true})
	if err != nil {
		return errors.Wrap(err, "ReoslveRevision")
	}
	ruleset, err := ownService.RulesetForRepo(ctx, repo.Name, repo.ID, commitID)
	if ruleset == nil || err != nil {
		// TODO: Handle errors
		return nil
	}
	r.logger.Info("CODEOWNERS", logger.String("repo", string(repo.Name)), logger.Int("rules", len(ruleset.GetFile().GetRule())))
	// Try compute deep owner counts
	if len(ruleset.GetFile().GetRule()) > 0 {
		for f := range ids {
			var owners []string
			for _, o := range ruleset.Match(f).GetOwner() {
				if o.GetEmail() != "" {
					owners = append(owners, o.GetEmail())
				} else if o.GetHandle() != "" {
					owners = append(owners, "@"+o.GetHandle())
				}
			}
			root.AssignOwner(f, owners)
		}
	}
	_, err = r.db.OwnershipStats().UpdateCodeownersCounts(ctx, repo.ID, walkOwnership{root})
	if err != nil {
		return err
	}
	return nil
}

// TODO can make private
type RepoFileNode struct {
	BaseName  string
	ID        int
	DeepCount int
	Owners    map[string]int
	Children  []*RepoFileNode
}

func (t *RepoFileNode) Add(path string, id int) {
	for path != "" {
		sep := strings.Index(path, "/")
		var base, rest string
		if sep == -1 {
			base = path
		} else {
			base = path[:sep]
			rest = path[sep+1:]
		}
		var node *RepoFileNode
		for _, n := range t.Children {
			if n.BaseName == base {
				node = n
				break
			}
		}
		if node == nil {
			node = &RepoFileNode{BaseName: base, Owners: map[string]int{}}
			t.Children = append(t.Children, node)
		}
		t.DeepCount++
		// tail-recurse
		t = node
		path = rest
	}
	t.ID = id
}

func (t *RepoFileNode) AssignOwner(filePath string, owners []string) {
	if len(owners) == 0 {
		return
	}
	for _, o := range owners {
		t.Owners[o]++
	}
	if filePath != "" {
		idx := strings.Index(filePath, "/")
		var name, sub string
		if idx == -1 {
			name = filePath
		} else {
			name = filePath[:idx]
			sub = filePath[idx+1:]
		}

		var child *RepoFileNode
		for _, c := range t.Children {
			if c.BaseName == name {
				child = c
				break
			}
		}
		// Recurse
		if child != nil {
			child.AssignOwner(sub, owners)
		}
	}
}

func (t *RepoFileNode) Walk(f func(path string, count int) error) error {
	s := &stack{}
	s.push("", t)
	for !s.empty() {
		path, n := s.pop()
		if err := f(path, n.DeepCount); err != nil {
			return err
		}
		if path != "" {
			path = path + "/"
		}
		for _, c := range n.Children {
			s.push(path+c.BaseName, c)
		}
	}
	return nil
}

type walkOwnership struct{ *RepoFileNode }

func (w walkOwnership) Walk(f func(path string, ownerCounts map[string]int) error) error {
	s := &stack{}
	s.push("", w.RepoFileNode)
	for !s.empty() {
		path, n := s.pop()
		if err := f(path, n.Owners); err != nil {
			return err
		}
		if path != "" {
			path = path + "/"
		}
		for _, c := range n.Children {
			s.push(path+c.BaseName, c)
		}
	}
	return nil
}

type stackItem struct {
	fullPath string
	node     *RepoFileNode
}

type stack []stackItem

func (s *stack) push(path string, node *RepoFileNode) {
	*s = append(*s, stackItem{path, node})
}

func (s *stack) empty() bool {
	return len(*s) == 0
}

func (s *stack) pop() (string, *RepoFileNode) {
	i := len(*s) - 1
	var item stackItem
	*s, item = (*s)[:i], (*s)[i]
	return item.fullPath, item.node
}
