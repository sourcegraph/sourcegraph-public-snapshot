package graphqlbackend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

type sharedItemResolver struct {
	authorUserID string
	thread       *threadResolver
	comment      *commentResolver
}

func (s *sharedItemResolver) Author(ctx context.Context) (*userResolver, error) {
	user, err := store.Users.GetByAuth0ID(s.authorUserID)
	if err != nil {
		return nil, err
	}
	return &userResolver{user, actor.FromContext(ctx)}, nil
}

func (s *sharedItemResolver) Thread(ctx context.Context) *threadResolver {
	return s.thread
}

func (s *sharedItemResolver) Comment(ctx context.Context) *commentResolver {
	return s.comment
}

func (r *rootResolver) SharedItem(ctx context.Context, args *struct {
	ULID string
}) (*sharedItemResolver, error) {
	item, err := store.SharedItems.Get(ctx, args.ULID)
	if err != nil {
		if legacyerr.ErrCode(err) == legacyerr.NotFound {
			// shared item does not exist.
			return nil, nil
		}
		return nil, err
	}

	switch {
	case item.CommentID != nil:
		comment, err := store.Comments.GetByID(ctx, *item.CommentID)
		if err != nil {
			return nil, err
		}
		thread, err := store.Threads.Get(ctx, comment.ThreadID)
		if err != nil {
			return nil, err
		}
		orgRepo, err := store.OrgRepos.GetByID(ctx, thread.OrgRepoID)
		if err != nil {
			return nil, err
		}

		// 🚨 SECURITY: verify that the user is in the org.
		actor := actor.FromContext(ctx)
		_, err = store.OrgMembers.GetByOrgIDAndUserID(ctx, orgRepo.OrgID, actor.UID)
		if err != nil {
			return nil, err
		}

		org, err := store.Orgs.GetByID(ctx, orgRepo.OrgID)
		if err != nil {
			return nil, err
		}
		return &sharedItemResolver{
			item.AuthorUserID,
			&threadResolver{org, orgRepo, thread},
			&commentResolver{org, orgRepo, thread, comment},
		}, nil
	case item.ThreadID != nil:
		thread, err := store.Threads.Get(ctx, *item.ThreadID)
		if err != nil {
			return nil, err
		}
		orgRepo, err := store.OrgRepos.GetByID(ctx, thread.OrgRepoID)
		if err != nil {
			return nil, err
		}

		// 🚨 SECURITY: verify that the current user is in the org.
		actor := actor.FromContext(ctx)
		_, err = store.OrgMembers.GetByOrgIDAndUserID(ctx, orgRepo.OrgID, actor.UID)
		if err != nil {
			return nil, err
		}

		org, err := store.Orgs.GetByID(ctx, orgRepo.OrgID)
		if err != nil {
			return nil, err
		}
		return &sharedItemResolver{
			item.AuthorUserID,
			&threadResolver{org, orgRepo, thread},
			nil,
		}, nil
	default:
		panic("SharedItem: never here")
	}
}
