package changesets

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/sourcegraph/mux"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/notif"
	"src.sourcegraph.com/sourcegraph/platform/pctx"
	"src.sourcegraph.com/sourcegraph/platform/putil"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
)

// serveCreate creates a new changeset. It takes the changeset data from within
// the request's body and returns the changeset in JSON form when successful.
func serveCreate(w http.ResponseWriter, r *http.Request) error {
	var newChangeset *sourcegraph.Changeset
	if err := json.NewDecoder(r.Body).Decode(&newChangeset); err != nil {
		return err
	}
	defer r.Body.Close()
	ctx := putil.Context(r)
	repo, ok := pctx.RepoRevSpec(ctx)
	if !ok {
		return errors.New("no repo found in context")
	}
	uri := repo.URI

	user := putil.UserFromRequest(r)
	if user == nil {
		return &handlerutil.HTTPErr{Status: http.StatusUnauthorized}
	}
	newChangeset.Author = user.Spec()

	sg := sourcegraph.NewClientFromContext(ctx)
	cs, err := sg.Changesets.Create(ctx, &sourcegraph.ChangesetCreateOp{
		Repo:      sourcegraph.RepoSpec{URI: uri},
		Changeset: newChangeset,
	})
	if err != nil {
		return err
	}
	if err := writeJSON(w, struct {
		Repo string
		ID   int64
	}{uri, cs.ID}); err != nil {
		return err
	}

	if flags.JiraURL != "" {
		jiraOnChangesetUpdate(ctx, cs)
	}
	notifyCreation(ctx, user, uri, cs)
	{
		events.Publish(events.Event{
			EventID: notif.ChangesetCreateEvent,
			Payload: notif.Payload{
				Type:        notif.ChangesetCreateEvent,
				UserSpec:    user.Spec(),
				ActionType:  "created",
				ObjectID:    cs.ID,
				ObjectRepo:  uri,
				ObjectTitle: cs.Title,
				ObjectType:  "changeset",
				ObjectURL:   urlToChangeset(ctx, cs.ID),
				Object:      cs,
			},
		})
	}
	return nil
}

// serveUpdate updates a changeset based on the data received in the request's
// body. The data is in JSON form and is decoded against `sourcegraph.ChangesetUpdateOp`.
func serveUpdate(w http.ResponseWriter, r *http.Request) (err error) {
	defer func() {
		if err != nil {
			err = writeJSON(w, err)
		}
	}()

	ctx := putil.Context(r)
	repo, ok := pctx.RepoRevSpec(ctx)
	if !ok {
		return errors.New("no repo found in context")
	}
	uri := repo.URI
	id, err := strconv.ParseInt(mux.Vars(r)["ID"], 10, 64)
	if err != nil {
		return err
	}
	var op sourcegraph.ChangesetUpdateOp
	if err := json.NewDecoder(r.Body).Decode(&op); err != nil {
		return err
	}
	defer r.Body.Close()
	op.ID = id
	op.Repo = sourcegraph.RepoSpec{URI: uri}

	user := putil.UserFromRequest(r)
	if user == nil {
		return &handlerutil.HTTPErr{Status: http.StatusUnauthorized}
	}
	op.Author = user.Spec()
	sg := sourcegraph.NewClientFromContext(ctx)
	result, err := sg.Changesets.Update(ctx, &op)
	if err != nil {
		return err
	}

	if flags.JiraURL != "" {
		cs, err := sg.Changesets.Get(ctx, &sourcegraph.ChangesetSpec{
			Repo: sourcegraph.RepoSpec{URI: uri},
			ID:   id,
		})
		if err != nil {
			return err
		}
		jiraOnChangesetUpdate(ctx, cs)
	}

	{
		events.Publish(events.Event{
			EventID: notif.ChangesetUpdateEvent,
			Payload: notif.Payload{
				Type:        notif.ChangesetUpdateEvent,
				UserSpec:    user.Spec(),
				ActionType:  "updated",
				ObjectID:    op.ID,
				ObjectRepo:  uri,
				ObjectTitle: op.Title,
				ObjectType:  "changeset",
				ObjectURL:   urlToChangeset(ctx, op.ID),
				Object:      op,
			},
		})
	}
	return writeJSON(w, result)
}

// serveSubmitReview submits a new review. The request's body contains the review
// information in JSON form.
func serveSubmitReview(w http.ResponseWriter, r *http.Request) error {
	v := mux.Vars(r)
	id, err := strconv.ParseInt(v["ID"], 10, 64)
	if err != nil {
		return err
	}

	ctx := putil.Context(r)
	repo, ok := pctx.RepoRevSpec(ctx)
	if !ok {
		return errors.New("no repo found in context")
	}
	uri := repo.URI
	newReview := &sourcegraph.ChangesetReview{}
	if err := json.NewDecoder(r.Body).Decode(newReview); err != nil {
		return err
	}
	defer r.Body.Close()
	user := putil.UserFromRequest(r)
	if user == nil {
		return &handlerutil.HTTPErr{Status: http.StatusUnauthorized}
	}
	newReview.Author = user.Spec()

	sg := sourcegraph.NewClientFromContext(ctx)
	op := &sourcegraph.ChangesetCreateReviewOp{
		Repo:        sourcegraph.RepoSpec{URI: uri},
		ChangesetID: id,
		Review:      newReview,
	}
	review, err := sg.Changesets.CreateReview(ctx, op)
	if err != nil {
		return err
	}
	if err := writeJSON(w, review); err != nil {
		return err
	}
	cs, err := sg.Changesets.Get(ctx, &sourcegraph.ChangesetSpec{Repo: repo.RepoSpec, ID: id})
	if err != nil {
		return err
	}
	notifyReview(ctx, user, uri, cs, op)
	{
		events.Publish(events.Event{
			EventID: notif.ChangesetReviewEvent,
			Payload: notif.Payload{
				Type:        notif.ChangesetReviewEvent,
				UserSpec:    user.Spec(),
				ActionType:  "reviewed",
				ObjectID:    cs.ID,
				ObjectRepo:  uri,
				ObjectTitle: cs.Title,
				ObjectType:  "changeset",
				ObjectURL:   urlToChangeset(ctx, cs.ID),
				Object:      op,
			},
		})
	}
	return nil
}

// serverMerge initiates a merge from the changeset's head branch to its base
// branch.
func serveMerge(w http.ResponseWriter, r *http.Request) (err error) {
	defer func() {
		if err != nil {
			err = writeJSON(w, err)
		}
	}()

	ctx := putil.Context(r)
	repo, ok := pctx.RepoRevSpec(ctx)
	if !ok {
		return errors.New("no repo found in context")
	}
	uri := repo.URI
	id, err := strconv.ParseInt(mux.Vars(r)["ID"], 10, 64)
	if err != nil {
		return err
	}

	var op sourcegraph.ChangesetMergeOp
	if err := json.NewDecoder(r.Body).Decode(&op); err != nil {
		return err
	}
	op.ID = id
	op.Repo = sourcegraph.RepoSpec{URI: uri}

	sg := sourcegraph.NewClientFromContext(ctx)
	event, err := sg.Changesets.Merge(ctx, &op)
	if err != nil {
		return err
	}

	return writeJSON(w, event)
}
