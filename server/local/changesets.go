package local

import (
	"bytes"
	"fmt"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/ext/slack"
	"src.sourcegraph.com/sourcegraph/notif"
	"src.sourcegraph.com/sourcegraph/server/internal/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/svc"
	"src.sourcegraph.com/sourcegraph/util/mdutil"
)

var Changesets sourcegraph.ChangesetsServer = &changesets{}

var _ sourcegraph.ChangesetsServer = (*changesets)(nil)

type changesets struct{}

func (s *changesets) Create(ctx context.Context, op *sourcegraph.ChangesetCreateOp) (*sourcegraph.Changeset, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Changesets.Create"); err != nil {
		return nil, err
	}

	defer noCache(ctx)

	{
		// Enqueue builds (if they don't yet exist) for the newly
		// created changeset's base and head.
		//
		// Do this before creating the changeset in case this step
		// fails.
		enqueueBuild := func(rev sourcegraph.RepoRevSpec) error {
			if err := (&repos{}).resolveRepoRev(ctx, &rev); err != nil {
				return err
			}
			_, err := svc.Builds(ctx).Create(ctx, &sourcegraph.BuildsCreateOp{
				RepoRev: rev,
				Opt:     &sourcegraph.BuildCreateOptions{BuildConfig: sourcegraph.BuildConfig{Import: true, Queue: true}},
			})
			return err
		}
		if err := enqueueBuild(op.Changeset.DeltaSpec.Base); err != nil {
			return nil, err
		}
		if err := enqueueBuild(op.Changeset.DeltaSpec.Head); err != nil {
			return nil, err
		}
	}

	if err := store.ChangesetsFromContext(ctx).Create(ctx, op.Repo.URI, op.Changeset); err != nil {
		return nil, err
	}

	{
		// Send Slack notification.
		userStr, err := getUserDisplayName(ctx)
		if err != nil {
			return nil, err
		}
		msg := fmt.Sprintf("*%s* created <%s|%s changeset #%d>: %s\n\n%s",
			userStr,
			appURL(ctx, router.Rel.URLToRepoChangeset(op.Repo.URI, op.Changeset.ID)),
			op.Repo.URI, op.Changeset.ID, op.Changeset.Title, op.Changeset.Description,
		)
		go slack.PostMessage(slack.PostOpts{Msg: msg})

		// Notify mentioned people.
		ppl, err := mdutil.Mentions(ctx, []byte(op.Changeset.Description))
		if err != nil {
			return nil, err
		}
		for _, p := range ppl {
			msg := fmt.Sprintf(
				"*%s* mentioned @%s in <%s|%s changeset #%d>: %s\n\n%s",
				userStr, p.Login,
				appURL(ctx, router.Rel.URLToRepoChangeset(op.Repo.URI, op.Changeset.ID)),
				op.Repo.URI, op.Changeset.ID, op.Changeset.Title, op.Changeset.Description,
			)
			notif.Mention(p, notif.Context{
				Mentioner:    userStr,
				MentionerURL: appURL(ctx, router.Rel.URLToUser(userStr)),
				Where:        fmt.Sprintf("in a changeset %s/%d", op.Repo.URI, op.Changeset.ID),
				WhereURL:     appURL(ctx, router.Rel.URLToRepoChangeset(op.Repo.URI, op.Changeset.ID)),
				SlackMsg:     msg,
			})
		}
	}

	return op.Changeset, nil
}

func (s *changesets) Get(ctx context.Context, op *sourcegraph.ChangesetSpec) (*sourcegraph.Changeset, error) {
	return store.ChangesetsFromContext(ctx).Get(ctx, op.Repo.URI, op.ID)
}

func (s *changesets) CreateReview(ctx context.Context, op *sourcegraph.ChangesetCreateReviewOp) (*sourcegraph.ChangesetReview, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Changesets.CreateReview"); err != nil {
		return nil, err
	}

	defer noCache(ctx)

	review, err := store.ChangesetsFromContext(ctx).CreateReview(ctx, op.Repo.URI, op.ChangesetID, op.Review)
	if err != nil {
		return nil, err
	}

	{
		// Send Slack notification.
		userStr, err := getUserDisplayName(ctx)
		if err != nil {
			return nil, err
		}
		cs, err := s.Get(ctx, &sourcegraph.ChangesetSpec{Repo: op.Repo, ID: op.ChangesetID})
		if err != nil {
			return nil, err
		}

		// TODO: We assume that Sourcegraph author's login is identical
		// on Slack for the '/cc @login' portion of the message. Provide a way
		// for users to specify their Slack name, and a CLI option to disable
		// Slack /cc's.
		msg := bytes.NewBufferString(fmt.Sprintf("*%s* reviewed <%s|%s changeset #%d>: %s /cc @%s\n\"%s\"",
			userStr,
			appURL(ctx, router.Rel.URLToRepoChangeset(op.Repo.URI, op.ChangesetID)),
			op.Repo.URI, op.ChangesetID, cs.Title, cs.Author.Login, op.Review.Body,
		))
		ppl, err := mdutil.Mentions(ctx, []byte(op.Review.Body))
		if err != nil {
			return nil, err
		}
		for _, c := range op.Review.Comments {
			msg.WriteString(fmt.Sprintf("\n*%s:%d* - %s", c.Filename, c.LineNumber, c.Body))
			ppll, err := mdutil.Mentions(ctx, []byte(c.Body))
			if err != nil {
				return nil, err
			}
			ppl = append(ppl, ppll...)
		}
		go slack.PostMessage(slack.PostOpts{Msg: msg.String()})

		// Notify mentions.
		for _, p := range ppl {
			msg := fmt.Sprintf(
				"*%s* mentioned @%s in a review on <%s|changeset #%d>",
				userStr, p.Login,
				appURL(ctx, router.Rel.URLToRepoChangeset(op.Repo.URI, op.ChangesetID)),
				op.ChangesetID,
			)
			notif.Mention(p, notif.Context{
				Mentioner:    userStr,
				MentionerURL: appURL(ctx, router.Rel.URLToUser(userStr)),
				Where:        fmt.Sprintf("in review %s/%d", op.Repo.URI, op.ChangesetID),
				WhereURL:     appURL(ctx, router.Rel.URLToRepoChangeset(op.Repo.URI, op.ChangesetID)),
				SlackMsg:     msg,
			})
		}
	}

	return review, err
}

func (s *changesets) ListReviews(ctx context.Context, op *sourcegraph.ChangesetListReviewsOp) (*sourcegraph.ChangesetReviewList, error) {
	return store.ChangesetsFromContext(ctx).ListReviews(ctx, op.Repo.URI, op.ChangesetID)
}

func (s *changesets) Update(ctx context.Context, op *sourcegraph.ChangesetUpdateOp) (*sourcegraph.ChangesetEvent, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Changesets.Update"); err != nil {
		return nil, err
	}

	defer noCache(ctx)

	return store.ChangesetsFromContext(ctx).Update(ctx, &store.ChangesetUpdateOp{Op: op})
}

func (s *changesets) List(ctx context.Context, op *sourcegraph.ChangesetListOp) (*sourcegraph.ChangesetList, error) {
	return store.ChangesetsFromContext(ctx).List(ctx, op)
}

func (s *changesets) ListEvents(ctx context.Context, spec *sourcegraph.ChangesetSpec) (*sourcegraph.ChangesetEventList, error) {
	return store.ChangesetsFromContext(ctx).ListEvents(ctx, spec)
}
