package graphqlbackend

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type discussionThreadTargetInput struct {
	Repo *discussionThreadTargetRepoInput `json:"repo,omitempty"`
}

// validate checks the validity of the input and returns an error, if any.
func (d *discussionThreadTargetInput) validate() error {
	count := 0
	if d.Repo != nil {
		count++
		if err := d.Repo.validate(); err != nil {
			return err
		}
	}
	if count != 1 {
		return errors.New("exactly 1 field in DiscussionThreadTargetInput must be non-null")
	}
	return nil
}

func (d *discussionThreadTargetInput) validateAndGetTarget(ctx context.Context) (*types.DiscussionThreadTargetRepo, error) {
	if err := d.validate(); err != nil {
		return nil, err
	}

	switch {
	case d.Repo != nil:
		return d.Repo.convert(ctx)
	default:
		return nil, errors.New("exactly 1 field in DiscussionThreadTargetInput must be non-null (or an unrecognized target type was specified)")
	}
}

func (r *discussionsMutationResolver) AddTargetToThread(ctx context.Context, args *struct {
	ThreadID graphql.ID
	Target   *discussionThreadTargetInput
}) (*discussionThreadTargetResolver, error) {
	// 🚨 SECURITY: Only signed in users may add a target to a thread.
	currentUser, err := CurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	if currentUser == nil {
		return nil, errors.New("no current user")
	}

	threadID, err := unmarshalDiscussionThreadID(args.ThreadID)
	if err != nil {
		return nil, err
	}
	target, err := args.Target.validateAndGetTarget(ctx)
	if err != nil {
		return nil, err
	}
	target.ThreadID = threadID

	// Avoid adding duplicates.
	{
		targets, err := db.DiscussionThreads.ListTargets(ctx, db.DiscussionThreadsListTargetsOptions{
			ThreadID: threadID,
			RepoID:   target.RepoID,
			Path:     *target.Path,
		})
		if err != nil {
			return nil, errors.Wrap(err, "DiscussionThreads.ListTargets")
		}
		if len(targets) > 0 {
			return &discussionThreadTargetResolver{t: targets[0]}, nil
		}
	}

	if _, err := db.DiscussionThreads.AddTarget(ctx, target); err != nil {
		return nil, errors.Wrap(err, "DiscussionThreads.AddTarget")
	}
	return &discussionThreadTargetResolver{t: target}, nil
}

type discussionThreadUpdateTargetInput struct {
	TargetID  graphql.ID
	Remove    *bool
	IsIgnored *bool
}

func (r *discussionsMutationResolver) UpdateTargetInThread(ctx context.Context, args *struct {
	Input discussionThreadUpdateTargetInput
}) (*discussionThreadTargetResolver, error) {
	// 🚨 SECURITY: Only signed in users may update a target in a thread.
	currentUser, err := CurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	if currentUser == nil {
		return nil, errors.New("no current user")
	}

	targetID, err := unmarshalDiscussionThreadTargetID(args.Input.TargetID)
	if err != nil {
		return nil, err
	}
	if args.Input.Remove != nil && *args.Input.Remove {
		if err := db.DiscussionThreads.RemoveTarget(ctx, targetID); err != nil {
			return nil, errors.Wrap(err, "DiscussionThreads.RemoveTarget")
		}
		return nil, nil
	}
	if args.Input.IsIgnored != nil {
		if err := db.DiscussionThreads.SetTargetIsIgnored(ctx, targetID, *args.Input.IsIgnored); err != nil {
			return nil, errors.Wrap(err, "DiscussionThreads.SetTargetIsIgnored")
		}
	}

	target, err := db.DiscussionThreads.GetTarget(ctx, targetID)
	if err != nil {
		return nil, err
	}
	return &discussionThreadTargetResolver{t: target}, nil
}
