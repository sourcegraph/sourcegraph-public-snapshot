package local

import (
	"fmt"

	"src.sourcegraph.com/sourcegraph/notif"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
)

var Notify sourcegraph.NotifyServer = &notify{}

type notify struct{}

var _ sourcegraph.NotifyServer = (*notify)(nil)

func (s *notify) GenericEvent(ctx context.Context, e *sourcegraph.NotifyGenericEvent) (*pbtypes.Void, error) {
	defer noCache(ctx)

	// TODO verify we can act as actor and notify recipients
	actors := s.getPeople(ctx, e.Actor)
	recipients := s.getPeople(ctx, e.Recipients...)

	notif.ActionEmailMessage(notif.ActionContext{
		Person:        actors[0],
		Recipients:    recipients,
		ActionType:    e.ActionType,
		ActionContent: e.ActionContent,
		ObjectID:      e.ObjectID,
		ObjectRepo:    e.ObjectRepo,
		ObjectType:    e.ObjectType,
		ObjectTitle:   e.ObjectTitle,
		ObjectURL:     e.ObjectURL,
	})

	return &pbtypes.Void{}, nil
}

func (s *notify) Mention(ctx context.Context, m *sourcegraph.NotifyMention) (*pbtypes.Void, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *notify) getPeople(ctx context.Context, users ...*sourcegraph.UserSpec) []*sourcegraph.Person {
	cl := sourcegraph.NewClientFromContext(ctx)
	people := make([]*sourcegraph.Person, len(users))
	for i, u := range users {
		people[i] = notif.Person(ctx, cl, u)
	}
	return people
}
