package fs

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"text/template"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sqs/pbtypes"
	approuter "src.sourcegraph.com/sourcegraph/app/router"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/ext"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/notif"
	platformstorage "src.sourcegraph.com/sourcegraph/platform/storage"
	"src.sourcegraph.com/sourcegraph/sgx/client"
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

// changesetIndexNeedsReview returns a bucket name for the given UID that
// represents the "changesets that need review" (by the given UID) index.
func changesetIndexNeedsReview(uid int32) string {
	return fmt.Sprintf("index_needs_review_%d", uid)
}

type Changesets struct {
	fsLock sync.RWMutex // guards FS
}

func (s *Changesets) Create(ctx context.Context, repoPath string, cs *sourcegraph.Changeset) error {
	fs := s.storage(ctx, repoPath)
	repoVCS, err := store.RepoVCSFromContext(ctx).Open(ctx, repoPath)
	if err != nil {
		return err
	}

	cs.ID, err = claimNextChangesetID(fs)
	if err != nil {
		return err
	}

	// It's fine to specify reviewers at creation time, but they can't set their
	// LGTM status unless they are in fact that user.
	a := authpkg.ActorFromContext(ctx)
	if !a.HasAdminAccess() {
		for _, reviewer := range cs.Reviewers {
			// Get user to ensure we have a UID for comparison and storage (in case
			// caller specifies just login).
			reviewerUser, err := svc.Users(ctx).Get(ctx, &reviewer.UserSpec)
			if err != nil {
				return err
			}
			reviewer.UserSpec = reviewerUser.Spec()

			// If callee isn't that user, clear the LGTM status.
			if reviewer.UserSpec.Domain != a.Domain || reviewer.UserSpec.UID != int32(a.UID) {
				reviewer.LGTM = false
			}

			// The CS needs review by the reviewer, so update the index.
			if err := s.updateNeedsReviewIndex(ctx, fs, cs.ID, reviewer.UserSpec, true); err != nil {
				return err
			}
		}
	}

	ts := pbtypes.NewTimestamp(time.Now())
	cs.CreatedAt = &ts
	id := strconv.FormatInt(cs.ID, 10)
	head, err := repoVCS.ResolveBranch(cs.DeltaSpec.Head.Rev)
	if err != nil {
		return err
	}
	cs.DeltaSpec.Head.CommitID = string(head)

	err = platformstorage.PutJSON(fs, id, changesetMetadataFile, cs)
	if err == nil {
		// Update the index with the changeset's current state.
		s.updateIndex(ctx, fs, cs, true)
	}
	return err
}

