package fs

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"
	"golang.org/x/tools/godoc/vfs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/ext"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/notif"
	"src.sourcegraph.com/sourcegraph/platform/storage"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/svc"
	"src.sourcegraph.com/sourcegraph/util"
)

const (
	changesetMetadataFile = "changeset.json"
	changesetReviewsFile  = "reviews.json"
	changesetEventsFile   = "events.json"

	changesetIndexAllDir    = "index_all"
	changesetIndexOpenDir   = "index_open"
	changesetIndexClosedDir = "index_closed"

	defaultCommitMessageTmpl = "{{.Title}}\n\nMerge changeset #{{.ID}}\n\n{{.Description}}"
)

type Changesets struct {
	fsLock sync.RWMutex // guards FS
}

func (s *Changesets) Create(ctx context.Context, repoPath string, cs *sourcegraph.Changeset) error {
	s.migrate(ctx, repoPath)
	fs := s.storage(ctx, repoPath)
	repoVCS, err := store.RepoVCSFromContext(ctx).Open(ctx, repoPath)
	if err != nil {
		return err
	}

	cs.ID, err = claimNextChangesetID(fs)
	if err != nil {
		return err
	}

	ts := pbtypes.NewTimestamp(time.Now())
	cs.CreatedAt = &ts
	id := strconv.FormatInt(cs.ID, 10)
	head, err := repoVCS.ResolveBranch(cs.DeltaSpec.Head.Rev)
	if err != nil {
		return err
	}
	cs.DeltaSpec.Head.CommitID = string(head)

	refStore := s.GitRefStoreFromContext(ctx, repoPath)
	if err = refStore.UpdateRef("refs/changesets/"+id+"/head", string(head)); err != nil {
		return err
	}
	err = storage.PutJSON(fs, id, changesetMetadataFile, cs)
	if err == nil {
		// Update the index with the changeset's current state.
		s.updateIndex(ctx, fs, cs.ID, true)
	}
	return err
}

func (s *Changesets) GitRefStoreFromContext(ctx context.Context, repo string) GitRefStore {
	dir := absolutePathForRepo(ctx, repo)
	if _, err := os.Stat(dir); err != nil {
		return &noopGitRefStore{}
	}
	return &localGitRefStore{dir: dir}
}

func claimNextChangesetID(sys storage.System) (int64, error) {
	id, err := resolveNextChangesetID(sys)
	if err != nil {
		return 0, err
	}
	maxID := id + 100

	for ; id <= maxID; id++ {
		err = sys.PutNoOverwrite(changesetIndexAllDir, strconv.FormatInt(id, 10), []byte{})
		if err != nil && os.IsExist(err) {
			continue
		}
		break
	}
	return id, err
}

