package fs

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/tools/godoc/vfs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/vfsutil"
)

const (
	reviewRef             = "refs/src/review"
	changesetMetadataFile = "changeset.json"
	changesetReviewsFile  = "reviews.json"
	changesetEventsFile   = "events.json"
)

type Changesets struct {
	fsLock sync.RWMutex // guards FS
}

func (s *Changesets) Create(ctx context.Context, repoPath string, cs *sourcegraph.Changeset) error {
	s.fsLock.Lock()
	defer s.fsLock.Unlock()

	// We only need a commit ID on base after the changeset is closed, or
	// head is deleted/merged.
	cs.DeltaSpec.Base.CommitID = ""
	dir := absolutePathForRepo(ctx, repoPath)
	repo, err := vcs.Open("git", dir)
	if err != nil {
		return err
	}
	cid, err := s.getReviewRefTip(repo)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		cs.ID = 1
	} else {
		cs.ID, err = resolveNextChangesetID(repo, cid)
		if err != nil {
			return err
		}
	}
	ts := pbtypes.NewTimestamp(time.Now())
	cs.CreatedAt = &ts
	id := strconv.FormatInt(cs.ID, 10)
	head, err := repo.ResolveBranch(cs.DeltaSpec.Head.Rev)
	if err != nil {
		return err
	}
	cs.DeltaSpec.Head.CommitID = string(head)
	if err := updateRef(dir, "refs/changesets/"+id+"/head", string(head)); err != nil {
		return err
	}
	return writeChangeset(ctx, repoPath, cs, "Created changeset")
}

// writeChangeset writes the given changeset into the repository specified by
// repoPath. When committing the change, msg is used as the commit message to
// which the changeset ID is appended.
//
// Callers must guard by holding the s.fsLock lock.
func writeChangeset(ctx context.Context, repoPath string, cs *sourcegraph.Changeset, msg string) error {
	dir := absolutePathForRepo(ctx, repoPath)
	rs, err := NewRepoStage(dir, reviewRef)
	if err != nil {
		return err
	}
	defer rs.Free()
	b, err := json.MarshalIndent(cs, "", "\t")
	if err != nil {
		return err
	}
	sid := strconv.FormatInt(cs.ID, 10)
	if err := rs.Add(filepath.Join(sid, changesetMetadataFile), b); err != nil {
		return err
	}
	RefAuthor.Date, RefCommitter.Date = pbtypes.NewTimestamp(time.Now()), pbtypes.NewTimestamp(time.Now())
	return rs.Commit(RefAuthor, RefCommitter, msg+" (#"+sid+")")
}

// resolveNextChangesetID tries to resolve the next available changeset ID by recursing
// through the files in the given repository at the given commit ID. The filesystem
// is expected to be that of a correct changeset storage (meaning a set of folders having
// a numeric name that corresponds to the ID of the changeset data contained within them)
func resolveNextChangesetID(repo vcs.Repository, commitID vcs.CommitID) (int64, error) {
	repoFS, err := repo.FileSystem(commitID)
	if err != nil {
		return 0, err
	}
	fis, err := repoFS.ReadDir(".")
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

func (s *Changesets) Get(ctx context.Context, repoPath string, ID int64) (*sourcegraph.Changeset, error) {
	s.fsLock.RLock()
	defer s.fsLock.RUnlock()

	return s.get(ctx, repoPath, ID)
}

// callers must guard
func (s *Changesets) get(ctx context.Context, repoPath string, ID int64) (*sourcegraph.Changeset, error) {
	dir := absolutePathForRepo(ctx, repoPath)
	repo, err := vcs.Open("git", dir)
	if err != nil {
		return nil, err
	}
	cid, err := s.getReviewRefTip(repo)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, grpc.Errorf(codes.NotFound, "changeset metadata not present")
		}
		return nil, err
	}
	fs, err := repo.FileSystem(cid)
	if err != nil {
		return nil, err
	}
	b, err := vfs.ReadFile(fs, filepath.Join(strconv.FormatInt(ID, 10), changesetMetadataFile))
	if err != nil {
		return nil, err
	}
	cs := &sourcegraph.Changeset{}
	if err := json.Unmarshal(b, &cs); err != nil {
		return nil, err
	}
	return cs, nil
}

// getReviewRefTip returns the commit ID that is at the tip of the reference where
// the changeset data is stored.
//
// Callers must guard by holding the s.fsLock lock.
func (s *Changesets) getReviewRefTip(repo vcs.Repository) (vcs.CommitID, error) {
	refResolver, ok := repo.(refResolver)
	if !ok {
		return "", &os.PathError{Op: "getReviewRefTip", Path: reviewRef, Err: os.ErrNotExist}
	}
	id, err := refResolver.ResolveRef(reviewRef)
	if err == vcs.ErrRefNotFound {
		return "", &os.PathError{Op: "getReviewRefTip", Path: reviewRef, Err: os.ErrNotExist}
	}
	return id, err
}