func claimNextChangesetID(sys platformstorage.System) (int64, error) {
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

func resolveNextChangesetID(fs platformstorage.System) (int64, error) {
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
	fs := s.storage(ctx, repoPath)
	cs := &sourcegraph.Changeset{}
	err := s.unmarshal(fs, ID, changesetMetadataFile, cs)
	return cs, err
}

func (s *Changesets) CreateReview(ctx context.Context, repoPath string, changesetID int64, newReview *sourcegraph.ChangesetReview) (*sourcegraph.ChangesetReview, error) {
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

	err = platformstorage.PutJSON(fs, strconv.FormatInt(changesetID, 10), changesetReviewsFile, all.Reviews)
	return newReview, err
}

func (s *Changesets) ListReviews(ctx context.Context, repo string, changesetID int64) (*sourcegraph.ChangesetReviewList, error) {
	fs := s.storage(ctx, repo)

	list := &sourcegraph.ChangesetReviewList{Reviews: []*sourcegraph.ChangesetReview{}}
	err := s.unmarshal(fs, changesetID, changesetReviewsFile, &list.Reviews)
	if os.IsNotExist(err) {
		err = nil
	}
	return list, err
}

// unmarshal JSON decodes the contents of the file for a changeset into v
func (s *Changesets) unmarshal(fs platformstorage.System, changesetID int64, filename string, v interface{}) error {
	return platformstorage.GetJSON(fs, strconv.FormatInt(changesetID, 10), filename, v)
}

var errInvalidUpdateOp = errors.New("invalid update operation")

func (s *Changesets) Update(ctx context.Context, opt *store.ChangesetUpdateOp) (*sourcegraph.ChangesetEvent, error) {
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

	addReviewer := func(add sourcegraph.UserSpec, lgtm bool) error {
		// Check they aren't already a reviewer.
		already := false
		for _, reviewer := range after.Reviewers {
			if reviewer.UserSpec.Domain == add.Domain && reviewer.UserSpec.UID == add.UID {
				already = true
				break
			}
		}
		if !already {
			after.Reviewers = append(after.Reviewers, &sourcegraph.ChangesetReviewer{
				UserSpec: add,
				LGTM:     lgtm,
			})

			// The CS needs review by the reviewer, so update the index.
			if err := s.updateNeedsReviewIndex(ctx, fs, current.ID, add, true); err != nil {
				return err
			}
		}
		return nil
	}

	// Add a reviewer to the CS.
	if add := op.AddReviewer; add != nil {
		// Get user to ensure we have a UID for comparison and storage (in case
		// caller specifies just login).
		addUser, err := svc.Users(ctx).Get(ctx, add)
		if err != nil {
			return nil, err
		}
		if err := addReviewer(addUser.Spec(), false); err != nil {
			return nil, err
		}
	}

	// Remove a reviewer from the CS.
	if remove := op.RemoveReviewer; remove != nil {
		// Get user to ensure we have a UID for comparison and storage (in case
		// caller specifies just login).
		removeUser, err := svc.Users(ctx).Get(ctx, remove)
		if err != nil {
			return nil, err
		}
		*remove = removeUser.Spec()
		for i, reviewer := range after.Reviewers {
			if reviewer.UserSpec.Domain == remove.Domain && reviewer.UserSpec.UID == remove.UID {
				// Delete from slice.
				after.Reviewers = append(after.Reviewers[:i], after.Reviewers[i+1:]...)

				// The CS no longer needs review by the reviewer, so update the index.
				if err := s.updateNeedsReviewIndex(ctx, fs, current.ID, reviewer.UserSpec, false); err != nil {
					return nil, err
				}
				break
			}
		}
	}

	// LGTM and not-LGTM operations.
	if op.LGTM && op.NotLGTM {
		return nil, errInvalidUpdateOp
	}
	if op.LGTM || op.NotLGTM {
		// Confirm they have access to modify LGTM status for op.Author
		a := authpkg.ActorFromContext(ctx)
		authorUser, err := svc.Users(ctx).Get(ctx, &op.Author)
		if err != nil {
			return nil, err
		}
		hasAccess := a.HasAdminAccess() || authorUser.Domain == a.Domain && authorUser.UID == int32(a.UID)
		if hasAccess {
			updated := false
			for _, reviewer := range after.Reviewers {
				if reviewer.UserSpec.Domain == a.Domain && reviewer.UserSpec.UID == int32(a.UID) {
					reviewer.LGTM = op.LGTM
					updated = true

					if op.LGTM {
						// The CS no longer needs review by the reviewer, so update the index.
						if err := s.updateNeedsReviewIndex(ctx, fs, current.ID, reviewer.UserSpec, false); err != nil {
							return nil, err
						}
					} else {
						// The CS needs review by the reviewer, so update the index.
						if err := s.updateNeedsReviewIndex(ctx, fs, current.ID, reviewer.UserSpec, true); err != nil {
							return nil, err
						}
					}
					break
				}
			}
			// Per the gRPC definition, if they aren't already a reviewer we will make
			// them one.
			if !updated {
				if err := addReviewer(authorUser.Spec(), op.LGTM); err != nil {
					return nil, err
				}
			}
		}
	}
	if opt.Head != "" || opt.Base != "" {
		// Updating the head or base of the CS? Then reviewers need to follow-up,
		// undo their LGTM statuses.
		for _, reviewer := range after.Reviewers {
			reviewer.LGTM = false
		}
	}

	id := strconv.FormatInt(current.ID, 10)
	if opt.Head != "" {
		// We need to track the tip of this branch so that we can access its
		// data even after a potential deletion or merge.
		after.DeltaSpec.Head.CommitID = opt.Head
	}
	if opt.Base != "" {
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
	err = platformstorage.PutJSON(fs, id, changesetMetadataFile, &after)
	if err != nil {
		return nil, err
	}
	s.updateIndex(ctx, fs, &after, after.ClosedAt == nil)

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
		err = platformstorage.PutJSON(fs, id, changesetEventsFile, evts)
		return evt, err
	}

	return evt, nil
}

func (s *Changesets) Merge(ctx context.Context, opt *sourcegraph.ChangesetMergeOp) error {
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
	} else {
		auth = client.Credentials.GetAccessToken()
		if auth == "" {
			return errors.New("can't generate local access token: token is empty")
		}
	}

	// Determine the person who is merging.
	p := notif.PersonFromContext(ctx)
	if p == nil {
		return grpc.Errorf(codes.Unauthenticated, "merging requires authentication")
	}
	merger := vcs.Signature{
		Name:  p.FullName,
		Email: p.Email,
		Date:  pbtypes.NewTimestamp(time.Now()),
	}

	// Determine the clone URL.
	cloneURL, err := url.Parse(repo.HTTPCloneURL)
	if err != nil {
		return err
	}
	cloneURL.User = url.User("x-oauth-basic")
	repoPath = cloneURL.String()

	// Create a repo stage to perform the merge operation.
	rs, err := changesetsNewRepoStage(repoPath, base, auth)
	if err != nil {
		return err
	}
	defer rs.free()

	// Determine the commit message template.
	msgTmpl := opt.Message
	if msgTmpl == "" {
		// Caller didn't specify a template to use, so check the repository for one.
		data, err := rs.readFile(".sourcegraph-merge-template")
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		msgTmpl = string(data)
	}
	if msgTmpl == "" {
		// Caller didn't specify the template, and the repository doesn't have one,
		// so fallback to the default template.
		msgTmpl = defaultCommitMessageTmpl
	}
	msg, err := mergeCommitMessage(ctx, cs, msgTmpl)
	if err != nil {
		return err
	}

	if err := rs.pull(head, opt.Squash); err != nil {
		return err
	}
	if err := rs.commit(merger, merger, msg); err != nil {
		return err
	}

	// If this was a squash, then we do not push the squash commit back into the
	// HEAD repo (or else the githooks would detect this and the CS would only
	// show the single "Merge changeset #33" commit message). Instead, we just
	// mark the CS as merged here and confirm it has a resolved Base/Head for
	// persistence after the branch has been deleted.
	if opt.Squash {
		_, err = s.Update(ctx, &store.ChangesetUpdateOp{
			Op: &sourcegraph.ChangesetUpdateOp{
				Repo:   opt.Repo,
				ID:     opt.ID,
				Merged: true,
				Close:  true,
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// mergeCommitMessage takes a user-provided commit message template and renders
// it. The returned string should be treated as arbitrary user input, and e.g.
// sanitized before display as it may contain HTML.
//
// The following fields are usable within the template:
//
//  "{{.ID}}" -> Replaced with the Changeset ID number.
//  "{{.Title}}" -> Replaced with the Changeset Title string.
//  "{{.Description}}" -> Replaced with the Changeset description string.
//  "{{.URL}}" -> Replaced with the URL to the Changeset on Sourcegraph.
//  "{{.Author}}" -> Replaced with the login name of the Changeset author.
//
func mergeCommitMessage(ctx context.Context, cs *sourcegraph.Changeset, msgTmpl string) (string, error) {
	t, err := template.New("commit message").Parse(msgTmpl)
	if err != nil {
		return "", err
	}

	// Hack to construct the URL to the changeset, not perfect but we don't have
	// a better way just yet.
	//
	// TODO(slimsag): clean this up.
	url := fmt.Sprintf("%s%s/.changes/%d", conf.AppURL(ctx).String(), approuter.Rel.URLToRepo(cs.DeltaSpec.Base.RepoSpec.URI).String(), cs.ID)

	data := &struct {
		ID          int64
		Title       string
		Description string
		Author      string
		URL         string
	}{
		ID:          cs.ID,
		Title:       cs.Title,
		Description: cs.Description,
		Author:      cs.Author.Login,
		URL:         url,
	}

	var b bytes.Buffer
	err = t.Execute(&b, data)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

// shouldRegisterEvent returns true if the passed update operation will cause an event
// that needs to be registered. Registered events are be visible in the changeset's
// timeline.
func shouldRegisterEvent(op *sourcegraph.ChangesetUpdateOp) bool {
	return op.Merged || op.Close || op.Open || op.Title != "" || op.Description != "" || op.LGTM || op.NotLGTM || op.AddReviewer != nil || op.RemoveReviewer != nil
}

var errInvalidListOp = errors.New("invalid list operation")

func (s *Changesets) List(ctx context.Context, op *sourcegraph.ChangesetListOp) (*sourcegraph.ChangesetList, error) {
	fs := s.storage(ctx, op.Repo)

	list := sourcegraph.ChangesetList{Changesets: []*sourcegraph.Changeset{}}
	fis, err := fs.List(changesetIndexAllDir)
	if err != nil {
		if os.IsNotExist(err) {
			return &list, nil
		}
		return nil, err
	}

	var open, closed, needsReview map[int64]bool
	if op.Open && op.Closed {
		// TODO(slimsag): consider supporting this
		return nil, errInvalidListOp
	}
	if op.NeedsReview != nil {
		needsReview, err = s.indexList(ctx, fs, changesetIndexNeedsReview(op.NeedsReview.UID))
	}
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

		if (op.Open && open[int64(id)]) || (op.Closed && closed[int64(id)]) || (op.NeedsReview != nil && needsReview[int64(id)]) {
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
		err = platformstorage.GetJSON(fs, strconv.Itoa(id), changesetMetadataFile, &cs)
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
	fs := s.storage(ctx, spec.Repo.URI)
	list := sourcegraph.ChangesetEventList{Events: []*sourcegraph.ChangesetEvent{}}
	err := s.unmarshal(fs, spec.ID, changesetEventsFile, &list.Events)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return &list, nil
}

// updateNeedsReviewIndex updates the review index with whether or not the user,
// u, needs to review the specified changeset or not.
func (s *Changesets) updateNeedsReviewIndex(ctx context.Context, fs platformstorage.System, csID int64, u sourcegraph.UserSpec, needsReview bool) error {
	bucket := changesetIndexNeedsReview(u.UID)
	if needsReview {
		return s.indexAdd(ctx, fs, csID, bucket)
	}
	return s.indexRemove(ctx, fs, csID, bucket)
}

// updateIndex updates the index with the given changeset state (whether or not
// it is opened or closed).
func (s *Changesets) updateIndex(ctx context.Context, fs platformstorage.System, cs *sourcegraph.Changeset, open bool) error {
	var adds, removes []string
	if open {
		// Changeset opened.
		adds = append(adds, changesetIndexOpenDir)
		removes = append(removes, changesetIndexClosedDir)

		// All reviewers in the CS who have not already LGTM'd it need to review it.
		for _, r := range cs.Reviewers {
			if !r.LGTM {
				if err := s.updateNeedsReviewIndex(ctx, fs, cs.ID, r.UserSpec, true); err != nil {
					return err
				}
			}
		}
	} else {
		// Changeset closed.
		adds = append(adds, changesetIndexClosedDir)
		removes = append(removes, changesetIndexOpenDir)

		// All reviewers in the CS no longer need to review it.
		for _, r := range cs.Reviewers {
			if err := s.updateNeedsReviewIndex(ctx, fs, cs.ID, r.UserSpec, false); err != nil {
				return err
			}
		}
	}

	// We need the state between open and closed to stay consistent. We
	// grab a lock to prevent race conditions for concurrent modifications
	// to the same CS's state
	s.fsLock.Lock()
	defer s.fsLock.Unlock()

	// Perform additions.
	for _, indexDir := range adds {
		if err := s.indexAdd(ctx, fs, cs.ID, indexDir); err != nil {
			return err
		}
	}

	// Perform removals.
	for _, indexDir := range removes {
		if err := s.indexRemove(ctx, fs, cs.ID, indexDir); err != nil {
			return err
		}
	}
	return nil
}

// indexAdd adds the given changeset ID to the index directory if it does not
// already exist.
//
// Callers must guard by holding the s.fsLock lock.
func (s *Changesets) indexAdd(ctx context.Context, fs platformstorage.System, cid int64, indexDir string) error {
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
func (s *Changesets) indexRemove(ctx context.Context, fs platformstorage.System, cid int64, indexDir string) error {
	// If the file does not exist nothing needs to be done.
	if exists, err := s.indexHas(ctx, fs, cid, indexDir); err != nil {
		return err
	} else if !exists {
		return nil
	}
	return fs.Delete(indexDir, strconv.FormatInt(cid, 10))
}

// indexList returns a list of changeset IDs found in the given index directory.
func (s *Changesets) indexList(ctx context.Context, fs platformstorage.System, indexDir string) (map[int64]bool, error) {
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
func (s *Changesets) indexHas(ctx context.Context, fs platformstorage.System, cid int64, indexDir string) (bool, error) {
	return fs.Exists(indexDir, strconv.FormatInt(cid, 10))
}

func (s *Changesets) storage(ctx context.Context, repoPath string) platformstorage.System {
	return platformstorage.Namespace(ctx, "changesets", repoPath)
}
