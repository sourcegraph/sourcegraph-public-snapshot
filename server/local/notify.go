package local

import (
	"fmt"

	"src.sourcegraph.com/sourcegraph/fed"
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
	}

	if s.shouldSendSlack() {
		notif.ActionSlackMessage(nctx)
	}

	if s.shouldFederateEmail() {
		// Forward request to mothership since we are not setup to send email
		notify, err := s.mothershipNotifyClient(ctx)
		if err != nil {
			return nil, err
		}
		return notify.GenericEvent(ctx, e)
	} else {
		notif.ActionEmailMessage(nctx)
	}

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

func (s *notify) verifyCanNotify(ctx context.Context, actor *sourcegraph.UserSpec, recipients []*sourcegraph.UserSpec) error {
	// TODO(keegan) implement some sort of verification to prevent abuse
	return nil
}

func (s *notify) mothershipNotifyClient(ctx context.Context) (sourcegraph.NotifyClient, error) {
	mothership, err := fed.Config.RootGRPCEndpoint()
	if err != nil {
		return nil, err
	}
	mctx := sourcegraph.WithGRPCEndpoint(ctx, mothership)
	return sourcegraph.NewClientFromContext(mctx).Notify, nil
}

func (s *notify) shouldFederateEmail() bool {
	// If we are configured to send email or we are the mothership then
	// our instance should send the email, rather than federating out
	return !(notif.EmailIsConfigured() || fed.Config.IsRoot)
}

func (s *notify) shouldSendSlack() bool {
	// TODO(keegan) This is a temporary hack since we currently don't have
	// any slack notifications setup for our fedroot, but do have slack
	// configured for other notifications. We don't want to send slack
	// notifications for private instances to our own slack channel
	return !fed.Config.IsRoot
}