func (s *Changesets) CreateReview(ctx context.Context, repoPath string, changesetID int64, newReview *sourcegraph.ChangesetReview) (*sourcegraph.ChangesetReview, error) {
	s.fsLock.Lock()
	defer s.fsLock.Unlock()

	// Read current reviews into structure
	all := sourcegraph.ChangesetReviewList{Reviews: []*sourcegraph.ChangesetReview{}}
	err := s.readFile(ctx, repoPath, changesetID, changesetReviewsFile, &all.Reviews)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// Append new review to list
	newReview.ID = int64(len(all.Reviews) + 1)
	all.Reviews = append(all.Reviews, newReview)

	// Marshal to JSON and write new structure to repo
	dir := absolutePathForRepo(ctx, repoPath)
	rs, err := NewRepoStage(dir, reviewRef)
	if err != nil {
		return nil, err
	}
	defer rs.Free()
	b, err := json.MarshalIndent(all.Reviews, "", "\t")
	if err != nil {
		return nil, err
	}
	sid := strconv.FormatInt(changesetID, 10)
	if err := rs.Add(filepath.Join(sid, changesetReviewsFile), b); err != nil {
		return nil, err
	}
	RefAuthor.Date, RefCommitter.Date = pbtypes.NewTimestamp(time.Now()), pbtypes.NewTimestamp(time.Now())
	if err := rs.Commit(RefAuthor, RefCommitter, "Submitted a new review"); err != nil {
		return nil, err
	}
	return newReview, err
}

func (s *Changesets) ListReviews(ctx context.Context, repo string, changesetID int64) (*sourcegraph.ChangesetReviewList, error) {
	s.fsLock.RLock()
	defer s.fsLock.RUnlock()

	list := &sourcegraph.ChangesetReviewList{Reviews: []*sourcegraph.ChangesetReview{}}
	err := s.readFile(ctx, repo, changesetID, changesetReviewsFile, &list.Reviews)
	if os.IsNotExist(err) {
		err = nil
	}
	return list, err
}

