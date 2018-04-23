package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
)

type sharedItemResolver struct {
	authorUserID int32
	public       bool
	thread       *threadResolver
	comment      *commentResolver
}

func (s *sharedItemResolver) Author(ctx context.Context) (*userResolver, error) {
	user, err := db.Users.GetByID(ctx, s.authorUserID)
	if err != nil {
		return nil, err
	}
	return &userResolver{user}, nil
}

func (s *sharedItemResolver) Public(ctx context.Context) bool {
	return s.public
}

func (s *sharedItemResolver) Thread(ctx context.Context) *threadResolver {
	return s.thread
}

func (s *sharedItemResolver) Comment(ctx context.Context) *commentResolver {
	return s.comment
}

func (r *schemaResolver) SharedItem(ctx context.Context, args *struct {
	ULID string
}) (*sharedItemResolver, error) {
	item, err := db.SharedItems.Get(ctx, args.ULID)
	if err != nil {
		if _, ok := err.(db.ErrSharedItemNotFound); ok {
			// shared item does not exist.
			return nil, nil
		}
		return nil, err
	}

	switch {
	case item.CommentID != nil:
		comment, err := db.Comments.GetByID(ctx, *item.CommentID)
		if err != nil {
			return nil, err
		}
		thread, err := db.Threads.Get(ctx, comment.ThreadID)
		if err != nil {
			return nil, err
		}
		orgRepo, err := db.OrgRepos.GetByID(ctx, thread.OrgRepoID)
		if err != nil {
			return nil, err
		}

		if !item.Public {
			// 🚨 SECURITY: Check that the current user is a member of the org.
			if err := backend.CheckOrgAccess(ctx, orgRepo.OrgID); err != nil {
				return nil, err
			}
		}

		org, err := db.Orgs.GetByID(ctx, orgRepo.OrgID)
		if err != nil {
			return nil, err
		}
		return &sharedItemResolver{
			item.AuthorUserID,
			item.Public,
			&threadResolver{org, orgRepo, thread},
			&commentResolver{org, orgRepo, thread, comment},
		}, nil
	case item.ThreadID != nil:
		thread, err := db.Threads.Get(ctx, *item.ThreadID)
		if err != nil {
			return nil, err
		}
		orgRepo, err := db.OrgRepos.GetByID(ctx, thread.OrgRepoID)
		if err != nil {
			return nil, err
		}

		if !item.Public {
			// 🚨 SECURITY: Check that the current user is a member of the org.
			if err := backend.CheckOrgAccess(ctx, orgRepo.OrgID); err != nil {
				return nil, err
			}
		}

		org, err := db.Orgs.GetByID(ctx, orgRepo.OrgID)
		if err != nil {
			return nil, err
		}
		return &sharedItemResolver{
			item.AuthorUserID,
			item.Public,
			&threadResolver{org, orgRepo, thread},
			nil,
		}, nil
	default:
		panic("SharedItem: never here")
	}
}