func resolveNextChangesetID(fs storage.System) (int64, error) {
	fis, err := fs.List(changesetIndexAllDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}
	max := int64(0)
	for _, name := range fis {
		n, err := strconv.Atoi(name)
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
	s.migrate(ctx, repoPath)
	fs := s.storage(ctx, repoPath)
	cs := &sourcegraph.Changeset{}
	err := s.unmarshal(fs, ID, changesetMetadataFile, cs)
	return cs, err
}

func (s *Changesets) CreateReview(ctx context.Context, repoPath string, changesetID int64, newReview *sourcegraph.ChangesetReview) (*sourcegraph.ChangesetReview, error) {
	s.migrate(ctx, repoPath)
	s.fsLock.Lock()
	defer s.fsLock.Unlock()
	fs := s.storage(ctx, repoPath)

	// Read current reviews into structure
	all := sourcegraph.ChangesetReviewList{Reviews: []*sourcegraph.ChangesetReview{}}
	err := s.unmarshal(fs, changesetID, changesetReviewsFile, &all.Reviews)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// Append new review to list
	newReview.ID = int64(len(all.Reviews) + 1)
	all.Reviews = append(all.Reviews, newReview)

	err = storage.PutJSON(fs, strconv.FormatInt(changesetID, 10), changesetReviewsFile, all.Reviews)
	return newReview, err
}

func (s *Changesets) ListReviews(ctx context.Context, repo string, changesetID int64) (*sourcegraph.ChangesetReviewList, error) {
	s.migrate(ctx, repo)
	fs := s.storage(ctx, repo)

	list := &sourcegraph.ChangesetReviewList{Reviews: []*sourcegraph.ChangesetReview{}}
	err := s.unmarshal(fs, changesetID, changesetReviewsFile, &list.Reviews)
	if os.IsNotExist(err) {
		err = nil
	}
	return list, err
}

// unmarshal JSON decodes the contents of the file for a changeset into v
func (s *Changesets) unmarshal(fs storage.System, changesetID int64, filename string, v interface{}) error {
	return storage.GetJSON(fs, strconv.FormatInt(changesetID, 10), filename, v)
}

var errInvalidUpdateOp = errors.New("invalid update operation")

func (s *Changesets) Update(ctx context.Context, opt *store.ChangesetUpdateOp) (*sourcegraph.ChangesetEvent, error) {
	s.migrate(ctx, opt.Op.Repo.URI)
	fs := s.storage(ctx, opt.Op.Repo.URI)

	op := opt.Op

	if (op.Close && op.Open) || (op.Open && op.Merged) {
		return nil, errInvalidUpdateOp
	}
	current, err := s.Get(ctx, op.Repo.URI, op.ID)
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

	id := strconv.FormatInt(current.ID, 10)
	refStore := s.GitRefStoreFromContext(ctx, opt.Op.Repo.URI)
	if opt.Head != "" {
		// We need to track the tip of this branch so that we can access its
		// data even after a potential deletion or merge.
		if err := refStore.UpdateRef("refs/changesets/"+id+"/head", opt.Head); err != nil {
			return nil, err
		}
		after.DeltaSpec.Head.CommitID = opt.Head
	}
	if opt.Base != "" {
		if err := refStore.UpdateRef("refs/changesets/"+id+"/base", opt.Base); err != nil {
			return nil, err
		}
		after.DeltaSpec.Base.CommitID = opt.Base
	}

	// If the resulting changeset is the same as the inital one, no event
	// has occurred.
	if reflect.DeepEqual(after, current) {
		return &sourcegraph.ChangesetEvent{}, nil
	}

	// HERE BE DRAGONS: We are doing a read followed by a write without
	// any concurrency control. We are relying on the fact that concurrent
	// changes to CS metadata should be exceedingly rare, and we do as
	// little as possible between read and write. This is done in the vain
	// of performance wins.
	err = storage.PutJSON(fs, id, changesetMetadataFile, &after)
	if err != nil {
		return nil, err
	}
	s.updateIndex(ctx, fs, op.ID, after.ClosedAt == nil)

	var evt *sourcegraph.ChangesetEvent
	if shouldRegisterEvent(op) {
		evt = &sourcegraph.ChangesetEvent{
			Before:    current,
			After:     &after,
			Op:        op,
			CreatedAt: &ts,
		}
		evts := []*sourcegraph.ChangesetEvent{}
		s.fsLock.Lock()
		defer s.fsLock.Unlock()
		err = s.unmarshal(fs, op.ID, changesetEventsFile, &evts)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		evts = append(evts, evt)
		err = storage.PutJSON(fs, id, changesetEventsFile, evts)
		return evt, err
	}

	return evt, nil
}

func (s *Changesets) Merge(ctx context.Context, opt *sourcegraph.ChangesetMergeOp) error {
	s.migrate(ctx, opt.Repo.URI)
	cs, _ := s.Get(ctx, opt.Repo.URI, opt.ID)
	base, head := cs.DeltaSpec.Base.Rev, cs.DeltaSpec.Head.Rev
	if cs.Merged {
		return grpc.Errorf(codes.FailedPrecondition, "changeset #%d already merged", cs.ID)
	}
	if cs.ClosedAt != nil {
		return grpc.Errorf(codes.FailedPrecondition, "changeset #%d already closed", cs.ID)
	}

	repo, err := svc.Repos(ctx).Get(ctx, &sourcegraph.RepoSpec{
		URI: opt.Repo.URI,
	})
	if err != nil {
		return err
	}

	var msgTmpl string
	if opt.Message != "" {
		msgTmpl = opt.Message
	} else {
		msgTmpl = defaultCommitMessageTmpl
	}
	msg, err := mergeCommitMessage(cs, msgTmpl)
	if err != nil {
		return err
	}

	var repoPath string
	var auth string
	if repo.Mirror {
		gitHost := util.RepoURIHost(repo.URI)
		authStore := ext.AuthStore{}
		cred, err := authStore.Get(ctx, gitHost)
		if err != nil {
			return fmt.Errorf("unable to fetch git credentials for host %q: %v", gitHost, err)
		}
		auth = cred.Token

		cloneURL, err := url.Parse(repo.HTTPCloneURL)
		if err != nil {
			return err
		}
		cloneURL.User = url.User("x-oauth-basic")
		repoPath = cloneURL.String()
	} else {
		repoPath = absolutePathForRepo(ctx, repo.URI)
	}

	rs, err := NewRepoStage(repoPath, base, auth)
	if err != nil {
		return err
	}
	defer rs.Free()

	p := notif.PersonFromContext(ctx)
	if err != nil {
		return err
	}
	merger := vcs.Signature{
		Name:  p.FullName,
		Email: p.Email,
		Date:  pbtypes.NewTimestamp(time.Now()),
	}

	if err := rs.Pull(head, opt.Squash); err != nil {
		return err
	}
	if err := rs.Commit(merger, merger, msg); err != nil {
		return err
	}

	return nil
}

func mergeCommitMessage(cs *sourcegraph.Changeset, msgTmpl string) (string, error) {
	t, err := template.New("commit message").Parse(msgTmpl)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	err = t.Execute(&b, cs)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

// shouldRegisterEvent returns true if the passed update operation will cause an event
// that needs to be registered. Registered events are be visible in the changeset's
// timeline.
func shouldRegisterEvent(op *sourcegraph.ChangesetUpdateOp) bool {
	return op.Merged || op.Close || op.Open || op.Title != "" || op.Description != ""
}

func (s *Changesets) List(ctx context.Context, op *sourcegraph.ChangesetListOp) (*sourcegraph.ChangesetList, error) {
	s.migrate(ctx, op.Repo)
	fs := s.storage(ctx, op.Repo)

	list := sourcegraph.ChangesetList{Changesets: []*sourcegraph.Changeset{}}
	fis, err := fs.List(changesetIndexAllDir)
	if err != nil {
		if os.IsNotExist(err) {
			return &list, nil
		}
		return nil, err
	}

	var open, closed map[int64]bool
	if op.Open {
		open, err = s.indexList(ctx, fs, changesetIndexOpenDir)
	}
	if op.Closed {
		closed, err = s.indexList(ctx, fs, changesetIndexClosedDir)
	}
	if err != nil {
		return nil, err
	}

	var ids []int
	for _, name := range fis {
		id, err := strconv.Atoi(name)
		if err != nil {
			log15.Debug("Non numeric changeset id", "name", name)
			continue
		}

		if (op.Open && open[int64(id)]) || (op.Closed && closed[int64(id)]) {
			ids = append(ids, id)
		}
	}

	// Sort the filepaths by ID in reverse, i.e. descending order. We use
	// descending order as higher IDs were created most recently.
	//
	// TODO(slimsag): let the user choose the sorting order.
	sort.Sort(sort.Reverse(sort.IntSlice(ids)))

	// TODO(slimsag): we currently do not index by Head/Base like we should. For
	// now, we special case it by iterating over each path element, which is
	// slower.
	headBaseSearch := op.Head != "" || op.Base != ""
	if !headBaseSearch {
		start := op.Offset()
		if start < 0 {
			start = 0
		}
		if start > len(ids) {
			start = len(ids)
		}
		end := start + op.Limit()
		if end < 0 {
			end = 0
		}
		if end > len(ids) {
			end = len(ids)
		}
		ids = ids[start:end]
	}

	// Read each metadata file from the VFS.
	skip := op.Offset()
	for _, id := range ids {
		var cs sourcegraph.Changeset
		err = storage.GetJSON(fs, strconv.Itoa(id), changesetMetadataFile, &cs)
		if err != nil {
			return nil, err
		}

		// If we're only interested in a changeset with a specific branch for head
		// or base, check that now or simply continue.
		if (op.Head != "" && op.Head != cs.DeltaSpec.Head.Rev) || (op.Base != "" && op.Base != cs.DeltaSpec.Base.Rev) {
			continue
		}

		// Handle offset.
		if headBaseSearch && skip > 0 {
			skip--
			continue
		}

		list.Changesets = append(list.Changesets, &cs)

		// If we're not migrating data, abort early once we have gotten enough
		// changesets.
		if headBaseSearch && len(list.Changesets) >= op.Limit() {
			break
		}
	}
	return &list, nil
}

func (s *Changesets) ListEvents(ctx context.Context, spec *sourcegraph.ChangesetSpec) (*sourcegraph.ChangesetEventList, error) {
	s.migrate(ctx, spec.Repo.URI)
	fs := s.storage(ctx, spec.Repo.URI)
	list := sourcegraph.ChangesetEventList{Events: []*sourcegraph.ChangesetEvent{}}
	err := s.unmarshal(fs, spec.ID, changesetEventsFile, &list.Events)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return &list, nil
}

// updateIndex updates the index with the given changeset state (whether or not
// it is opened or closed).
func (s *Changesets) updateIndex(ctx context.Context, fs storage.System, cid int64, open bool) error {
	var adds, removes []string
	if open {
		// Changeset opened.
		adds = append(adds, changesetIndexOpenDir)
		removes = append(removes, changesetIndexClosedDir)
	} else {
		// Changeset closed.
		adds = append(adds, changesetIndexClosedDir)
		removes = append(removes, changesetIndexOpenDir)
	}

	// We need the state between open and closed to stay consistent. We
	// grab a lock to prevent race conditions for concurrent modifications
	// to the same CS's state
	s.fsLock.Lock()
	defer s.fsLock.Unlock()

	// Perform additions.
	for _, indexDir := range adds {
		if err := s.indexAdd(ctx, fs, cid, indexDir); err != nil {
			return err
		}
	}

	// Perform removals.
	for _, indexDir := range removes {
		if err := s.indexRemove(ctx, fs, cid, indexDir); err != nil {
			return err
		}
	}
	return nil
}

// indexAdd adds the given changeset ID to the index directory if it does not
// already exist.
//
// Callers must guard by holding the s.fsLock lock.
func (s *Changesets) indexAdd(ctx context.Context, fs storage.System, cid int64, indexDir string) error {
	// If the file exists nothing needs to be done.
	if exists, err := s.indexHas(ctx, fs, cid, indexDir); err != nil {
		return err
	} else if exists {
		return nil
	}

	return fs.Put(indexDir, strconv.FormatInt(cid, 10), []byte{})
}

// indexRemove removes the given changeset ID from the index directory if it
// exists.
//
// Callers must guard by holding the s.fsLock lock.
func (s *Changesets) indexRemove(ctx context.Context, fs storage.System, cid int64, indexDir string) error {
	// If the file does not exist nothing needs to be done.
	if exists, err := s.indexHas(ctx, fs, cid, indexDir); err != nil {
		return err
	} else if !exists {
		return nil
	}
	return fs.Delete(indexDir, strconv.FormatInt(cid, 10))
}

// indexList returns a list of changeset IDs found in the given index directory.
func (s *Changesets) indexList(ctx context.Context, fs storage.System, indexDir string) (map[int64]bool, error) {
	infos, err := fs.List(indexDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	ids := make(map[int64]bool)
	for _, name := range infos {
		id, err := strconv.ParseInt(name, 10, 64)
		if err != nil {
			return nil, err
		}
		ids[id] = true
	}
	return ids, nil
}

// indexHas stats the changeset ID in the given index directory.
func (s *Changesets) indexHas(ctx context.Context, fs storage.System, cid int64, indexDir string) (bool, error) {
	return fs.Exists(indexDir, strconv.FormatInt(cid, 10))
}

func (s *Changesets) storage(ctx context.Context, repoPath string) storage.System {
	return storage.Namespace(ctx, "changesets", repoPath)
}

var (
	csMigrateMusMu sync.Mutex
	csMigrateMus   map[string]*sync.Once
)

func csGetMigrateOnce(repo string) *sync.Once {
	csMigrateMusMu.Lock()
	defer csMigrateMusMu.Unlock()
	if csMigrateMus == nil {
		csMigrateMus = map[string]*sync.Once{}
	}
	if once, ok := csMigrateMus[repo]; ok {
		return once
	}
	once := &sync.Once{}
	csMigrateMus[repo] = once
	return once
}

func (s *Changesets) migrate(ctx context.Context, repoPath string) {
	once := csGetMigrateOnce(repoPath)
	once.Do(func() {
		system := s.storage(ctx, repoPath)
		exists, err := system.Exists("meta", "version")
		if err != nil {
			log15.Error("Changesets.migrate checking for meta version got", err)
			return
		}
		if exists {
			// Migration has already happened
			return
		}
		err = s.doMigration(ctx, repoPath, system)
		if err != nil {
			log15.Info("Changesets Migration failed", "repo", repoPath, "error", err)
		}
	})
}

func vfsRecursiveCopy(from vfs.FileSystem, to storage.System, dir string) error {
	children, err := from.ReadDir(dir)
	if err != nil {
		return err
	}
	bucket := strings.Replace(strings.Trim(dir, "/"), "/", "_", -1)
	for _, c := range children {
		p := filepath.Join(dir, c.Name())
		if c.IsDir() {
			err = vfsRecursiveCopy(from, to, p)
			if err != nil {
				return err
			}
			if _, err := strconv.Atoi(c.Name()); err == nil && bucket == "" {
				// Top-level directories we need to remember
				// since they are changesets
				err = to.Put(changesetIndexAllDir, c.Name(), []byte{})
				if err != nil {
					return err
				}
			}
		} else {
			if bucket == "" {
				// We ignore top-level files
				continue
			}
			src, err := from.Open(p)
			if err != nil {
				return err
			}
			b, err := ioutil.ReadAll(src)
			src.Close()
			if err != nil {
				return err
			}
			err = to.Put(bucket, c.Name(), b)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type refResolver interface {
	ResolveRef(name string) (vcs.CommitID, error)
}

func getOldChangesets(ctx context.Context, repo string) (vfs.FileSystem, error) {
	dir := absolutePathForRepo(ctx, repo)
	r, err := vcs.Open("git", dir)
	if err != nil {
		return nil, err
	}
	refResolver, ok := r.(refResolver)
	if !ok {
		return nil, os.ErrNotExist
	}
	cid, err := refResolver.ResolveRef("refs/src/review")
	if err == vcs.ErrRefNotFound {
		return nil, os.ErrNotExist
	}
	return r.FileSystem(cid)
}

func (s *Changesets) doMigration(ctx context.Context, repo string, to storage.System) error {
	from, err := getOldChangesets(ctx, repo)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// Changesets is empty, no migration necessary
	} else {
		log15.Info("Starting migration of Changesets to Platform Storage", "repo", repo)
		err = vfsRecursiveCopy(from, to, "/")
		if err != nil {
			return err
		}
		log15.Info("Finished migration of Changesets to Platform Storage", "repo", repo)
	}
	version := struct {
		Version int
	}{1}
	return storage.PutJSON(to, "meta", "version", version)
}
