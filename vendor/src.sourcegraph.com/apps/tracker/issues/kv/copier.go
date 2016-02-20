package kv

import (
	"fmt"

	"src.sourcegraph.com/apps/tracker/issues"
	"src.sourcegraph.com/sourcegraph/platform/storage"
)

const TODO = true

func (s service) CopyFrom(src issues.Service, repo issues.RepoSpec) error {
	ctx := s.appCtx
	sys := storage.Namespace(s.appCtx, s.appName, repo.URI)

	is, err := src.List(ctx, repo, issues.IssueListOptions{State: issues.AllStates})
	if err != nil {
		return err
	}
	fmt.Printf("Copying %v issues.\n", len(is))
	for _, i := range is {
		i, err = src.Get(ctx, repo, i.ID) // Needed to get the reference, since List operation doesn't include all details.
		if err != nil {
			return err
		}
		{
			// Copy issue.
			issue := issue{
				State:     i.State,
				Title:     i.Title,
				CreatedAt: i.CreatedAt,
			}
			if i.User.Domain == "sourcegraph.com" || TODO {
				issue.AuthorUID = int32(i.User.ID)
			}
			if ref := i.Reference; ref != nil {
				issue.Reference = &reference{
					Repo:      ref.Repo,
					Path:      ref.Path,
					CommitID:  ref.CommitID,
					StartLine: ref.StartLine,
					EndLine:   ref.EndLine,
				}
			}

			// Put in storage.
			if err := storage.PutJSON(sys, issuesBucket, formatUint64(i.ID), issue); err != nil {
				return err
			}
		}

		comments, err := src.ListComments(ctx, repo, i.ID, nil)
		if err != nil {
			return err
		}
		fmt.Printf("Issue %v: Copying %v comments.\n", i.ID, len(comments))
		for _, c := range comments {
			// Copy comment.
			comment := comment{
				CreatedAt: c.CreatedAt,
				Body:      c.Body,
			}
			if c.User.Domain == "sourcegraph.com" || TODO {
				comment.AuthorUID = int32(c.User.ID)
			}

			// Put in storage.
			err = storage.PutJSON(sys, issueCommentsBucket(i.ID), formatUint64(c.ID), comment)
			if err != nil {
				return err
			}
		}

		events, err := src.ListEvents(ctx, repo, i.ID, nil)
		if err != nil {
			return err
		}
		fmt.Printf("Issue %v: Copying %v events.\n", i.ID, len(events))
		for _, e := range events {
			// Copy event.
			event := event{
				CreatedAt: e.CreatedAt,
				Type:      e.Type,
				Rename:    e.Rename,
			}
			if e.Actor.Domain == "sourcegraph.com" || TODO {
				event.ActorUID = int32(e.Actor.ID)
			}

			// Put in storage.
			err = storage.PutJSON(sys, issueEventsBucket(i.ID), formatUint64(e.ID), event)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
