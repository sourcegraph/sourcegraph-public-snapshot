package local

import (
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/notif"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/server/accesscontrol"
	"sourcegraph.com/sqs/pbtypes"
)

var Notify sourcegraph.NotifyServer = &notify{}

type notify struct{}

var _ sourcegraph.NotifyServer = (*notify)(nil)

func (s *notify) GenericEvent(ctx context.Context, e *sourcegraph.NotifyGenericEvent) (*pbtypes.Void, error) {
	// Dedup recipients. We do this here as a convenience to users of the
	// API
	e.Recipients = dedupUsers(e.Recipients)

	if err := s.verifyCanNotify(ctx, e.Actor, e.Recipients); err != nil {
		return nil, err
	}

	actors := s.getPeople(ctx, e.Actor)
	recipients := s.getPeople(ctx, e.Recipients...)

	nctx := notif.ActionContext{
		Person:        actors[0],
		Recipients:    recipients,
		ActionType:    e.ActionType,
		ActionContent: e.ActionContent,
		ObjectID:      e.ObjectID,
		ObjectRepo:    e.ObjectRepo,
		ObjectType:    e.ObjectType,
		ObjectTitle:   e.ObjectTitle,
		ObjectURL:     e.ObjectURL,
		SlackMsg:      e.SlackMsg,
		EmailHTML:     e.EmailHTML,
	}

	notif.ActionSlackMessage(nctx)

	if !e.NoEmail {
		notif.ActionEmailMessage(nctx)
	}

	return &pbtypes.Void{}, nil
}

func (s *notify) getPeople(ctx context.Context, users ...*sourcegraph.UserSpec) []*sourcegraph.Person {
	people := make([]*sourcegraph.Person, len(users))
	store := store.UsersFromContext(ctx)
	for i, u := range users {
		people[i] = notif.Person(ctx, u)
		if people[i] == nil {
			people[i] = &sourcegraph.Person{}
		}
		if people[i].Email == "" && store != nil {
			// We directly query the user store, since the gRPC
			// layer enforces that the actor can only query there
			// own emails. The emails here will not be leaked back
			// to the actor, but instead used to send the emails.
			emails, err := store.ListEmails(ctx, *u)
			if err == nil {
				email := ""
				for _, emailAddr := range emails {
					if emailAddr.Blacklisted {
						continue
					}
					email = emailAddr.Email
					if emailAddr.Primary {
						break
					}
				}
				people[i].Email = email
			}
		}
	}
	return people
}

func (s *notify) verifyCanNotify(ctx context.Context, actor *sourcegraph.UserSpec, recipients []*sourcegraph.UserSpec) error {
	var uid int32
	if actor != nil {
		uid = actor.UID
	}
	return accesscontrol.VerifyUserSelfOrAdmin(ctx, "Notify.GenericEvent", uid)
}

func dedupUsers(users []*sourcegraph.UserSpec) []*sourcegraph.UserSpec {
	seen := map[int32]struct{}{}
	var dedup []*sourcegraph.UserSpec
	for _, u := range users {
		if _, ok := seen[u.UID]; !ok {
			dedup = append(dedup, u)
			seen[u.UID] = struct{}{}
		}
	}
	return dedup
}