// readFile JSON decodes the contents of the file named 'filename' from
// the folder of the given changeset into v.
//
// Callers must guard by holding the s.fsLock lock.
func (s *Changesets) readFile(ctx context.Context, repoPath string, changesetID int64, filename string, v interface{}) error {
	dir := absolutePathForRepo(ctx, repoPath)
	repo, err := vcs.Open("git", dir)
	if err != nil {
		return err
	}
	cid, err := s.getReviewRefTip(repo)
	if err != nil {
		return err
	}
	fs, err := repo.FileSystem(cid)
	if err != nil {
		return err
	}
	b, err := vfs.ReadFile(fs, filepath.Join(strconv.FormatInt(changesetID, 10), filename))
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

var errInvalidUpdateOp = errors.New("invalid update operation")

func (s *Changesets) Update(ctx context.Context, opt *store.ChangesetUpdateOp) (*sourcegraph.ChangesetEvent, error) {
	s.fsLock.Lock()
	defer s.fsLock.Unlock()

	op := opt.Op

	dir := absolutePathForRepo(ctx, op.Repo.URI)
	if (op.Close && op.Open) || (op.Open && op.Merged) {
		return nil, errInvalidUpdateOp
	}
	current, err := s.get(ctx, op.Repo.URI, op.ID)
	if err != nil {
		return nil, err
	}
	after := *current
	ts := pbtypes.NewTimestamp(time.Now())

	if op.Title != "" {
		after.Title = op.Title
	}
	if op.Description != "" {
		after.Description = op.Description
	}
	if op.Open {
		after.ClosedAt = nil
	}
	if op.Close && (current.ClosedAt == nil || (op.Merged && !current.Merged)) {
		after.ClosedAt = &ts
	}
	if op.Merged {
		after.Merged = true
	}

	if opt.Head != "" {
		// We need to track the tip of this branch so that we can access its
		// data even after a potential deletion or merge.
		id := strconv.FormatInt(current.ID, 10)
		if err := updateRef(dir, "refs/changesets/"+id+"/head", opt.Head); err != nil {
			return nil, err
		}
		after.DeltaSpec.Head.CommitID = opt.Head
	}
	if opt.Base != "" {
		id := strconv.FormatInt(current.ID, 10)
		if err := updateRef(dir, "refs/changesets/"+id+"/base", opt.Base); err != nil {
			return nil, err
		}
		after.DeltaSpec.Base.CommitID = opt.Base
	}

	// If the resulting changeset is the same as the inital one, no event
	// has occurred.
	if reflect.DeepEqual(after, current) {
		return &sourcegraph.ChangesetEvent{}, nil
	}

	rs, err := NewRepoStage(dir, reviewRef)
	if err != nil {
		return nil, err
	}
	defer rs.Free()
	sid := strconv.FormatInt(op.ID, 10)
	b, err := json.MarshalIndent(after, "", "\t")
	if err != nil {
		return nil, err
	}
	if err := rs.Add(filepath.Join(sid, changesetMetadataFile), b); err != nil {
		return nil, err
	}

	var evt *sourcegraph.ChangesetEvent
	if shouldRegisterEvent(op) {
		evt = &sourcegraph.ChangesetEvent{
			Before:    current,
			After:     &after,
			Op:        op,
			CreatedAt: &ts,
		}
		evts := []*sourcegraph.ChangesetEvent{}
		err = s.readFile(ctx, op.Repo.URI, op.ID, changesetEventsFile, &evts)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		evts = append(evts, evt)
		b, err = json.MarshalIndent(evts, "", "\t")
		if err != nil {
			return nil, err
		}
		if err := rs.Add(filepath.Join(sid, changesetEventsFile), b); err != nil {
			return nil, err
		}
	}

	RefAuthor.Date, RefCommitter.Date = pbtypes.NewTimestamp(time.Now()), pbtypes.NewTimestamp(time.Now())
	if err := rs.Commit(RefAuthor, RefCommitter, "Updated changeset (#"+sid+")"); err != nil {
		return nil, err
	}
	return evt, nil
}

// shouldRegisterEvent returns true if the passed update operation will cause an event
// that needs to be registered. Registered events are be visible in the changeset's
// timeline.
func shouldRegisterEvent(op *sourcegraph.ChangesetUpdateOp) bool {
	return op.Merged || op.Close || op.Open || op.Title != "" || op.Description != ""
}

func (s *Changesets) List(ctx context.Context, op *sourcegraph.ChangesetListOp) (*sourcegraph.ChangesetList, error) {
	s.fsLock.RLock()
	defer s.fsLock.RUnlock()

	// by default, retrieve all changesets
	if !op.Open && !op.Closed {
		op.Open = true
		op.Closed = true
	}
	dir := absolutePathForRepo(ctx, op.Repo)
	repo, err := vcs.Open("git", dir)
	if err != nil {
		return nil, err
	}
	list := sourcegraph.ChangesetList{Changesets: []*sourcegraph.Changeset{}}
	cid, err := s.getReviewRefTip(repo)
	if err != nil {
		if os.IsNotExist(err) {
			return &list, nil
		}
		return nil, err
	}
	fs, err := repo.FileSystem(cid)
	if err != nil {
		return nil, err
	}
	fis, err := fs.ReadDir(".")
	if err != nil {
		if os.IsNotExist(err) {
			return &list, nil
		}
		return nil, err
	}

	skip := op.Offset()
	limit := op.Limit()
	var paths []string
	for _, fi := range fis {
		// skip files
		if !fi.IsDir() {
			continue
		}
		// skip non-numeric folders
		if _, err := strconv.Atoi(fi.Name()); err != nil {
			continue
		}
		paths = append(paths, filepath.Join(fi.Name(), changesetMetadataFile))
	}
	readCh, done := vfsutil.ConcurrentRead(fs, paths)
	defer done.Done()
	for readRet := range readCh {
		if readRet.Error != nil {
			return nil, readRet.Error
		}
		var cs sourcegraph.Changeset
		if err := json.Unmarshal(readRet.Bytes, &cs); err != nil {
			return nil, err
		}
		if op.Open && cs.ClosedAt == nil || op.Closed && cs.ClosedAt != nil {
			// check if the request was only for changesets with a specific
			// branch for head or base.
			if (op.Head == "" || op.Head == cs.DeltaSpec.Head.Rev) &&
				(op.Base == "" || op.Base == cs.DeltaSpec.Base.Rev) {
				if skip > 0 {
					skip--
					continue
				}
				if len(list.Changesets) == limit {
					break
				}
				list.Changesets = append(list.Changesets, &cs)
			}
		}
	}
	return &list, nil
}

func (s *Changesets) ListEvents(ctx context.Context, spec *sourcegraph.ChangesetSpec) (*sourcegraph.ChangesetEventList, error) {
	list := sourcegraph.ChangesetEventList{Events: []*sourcegraph.ChangesetEvent{}}
	err := s.readFile(ctx, spec.Repo.URI, spec.ID, changesetEventsFile, &list.Events)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return &list, nil
}
