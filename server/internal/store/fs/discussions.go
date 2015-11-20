package fs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/tools/godoc/vfs"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/vfsutil"
)

// TODO(keegan) Lots of duplication with changesets with storage layer in git

const (
	discussionsRef = "refs/src/discussion"
	discussionsDir = "discussions"
	discussionFile = "discussion.json"
)

type Discussions struct {
	fsLock sync.RWMutex
}

var _ store.Discussions = (*Discussions)(nil)

func (s *Discussions) Create(ctx context.Context, d *sourcegraph.Discussion) error {
	s.fsLock.Lock()
	defer s.fsLock.Unlock()

	fs, err := s.repoFS(ctx, d.DefKey.Repo)
	if err != nil {
		if os.IsNotExist(err) {
			d.ID = 1
		} else {
			return err
		}
	} else {
		d.ID, err = maxDirID(fs, discussionsDir)
		if err != nil {
			return err
		}
	}

	ts := pbtypes.NewTimestamp(time.Now())
	d.CreatedAt = &ts
	b, err := json.MarshalIndent(d, "", "\t")
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("Created discussion %d", d.ID)
	return s.put(ctx, d.DefKey.Repo, discussionPath(d.ID), msg, b)
}

func (s *Discussions) Get(ctx context.Context, repo sourcegraph.RepoSpec, id int64) (*sourcegraph.Discussion, error) {
	s.fsLock.RLock()
	defer s.fsLock.RUnlock()

	fs, err := s.repoFS(ctx, repo.URI)
	if err != nil {
		return nil, err
	}

	return s.get(ctx, fs, repo, id)
}

// callers must guard
func (s *Discussions) get(ctx context.Context, fs vfs.FileSystem, repo sourcegraph.RepoSpec, id int64) (*sourcegraph.Discussion, error) {
	b, err := vfs.ReadFile(fs, discussionPath(id))
	if err != nil {
		return nil, err
	}

	var d sourcegraph.Discussion
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, err
	} else {
		return &d, nil
	}
}

func (s *Discussions) List(ctx context.Context, op *sourcegraph.DiscussionListOp) (*sourcegraph.DiscussionList, error) {
	s.fsLock.RLock()
	defer s.fsLock.RUnlock()

	var (
		repo   string
		filter func(graph.DefKey) bool

		filterAll     = func(_ graph.DefKey) bool { return true }
		filterSameDef = func(a graph.DefKey) bool { return sameDef(op.DefKey, a) }
	)
	if op.Repo.URI != "" {
		repo = op.Repo.URI
		filter = filterAll
	} else {
		repo = op.DefKey.Repo
		filter = filterSameDef
	}

	fs, err := s.repoFS(ctx, repo)
	if err != nil {
		if os.IsNotExist(err) {
			return &sourcegraph.DiscussionList{
				Discussions: []*sourcegraph.Discussion{},
			}, nil
		}
		return nil, err
	}

	return s.list(ctx, fs, filter)
}

// callers must guard
func (s *Discussions) list(ctx context.Context, fs vfs.FileSystem, filter func(graph.DefKey) bool) (*sourcegraph.DiscussionList, error) {
	var ds []*sourcegraph.Discussion
	fis, err := fs.ReadDir(discussionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			log15.Debug("No discussions", "fs", fs)
			list := sourcegraph.DiscussionList{Discussions: ds}
			return &list, nil
		}
		return nil, err
	}
	var paths []string
	for _, fi := range fis {
		if !fi.IsDir() {
			continue
		}
		id, err := strconv.Atoi(fi.Name())
		if err != nil {
			continue
		}
		paths = append(paths, discussionPath(int64(id)))
	}
	readCh, done := vfsutil.ConcurrentRead(fs, paths)
	defer done.Done()
	for readRet := range readCh {
		if readRet.Error != nil {
			log15.Warn("Could not read discussion", "path", readRet.Path, "error", readRet.Error)
			continue
		}
		var d sourcegraph.Discussion
		if err := json.Unmarshal(readRet.Bytes, &d); err != nil {
			log15.Warn("Could not unmarshal discussion", "path", readRet.Path, "error", readRet.Error)
		} else if filter(d.DefKey) {
			ds = append(ds, &d)
		}
	}

	list := sourcegraph.DiscussionList{Discussions: ds}
	return &list, nil
}

func (s *Discussions) CreateComment(ctx context.Context, discussionID int64, comment *sourcegraph.DiscussionComment) error {
	s.fsLock.Lock()
	defer s.fsLock.Unlock()

	fs, err := s.repoFS(ctx, comment.DefKey.Repo)
	if err != nil {
		return err
	}

	repo := sourcegraph.RepoSpec{URI: comment.DefKey.Repo}
	d, err := s.get(ctx, fs, repo, discussionID)
	if err != nil {
		return err
	}

	now := pbtypes.NewTimestamp(time.Now())
	comment.ID = int64(len(d.Comments) + 1)
	comment.CreatedAt = &now
	d.Comments = append(d.Comments, comment)

	b, err := json.MarshalIndent(d, "", "\t")
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("Discussion comment %d %d", d.ID, comment.ID)
	return s.put(ctx, d.DefKey.Repo, discussionPath(discussionID), msg, b)
}

// callers must guard
func (s *Discussions) repoFS(ctx context.Context, repoPath string) (vfs.FileSystem, error) {
	dir := absolutePathForRepo(ctx, repoPath)
	repo, err := vcs.Open("git", dir)
	if err != nil {
		return nil, err
	}
	refResolver, ok := repo.(refResolver)
	if !ok {
		return nil, &os.PathError{Op: "getDiscussionsRefTip", Path: discussionsRef, Err: os.ErrNotExist}
	}
	cid, err := refResolver.ResolveRef(discussionsRef)
	if err == vcs.ErrRefNotFound {
		return nil, &os.PathError{Op: "getDiscussionsRefTip", Path: discussionsRef, Err: os.ErrNotExist}
	}
	return repo.FileSystem(cid)
}

// callers must guard
func (s *Discussions) put(ctx context.Context, repo, path, message string, contents []byte) error {
	dir := absolutePathForRepo(ctx, repo)
	rs, err := NewRepoStage(dir, discussionsRef)
	if err != nil {
		return err
	}
	defer rs.Free()
	if err := rs.Add(path, contents); err != nil {
		return err
	}
	RefAuthor.Date, RefCommitter.Date = pbtypes.NewTimestamp(time.Now()), pbtypes.NewTimestamp(time.Now())
	return rs.Commit(RefAuthor, RefCommitter, message)
}

// callers must guard
func maxDirID(fs vfs.FileSystem, dir string) (int64, error) {
	fis, err := fs.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}
	max := int64(0)
	for _, fi := range fis {
		if !fi.IsDir() {
			continue
		}
		n, err := strconv.Atoi(fi.Name())
		if err != nil {
			continue
		}
		if n := int64(n); n > max {
			max = n
		}
	}
	return max + 1, nil
}

func discussionPath(id int64) string {
	return filepath.Join(discussionsDir, strconv.FormatInt(id, 10), discussionFile)
}

// sameDef checks if we are talking about the same Def, ignoring CommitID
func sameDef(a, b graph.DefKey) bool {
	a.CommitID = b.CommitID
	return a == b
}

func getFilter(query graph.DefKey) func(graph.DefKey) bool {
	isListAll := sameDef(query, graph.DefKey{Repo: query.Repo})
	if isListAll {
		return func(_ graph.DefKey) bool {
			return true
		}
	}
	return func(a graph.DefKey) bool {
		return sameDef(query, a)
	}
}
